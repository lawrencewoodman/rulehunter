// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/progress"
	"html/template"
	"path/filepath"
	"time"
)

func generateActivityPage(
	config *config.Config,
	progressMonitor *progress.Monitor,
) error {
	type TplExperiment struct {
		Title       string
		Category    string
		CategoryURL string
		Tags        map[string]string
		Stamp       string
		Filename    string
		Status      string
		Msg         string
		Percent     float64
	}

	type TplData struct {
		Experiments []*TplExperiment
		Html        map[string]template.HTML
	}

	experiments := progressMonitor.GetExperiments()
	tplExperiments := make([]*TplExperiment, len(experiments))

	for i, experiment := range experiments {
		tplExperiments[i] = &TplExperiment{
			experiment.Title,
			experiment.Category,
			makeCategoryLink(experiment.Category),
			makeTagLinks(experiment.Tags),
			experiment.Status.Stamp.Format(time.RFC822),
			experiment.Filename,
			experiment.Status.State.String(),
			experiment.Status.Msg,
			experiment.Status.Percent,
		}
	}
	tplData := TplData{tplExperiments, makeHtml(config, "activity")}

	outputFilename := filepath.Join("activity", "index.html")
	return writeTemplate(config, outputFilename, activityTpl, tplData)
}
