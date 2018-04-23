// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package assessment

import (
	"fmt"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/aggregator"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
)

type RuleAssessment struct {
	Rule        rule.Rule                `json:"rule"`
	Aggregators map[string]*dlit.Literal `json:"aggregators"`
	Goals       []*GoalAssessment        `json:"goals"`
	aggregators []aggregator.Instance
	goals       []*goal.Goal
}

type AggregatorError struct {
	Name string
	Err  error
}

func (ae AggregatorError) Error() string {
	return fmt.Sprintf("problem with aggregator: %s, %s", ae.Name, ae.Err)
}

func newRuleAssessment(
	rule rule.Rule,
	aggregatorSpecs []aggregator.Spec,
	goals []*goal.Goal,
) *RuleAssessment {
	aggregatorInstances := make([]aggregator.Instance, len(aggregatorSpecs))
	for i, ad := range aggregatorSpecs {
		aggregatorInstances[i] = ad.New()
	}
	return &RuleAssessment{
		Rule:        rule,
		aggregators: aggregatorInstances,
		goals:       goals,
	}
}

func (r *RuleAssessment) String() string {
	return fmt.Sprintf("{Rule: %s, Aggregators: %v, Goals: %v}",
		r.Rule, r.Aggregators, r.Goals)
}

func (r *RuleAssessment) NextRecord(record ddataset.Record) error {
	var ruleIsTrue bool
	var err error
	for _, aggregator := range r.aggregators {
		ruleIsTrue, err = r.Rule.IsTrue(record)
		if err != nil {
			return err
		}
		err = aggregator.NextRecord(record, ruleIsTrue)
		if err != nil {
			return AggregatorError{Name: aggregator.Name(), Err: err}
		}
	}
	return nil
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
	return &RuleAssessment{
		Rule:        r.Rule,
		Aggregators: r.Aggregators,
		Goals:       r.Goals,
		aggregators: r.aggregators,
		goals:       r.goals,
	}
}

func (r *RuleAssessment) update(numRecords int64) error {
	aggregatorInstancesMap, err :=
		aggregator.InstancesToMap(r.aggregators, r.goals, numRecords)
	if err != nil {
		return err
	}
	goalAssessments := make([]*GoalAssessment, len(r.goals))
	for j, goal := range r.goals {
		passed, err := goal.Assess(aggregatorInstancesMap)
		if err != nil {
			return err
		}
		goalAssessments[j] = &GoalAssessment{goal.String(), passed}
	}
	// TODO: Work out why this is here
	delete(aggregatorInstancesMap, "numRecords")
	r.Aggregators = aggregatorInstancesMap
	r.Goals = goalAssessments
	return nil
}
