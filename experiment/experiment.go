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
	"fmt"
	"github.com/vlifesystems/rulehunter"
	"github.com/vlifesystems/rulehunter/assessment"
	"github.com/vlifesystems/rulehunter/csvinput"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/rule"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/report"
	"os"
	"path/filepath"
)

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
	defer experiment.Close()
	err = epr.UpdateDetails(experiment.Title, tags)
	if err != nil {
		return err
	}

	if err := epr.ReportInfo("Describing input"); err != nil {
		return err
	}
	fieldDescriptions, err := rulehunter.DescribeInput(experiment.Input)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't describe input: %s", err)
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

	assessment6 := assessment5.LimitRuleAssessments(cfg.NumRulesInReport)

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

type experimentFile struct {
	Title                 string
	Tags                  []string
	InputFilename         string
	FieldNames            []string
	ExcludeFieldNames     []string
	IsFirstLineFieldNames bool
	Separator             string
	Aggregators           []*experiment.AggregatorDesc
	Goals                 []string
	SortOrder             []*experiment.SortDesc
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

	input, err := csvinput.New(
		e.FieldNames,
		e.InputFilename,
		rune(e.Separator[0]),
		e.IsFirstLineFieldNames,
	)
	if err != nil {
		return nil, []string{}, err
	}
	experimentDesc := &experiment.ExperimentDesc{
		Title:         e.Title,
		Input:         input,
		ExcludeFields: e.ExcludeFieldNames,
		Aggregators:   e.Aggregators,
		Goals:         e.Goals,
		SortOrder:     e.SortOrder,
	}
	experiment, err := experiment.New(experimentDesc)
	return experiment, e.Tags, err
}

func assessRules(
	rules []*rule.Rule,
	experiment *experiment.Experiment,
	epr *progress.ExperimentProgressReporter,
	cfg *config.Config,
) (*assessment.Assessment, error) {
	var assessment *assessment.Assessment
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
