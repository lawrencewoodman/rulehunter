// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/report"
)

func generateReports(
	cfg *config.Config,
	pm *progress.Monitor,
) error {
	type TplData struct {
		Reports []*TplReport
		Html    map[string]template.HTML
	}

	reportFiles, err := ioutil.ReadDir(filepath.Join(cfg.BuildDir, "reports"))
	if err != nil {
		return err
	}

	numReportFiles := countFiles(reportFiles)
	tplReports := make([]*TplReport, numReportFiles)

	for i, file := range reportFiles {
		if !file.IsDir() {
			report, err := report.LoadJSON(cfg, file.Name(), maxReportLoadAttempts)
			if err != nil {
				return err
			}
			reportURLDir, err := generateReport(report, cfg)
			if err != nil {
				return err
			}
			tplReports[i] = newTplReport(
				report.Title,
				makeTagLinks(report.Tags),
				report.Category,
				makeCategoryLink(report.Category),
				reportURLDir,
				report.Stamp,
			)
		}
	}
	sortTplReportsByDate(tplReports)
	tplData := TplData{
		Reports: tplReports,
		Html:    makeHtml(cfg, "reports"),
	}

	outputFilename := filepath.Join("reports", "index.html")
	return writeTemplate(cfg, outputFilename, reportsTpl, tplData)
}
