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
	"github.com/vlifesystems/rhkit"
	"github.com/vlifesystems/rhkit/experiment"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/report"
	"html/template"
	"path/filepath"
	"time"
)

func generateReport(
	_report *report.Report,
	description *rhkit.Description,
	config *config.Config,
) (string, error) {
	type TplData struct {
		Title              string
		Tags               map[string]string
		DateTime           string
		ExperimentFilename string
		NumRecords         int64
		Description        *rhkit.Description
		SortOrder          []experiment.SortField
		Assessments        []*report.Assessment
		Html               map[string]template.HTML
	}

	tagLinks := makeTagLinks(_report.Tags)

	tplData := TplData{
		Title:              _report.Title,
		Tags:               tagLinks,
		DateTime:           _report.Stamp.Format(time.RFC822),
		ExperimentFilename: _report.ExperimentFilename,
		NumRecords:         _report.NumRecords,
		Description:        description,
		SortOrder:          _report.SortOrder,
		Assessments:        _report.Assessments,
		Html:               makeHtml(config, "reports"),
	}

	reportURLDir := genReportURLDir(_report.Stamp, _report.Title)
	reportFilename := genReportFilename(_report.Stamp, _report.Title)
	err := writeTemplate(config, reportFilename, reportTpl, tplData)
	return reportURLDir, err
}

func genReportFilename(stamp time.Time, title string) string {
	magicNumber := genStampMagicString(stamp)
	escapedTitle := escapeString(title)
	return filepath.Join(
		"reports",
		fmt.Sprintf("%d", stamp.Year()),
		fmt.Sprintf("%02d", stamp.Month()),
		fmt.Sprintf("%02d", stamp.Day()),
		fmt.Sprintf("%s_%s", magicNumber, escapedTitle),
		"index.html",
	)
}
