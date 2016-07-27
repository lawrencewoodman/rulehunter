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
package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehuntersrv/config"
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
	SortOrder          []experiment.SortField
	Assessments        []*Assessment
}

func WriteJson(
	assessment *rulehunter.Assessment,
	experiment *experiment.Experiment,
	experimentFilename string,
	tags []string,
	config *config.Config,
) error {
	_assessment := assessment
	_assessment.Sort(experiment.SortOrder)
	_assessment.Refine(1)

	trueAggregators, err := getTrueAggregators(_assessment)
	if err != nil {
		return err
	}

	assessments := make([]*Assessment, len(_assessment.RuleAssessments))
	for i, ruleAssessment := range _assessment.RuleAssessments {
		aggregatorNames := getSortedAggregatorNames(ruleAssessment.Aggregators)
		aggregators := make([]*Aggregator, len(ruleAssessment.Aggregators))
		j := 0
		for _, aggregatorName := range aggregatorNames {
			aggregator := ruleAssessment.Aggregators[aggregatorName]
			difference :=
				calcTrueAggregatorDiff(trueAggregators, aggregatorName, aggregator)
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

func calcTrueAggregatorDiff(
	trueAggregators map[string]*dlit.Literal,
	aggregatorName string,
	aggregatorValue *dlit.Literal,
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
	differenceL := diffExpr.Eval(vars, funcs)
	difference := differenceL.String()
	if err := differenceL.Err(); err != nil {
		difference = "N/A"
	}
	return difference
}
