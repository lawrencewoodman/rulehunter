/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package experiment

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencewoodman/rulehunter"
	"github.com/lawrencewoodman/rulehunter/csvinput"
	"github.com/lawrencewoodman/rulehuntersrv/config"
	"github.com/lawrencewoodman/rulehuntersrv/progress"
	"github.com/lawrencewoodman/rulehuntersrv/report"
	"os"
	"path/filepath"
	"runtime"
)

func Process(
	experimentFilename string,
	config *config.Config,
	progressMonitor *progress.ProgressMonitor,
) error {
	epr, err :=
		progress.NewExperimentProgressReporter(progressMonitor, experimentFilename)
	if err != nil {
		return err
	}

	experimentFullFilename :=
		filepath.Join(config.ExperimentsDir, experimentFilename)
	experiment, categories, err := loadExperiment(experimentFullFilename)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't load experiment file: %s", err)
		return epr.ReportError(fullErr)
	}
	defer experiment.Close()
	err = epr.UpdateDetails(experiment.Title, categories)
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

	assessment, err := assessRules(rules, experiment, epr)
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

	assessment2, err := assessRules(tweakableRules, experiment, epr)
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

	bestNonCombinedRules := assessment3.GetRules()
	numRulesToCombine := 50
	combinedRules := rulehunter.CombineRules(
		truncateRules(bestNonCombinedRules, numRulesToCombine),
	)

	assessment4, err := assessRules(combinedRules, experiment, epr)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't assess rules: %s", err)
		return epr.ReportError(fullErr)
	}

	assessment5, err := assessment3.Merge(assessment4)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't merge assessments: %s", err)
		return epr.ReportError(fullErr)
	}

	err = report.WriteJson(
		assessment5,
		experiment,
		experimentFilename,
		categories,
		config,
	)
	if err != nil {
		fullErr := fmt.Errorf("Couldn't write json report: %s", err)
		return epr.ReportError(fullErr)
	}

	return epr.ReportSuccess()
}

type experimentFile struct {
	Title                 string
	Categories            []string
	InputFilename         string
	FieldNames            []string
	ExcludeFieldNames     []string
	IsFirstLineFieldNames bool
	Separator             string
	Aggregators           []*rulehunter.AggregatorDesc
	Goals                 []string
	SortOrder             []*rulehunter.SortDesc
}

func loadExperiment(filename string) (
	*rulehunter.Experiment,
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
	experimentDesc := &rulehunter.ExperimentDesc{
		Title:         e.Title,
		Input:         input,
		Fields:        e.FieldNames,
		ExcludeFields: e.ExcludeFieldNames,
		Aggregators:   e.Aggregators,
		Goals:         e.Goals,
		SortOrder:     e.SortOrder,
	}
	experiment, err := rulehunter.MakeExperiment(experimentDesc)
	return experiment, e.Categories, err
}

func prettyPrintFieldDescriptions(fds map[string]*rulehunter.FieldDescription) {
	fmt.Println("Input Description\n")
	for field, fd := range fds {
		fmt.Println("--------------------------")
		fmt.Printf("%s\n--------------------------\n", field)
		prettyPrintFieldDescription(fd)
	}
	fmt.Println("\n")
}

func prettyPrintFieldDescription(fd *rulehunter.FieldDescription) {
	fmt.Printf("Kind: %s\n", fd.Kind)
	fmt.Printf("Min: %s\n", fd.Min)
	fmt.Printf("Max: %s\n", fd.Max)
	fmt.Printf("MaxDP: %d\n", fd.MaxDP)
	fmt.Printf("Values: %s\n", fd.Values)
}

func assessRules(
	rules []*rulehunter.Rule,
	experiment *rulehunter.Experiment,
	epr *progress.ExperimentProgressReporter,
) (*rulehunter.Assessment, error) {
	var assessment *rulehunter.Assessment
	// TODO: Make this part of the config
	maxProcesses := runtime.NumCPU()
	c := make(chan *rulehunter.AssessRulesMPOutcome)

	msg := fmt.Sprintf("Assessing rules using %d CPUs...\n", maxProcesses)
	if err := epr.ReportInfo(msg); err != nil {
		return nil, err
	}

	go rulehunter.AssessRulesMP(
		rules,
		experiment.Aggregators,
		experiment.Goals,
		experiment.Input,
		maxProcesses,
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

func truncateRules(rules []*rulehunter.Rule, numRules int) []*rulehunter.Rule {
	if len(rules) < numRules {
		numRules = len(rules)
	}
	return rules[:numRules]
}
