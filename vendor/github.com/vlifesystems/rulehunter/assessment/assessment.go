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

// Package assessment implements functions to handle Assessments
package assessment

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/aggregators"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/goal"
	"github.com/vlifesystems/rulehunter/internal/ruleassessor"
	"github.com/vlifesystems/rulehunter/rule"
	"sort"
)

type Assessment struct {
	NumRecords      int64
	RuleAssessments []*RuleAssessment
	Flags           map[string]bool
}

type RuleAssessment struct {
	Rule        *rule.Rule
	Aggregators map[string]*dlit.Literal
	Goals       []*GoalAssessment
}

type GoalAssessment struct {
	Expr   string
	Passed bool
}

type ErrNameConflict string

func New(
	numRecords int64,
	goodRuleAssessors []*ruleassessor.RuleAssessor,
	goals []*goal.Goal,
) (*Assessment, error) {
	ruleAssessments := make([]*RuleAssessment, len(goodRuleAssessors))
	for i, ruleAssessment := range goodRuleAssessors {
		rule := ruleAssessment.Rule
		aggregatorsMap, err :=
			aggregators.AggregatorsToMap(
				ruleAssessment.Aggregators,
				ruleAssessment.Goals,
				numRecords,
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

func (a *Assessment) Sort(s []experiment.SortField) {
	sort.Sort(by{a.RuleAssessments, s})
	a.Flags["sorted"] = true
}

// TODO: Test this
func (a *Assessment) IsEqual(o *Assessment) bool {
	if a.NumRecords != o.NumRecords {
		return false
	}

	if len(a.RuleAssessments) != len(o.RuleAssessments) {
		return false
	}
	for i, ruleAssessment := range a.RuleAssessments {
		if !ruleAssessment.IsEqual(o.RuleAssessments[i]) {
			return false
		}
	}

	if len(a.Flags) != len(o.Flags) {
		return false
	}
	for k, v := range a.Flags {
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

// Can optionally pass maximum number of rules to return
func (a *Assessment) GetRules(args ...int) []*rule.Rule {
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

	r := make([]*rule.Rule, numRules)
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
		isInNiRule, operator, field := rule.GetInNiParts()
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
		isTweakable, field, operator, _ := rule.GetTweakableParts()
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

// by implements sort.Interface for []*RuleAssessments based
// on the sortFields
type by struct {
	ruleAssessments []*RuleAssessment
	sortFields      []experiment.SortField
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
		c := compareDlitNums(vI, vJ)

		if direction == experiment.DESCENDING {
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

func (r *RuleAssessment) IsEqual(o *RuleAssessment) bool {
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
		if !goal.IsEqual(o.Goals[i]) {
			return false
		}
	}
	return true
}

func (g *GoalAssessment) IsEqual(o *GoalAssessment) bool {
	return g.Expr == o.Expr && g.Passed == o.Passed
}

func (g *GoalAssessment) String() string {
	return fmt.Sprintf("Expr: %s, Passed: %t", g.Expr, g.Passed)
}
