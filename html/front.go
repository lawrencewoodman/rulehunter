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
	"html/template"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/report"
)

func generateFront(
	config *config.Config,
	progressMonitor *progress.Monitor,
) error {
	const maxNumReports = 10
	type TplExperiment struct {
		Title    string
		Category string
		Status   string
		Msg      string
		Percent  float64
	}
	type TplData struct {
		Categories  map[string]string
		Tags        map[string]string
		Reports     []*TplReport
		Experiments []*TplExperiment
		Html        map[string]template.HTML
	}
	categories := map[string]string{}
	tags := map[string]string{}

	reportFiles, err := ioutil.ReadDir(filepath.Join(config.BuildDir, "reports"))
	if err != nil {
		return err
	}

	tplReports := []*TplReport{}
	numReports := 0
	sort.Slice(reportFiles, func(i, j int) bool {
		return reportFiles[i].ModTime().After(reportFiles[j].ModTime())
	})
	for _, file := range reportFiles {
		if !file.IsDir() {
			r, err := report.LoadJSON(config, file.Name())
			if err != nil {
				return err
			}
			tags = joinURLMaps(tags, makeTagLinks(r.Tags))
			if categoryName := escapeString(r.Category); categoryName != "" {
				categories[r.Category] = makeCategoryLink(r.Category)
			}

			if numReports < maxNumReports {
				tplReports = append(
					tplReports,
					newTplReport(
						r.Title,
						makeTagLinks(r.Tags),
						r.Category,
						makeCategoryLink(r.Category),
						genReportURLDir(r.Category, r.Title),
						r.Stamp,
					),
				)
			}
			numReports++
		}
	}

	tplExperiments := []*TplExperiment{}
	experiments := progressMonitor.GetExperiments()
	for _, experiment := range experiments {
		if experiment.Status.State == progress.Processing {
			tplExperiments = append(
				tplExperiments,
				&TplExperiment{
					experiment.Title,
					experiment.Category,
					experiment.Status.State.String(),
					experiment.Status.Msg,
					experiment.Status.Percent,
				},
			)
			break
		}
	}

	tplData := TplData{
		Categories:  categories,
		Tags:        tags,
		Reports:     tplReports,
		Experiments: tplExperiments,
		Html:        makeHtml(config, "front"),
	}

	outputFilename := "index.html"
	return writeTemplate(config, outputFilename, frontTpl, tplData)
}

func joinURLMaps(a, b map[string]string) map[string]string {
	r := map[string]string{}
	for n, u := range a {
		r[n] = u
	}
	for n, u := range b {
		r[n] = u
	}
	return r
}
