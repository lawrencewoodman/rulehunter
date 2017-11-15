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

func generateCategoryPages(
	cfg *config.Config,
	pm *progress.Monitor,
) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(cfg.BuildDir, "reports"))
	if err != nil {
		return err
	}

	categorysLen := make(map[string]int)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(cfg, file.Name(), maxReportLoadAttempts)
			if err != nil {
				return err
			}
			escapedCategory := escapeString(report.Category)
			if _, ok := categorysLen[escapedCategory]; !ok ||
				len(report.Category) < categorysLen[escapedCategory] {
				if err := generateCategoryPage(cfg, report.Category); err != nil {
					return err
				}
				// Use the shortest category out of those that resolve to the
				// same escaped category
				categorysLen[escapedCategory] = len(report.Category)
			}
		}
	}

	if err := generateCategoryPage(cfg, ""); err != nil {
		return err
	}
	categorysLen[""] = 0
	return nil
}

func generateCategoryPage(cfg *config.Config, categoryName string) error {
	type TplData struct {
		Category string
		Reports  []*TplReport
		Html     map[string]template.HTML
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
			report, err := report.LoadJSON(cfg, file.Name(), maxReportLoadAttempts)
			if err != nil {
				return err
			}
			if escapeString(categoryName) == escapeString(report.Category) {
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
	tplReports = tplReports[:i]
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		Category: categoryName,
		Reports:  tplReports,
		Html:     makeHtml(cfg, "category"),
	}
	outputFilename := filepath.Join(
		"reports",
		"category",
		escapeString(categoryName),
		"index.html",
	)
	if len(escapeString(categoryName)) == 0 {
		outputFilename = filepath.Join("reports", "nocategory", "index.html")
	}
	return writeTemplate(cfg, outputFilename, categoryTpl, tplData)
}

func makeCategoryLink(category string) string {
	return fmt.Sprintf("reports/category/%s/", escapeString(category))
}
