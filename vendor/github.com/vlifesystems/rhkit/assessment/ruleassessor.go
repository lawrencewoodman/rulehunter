// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package assessment

import (
	"errors"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/aggregators"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
)

type ruleAssessor struct {
	Rule        rule.Rule
	Aggregators []aggregators.AggregatorInstance
	Goals       []*goal.Goal
}

func newRuleAssessor(
	rule rule.Rule,
	aggregatorSpecs []aggregators.AggregatorSpec,
	goals []*goal.Goal,
) *ruleAssessor {
	aggregatorInstances :=
		make([]aggregators.AggregatorInstance, len(aggregatorSpecs))
	for i, ad := range aggregatorSpecs {
		aggregatorInstances[i] = ad.New()
	}
	return &ruleAssessor{
		Rule:        rule,
		Aggregators: aggregatorInstances,
		Goals:       goals,
	}
}

func (ra *ruleAssessor) NextRecord(record ddataset.Record) error {
	var ruleIsTrue bool
	var err error
	for _, aggregator := range ra.Aggregators {
		ruleIsTrue, err = ra.Rule.IsTrue(record)
		if err != nil {
			return err
		}
		err = aggregator.NextRecord(record, ruleIsTrue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ra *ruleAssessor) AggregatorValue(
	name string,
	numRecords int64,
) (*dlit.Literal, bool) {
	for _, aggregator := range ra.Aggregators {
		if aggregator.Name() == name {
			return aggregator.Result(ra.Aggregators, ra.Goals, numRecords), true
		}
	}
	// TODO: Test and create specific error type
	err := errors.New("Aggregator doesn't exist")
	l := dlit.MustNew(err)
	return l, false
}
