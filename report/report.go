// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package report

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	rhkaggregator "github.com/vlifesystems/rhkit/aggregator"
	rhkassessment "github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal"
)

type Aggregator struct {
	Name       string
	Value      string
	Difference string
}

type Assessment struct {
	Rule        string
	Aggregators []*Aggregator
	Goals       []*rhkassessment.GoalAssessment
}

type Report struct {
	Title              string
	Tags               []string
	Category           string
	Stamp              time.Time
	ExperimentFilename string
	NumRecords         int64
	SortOrder          []rhkassessment.SortOrder
	Aggregators        []AggregatorDesc
	Assessments        []*Assessment
}

type AggregatorDesc struct {
	Name string
	Kind string
	Arg  string
}

func New(
	title string,
	assessment *rhkassessment.Assessment,
	aggregators []rhkaggregator.Spec,
	sortOrder []rhkassessment.SortOrder,
	experimentFilename string,
	tags []string,
	category string,
) *Report {
	assessment.Sort(sortOrder)
	assessment.Refine()

	trueAggregators, err := getTrueAggregators(assessment)
	if err != nil {
		panic(err)
	}

	aggregatorDescs := make([]AggregatorDesc, len(aggregators))
	for i, as := range aggregators {
		aggregatorDescs[i] = AggregatorDesc{
			Name: as.Name(),
			Kind: as.Kind(),
			Arg:  as.Arg(),
		}
	}

	assessments := make([]*Assessment, len(assessment.RuleAssessments))
	for i, ruleAssessment := range assessment.RuleAssessments {
		aggregatorNames := getSortedAggregatorNames(ruleAssessment.Aggregators)
		aggregators := make([]*Aggregator, len(ruleAssessment.Aggregators))
		for j, aggregatorName := range aggregatorNames {
			aggregator := ruleAssessment.Aggregators[aggregatorName]
			difference :=
				calcTrueAggregatorDiff(trueAggregators, aggregatorName, aggregator)
			aggregators[j] = &Aggregator{
				Name:       aggregatorName,
				Value:      aggregator.String(),
				Difference: difference,
			}
		}
		assessments[i] = &Assessment{
			ruleAssessment.Rule.String(),
			aggregators,
			ruleAssessment.Goals,
		}
	}
	return &Report{
		Title:              title,
		Tags:               tags,
		Category:           category,
		Stamp:              time.Now(),
		ExperimentFilename: experimentFilename,
		NumRecords:         assessment.NumRecords,
		SortOrder:          sortOrder,
		Aggregators:        aggregatorDescs,
		Assessments:        assessments,
	}
}

func (r *Report) WriteJSON(config *config.Config) error {
	// File mode permission:
	// No special permission bits
	// User: Read, Write
	// Group: Read
	// Other: None
	const modePerm = 0640
	json, err := json.Marshal(r)
	if err != nil {
		return err
	}
	buildFilename := internal.MakeBuildFilename(r.Category, r.Title)
	reportFilename := filepath.Join(config.BuildDir, "reports", buildFilename)
	return ioutil.WriteFile(reportFilename, json, modePerm)
}

func LoadJSON(config *config.Config, reportFilename string) (*Report, error) {
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
	assessment *rhkassessment.Assessment,
) (map[string]*dlit.Literal, error) {
	trueRuleAssessment :=
		assessment.RuleAssessments[len(assessment.RuleAssessments)-1]
	if _, isTrueRule := trueRuleAssessment.Rule.(rule.True); !isTrueRule {
		return map[string]*dlit.Literal{}, errors.New("can't find true() rule")
	}
	trueAggregators := trueRuleAssessment.Aggregators
	return trueAggregators, nil
}

func calcTrueAggregatorDiff(
	trueAggregators map[string]*dlit.Literal,
	aggregatorName string,
	aggregatorValue *dlit.Literal,
) string {
	funcs := map[string]dexpr.CallFun{}
	diffExpr := dexpr.MustNew("r - t", funcs)
	vars := map[string]*dlit.Literal{
		"r": aggregatorValue,
		"t": trueAggregators[aggregatorName],
	}
	differenceL := diffExpr.Eval(vars)
	if err := differenceL.Err(); err != nil {
		return "N/A"
	}
	return differenceL.String()
}
