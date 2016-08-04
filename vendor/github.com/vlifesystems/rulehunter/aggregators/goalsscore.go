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

package aggregators

import (
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/goal"
)

type goalsScoreAggregator struct{}

type goalsScoreSpec struct {
	name string
}

type goalsScoreInstance struct {
	spec *goalsScoreSpec
}

func init() {
	Register("goalsscore", &goalsScoreAggregator{})
}

func (a *goalsScoreAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	d := &goalsScoreSpec{name: name}
	return d, nil
}

func (ad *goalsScoreSpec) New() AggregatorInstance {
	return &goalsScoreInstance{spec: ad}
}

func (ad *goalsScoreSpec) GetName() string {
	return ad.name
}

func (ad *goalsScoreSpec) GetArg() string {
	return ""
}

func (ai *goalsScoreInstance) GetName() string {
	return ai.spec.name
}

func (ai *goalsScoreInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	return nil
}

func (ai *goalsScoreInstance) GetResult(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	instancesMap, err :=
		InstancesToMap(aggregatorInstances, goals, numRecords, ai.GetName())
	if err != nil {
		return dlit.MustNew(err)
	}
	goalsScore := 0.0
	increment := 1.0
	for _, goal := range goals {
		hasPassed, err := goal.Assess(instancesMap)
		if err != nil {
			return dlit.MustNew(err)
		}

		if hasPassed {
			goalsScore += increment
		} else {
			increment = 0.001
		}
	}
	return dlit.MustNew(goalsScore)
}

func (ai *goalsScoreInstance) IsEqual(o AggregatorInstance) bool {
	if _, ok := o.(*goalsScoreInstance); !ok {
		return false
	}
	return ai.GetName() == o.GetName()
}
