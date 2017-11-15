// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/report"
)

func generateReport(r *report.Report, config *config.Config) (string, error) {
	type TplData struct {
		Mode               string
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
		Mode:               r.Mode.String(),
		Title:              r.Title,
		Tags:               makeTagLinks(r.Tags),
		Category:           r.Category,
		CategoryURL:        makeCategoryLink(r.Category),
		DateTime:           r.Stamp.Format(time.RFC822),
		ExperimentFilename: r.ExperimentFilename,
		NumRecords:         r.NumRecords,
		Description:        r.Description,
		SortOrder:          r.SortOrder,
		Aggregators:        r.Aggregators,
		Assessments:        r.Assessments,
		Html:               makeHtml(config, "reports"),
	}

	reportURLDir := genReportURLDir(r.Mode, r.Category, r.Title)
	reportFilename := genReportFilename(r.Mode, r.Category, r.Title)
	err := writeTemplate(config, reportFilename, reportTpl, tplData)
	return reportURLDir, err
}

func genReportFilename(
	mode report.ModeKind,
	category string,
	title string,
) string {
	escapedTitle := escapeString(title)
	escapedCategory := escapeString(category)
	if category != "" {
		return filepath.Join(
			"reports",
			"category",
			escapedCategory,
			escapedTitle,
			mode.String(),
			"index.html",
		)
	}
	return filepath.Join(
		"reports",
		"nocategory",
		escapedTitle,
		mode.String(),
		"index.html",
	)
}
