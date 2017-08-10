/*
	rulehunter - A server to find rules in data based on user specified goals
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
package report

import (
	"encoding/json"
	"errors"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	rhkaggregator "github.com/vlifesystems/rhkit/aggregator"
	rhkassessment "github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
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
	Goals       []*rhkassessment.GoalAssessment
}

type Report struct {
	Title              string
	Tags               []string
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
		title,
		tags,
		time.Now(),
		experimentFilename,
		assessment.NumRecords,
		sortOrder,
		aggregatorDescs,
		assessments,
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
	reportFilename :=
		filepath.Join(config.BuildDir, "reports", r.ExperimentFilename)
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
