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
	"github.com/vlifesystems/rhkit"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/report"
	"html/template"
	"io/ioutil"
	"path/filepath"
)

func generateReports(
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {

	type TplData struct {
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
			description, err := rhkit.LoadDescriptionJSON(
				filepath.Join(config.BuildDir, "descriptions", file.Name()),
			)
			if err != nil {
				return err
			}
			reportURLDir, err := generateReport(report, description, config)
			if err != nil {
				return err
			}
			tplReports[i] = newTplReport(
				report.Title,
				makeTagLinks(report.Tags),
				reportURLDir,
				report.Stamp,
			)
		}
		i++
	}
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		Reports: tplReports,
		Html:    makeHtml(config, "reports"),
	}

	outputFilename := "index.html"
	return writeTemplate(config, outputFilename, reportsTpl, tplData)
}
