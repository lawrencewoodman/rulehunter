/*
	rulehuntersrv - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

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
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/report"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

func generateTagPages(config *config.Config) error {
	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	tagsSeen := make(map[string]bool)
	for _, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			for _, tag := range report.Tags {
				if _, ok := tagsSeen[tag]; !ok {
					if err := generateTagPage(config, tag); err != nil {
						return err
					}
					tagsSeen[tag] = true
				}
			}
		}
	}
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
			report, err := report.LoadJson(config, file.Name())
			if err != nil {
				return err
			}
			if inStrings(tagName, report.Tags) {
				reportURLDir := genReportURLDir(report.Stamp, report.Title)
				tplReports[i] = newTplReport(
					report.Title,
					makeTagLinks(report.Tags),
					reportURLDir,
					report.Stamp,
				)
				i++
			}
		}
	}
	tplReports = tplReports[:i]
	sortTplReportsByDate(tplReports)
	tplData := TplData{tagName, tplReports, makeHtml("tag")}
	fullTagDir := filepath.Join(
		config.WWWDir,
		"reports",
		"tag",
		escapeString(tagName),
	)

	if err := os.MkdirAll(fullTagDir, modePerm); err != nil {
		return err
	}
	outputFilename := filepath.Join(fullTagDir, "index.html")
	return writeTemplate(outputFilename, tagTpl, tplData)
}

func makeTagLinks(tags []string) map[string]string {
	links := make(map[string]string, len(tags))
	for _, tag := range tags {
		links[tag] = makeTagLink(tag)
	}
	return links
}

func makeTagLink(tag string) string {
	return fmt.Sprintf(
		"/reports/tag/%s/",
		escapeString(tag),
	)
}
