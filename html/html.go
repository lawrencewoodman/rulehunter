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

// This should be run as a goroutine and each time a command is passed to
// cmds the html will be generated
func Run(
	config *config.Config,
	pm *progress.Monitor,
	l logger.Logger,
	quit *quitter.Quitter,
	cmds <-chan cmd.Cmd,
) {
	quit.Add()
	defer quit.Done()
	if err := generate(cmd.All, config, pm); err != nil {
		l.Error(fmt.Errorf("Couldn't generate report: %s", err))
	}

	for {
		select {
		case c := <-cmds:
			if err := generate(c, config, pm); err != nil {
				l.Error(fmt.Errorf("Couldn't generate report: %s", err))
			}
		case <-quit.C:
			if err := generate(cmd.All, config, pm); err != nil {
				l.Error(fmt.Errorf("Couldn't generate report: %s", err))
			}
			return
		}
	}
}

func generate(
	c cmd.Cmd,
	config *config.Config,
	pm *progress.Monitor,
) error {
	switch c {
	case cmd.Progress:
		if err := generateActivityPage(config, pm); err != nil {
			return err
		}
		if err := generateFront(config, pm); err != nil {
			return err
		}
	case cmd.Reports:
		if err := generateFront(config, pm); err != nil {
			return err
		}
		if err := generateReports(config); err != nil {
			return err
		}
		if err := generateTagPages(config); err != nil {
			return err
		}
		if err := generateCategoryPages(config); err != nil {
			return err
		}
		if err := generateActivityPage(config, pm); err != nil {
			return err
		}
	case cmd.All:
		if err := generateFront(config, pm); err != nil {
			return err
		}
		if err := generateReports(config); err != nil {
			return err
		}
		if err := generateTagPages(config); err != nil {
			return err
		}
		if err := generateCategoryPages(config); err != nil {
			return err
		}
		if err := generateActivityPage(config, pm); err != nil {
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
		"can't create html page for filename: %s, error: %s, Op: %s",
		wpe.Filename, wpe.Err, wpe.Op)
}

func writeTemplate(
	config *config.Config,
	filename string,
	tpl string,
	tplData interface{},
) error {
	funcMap := template.FuncMap{
		"ToTitle": strings.Title,
	}
	t, err := template.New("webpage").Funcs(funcMap).Parse(tpl)
	if err != nil {
		return CreatePageError{Filename: filename, Op: "Parse", Err: err}
	}

	fullFilename := filepath.Join(config.WWWDir, filename)
	dir := filepath.Dir(fullFilename)
	if err := os.MkdirAll(dir, modePerm); err != nil {
		return CreatePageError{Filename: filename, Op: "MkdirAll", Err: err}
	}

	f, err := os.Create(fullFilename)
	if err != nil {
		return CreatePageError{Filename: filename, Op: "Create", Err: err}
	}
	defer f.Close()

	if err := t.Execute(f, tplData); err != nil {
		return CreatePageError{Filename: filename, Op: "Execute", Err: err}
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

func makeHtmlHead(c *config.Config) template.HTML {
	var doc bytes.Buffer
	t, err := template.New("webpage").Parse(headTpl)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create head html: %s", err))
	}

	tplData := struct {
		BaseURL   string
		JSComment template.HTML
	}{BaseURL: c.BaseURL, JSComment: template.HTML(headJSComment)}

	if err := t.Execute(&doc, tplData); err != nil {
		panic(fmt.Sprintf("Couldn't create head html: %s", err))
	}
	return template.HTML(doc.String())
}

func makeHtml(c *config.Config, menuItem string) map[string]template.HTML {
	r := make(map[string]template.HTML)
	r["head"] = makeHtmlHead(c)
	r["nav"] = makeHtmlNav(menuItem)
	r["footer"] = template.HTML(footerHtml)
	r["bootstrapJS"] = template.HTML(bootstrapJSHtml)
	return r
}
