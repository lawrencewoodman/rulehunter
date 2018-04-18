// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package assessment assesses rules to meet user defined goals
package assessment

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/aggregator"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
)

var ErrNumRecordsChanged = errors.New("number of records changed in dataset")

type Assessment struct {
	NumRecords      int64             `json:"numRecords"`
	RuleAssessments []*RuleAssessment `json:"ruleAssessments"`
	aggregatorSpecs []aggregator.Spec
	goals           []*goal.Goal
	flags           map[string]bool
	mux             sync.RWMutex
}

type GoalAssessment struct {
	Expr   string `json:"expr"`
	Passed bool   `json:"passed"`
}

func New(aggregatorSpecs []aggregator.Spec, goals []*goal.Goal) *Assessment {
	a := &Assessment{
		NumRecords:      0,
		RuleAssessments: []*RuleAssessment{},
		aggregatorSpecs: aggregatorSpecs,
		goals:           goals,
	}
	a.resetFlags()
	return a
}

func (a *Assessment) AddRules(rules []rule.Rule) {
	a.mux.Lock()
	defer a.mux.Unlock()
	for _, rule := range rules {
		a.RuleAssessments = append(
			a.RuleAssessments,
			newRuleAssessment(rule, a.aggregatorSpecs, a.goals),
		)
	}
}

func (a *Assessment) Sort(s []SortOrder) {
	sort.Sort(by{a.RuleAssessments, s})
	a.flags["sorted"] = true
}

func (a *Assessment) IsSorted() bool {
	return a.flags["sorted"]
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

// Refine removes ruleAssessments that are poorer than similar rules
func (sortedAssessment *Assessment) Refine() {
	if !sortedAssessment.IsSorted() {
		panic("Assessment isn't sorted")
	}
	sortedAssessment.excludePoorRules()
	sortedAssessment.excludeSameRecordsRules()
	sortedAssessment.excludePoorerOverlappingRules()
}

func (a *Assessment) Merge(o *Assessment) (*Assessment, error) {
	if a.NumRecords != o.NumRecords {
		return nil, ErrNumRecordsChanged
	}
	newRuleAssessments := append(a.RuleAssessments, o.RuleAssessments...)
	r := &Assessment{
		NumRecords:      a.NumRecords,
		RuleAssessments: newRuleAssessments,
	}
	r.resetFlags()
	return r, nil
}

func getTrueRuleAssessment(ruleAssessments []*RuleAssessment) *RuleAssessment {
	for _, ra := range ruleAssessments {
		if _, isTrueRule := ra.Rule.(rule.True); isTrueRule {
			return ra
		}
	}
	return nil
}

// Assessment must be sorted first
func (a *Assessment) TruncateRuleAssessments(
	maxRuleAssessments int,
) *Assessment {
	if !a.IsSorted() {
		panic("Assessment isn't sorted")
	}
	trueRuleAssessment := getTrueRuleAssessment(a.RuleAssessments)

	if trueRuleAssessment == nil {
		panic("Assessment doesn't have True rule")
	}

	ruleAssessments := []*RuleAssessment{}
	for i, ra := range a.RuleAssessments {
		if i >= maxRuleAssessments {
			break
		}
		ruleAssessments = append(ruleAssessments, ra.clone())
	}

	if getTrueRuleAssessment(ruleAssessments) != nil && maxRuleAssessments > 0 {
		if len(ruleAssessments) >= maxRuleAssessments {
			ruleAssessments[len(ruleAssessments)-1] = trueRuleAssessment.clone()
		} else {
			ruleAssessments = append(ruleAssessments, trueRuleAssessment.clone())
		}
	}

	flags := map[string]bool{
		"sorted": true,
	}
	return &Assessment{
		NumRecords:      a.NumRecords,
		RuleAssessments: ruleAssessments,
		flags:           flags,
	}
}

// Can optionally pass maximum number of rules to return
func (a *Assessment) Rules(args ...int) []rule.Rule {
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

// AssessRules assesses the given rules against a Dataset and
// adds their assessment to the existing assessment.
// This function is thread safe.
func (a *Assessment) AssessRules(
	dataset ddataset.Dataset,
	rules []rule.Rule,
) error {
	ruleAssessments := make([]*RuleAssessment, len(rules))
	for i, rule := range rules {
		ruleAssessments[i] = newRuleAssessment(rule, a.aggregatorSpecs, a.goals)
	}
	numRecords, err := processDataset(dataset, ruleAssessments)
	if err != nil {
		return err
	}
	if a.NumRecords == 0 {
		a.mux.Lock()
		a.NumRecords = numRecords
		a.mux.Unlock()
	} else if numRecords != a.NumRecords {
		return ErrNumRecordsChanged
	}
	return a.addRuleAssessments(ruleAssessments)
}

// ProcessRecord assesses all the Assessment rules against
// the supplied record
func (a *Assessment) ProcessRecord(r ddataset.Record) error {
	for _, ruleAssessment := range a.RuleAssessments {
		err := ruleAssessment.NextRecord(r)
		if err != nil {
			return err
		}
	}
	a.mux.Lock()
	a.NumRecords++
	a.mux.Unlock()
	return nil
}

// Update the internal analysis for all the RuleAssessments
func (a *Assessment) Update() error {
	a.mux.Lock()
	defer a.mux.Unlock()
	for _, ruleAssessment := range a.RuleAssessments {
		if err := ruleAssessment.update(a.NumRecords); err != nil {
			return err
		}
	}
	return nil
}

func processDataset(
	dataset ddataset.Dataset,
	ruleAssessments []*RuleAssessment,
) (int64, error) {
	numRecords := int64(0)
	conn, err := dataset.Open()
	if err != nil {
		return numRecords, err
	}
	defer conn.Close()

	for conn.Next() {
		record := conn.Read()
		numRecords++
		for _, ruleAssessment := range ruleAssessments {
			err := ruleAssessment.NextRecord(record)
			if err != nil {
				return numRecords, err
			}
		}
	}
	return numRecords, conn.Err()
}

func (a *Assessment) resetFlags() {
	a.flags = map[string]bool{
		"sorted": false,
	}
}

func (a *Assessment) addRuleAssessments(
	ruleAssessments []*RuleAssessment,
) error {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.resetFlags()
	for _, ruleAssessment := range ruleAssessments {
		if err := ruleAssessment.update(a.NumRecords); err != nil {
			return err
		}
		numMatches, ok := ruleAssessment.Aggregators["numMatches"]
		if !ok {
			panic("numMatches doesn't exist in aggregators")
		}
		numMatchesInt, isInt := numMatches.Int()
		if !isInt {
			panic(fmt.Sprintf("can't cast numMatches to Int: %s", numMatches))
		}
		if numMatchesInt > 0 {
			a.RuleAssessments = append(a.RuleAssessments, ruleAssessment.clone())
		}
	}
	return nil
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
				break
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
	goodRuleAssessments := []*RuleAssessment{}
	for _, a := range sortedAssessment.RuleAssessments {
		percentMatches, percentMatchesIsFloat :=
			a.Aggregators["percentMatches"].Float()
		if !percentMatchesIsFloat {
			panic("percentMatches aggregator isn't a float")
		}
		if percentMatches >= 0.5 {
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

func (g *GoalAssessment) IsEqual(o *GoalAssessment) bool {
	return g.Expr == o.Expr && g.Passed == o.Passed
}
