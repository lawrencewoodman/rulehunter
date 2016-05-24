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

type goalsScore struct {
	name string
}

func newGoalsScore(name string) (*goalsScore, error) {
	a := &goalsScore{name: name}
	return a, nil
}

func (a *goalsScore) CloneNew() Aggregator {
	return &goalsScore{name: a.name}
}

func (a *goalsScore) GetName() string {
	return a.name
}

func (a *goalsScore) GetArg() string {
	return ""
}

func (a *goalsScore) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	return nil
}

func (a *goalsScore) GetResult(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	aggregatorsMap, err :=
		AggregatorsToMap(aggregators, goals, numRecords, a.name)
	if err != nil {
		return dlit.MustNew(err)
	}
	goalsScore := 0.0
	increment := 1.0
	for _, goal := range goals {
		hasPassed, err := goal.Assess(aggregatorsMap)
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

func (a *goalsScore) IsEqual(o Aggregator) bool {
	if _, ok := o.(*goalsScore); !ok {
		return false
	}
	return a.name == o.GetName()
}
