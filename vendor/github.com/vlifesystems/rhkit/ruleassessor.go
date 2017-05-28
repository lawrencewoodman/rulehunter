/*
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
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
