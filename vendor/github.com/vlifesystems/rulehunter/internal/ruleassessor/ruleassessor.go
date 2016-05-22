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

package ruleassessor

import (
	"errors"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/aggregators"
	"github.com/vlifesystems/rulehunter/goal"
	"github.com/vlifesystems/rulehunter/rule"
)

type RuleAssessor struct {
	Rule        *rule.Rule
	Aggregators []aggregators.Aggregator
	Goals       []*goal.Goal
}

func New(
	rule *rule.Rule,
	_aggregators []aggregators.Aggregator,
	goals []*goal.Goal,
) *RuleAssessor {
	// Clone the aggregators and goals to ensure the results are
	// specific to this rule
	cloneAggregators := make([]aggregators.Aggregator, len(_aggregators))
	for i, a := range _aggregators {
		cloneAggregators[i] = a.CloneNew()
	}
	cloneGoals := make([]*goal.Goal, len(goals))
	for i, g := range goals {
		cloneGoals[i] = g.Clone()
	}
	return &RuleAssessor{Rule: rule, Aggregators: cloneAggregators,
		Goals: cloneGoals}
}

func (ra *RuleAssessor) NextRecord(record map[string]*dlit.Literal) error {
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

func (ra *RuleAssessor) GetAggregatorValue(
	name string,
	numRecords int64,
) (*dlit.Literal, bool) {
	for _, aggregator := range ra.Aggregators {
		if aggregator.GetName() == name {
			return aggregator.GetResult(ra.Aggregators, ra.Goals, numRecords), true
		}
	}
	// TODO: Test and create specific error type
	err := errors.New("Aggregator doesn't exist")
	l := dlit.MustNew(err)
	return l, false
}
