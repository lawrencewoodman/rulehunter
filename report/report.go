/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dexpr_go"
	"github.com/lawrencewoodman/dlit_go"
	"github.com/lawrencewoodman/rulehunter"
	"github.com/lawrencewoodman/rulehuntersrv/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Aggregator struct {
	Name       string
	Value      string
	Difference string
}

type Assessment struct {
	Rule        string
	Aggregators []*Aggregator
	Goals       []*rulehunter.GoalAssessment
}

type Report struct {
	Title              string
	Tags               []string
	Stamp              time.Time
	ExperimentFilename string
	NumRecords         int64
	SortOrder          []rulehunter.SortField
	Assessments        []*Assessment
}

func WriteJson(
	assessment *rulehunter.Assessment,
	experiment *rulehunter.Experiment,
	experimentFilename string,
	tags []string,
	config *config.Config,
) error {
	assessment.Sort(experiment.SortOrder)
	assessment.Refine(1)

	trueAggregators, err := getTrueAggregators(assessment)
	if err != nil {
		return err
	}

	assessments := make([]*Assessment, len(assessment.RuleAssessments))
	for i, ruleAssessment := range assessment.RuleAssessments {
		aggregatorNames := getSortedAggregatorNames(ruleAssessment.Aggregators)
		aggregators := make([]*Aggregator, len(ruleAssessment.Aggregators))
		j := 0
		for _, aggregatorName := range aggregatorNames {
			aggregator := ruleAssessment.Aggregators[aggregatorName]
			difference :=
				calcTrueAggregatorDifference(trueAggregators, aggregator, aggregatorName)
			aggregators[j] = &Aggregator{
				aggregatorName,
				aggregator.String(),
				difference,
			}
			j++
		}
		assessments[i] = &Assessment{
			ruleAssessment.Rule.String(),
			aggregators,
			ruleAssessment.Goals,
		}
	}
	report := Report{
		experiment.Title,
		tags,
		time.Now(),
		experimentFilename,
		assessment.NumRecords,
		experiment.SortOrder,
		assessments,
	}
	json, err := json.Marshal(report)
	if err != nil {
		return err
	}
	reportFilename :=
		filepath.Join(config.BuildDir, "reports", experimentFilename)
	return ioutil.WriteFile(reportFilename, json, 0640)
}

func LoadJson(config *config.Config, reportFilename string) (*Report, error) {
	var report Report
	filename := filepath.Join(config.BuildDir, "reports", reportFilename)

	f, err := os.Open(filename)
	if err != nil {
		return &Report{}, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&report); err != nil {
		return &Report{}, err
	}
	return &report, nil
}

func getSortedAggregatorNames(aggregators map[string]*dlit.Literal) []string {
	aggregatorNames := make([]string, len(aggregators))
	j := 0
	for aggregatorName, _ := range aggregators {
		aggregatorNames[j] = aggregatorName
		j++
	}
	sort.Strings(aggregatorNames)
	return aggregatorNames
}

func getTrueAggregators(
	assessment *rulehunter.Assessment,
) (map[string]*dlit.Literal, error) {
	trueRuleAssessment :=
		assessment.RuleAssessments[len(assessment.RuleAssessments)-1]
	if trueRuleAssessment.Rule.String() != "true()" {
		return map[string]*dlit.Literal{}, errors.New("Can't find true() rule")
	}
	trueAggregators := trueRuleAssessment.Aggregators
	return trueAggregators, nil
}

func calcTrueAggregatorDifference(
	trueAggregators map[string]*dlit.Literal,
	aggregatorValue *dlit.Literal,
	aggregatorName string,
) string {
	diffExpr, err := dexpr.New("r - t")
	if err != nil {
		panic(fmt.Sprintf("Couldn't create difference expression: %s", err))
	}
	funcs := map[string]dexpr.CallFun{}
	vars := map[string]*dlit.Literal{
		"r": aggregatorValue,
		"t": trueAggregators[aggregatorName],
	}
	difference := "N/A"
	differenceL := diffExpr.Eval(vars, funcs)
	if !differenceL.IsError() {
		difference = differenceL.String()
	}
	return difference
}
