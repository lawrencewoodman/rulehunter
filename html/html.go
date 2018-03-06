// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/report"
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
		"IsLast": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
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

func genReportURLDir(
	mode report.ModeKind,
	category string,
	title string,
) string {
	escapedTitle := escapeString(title)
	escapedCategory := escapeString(category)
	if category != "" {
		return fmt.Sprintf("reports/category/%s/%s/%s/",
			escapedCategory, escapedTitle, mode.String())
	}
	return fmt.Sprintf("reports/nocategory/%s/%s/", escapedTitle, mode.String())
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
