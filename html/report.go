// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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

	reportURLDir := genReportURLDir(_report.Category, _report.Title)
	reportFilename :=
		genReportFilename(_report.Category, _report.Title)
	err := writeTemplate(config, reportFilename, reportTpl, tplData)
	return reportURLDir, err
}

func genReportFilename(category string, title string) string {
	escapedTitle := escapeString(title)
	escapedCategory := escapeString(category)
	if category != "" {
		return filepath.Join(
			"reports",
			"category",
			escapedCategory,
			escapedTitle,
			"index.html",
		)
	}
	return filepath.Join("reports", "nocategory", escapedTitle, "index.html")
}
