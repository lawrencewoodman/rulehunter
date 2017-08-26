/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

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
				reportURLDir := genReportURLDir(report.Title)
				tplReports[i] = newTplReport(
					report.Title,
					makeTagLinks(report.Tags),
					report.Category,
					makeCategoryLink(report.Category),
					reportURLDir,
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
		"category",
		escapeString(categoryName),
		"index.html",
	)
	return writeTemplate(config, outputFilename, categoryTpl, tplData)
}

func makeCategoryLink(category string) string {
	return fmt.Sprintf("category/%s/", escapeString(category))
}
