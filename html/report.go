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
	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/report"
	"html/template"
	"path/filepath"
	"time"
)

func generateReport(
	_report *report.Report,
	_description *description.Description,
	config *config.Config,
) (string, error) {
	type TplData struct {
		Title              string
		Tags               map[string]string
		Category           string
		CategoryURL        string
		DateTime           string
		ExperimentFilename string
		NumRecords         int64
		Description        *description.Description
		SortOrder          []assessment.SortOrder
		Aggregators        []report.AggregatorDesc
		Assessments        []*report.Assessment
		Html               map[string]template.HTML
	}

	tplData := TplData{
		Title:              _report.Title,
		Tags:               makeTagLinks(_report.Tags),
		Category:           _report.Category,
		CategoryURL:        makeCategoryLink(_report.Category),
		DateTime:           _report.Stamp.Format(time.RFC822),
		ExperimentFilename: _report.ExperimentFilename,
		NumRecords:         _report.NumRecords,
		Description:        _description,
		SortOrder:          _report.SortOrder,
		Aggregators:        _report.Aggregators,
		Assessments:        _report.Assessments,
		Html:               makeHtml(config, "reports"),
	}

	reportURLDir := genReportURLDir(_report.Title)
	reportFilename := genReportFilename(_report.Stamp, _report.Title)
	err := writeTemplate(config, reportFilename, reportTpl, tplData)
	return reportURLDir, err
}

func genReportFilename(stamp time.Time, title string) string {
	escapedTitle := escapeString(title)
	return filepath.Join("reports", escapedTitle, "index.html")
}
