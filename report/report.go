// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	rhkaggregator "github.com/vlifesystems/rhkit/aggregator"
	rhkassessment "github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal"
)

type Aggregator struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Difference string `json:"difference"`
}

type Assessment struct {
	Rule        string                          `json:"rule"`
	Aggregators []*Aggregator                   `json:"aggregators"`
	Goals       []*rhkassessment.GoalAssessment `json:"goals"`
}

type Report struct {
	Mode               ModeKind                  `json:"mode"`
	Title              string                    `json:"title"`
	Tags               []string                  `json:"tags"`
	Category           string                    `json:"category"`
	Stamp              time.Time                 `json:"stamp"`
	ExperimentFilename string                    `json:"experimentFilename"`
	NumRecords         int64                     `json:"numRecords"`
	SortOrder          []rhkassessment.SortOrder `json:"sortOrder"`
	Aggregators        []AggregatorDesc          `json:"aggregators"`
	Description        *description.Description  `json:"description"`
	Assessments        []*Assessment             `json:"assessments"`
}

type AggregatorDesc struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Arg  string `json:"arg"`
}

// Which mode was the report run in
type ModeKind int

const (
	Train ModeKind = iota
	Test
)

func (m ModeKind) String() string {
	if m == Train {
		return "train"
	}
	return "test"
}

func New(
	mode ModeKind,
	title string,
	desc *description.Description,
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
		Mode:               mode,
		Title:              title,
		Tags:               tags,
		Category:           category,
		Stamp:              time.Now(),
		ExperimentFilename: experimentFilename,
		NumRecords:         assessment.NumRecords,
		SortOrder:          sortOrder,
		Aggregators:        aggregatorDescs,
		Assessments:        assessments,
		Description:        desc,
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
	buildFilename :=
		internal.MakeBuildFilename(r.Mode.String(), r.Category, r.Title)
	reportFilename := filepath.Join(config.BuildDir, "reports", buildFilename)
	return ioutil.WriteFile(reportFilename, json, modePerm)
}

// LoadJSON loads a report from the specified reportFilename in
// reports directory of cfg.BuildDir. Following the reportFilename you
// can specify an optional number of times to try to decode the JSON file.
// This can be useful in situations where you might try to read the
// JSON file before it has been completely written.
func LoadJSON(
	cfg *config.Config,
	reportFilename string,
	args ...int,
) (*Report, error) {
	const sleep = 200 * time.Millisecond
	var report Report
	filename := filepath.Join(cfg.BuildDir, "reports", reportFilename)

	maxTries := 1
	if len(args) == 1 {
		maxTries = args[0]
	} else if len(args) > 1 {
		panic("too many arguments for function")
	}
	for tries := 1; ; tries++ {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		if err = dec.Decode(&report); err != nil {
			if tries > maxTries {
				return nil, fmt.Errorf("can't decode JSON file: %s, %s", filename, err)
			} else {
				time.Sleep(sleep)
			}
		} else {
			break
		}
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
