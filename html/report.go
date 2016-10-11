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
	"github.com/vlifesystems/rhkit/experiment"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/report"
	"html/template"
	"time"
)

func generateReport(
	_report *report.Report,
	config *config.Config,
) (string, error) {
	type TplData struct {
		Title              string
		Tags               map[string]string
		DateTime           string
		ExperimentFilename string
		NumRecords         int64
		SortOrder          []experiment.SortField
		Assessments        []*report.Assessment
		Html               map[string]template.HTML
	}

	tagLinks := makeTagLinks(_report.Tags)

	tplData := TplData{
		_report.Title,
		tagLinks,
		_report.Stamp.Format(time.RFC822),
		_report.ExperimentFilename,
		_report.NumRecords,
		_report.SortOrder,
		_report.Assessments,
		makeHtml("reports"),
	}

	reportURLDir, err :=
		makeReportURLDir(config.WWWDir, _report.Stamp, _report.Title)
	if err != nil {
		return "", err
	}
	reportFilename :=
		genReportFilename(config.WWWDir, _report.Stamp, _report.Title)
	err = writeTemplate(reportFilename, reportTpl, tplData)
	return reportURLDir, err
}
