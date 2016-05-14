/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/
package rulehunter

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/input"
	"github.com/vlifesystems/rulehunter/internal"
	"sort"
)

type Assessment struct {
	NumRecords      int64
	RuleAssessments []*RuleAssessment
	Flags           map[string]bool
}

type RuleAssessment struct {
	Rule        *Rule
	Aggregators map[string]*dlit.Literal
	Goals       []*GoalAssessment
}

type GoalAssessment struct {
	Expr   string
	Passed bool
}

type ErrNameConflict string

func (r *Assessment) Sort(s []SortField) {
	sort.Sort(by{r.RuleAssessments, s})
	r.Flags["sorted"] = true
}

// TODO: Test this
func (r *Assessment) IsEqual(o *Assessment) bool {
	if r.NumRecords != o.NumRecords {
		return false
	}
	for i, ruleAssessment := range r.RuleAssessments {
		if !ruleAssessment.isEqual(o.RuleAssessments[i]) {
			return false
		}
	}
	if len(r.Flags) != len(o.Flags) {
		return false
	}
	for k, v := range r.Flags {
		if v != o.Flags[k] {
			return false
		}
	}
	return true
}

func (r *RuleAssessment) String() string {
	return fmt.Sprintf("Rule: %s, Aggregators: %s, Goals: %s",
		r.Rule, r.Aggregators, r.Goals)
}

// Tidy up rule assessments by removing poor and poorer similar rules
// For example this removes all rules poorer than the 'true()' rule
func (sortedAssessment *Assessment) Refine(numSimilarRules int) {
	if !sortedAssessment.Flags["sorted"] {
		panic("Assessment isn't sorted")
	}
	sortedAssessment.excludePoorRules()
	sortedAssessment.excludePoorerInNiRules(numSimilarRules)
	sortedAssessment.excludePoorerTweakableRules(numSimilarRules)
	sortedAssessment.Flags["refined"] = true
}

func (e ErrNameConflict) Error() string {
	return string(e)
}

func (a *Assessment) Merge(o *Assessment) (*Assessment, error) {
	if a.NumRecords != o.NumRecords {
		// TODO: Create error type
		err := errors.New("Can't merge assessments: Number of records don't match")
		return nil, err
	}
	newRuleAssessments := append(a.RuleAssessments, o.RuleAssessments...)
	flags := map[string]bool{
		"sorted":  false,
		"refined": false,
	}
	return &Assessment{a.NumRecords, newRuleAssessments, flags}, nil
}

// Assessment must be sorted and refined first
func (a *Assessment) LimitRuleAssessments(
	numRuleAssessments int,
) *Assessment {
	if !a.Flags["sorted"] {
		panic("Assessment isn't sorted")
	}
	if !a.Flags["refined"] {
		panic("Assessment isn't refined")
	}
	if len(a.RuleAssessments) < numRuleAssessments {
		numRuleAssessments = len(a.RuleAssessments)
	}

	ruleAssessments := a.RuleAssessments[:numRuleAssessments]

	if len(a.RuleAssessments) != numRuleAssessments {
		trueRuleAssessment := a.RuleAssessments[len(a.RuleAssessments)-1]
		if trueRuleAssessment.Rule.String() != "true()" {
			panic("Assessment doesn't have 'true()' rule last")
		}

		ruleAssessments = append(ruleAssessments, trueRuleAssessment)
	}

	flags := map[string]bool{
		"sorted":  true,
		"refined": true,
	}
	return &Assessment{a.NumRecords, ruleAssessments, flags}
}

// need a progress callback and a specifier for how often to report
func AssessRules(
	rules []*Rule,
	aggregators []internal.Aggregator,
	goals []*internal.Goal,
	input input.Input,
) (*Assessment, error) {
	var allAggregators []internal.Aggregator
	var numRecords int64
	var err error

	allAggregators, err = addDefaultAggregators(aggregators)
	if err != nil {
		return &Assessment{}, err
	}

	ruleAssessments := make([]*ruleAssessment, len(rules))
	for i, rule := range rules {
		ruleAssessments[i] = newRuleAssessment(rule, allAggregators, goals)
	}
	numRecords, err = processInput(input, ruleAssessments)
	if err != nil {
		return &Assessment{}, err
	}
	goodRuleAssessments, err := filterGoodReports(ruleAssessments, numRecords)
	if err != nil {
		return &Assessment{}, err
	}

	assessment, err := makeAssessment(numRecords, goodRuleAssessments, goals)
	return assessment, err
}

type AssessRulesMPOutcome struct {
	Assessment *Assessment
	Err        error
	Progress   float64
	Finished   bool
}

func AssessRulesMP(
	rules []*Rule,
	aggregators []internal.Aggregator,
	goals []*internal.Goal,
	input input.Input,
	maxProcesses int,
	ec chan *AssessRulesMPOutcome,
) {
	var assessment *Assessment
	var isError bool
	ic := make(chan *assessRulesCOutcome)
	numRules := len(rules)
	if numRules < 2 {
		assessment, err := AssessRules(rules, aggregators, goals, input)
		ec <- &AssessRulesMPOutcome{assessment, err, 1.0, true}
		close(ec)
		return
	}
	progressIntervals := 1000
	numProcesses := 0
	if numRules < progressIntervals {
		progressIntervals = numRules
	}
	step := numRules / progressIntervals
	collectedI := 0
	for i := 0; i < numRules; i += step {
		progress := float64(collectedI) / float64(numRules)
		nextI := i + step
		if nextI > numRules {
			nextI = numRules
		}
		rulesPartial := rules[i:nextI]
		inputClone, inputCloneError := input.Clone()
		if inputCloneError != nil {
			ec <- &AssessRulesMPOutcome{nil, inputCloneError, progress, false}
			close(ec)
			return
		}
		go assessRulesC(rulesPartial, aggregators, goals, inputClone, ic)
		numProcesses++

		if numProcesses >= maxProcesses {
			assessment, isError = getCOutcome(ic, ec, assessment, progress)
			if isError {
				return
			}
			collectedI += step
			numProcesses--
		}
	}

	for p := 0; p < numProcesses; p++ {
		progress := float64(collectedI) / float64(numRules)
		assessment, isError = getCOutcome(ic, ec, assessment, progress)
		if isError {
			return
		}
		collectedI += step
	}

	ec <- &AssessRulesMPOutcome{assessment, nil, 1.0, true}
	close(ec)
}

func getCOutcome(
	ic chan *assessRulesCOutcome,
	ec chan *AssessRulesMPOutcome,
	assessment *Assessment,
	progress float64,
) (*Assessment, bool) {
	var retAssessment *Assessment
	var err error
	ec <- &AssessRulesMPOutcome{nil, nil, progress, false}
	assessmentOutcome := <-ic
	if assessmentOutcome.err != nil {
		ec <- &AssessRulesMPOutcome{nil, assessmentOutcome.err, progress, false}
		close(ec)
		return nil, true
	}
	if assessment == nil {
		retAssessment = assessmentOutcome.assessment
	} else {
		retAssessment, err = assessment.Merge(assessmentOutcome.assessment)
		if err != nil {
			ec <- &AssessRulesMPOutcome{nil, err, progress, false}
			close(ec)
			return nil, true
		}
	}
	return retAssessment, false
}

type assessRulesCOutcome struct {
	assessment *Assessment
	err        error
}

func assessRulesC(
	rules []*Rule,
	aggregators []internal.Aggregator,
	goals []*internal.Goal,
	input input.Input,
	c chan *assessRulesCOutcome,
) {
	assessment, err := AssessRules(rules, aggregators, goals, input)
	c <- &assessRulesCOutcome{assessment, err}
}

// by implements sort.Interface for []*RuleAssessments based
// on the sortFields
type by struct {
	ruleAssessments []*RuleAssessment
	sortFields      []SortField
}

func (b by) Len() int { return len(b.ruleAssessments) }
func (b by) Swap(i, j int) {
	b.ruleAssessments[i], b.ruleAssessments[j] =
		b.ruleAssessments[j], b.ruleAssessments[i]
}

func (b by) Less(i, j int) bool {
	var vI *dlit.Literal
	var vJ *dlit.Literal
	for _, sortField := range b.sortFields {
		field := sortField.Field
		direction := sortField.Direction
		vI = b.ruleAssessments[i].Aggregators[field]
		vJ = b.ruleAssessments[j].Aggregators[field]
		if vI == nil || vJ == nil {
			fmt.Printf("vI or vJ == nil field: %s", field)
		}
		c := compareDlitNums(vI, vJ)

		if direction == DESCENDING {
			c *= -1
		}
		if c < 0 {
			return true
		} else if c > 0 {
			return false
		}
	}

	ruleLenI := len(b.ruleAssessments[i].Rule.String())
	ruleLenJ := len(b.ruleAssessments[j].Rule.String())
	return ruleLenI < ruleLenJ
}

func compareDlitNums(l1 *dlit.Literal, l2 *dlit.Literal) int {
	i1, l1IsInt := l1.Int()
	i2, l2IsInt := l2.Int()
	if l1IsInt && l2IsInt {
		if i1 < i2 {
			return -1
		}
		if i1 > i2 {
			return 1
		}
		return 0
	}

	f1, l1IsFloat := l1.Float()
	f2, l2IsFloat := l2.Float()

	if l1IsFloat && l2IsFloat {
		if f1 < f2 {
			return -1
		}
		if f1 > f2 {
			return 1
		}
		return 0
	}
	panic(fmt.Sprintf("Can't compare numbers: %s, %s", l1, l2))
}

func (r *RuleAssessment) isEqual(o *RuleAssessment) bool {
	if r.Rule.String() != o.Rule.String() {
		return false
	}
	if len(r.Aggregators) != len(o.Aggregators) {
		return false
	}
	for aName, value := range r.Aggregators {
		if o.Aggregators[aName].String() != value.String() {
			return false
		}
	}
	if len(r.Goals) != len(o.Goals) {
		return false
	}
	for i, goal := range r.Goals {
		if o.Goals[i].Expr != goal.Expr || o.Goals[i].Passed != goal.Passed {
			return false
		}
	}
	return true
}

// Can optionally pass maximum number of rules to return
func (a *Assessment) GetRules(args ...int) []*Rule {
	var numRules int
	switch len(args) {
	case 0:
		numRules = len(a.RuleAssessments)
	case 1:
		numRules = args[0]
		if len(a.RuleAssessments) < numRules {
			numRules = len(a.RuleAssessments)
		}
	default:
		panic(fmt.Sprintf("incorrect number of arguments passed: %d", len(args)))
	}

	r := make([]*Rule, numRules)
	for i, ruleAssessment := range a.RuleAssessments {
		if i >= numRules {
			break
		}
		r[i] = ruleAssessment.Rule
	}
	return r
}

func (sortedAssessment *Assessment) excludePoorRules() {
	trueFound := false
	goodRuleAssessments := make([]*RuleAssessment, 0)
	for _, a := range sortedAssessment.RuleAssessments {
		numMatches, numMatchesIsInt := a.Aggregators["numMatches"].Int()
		if !numMatchesIsInt {
			panic("numMatches aggregator isn't an int")
		}
		if numMatches > 1 {
			goodRuleAssessments = append(goodRuleAssessments, a)
		}
		if a.Rule.String() == "true()" {
			trueFound = true
			break
		}
	}
	if !trueFound {
		panic("No 'true()' rule found")
	}
	sortedAssessment.RuleAssessments = goodRuleAssessments
}

func (sortedAssessment *Assessment) excludePoorerInNiRules(
	numSimilarRules int,
) {
	goodRuleAssessments := make([]*RuleAssessment, 0)
	inFields := make(map[string]int)
	niFields := make(map[string]int)
	for _, a := range sortedAssessment.RuleAssessments {
		rule := a.Rule
		isInNiRule, operator, field := rule.getInNiParts()
		if !isInNiRule {
			goodRuleAssessments = append(goodRuleAssessments, a)
		} else if operator == "in" {
			n, ok := inFields[field]
			if !ok {
				goodRuleAssessments = append(goodRuleAssessments, a)
				inFields[field] = 1
			} else if n < numSimilarRules {
				goodRuleAssessments = append(goodRuleAssessments, a)
				inFields[field]++
			}
		} else if operator == "ni" {
			n, ok := niFields[field]
			if !ok {
				goodRuleAssessments = append(goodRuleAssessments, a)
				niFields[field] = 1
			} else if n < numSimilarRules {
				goodRuleAssessments = append(goodRuleAssessments, a)
				niFields[field]++
			}
		}
	}
	sortedAssessment.RuleAssessments = goodRuleAssessments
}

func (sortedAssessment *Assessment) excludePoorerTweakableRules(
	numSimilarRules int,
) {
	goodRuleAssessments := make([]*RuleAssessment, 0)
	fieldOperatorIDs := make(map[string]int)
	for _, a := range sortedAssessment.RuleAssessments {
		rule := a.Rule
		isTweakable, field, operator, _ := rule.getTweakableParts()
		if !isTweakable {
			goodRuleAssessments = append(goodRuleAssessments, a)
		} else {
			fieldOperatorID := fmt.Sprintf("%s^%s", field, operator)
			n, ok := fieldOperatorIDs[fieldOperatorID]
			if !ok {
				goodRuleAssessments = append(goodRuleAssessments, a)
				fieldOperatorIDs[fieldOperatorID] = 1
			} else if n < numSimilarRules {
				goodRuleAssessments = append(goodRuleAssessments, a)
				fieldOperatorIDs[fieldOperatorID]++
			}
		}
	}
	sortedAssessment.RuleAssessments = goodRuleAssessments
}

func makeAssessment(
	numRecords int64,
	goodRuleAssessments []*ruleAssessment,
	goals []*internal.Goal,
) (*Assessment, error) {
	ruleAssessments := make([]*RuleAssessment, len(goodRuleAssessments))
	for i, ruleAssessment := range goodRuleAssessments {
		rule := ruleAssessment.Rule
		aggregatorsMap, err :=
			internal.AggregatorsToMap(
				ruleAssessment.Aggregators,
				ruleAssessment.Goals,
				numRecords,
				"",
			)
		if err != nil {
			return nil, err
		}
		goalAssessments := make([]*GoalAssessment, len(ruleAssessment.Goals))
		for j, goal := range ruleAssessment.Goals {
			passed, err := goal.Assess(aggregatorsMap)
			if err != nil {
				return &Assessment{}, err
			}
			goalAssessments[j] = &GoalAssessment{goal.String(), passed}
		}
		delete(aggregatorsMap, "numRecords")
		ruleAssessments[i] = &RuleAssessment{
			Rule:        rule,
			Aggregators: aggregatorsMap,
			Goals:       goalAssessments,
		}
	}
	flags := map[string]bool{
		"sorted":  false,
		"refined": false,
	}
	assessment := &Assessment{
		NumRecords:      numRecords,
		RuleAssessments: ruleAssessments,
		Flags:           flags,
	}
	return assessment, nil
}

func filterGoodReports(
	ruleAssessments []*ruleAssessment,
	numRecords int64) ([]*ruleAssessment, error) {
	goodRuleAssessments := make([]*ruleAssessment, 0)
	for _, ruleAssessment := range ruleAssessments {
		numMatches, exists :=
			ruleAssessment.getAggregatorValue("numMatches", numRecords)
		if !exists {
			// TODO: Create a proper error for this?
			err := errors.New("numMatches doesn't exist in aggregators")
			return goodRuleAssessments, err
		}
		numMatchesInt, isInt := numMatches.Int()
		if !isInt {
			// TODO: Create a proper error for this?
			err := errors.New(fmt.Sprintf("Can't cast to Int: %q", numMatches))
			return goodRuleAssessments, err
		}
		if numMatchesInt > 0 {
			goodRuleAssessments = append(goodRuleAssessments, ruleAssessment)
		}
	}
	return goodRuleAssessments, nil
}

func processInput(input input.Input,
	ruleAssessments []*ruleAssessment) (int64, error) {
	numRecords := int64(0)
	// TODO: test this rewinds properly
	if err := input.Rewind(); err != nil {
		return numRecords, err
	}

	for input.Next() {
		record, err := input.Read()
		if err != nil {
			return numRecords, err
		}
		numRecords++
		for _, ruleAssessment := range ruleAssessments {
			err := ruleAssessment.nextRecord(record)
			if err != nil {
				return numRecords, err
			}
		}
	}

	return numRecords, input.Err()
}

func addDefaultAggregators(
	aggregators []internal.Aggregator,
) ([]internal.Aggregator, error) {
	newAggregators := make([]internal.Aggregator, 2)
	numMatchesAggregator, err := internal.NewCountAggregator("numMatches", "1==1")
	if err != nil {
		return newAggregators, err
	}
	percentMatchesAggregator, err :=
		internal.NewCalcAggregator("percentMatches",
			"roundto(100.0 * numMatches / numRecords, 2)")
	if err != nil {
		return newAggregators, err
	}
	goalsPassedScoreAggregator, err :=
		internal.NewGoalsPassedScoreAggregator("numGoalsPassed")
	if err != nil {
		return newAggregators, err
	}
	newAggregators[0] = numMatchesAggregator
	newAggregators[1] = percentMatchesAggregator
	newAggregators = append(newAggregators, aggregators...)
	newAggregators = append(newAggregators, goalsPassedScoreAggregator)
	return newAggregators, nil
}
