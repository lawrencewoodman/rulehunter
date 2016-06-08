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
package experiment

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vlifesystems/rulehunter"
	"github.com/vlifesystems/rulehunter/csvdataset"
	"github.com/vlifesystems/rulehunter/dataset"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/report"
	"os"
	"path/filepath"
)

type experimentFile struct {
	Title             string
	Tags              []string
	Dataset           string
	Csv               *csvDesc
	FieldNames        []string
	ExcludeFieldNames []string
	Aggregators       []*experiment.AggregatorDesc
	Goals             []string
	SortOrder         []*experiment.SortDesc
}

type csvDesc struct {
	Filename  string
	HasHeader bool
	Separator string
}

func Process(
	experimentFilename string,
	cfg *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	epr, err :=
		progress.NewExperimentProgressReporter(progressMonitor, experimentFilename)
	if err != nil {
		return err
	}

	experimentFullFilename :=
		filepath.Join(cfg.ExperimentsDir, experimentFilename)
	experiment, tags, err := loadExperiment(experimentFullFilename)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't load experiment file: %s", err)
		return epr.ReportError(fullErr)
	}
	err = epr.UpdateDetails(experiment.Title, tags)
	if err != nil {
		return err
	}

	if err := epr.ReportInfo("Describing dataset"); err != nil {
		return err
	}
	fieldDescriptions, err := rulehunter.DescribeDataset(experiment.Dataset)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't describe dataset: %s", err)
		return epr.ReportError(fullErr)
	}

	if err := epr.ReportInfo("Generating rules"); err != nil {
		return err
	}
	rules, err :=
		rulehunter.GenerateRules(fieldDescriptions, experiment.ExcludeFieldNames)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't generate rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment, err := assessRules(rules, experiment, epr, cfg)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment.Sort(experiment.SortOrder)
	assessment.Refine(3)
	sortedRules := assessment.GetRules()

	if err := epr.ReportInfo("Tweaking rules"); err != nil {
		return err
	}
	tweakableRules := rulehunter.TweakRules(sortedRules, fieldDescriptions)

	assessment2, err := assessRules(tweakableRules, experiment, epr, cfg)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment3, err := assessment.Merge(assessment2)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}
	assessment3.Sort(experiment.SortOrder)
	assessment3.Refine(1)

	numRulesToCombine := 50
	bestNonCombinedRules := assessment3.GetRules(numRulesToCombine)
	combinedRules := rulehunter.CombineRules(bestNonCombinedRules)

	assessment4, err := assessRules(combinedRules, experiment, epr, cfg)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5, err := assessment3.Merge(assessment4)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5.Sort(experiment.SortOrder)
	assessment5.Refine(1)

	assessment6 := assessment5.TruncateRuleAssessments(cfg.NumRulesInReport)

	err = report.WriteJson(
		assessment6,
		experiment,
		experimentFilename,
		tags,
		cfg,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't write json report: %s", err)
		return epr.ReportError(fullErr)
	}

	return epr.ReportSuccess()
}

func loadExperiment(filename string) (
	*experiment.Experiment,
	[]string,
	error,
) {
	var f *os.File
	var e experimentFile
	var err error

	f, err = os.Open(filename)
	if err != nil {
		return nil, []string{}, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&e); err != nil {
		return nil, []string{}, err
	}

	if err := e.checkValid(); err != nil {
		return nil, []string{}, err
	}

	dataset, err := makeDataset(&e)
	if err != nil {
		return nil, []string{}, err
	}
	experimentDesc := &experiment.ExperimentDesc{
		Title:         e.Title,
		Dataset:       dataset,
		ExcludeFields: e.ExcludeFieldNames,
		Aggregators:   e.Aggregators,
		Goals:         e.Goals,
		SortOrder:     e.SortOrder,
	}
	experiment, err := experiment.New(experimentDesc)
	return experiment, e.Tags, err
}

func makeDataset(e *experimentFile) (dataset.Dataset, error) {
	var dataset dataset.Dataset
	var err error

	switch e.Dataset {
	case "csv":
		dataset, err = csvdataset.New(
			e.FieldNames,
			e.Csv.Filename,
			rune(e.Csv.Separator[0]),
			e.Csv.HasHeader,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil,
			fmt.Errorf("Experiment field: dataset, has invalid type: %s", e.Dataset)
	}
	return dataset, nil
}

func (e *experimentFile) checkValid() error {
	if len(e.Title) == 0 {
		return errors.New("Experiment field missing: title")
	}
	if len(e.Dataset) == 0 {
		return errors.New("Experiment field missing: dataset")
	}
	if e.Dataset == "csv" {
		if e.Csv == nil {
			return errors.New("Experiment field missing: csv")
		}
		if len(e.Csv.Filename) == 0 {
			return errors.New("Experiment field missing: csv > filename")
		}
		if len(e.Csv.Separator) == 0 {
			return errors.New("Experiment field missing: csv > separator")
		}
	}
	return nil
}

func assessRules(
	rules []*rulehunter.Rule,
	experiment *experiment.Experiment,
	epr *progress.ExperimentProgressReporter,
	cfg *config.Config,
) (*rulehunter.Assessment, error) {
	var assessment *rulehunter.Assessment
	c := make(chan *rulehunter.AssessRulesMPOutcome)

	msg := fmt.Sprintf("Assessing rules using %d CPUs...\n", cfg.MaxNumProcesses)
	if err := epr.ReportInfo(msg); err != nil {
		return nil, err
	}

	go rulehunter.AssessRulesMP(
		rules,
		experiment,
		cfg.MaxNumProcesses,
		c,
	)
	for o := range c {
		if o.Err != nil {
			return nil, o.Err
		}
		msg := fmt.Sprintf("Assessment progress: %.2f%%", o.Progress*100)
		if err := epr.ReportInfo(msg); err != nil {
			return nil, err
		}
		assessment = o.Assessment
	}
	msg = fmt.Sprintf("Assessment progress: 100%%")
	if err := epr.ReportInfo(msg); err != nil {
		return nil, err
	}
	return assessment, nil
}
