// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"bytes"
	"fmt"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// File mode permission used as standard for the html content:
// No special permission bits
// User: Read, Write Execute
// Group: Read
// Other: None
const modePerm = 0740

// The number of times to try to load a report.  This is useful because
// sometimes a report will be being written while trying to load it
// and therefore the load will fail.
const maxReportLoadAttempts = 5

// Builder represents an html website builder
type Builder struct {
	cfg       *config.Config
	pm        *progress.Monitor
	log       logger.Logger
	cmds      <-chan cmd.Cmd
	isRunning bool
}

func New(
	cfg *config.Config,
	pm *progress.Monitor,
	l logger.Logger,
	cmds <-chan cmd.Cmd,
) *Builder {
	return &Builder{
		cfg:       cfg,
		pm:        pm,
		log:       l,
		cmds:      cmds,
		isRunning: false,
	}
}

// This should be run as a goroutine and each time a command is passed to
// cmds the html will be generated
func (b *Builder) Run(q *quitter.Quitter) {
	q.Add()
	defer q.Done()
	b.isRunning = true
	defer func() { b.isRunning = false }()
	if err := b.generate(cmd.All); err != nil {
		b.log.Error(fmt.Errorf("Couldn't generate report: %s", err))
	}

	for {
		select {
		case c := <-b.cmds:
			if err := b.generate(c); err != nil {
				b.log.Error(fmt.Errorf("Couldn't generate report: %s", err))
			}
		case <-q.C:
			if err := b.generate(cmd.All); err != nil {
				b.log.Error(fmt.Errorf("Couldn't generate report: %s", err))
			}
			return
		}
	}
}

// Running returns whether the html builder is running
func (b *Builder) Running() bool {
	return b.isRunning
}

func (b *Builder) generate(c cmd.Cmd) error {
	type generator func(*config.Config, *progress.Monitor) error
	generators := []generator{}
	switch c {
	case cmd.Progress:
		generators = []generator{
			generateActivityPage,
			generateFront,
		}
	case cmd.Reports:
		fallthrough
	case cmd.All:
		generators = []generator{
			generateActivityPage,
			generateFront,
			generateReports,
			generateTagPages,
			generateCategoryPages,
		}
	}
	for _, g := range generators {
		if err := g(b.cfg, b.pm); err != nil {
			return err
		}
	}
	return nil
}

var nonAlphaNumOrSpaceRegexp = regexp.MustCompile("[^[:alnum:] ]")
var spaceRegexp = regexp.MustCompile(" ")
var multipleHyphenRegexp = regexp.MustCompile("-+")

func escapeString(s string) string {
	newS := nonAlphaNumOrSpaceRegexp.ReplaceAllString(s, "")
	newS = spaceRegexp.ReplaceAllString(newS, "-")
	newS = multipleHyphenRegexp.ReplaceAllString(newS, "-")
	newS = strings.Trim(newS, " -")
	return strings.ToLower(newS)
}

// CreatePageError indicates that an html page can't be created
type CreatePageError struct {
	Filename string
	Op       string
	Err      error
}

func (wpe CreatePageError) Error() string {
	return fmt.Sprintf(
		"can't create html page for filename: %s, %s (%s)",
		wpe.Filename, wpe.Err, wpe.Op)
}

func writeTemplate(
	cfg *config.Config,
	filename string,
	tpl string,
	tplData interface{},
) error {
	funcMap := template.FuncMap{
		"ToTitle": strings.Title,
	}
	t, err := template.New("webpage").Funcs(funcMap).Parse(tpl)
	if err != nil {
		return CreatePageError{Filename: filename, Op: "parse", Err: err}
	}

	fullFilename := filepath.Join(cfg.WWWDir, filename)
	dir := filepath.Dir(fullFilename)
	if err := os.MkdirAll(dir, modePerm); err != nil {
		return CreatePageError{Filename: filename, Op: "mkdir", Err: err}
	}

	f, err := os.Create(fullFilename)
	if err != nil {
		return CreatePageError{Filename: filename, Op: "create", Err: err}
	}
	defer f.Close()

	if err := t.Execute(f, tplData); err != nil {
		return CreatePageError{Filename: filename, Op: "execute", Err: err}
	}
	return nil
}

func genReportURLDir(category string, title string) string {
	escapedTitle := escapeString(title)
	escapedCategory := escapeString(category)
	if category != "" {
		return fmt.Sprintf("reports/category/%s/%s/", escapedCategory, escapedTitle)
	}
	return fmt.Sprintf("reports/nocategory/%s/", escapedTitle)
}

func countFiles(files []os.FileInfo) int {
	numFiles := 0
	for _, file := range files {
		if !file.IsDir() {
			numFiles++
		}
	}
	return numFiles
}

func makeHtmlNav(menuItem string) template.HTML {
	var doc bytes.Buffer
	validMenuItems := []string{
		"front",
		"reports",
		"tag",
		"category",
		"activity",
	}

	foundValidItem := false
	for _, validMenuItem := range validMenuItems {
		if validMenuItem == menuItem {
			foundValidItem = true
		}
	}
	if !foundValidItem {
		panic(fmt.Sprintf("menuItem not valid: %s", menuItem))
	}

	t, err := template.New("webpage").Parse(navTpl)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s", err))
	}

	tplData := struct{ MenuItem string }{menuItem}

	if err := t.Execute(&doc, tplData); err != nil {
		panic(fmt.Sprintf("Couldn't create nav html: %s", err))
	}
	return template.HTML(doc.String())
}

func makeHtmlHead(cfg *config.Config) template.HTML {
	var doc bytes.Buffer
	t, err := template.New("webpage").Parse(headTpl)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create head html: %s", err))
	}

	tplData := struct {
		BaseURL   string
		JSComment template.HTML
	}{BaseURL: cfg.BaseURL, JSComment: template.HTML(headJSComment)}

	if err := t.Execute(&doc, tplData); err != nil {
		panic(fmt.Sprintf("Couldn't create head html: %s", err))
	}
	return template.HTML(doc.String())
}

func makeHtml(cfg *config.Config, menuItem string) map[string]template.HTML {
	r := make(map[string]template.HTML)
	r["head"] = makeHtmlHead(cfg)
	r["nav"] = makeHtmlNav(menuItem)
	r["footer"] = template.HTML(footerHtml)
	r["bootstrapJS"] = template.HTML(bootstrapJSHtml)
	return r
}
