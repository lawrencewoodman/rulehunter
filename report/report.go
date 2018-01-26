// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	Name          string `json:"name"`
	OriginalValue string `json:"originalValue"`
	RuleValue     string `json:"ruleValue"`
	Difference    string `json:"difference"`
}

type Goal struct {
	Expr           string `json:"expr"`
	OriginalPassed bool   `json:"originalPassed"`
	RulePassed     bool   `json:"rulePassed"`
}

type Assessment struct {
	Rule        string        `json:"rule"`
	Aggregators []*Aggregator `json:"aggregators"`
	Goals       []*Goal       `json:"goals"`
}

func (a *Assessment) String() string {
	return fmt.Sprintf("{Rule: %s, Aggregators: %v, Goals: %v}",
		a.Rule, a.Aggregators, a.Goals)
}

func (g *Goal) String() string {
	return fmt.Sprintf("{Expr: %s, OriginalPassed: %t, RulePassed: %t}",
		g.Expr, g.OriginalPassed, g.RulePassed)
}

func (a *Aggregator) String() string {
	return fmt.Sprintf(
		"{Name: %s, OriginalValue: %s, RuleValue: %s, Difference: %s}",
		a.Name, a.OriginalValue, a.RuleValue, a.Difference,
	)
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

	aggregatorDescs := make([]AggregatorDesc, len(aggregators))
	for i, as := range aggregators {
		aggregatorDescs[i] = AggregatorDesc{
			Name: as.Name(),
			Kind: as.Kind(),
			Arg:  as.Arg(),
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
		Assessments:        makeAssessments(assessment),
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

func makeAssessments(assessment *rhkassessment.Assessment) []*Assessment {
	trueRuleAssessment, err := getTrueRuleAssessment(assessment)
	if err != nil {
		panic(err)
	}
	trueAggregators := trueRuleAssessment.Aggregators
	trueGoals := trueRuleAssessment.Goals

	assessments := make([]*Assessment, len(assessment.RuleAssessments)-1)
	for i, ruleAssessment := range assessment.RuleAssessments {
		if _, isTrueRule := ruleAssessment.Rule.(rule.True); !isTrueRule {
			assessments[i] = &Assessment{
				Rule: ruleAssessment.Rule.String(),
				Aggregators: makeAggregators(
					trueAggregators,
					ruleAssessment.Aggregators,
				),
				Goals: makeGoals(trueGoals, ruleAssessment.Goals),
			}
		}
	}
	return assessments
}

func makeAggregators(
	trueAggregators map[string]*dlit.Literal,
	ruleAggregators map[string]*dlit.Literal,
) []*Aggregator {
	aggregatorNames := getSortedAggregatorNames(ruleAggregators)
	aggregators := make([]*Aggregator, len(ruleAggregators))
	for j, aggregatorName := range aggregatorNames {
		aggregator := ruleAggregators[aggregatorName]
		difference :=
			calcTrueAggregatorDiff(trueAggregators, aggregatorName, aggregator)
		aggregators[j] = &Aggregator{
			Name:          aggregatorName,
			OriginalValue: trueAggregators[aggregatorName].String(),
			RuleValue:     aggregator.String(),
			Difference:    difference,
		}
	}
	return aggregators
}

func makeGoals(
	trueGoals []*rhkassessment.GoalAssessment,
	ruleGoals []*rhkassessment.GoalAssessment,
) []*Goal {
	goals := make([]*Goal, len(ruleGoals))
	for i, g := range ruleGoals {
		goals[i] = &Goal{
			Expr:           g.Expr,
			OriginalPassed: trueGoals[i].Passed,
			RulePassed:     g.Passed,
		}
	}
	return goals
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

func getTrueRuleAssessment(
	assessment *rhkassessment.Assessment,
) (*rhkassessment.RuleAssessment, error) {
	trueRuleAssessment :=
		assessment.RuleAssessments[len(assessment.RuleAssessments)-1]
	if _, isTrueRule := trueRuleAssessment.Rule.(rule.True); !isTrueRule {
		return nil, errors.New("can't find true() rule")
	}
	return trueRuleAssessment, nil
}

func numDecPlaces(s string) int {
	i := strings.IndexByte(s, '.')
	if i > -1 {
		s = strings.TrimRight(s, "0")
		return len(s) - i - 1
	}
	return 0
}

type CantConvertToTypeError struct {
	Kind  string
	Value *dlit.Literal
}

func (e CantConvertToTypeError) Error() string {
	return fmt.Sprintf("can't convert to %s: %s", e.Kind, e.Value)
}

// roundTo returns a number n,  rounded to a number of decimal places dp.
// This uses round half-up to tie-break
func roundTo(n *dlit.Literal, dp int) (*dlit.Literal, error) {

	if _, isInt := n.Int(); isInt {
		return n, nil
	}

	x, isFloat := n.Float()
	if !isFloat {
		if err := n.Err(); err != nil {
			return n, err
		}
		err := CantConvertToTypeError{Kind: "float", Value: n}
		r := dlit.MustNew(err)
		return r, err
	}

	// Prevent rounding errors where too high dp is used
	xNumDP := numDecPlaces(n.String())
	if dp > xNumDP {
		dp = xNumDP
	}
	shift := math.Pow(10, float64(dp))
	return dlit.New(math.Floor(.5+x*shift) / shift)
}

func calcTrueAggregatorDiff(
	trueAggregators map[string]*dlit.Literal,
	aggregatorName string,
	aggregatorValue *dlit.Literal,
) string {
	funcs := map[string]dexpr.CallFun{}
	maxDP := numDecPlaces(aggregatorValue.String())
	trueAggregatorValueDP := numDecPlaces(trueAggregators[aggregatorName].String())
	if trueAggregatorValueDP > maxDP {
		maxDP = trueAggregatorValueDP
	}
	diffExpr := dexpr.MustNew("r - t", funcs)
	vars := map[string]*dlit.Literal{
		"r": aggregatorValue,
		"t": trueAggregators[aggregatorName],
	}
	differenceL := diffExpr.Eval(vars)
	if err := differenceL.Err(); err != nil {
		return "N/A"
	}
	roundedDifferenceL, err := roundTo(differenceL, maxDP)
	if err != nil {
		return "N/A"
	}
	return roundedDifferenceL.String()
}
