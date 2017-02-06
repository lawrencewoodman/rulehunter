/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of rhkit.

	rhkit is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	rhkit is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with rhkit; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package rhkit

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/aggregators"
	"github.com/vlifesystems/rhkit/experiment"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
	"sort"
	"strings"
)

type Assessment struct {
	NumRecords      int64
	RuleAssessments []*RuleAssessment
	flags           map[string]bool
}

type RuleAssessment struct {
	Rule        rule.Rule
	Aggregators map[string]*dlit.Literal
	Goals       []*GoalAssessment
}

type GoalAssessment struct {
	Expr   string
	Passed bool
}

func newAssessment(
	numRecords int64,
	goodRuleAssessors []*ruleAssessor,
	goals []*goal.Goal,
) (*Assessment, error) {
	ruleAssessments := make([]*RuleAssessment, len(goodRuleAssessors))
	for i, ruleAssessment := range goodRuleAssessors {
		rule := ruleAssessment.Rule
		aggregatorInstancesMap, err :=
			aggregators.InstancesToMap(
				ruleAssessment.Aggregators,
				ruleAssessment.Goals,
				numRecords,
			)
		if err != nil {
			return nil, err
		}
		goalAssessments := make([]*GoalAssessment, len(ruleAssessment.Goals))
		for j, goal := range ruleAssessment.Goals {
			passed, err := goal.Assess(aggregatorInstancesMap)
			if err != nil {
				return &Assessment{}, err
			}
			goalAssessments[j] = &GoalAssessment{goal.String(), passed}
		}
		delete(aggregatorInstancesMap, "numRecords")
		ruleAssessments[i] = &RuleAssessment{
			Rule:        rule,
			Aggregators: aggregatorInstancesMap,
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
		flags:           flags,
	}
	return assessment, nil
}

func (a *Assessment) Sort(s []experiment.SortField) {
	sort.Sort(by{a.RuleAssessments, s})
	a.flags["sorted"] = true
}

func (a *Assessment) IsSorted() bool {
	return a.flags["sorted"]
}

func (a *Assessment) IsRefined() bool {
	return a.flags["refined"]
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

	if len(a.flags) != len(o.flags) {
		return false
	}
	for k, v := range a.flags {
		if v != o.flags[k] {
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
// For example this removes all rules poorer than the True rule
func (sortedAssessment *Assessment) Refine() {
	if !sortedAssessment.IsSorted() {
		panic("Assessment isn't sorted")
	}
	sortedAssessment.excludePoorRules()
	sortedAssessment.excludeSameRecordsRules()
	sortedAssessment.excludePoorerOverlappingRules()
	sortedAssessment.flags["refined"] = true
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
func (a *Assessment) TruncateRuleAssessments(
	numRuleAssessments int,
) *Assessment {
	if !a.IsSorted() {
		panic("Assessment isn't sorted")
	}
	if !a.IsRefined() {
		panic("Assessment isn't refined")
	}
	if len(a.RuleAssessments) < numRuleAssessments {
		numRuleAssessments = len(a.RuleAssessments)
	}
	numNonTrueRuleAssessments := numRuleAssessments - 1

	ruleAssessments := make([]*RuleAssessment, numRuleAssessments)
	for i := 0; i < numNonTrueRuleAssessments; i++ {
		ruleAssessments[i] = a.RuleAssessments[i].clone()
	}

	if numRuleAssessments > 0 {
		trueRuleAssessment := a.RuleAssessments[len(a.RuleAssessments)-1]
		if _, isTrueRule := trueRuleAssessment.Rule.(rule.True); !isTrueRule {
			panic("Assessment doesn't have True rule last")
		}

		ruleAssessments[numNonTrueRuleAssessments] = trueRuleAssessment
	}

	flags := map[string]bool{
		"sorted":  true,
		"refined": true,
	}
	return &Assessment{a.NumRecords, ruleAssessments, flags}
}

// Can optionally pass maximum number of rules to return
func (a *Assessment) GetRules(args ...int) []rule.Rule {
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

	r := make([]rule.Rule, numRules)
	for i, ruleAssessment := range a.RuleAssessments {
		if i >= numRules {
			break
		}
		r[i] = ruleAssessment.Rule
	}
	return r
}

func (sortedAssessment *Assessment) excludeSameRecordsRules() {
	if len(sortedAssessment.RuleAssessments) < 2 {
		return
	}
	lastAggregators := sortedAssessment.RuleAssessments[0].Aggregators
	if len(lastAggregators) <= 3 {
		return
	}

	goodRuleAssessments := make([]*RuleAssessment, 1)
	goodRuleAssessments[0] = sortedAssessment.RuleAssessments[0]
	for _, a := range sortedAssessment.RuleAssessments[1:] {
		aggregatorsMatch := true
		for k, v := range lastAggregators {
			if a.Aggregators[k].String() != v.String() {
				aggregatorsMatch = false
			}
		}
		switch a.Rule.(type) {
		case rule.True:
			if aggregatorsMatch {
				goodRuleAssessments[len(goodRuleAssessments)-1] = a
			} else {
				goodRuleAssessments = append(goodRuleAssessments, a)
			}
			break
		default:
			if !aggregatorsMatch {
				goodRuleAssessments = append(goodRuleAssessments, a)
			}
		}
		lastAggregators = a.Aggregators
	}
	sortedAssessment.RuleAssessments = goodRuleAssessments
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
		if _, isTrueRule := a.Rule.(rule.True); isTrueRule {
			trueFound = true
			break
		}
	}
	if !trueFound {
		panic("No True rule found")
	}
	sortedAssessment.RuleAssessments = goodRuleAssessments
}

func (sortedAssessment *Assessment) excludePoorerInRules(
	numSimilarRules int,
) {
	goodRuleAssessments := make([]*RuleAssessment, 0)
	inFields := make(map[string]int)
	for _, a := range sortedAssessment.RuleAssessments {
		switch x := a.Rule.(type) {
		case *rule.InFV:
			field := x.GetFields()[0]
			n, ok := inFields[field]
			if !ok {
				goodRuleAssessments = append(goodRuleAssessments, a)
				inFields[field] = 1
			} else if n < numSimilarRules {
				goodRuleAssessments = append(goodRuleAssessments, a)
				inFields[field]++
			}
		default:
			goodRuleAssessments = append(goodRuleAssessments, a)
		}
	}
	sortedAssessment.RuleAssessments = goodRuleAssessments
}

func (sortedAssessment *Assessment) excludePoorerOverlappingRules() {
	goodRuleAssessments := make([]*RuleAssessment, 0)
	for i, aI := range sortedAssessment.RuleAssessments {
		switch xI := aI.Rule.(type) {
		case rule.Overlapper:
			overlaps := false
			for j, aJ := range sortedAssessment.RuleAssessments {
				if j >= i {
					break
				}
				if xI.Overlaps(aJ.Rule) {
					overlaps = true
				}
			}
			if !overlaps {
				goodRuleAssessments = append(goodRuleAssessments, aI)
			}
		default:
			goodRuleAssessments = append(goodRuleAssessments, aI)
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

	ruleStrI := b.ruleAssessments[i].Rule.String()
	ruleStrJ := b.ruleAssessments[j].Rule.String()
	ruleLenI := len(ruleStrI)
	ruleLenJ := len(ruleStrJ)
	if ruleLenI != ruleLenJ {
		return ruleLenI < ruleLenJ
	}

	return strings.Compare(ruleStrI, ruleStrJ) == -1
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

func (r *RuleAssessment) clone() *RuleAssessment {
	return &RuleAssessment{r.Rule, r.Aggregators, r.Goals}
}

func (g *GoalAssessment) IsEqual(o *GoalAssessment) bool {
	return g.Expr == o.Expr && g.Passed == o.Passed
}

func (g *GoalAssessment) String() string {
	return fmt.Sprintf("Expr: %s, Passed: %t", g.Expr, g.Passed)
}
