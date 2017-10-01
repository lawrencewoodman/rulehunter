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

func generateCategoryPages(config *config.Config) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	categorysLen := make(map[string]int)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(config, file.Name())
			if err != nil {
				return err
			}
			escapedCategory := escapeString(report.Category)
			if _, ok := categorysLen[escapedCategory]; !ok ||
				len(report.Category) < categorysLen[escapedCategory] {
				if err := generateCategoryPage(config, report.Category); err != nil {
					return err
				}
				// Use the shortest category out of those that resolve to the
				// same escaped category
				categorysLen[escapedCategory] = len(report.Category)
			}
		}
	}
	return nil
}

func generateCategoryPage(config *config.Config, categoryName string) error {
	type TplData struct {
		Category string
		Reports  []*TplReport
		Html     map[string]template.HTML
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
			if escapeString(categoryName) == escapeString(report.Category) {
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
	tplReports = tplReports[:i]
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		Category: categoryName,
		Reports:  tplReports,
		Html:     makeHtml(config, "category"),
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
	return writeTemplate(config, outputFilename, categoryTpl, tplData)
}

func makeCategoryLink(category string) string {
	return fmt.Sprintf("reports/category/%s/", escapeString(category))
}
