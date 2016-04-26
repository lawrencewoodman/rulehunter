/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lawrencewoodman/rulehunter"
	"github.com/lawrencewoodman/rulehunter/csvinput"
	"os"
	"path/filepath"
	"runtime"
)

func processExperiment(experimentFilename string, config *config) error {
	var experiment *rulehunter.Experiment
	var p *os.File
	var err error

	progressFullFilename := filepath.Join(
		config.ProgressDir,
		fmt.Sprintf("%s.progress", experimentFilename),
	)
	p, err = os.Create(progressFullFilename)
	if err != nil {
		msg := fmt.Sprintf("Couldn't create progress file: %s", err)
		err := errors.New(msg)
		return err
	}
	defer p.Close()

	experimentFullFilename :=
		filepath.Join(config.ExperimentsDir, experimentFilename)
	experiment, err = loadExperiment(experimentFullFilename)
	if err != nil {
		msg := fmt.Sprintf("Couldn't load experiment file: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}
	defer experiment.Close()

	reportProgress(p, "Describing input")
	fieldDescriptions, err := rulehunter.DescribeInput(experiment.Input)
	if err != nil {
		msg := fmt.Sprintf("Couldn't describe input: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}

	reportProgress(p, "Generating rules")
	rules, err :=
		rulehunter.GenerateRules(fieldDescriptions, experiment.ExcludeFieldNames)
	if err != nil {
		msg := fmt.Sprintf("Couldn't make report: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}

	assessment, err := assessRules(rules, experiment, p)
	if err != nil {
		msg := fmt.Sprintf("Couldn't assess rules: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}

	assessment.Sort(experiment.SortOrder)
	assessment.Refine(3)
	sortedRules := assessment.GetRules()

	reportProgress(p, "Tweaking rules")
	tweakableRules := rulehunter.TweakRules(sortedRules, fieldDescriptions)

	assessment2, err := assessRules(tweakableRules, experiment, p)
	if err != nil {
		msg := fmt.Sprintf("Couldn't assess rules: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}

	assessment3, err := assessment.Merge(assessment2)
	if err != nil {
		msg := fmt.Sprintf("Couldn't merge assessments: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}
	assessment3.Sort(experiment.SortOrder)
	assessment3.Refine(1)

	bestNonCombinedRules := assessment3.GetRules()
	numRulesToCombine := 50
	if len(bestNonCombinedRules) < numRulesToCombine {
		numRulesToCombine = len(bestNonCombinedRules)
	}
	combinedRules :=
		rulehunter.CombineRules(bestNonCombinedRules[:numRulesToCombine])

	assessment4, err := assessRules(combinedRules, experiment, p)
	if err != nil {
		msg := fmt.Sprintf("Couldn't assess rules: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}

	assessment5, err := assessment3.Merge(assessment4)
	if err != nil {
		msg := fmt.Sprintf("Couldn't merge assessments: %s", err)
		reportProgress(p, msg)
		err := errors.New(msg)
		return err
	}

	err =
		writeReport(assessment5, experiment, experimentFilename, config.ReportsDir)
	if err != nil {
		reportProgress(p, err.Error())
		return err
	}
	return nil
}

type experimentFile struct {
	Title                 string
	InputFilename         string
	FieldNames            []string
	ExcludeFieldNames     []string
	IsFirstLineFieldNames bool
	Separator             string
	Aggregators           []*rulehunter.AggregatorDesc
	Goals                 []string
	SortOrder             []*rulehunter.SortDesc
}

func loadExperiment(filename string) (*rulehunter.Experiment, error) {
	var f *os.File
	var e experimentFile
	var err error

	f, err = os.Open(filename)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(f)
	if err = dec.Decode(&e); err != nil {
		return nil, err
	}

	input, err := csvinput.New(
		e.FieldNames,
		e.InputFilename,
		rune(e.Separator[0]),
		e.IsFirstLineFieldNames,
	)
	if err != nil {
		return nil, err
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
	return experiment, err
}

func reportProgress(f *os.File, msg string) {
	f.WriteString(fmt.Sprintf("%s\n", msg))
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
	progressFile *os.File,
) (*rulehunter.Assessment, error) {
	var assessment *rulehunter.Assessment
	// TODO: Make this part of the config
	maxProcesses := runtime.NumCPU()
	c := make(chan *rulehunter.AssessRulesMPOutcome)

	reportProgress(progressFile,
		fmt.Sprintf("Assessing rules using %d CPUs...\n", maxProcesses))
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
		reportProgress(progressFile, fmt.Sprintf("Progress: %.2f%%", o.Progress*100))
		assessment = o.Assessment
	}
	reportProgress(progressFile, "Progress: complete")
	return assessment, nil
}

func moveExperimentToSuccess(experimentFilename string, config *config) error {
	experimentFullFilename :=
		filepath.Join(config.ExperimentsDir, experimentFilename)
	experimentSuccessFullFilename :=
		filepath.Join(config.ExperimentsDir, "success", experimentFilename)
	return os.Rename(experimentFullFilename, experimentSuccessFullFilename)
}

func moveExperimentToFail(experimentFilename string, config *config) error {
	experimentFullFilename :=
		filepath.Join(config.ExperimentsDir, experimentFilename)
	experimentFailFullFilename :=
		filepath.Join(config.ExperimentsDir, "fail", experimentFilename)
	return os.Rename(experimentFullFilename, experimentFailFullFilename)
}
