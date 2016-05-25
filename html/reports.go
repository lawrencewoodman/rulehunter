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
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/report"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"
)

type TplReport struct {
	Title    string
	Tags     map[string]string
	DateTime string
	Filename string
	Stamp    time.Time
}

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
			reportFilename := makeReportFilename(report.Stamp, report.Title)
			if err = generateReport(report, reportFilename, config); err != nil {
				return err
			}
			tplReports[i] = &TplReport{
				report.Title,
				makeTagLinks(report.Tags),
				report.Stamp.Format(time.RFC822),
				reportFilename,
				report.Stamp,
			}
		}
		i++
	}
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		tplReports,
		makeHtml("reports"),
	}

	outputFilename := filepath.Join(config.WWWDir, "reports", "index.html")
	return writeTemplate(outputFilename, reportsTpl, tplData)
}

// byDate implements sort.Interface for []*TplReport
type byDate []*TplReport

func (r byDate) Len() int { return len(r) }
func (r byDate) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r byDate) Less(i, j int) bool {
	return r[j].Stamp.Before(r[i].Stamp)
}

func sortTplReportsByDate(tplReports []*TplReport) {
	sort.Sort(byDate(tplReports))
}
