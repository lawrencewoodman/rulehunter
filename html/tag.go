/*
	rulehunter - A server to find rules in data based on user specified goals
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
			for _, reportTag := range report.Tags {
				if escapedTagname == escapeString(reportTag) {
					reportURLDir := genReportURLDir(report.Title)
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
	}
	tplReports = tplReports[:i]
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		Tag:     tagName,
		Reports: tplReports,
		Html:    makeHtml(config, "tag"),
	}
	outputFilename := filepath.Join(
		"tag",
		escapeString(tagName),
		"index.html",
	)

	return writeTemplate(config, outputFilename, tagTpl, tplData)
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
		"tag/%s/",
		escapeString(tag),
	)
}
