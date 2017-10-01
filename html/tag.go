// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"fmt"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/report"
	"html/template"
	"io/ioutil"
	"path/filepath"
)

func generateTagPages(config *config.Config) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	tagsLen := make(map[string]int)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(config, file.Name())
			if err != nil {
				return err
			}
			for _, tag := range report.Tags {
				escapedTag := escapeString(tag)
				if _, ok := tagsLen[escapedTag]; !ok ||
					len(tag) < tagsLen[escapedTag] {
					if err := generateTagPage(config, tag); err != nil {
						return err
					}
					// Use the shortest tag out of those that resolve to the
					// same escaped tag
					tagsLen[escapedTag] = len(tag)
				}
			}
		}
	}

	if err := generateTagPage(config, ""); err != nil {
		return err
	}
	tagsLen[""] = 0
	return nil
}

func generateTagPage(config *config.Config, tagName string) error {
	type TplData struct {
		Tag     string
		Reports []*TplReport
		Html    map[string]template.HTML
	}

	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	numReportFiles := countFiles(reportFiles)
	tplReports := make([]*TplReport, numReportFiles)

	i := 0
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(config, file.Name())
			if err != nil {
				return err
			}
			escapedTagname := escapeString(tagName)
			if len(report.Tags) > 0 {
				for _, reportTag := range report.Tags {
					if escapedTagname == escapeString(reportTag) {
						tplReports[i] = newTplReport(
							report.Title,
							makeTagLinks(report.Tags),
							report.Category,
							makeCategoryLink(report.Category),
							genReportURLDir(report.Category, report.Title),
							report.Stamp,
						)
						i++
					}
				}
			} else {
				if len(escapedTagname) == 0 {
					tplReports[i] = newTplReport(
						report.Title,
						makeTagLinks(report.Tags),
						report.Category,
						makeCategoryLink(report.Category),
						genReportURLDir(report.Category, report.Title),
						report.Stamp,
					)
					i++
				}
			}
		}
	}
	tplReports = tplReports[:i]
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		Tag:     tagName,
		Reports: tplReports,
		Html:    makeHtml(config, "tag"),
	}
	outputFilename := filepath.Join(
		"reports",
		"tag",
		escapeString(tagName),
		"index.html",
	)
	if len(escapeString(tagName)) == 0 {
		outputFilename = filepath.Join("reports", "notag", "index.html")
	}
	return writeTemplate(config, outputFilename, tagTpl, tplData)
}

func makeTagLinks(tags []string) map[string]string {
	links := make(map[string]string, len(tags))
	for _, tag := range tags {
		if escapeString(tag) != "" {
			links[tag] = makeTagLink(tag)
		}
	}
	return links
}

func makeTagLink(tag string) string {
	return fmt.Sprintf(
		"reports/tag/%s/",
		escapeString(tag),
	)
}
