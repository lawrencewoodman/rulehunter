// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/report"
)

func generateTagPages(
	cfg *config.Config,
	pm *progress.Monitor,
) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(cfg.BuildDir, "reports"))
	if err != nil {
		return err
	}

	tagsLen := make(map[string]int)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(cfg, file.Name(), maxReportLoadAttempts)
			if err != nil {
				return err
			}
			for _, tag := range report.Tags {
				escapedTag := escapeString(tag)
				if _, ok := tagsLen[escapedTag]; !ok ||
					len(tag) < tagsLen[escapedTag] {
					if err := generateTagPage(cfg, tag); err != nil {
						return err
					}
					// Use the shortest tag out of those that resolve to the
					// same escaped tag
					tagsLen[escapedTag] = len(tag)
				}
			}
		}
	}

	if err := generateTagPage(cfg, ""); err != nil {
		return err
	}
	tagsLen[""] = 0
	return nil
}

func generateTagPage(cfg *config.Config, tagName string) error {
	type TplData struct {
		Tag     string
		Reports []*TplReport
		Html    map[string]template.HTML
	}

	reportFiles, err := ioutil.ReadDir(filepath.Join(cfg.BuildDir, "reports"))
	if err != nil {
		return err
	}

	numReportFiles := countFiles(reportFiles)
	tplReports := make([]*TplReport, numReportFiles)

	i := 0
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(cfg, file.Name())
			if err != nil {
				return err
			}
			escapedTagname := escapeString(tagName)
			if len(report.Tags) > 0 {
				for _, reportTag := range report.Tags {
					if escapedTagname == escapeString(reportTag) {
						tplReports[i] = newTplReport(
							report.Mode,
							report.Title,
							makeTagLinks(report.Tags),
							report.Category,
							makeCategoryLink(report.Category),
							genReportURLDir(report.Mode, report.Category, report.Title),
							report.Stamp,
						)
						i++
					}
				}
			} else {
				if len(escapedTagname) == 0 {
					tplReports[i] = newTplReport(
						report.Mode,
						report.Title,
						makeTagLinks(report.Tags),
						report.Category,
						makeCategoryLink(report.Category),
						genReportURLDir(report.Mode, report.Category, report.Title),
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
		Html:    makeHtml(cfg, "tag"),
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
	return writeTemplate(cfg, outputFilename, tagTpl, tplData)
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
