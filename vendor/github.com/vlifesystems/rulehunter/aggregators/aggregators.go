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

// Package aggregators describes and handles Aggregators
package aggregators

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/goal"
)

type Aggregator interface {
	CloneNew() Aggregator
	GetName() string
	GetArg() string
	GetResult([]Aggregator, []*goal.Goal, int64) *dlit.Literal
	NextRecord(map[string]*dlit.Literal, bool) error
	IsEqual(Aggregator) bool
}

// Create a new Aggregator where 'name' is what the aggregator will be
// known as, 'aggType' is the type of the Aggregator and 'args' are any
// arguments to pass to the Aggregator.  The valid values for 'aggType'
// are: accuracy, calc, count, percent, sum, goalsscore.
func New(name string, aggType string, args ...string) (Aggregator, error) {
	var r Aggregator
	var err error

	checkArgs := func(validNumArgs int) error {
		if len(args) != validNumArgs {
			return fmt.Errorf("Invalid number of arguments for aggregator: %s",
				aggType)
		}
		return nil

	}
	switch aggType {
	case "accuracy":
		if err = checkArgs(1); err != nil {
			return r, err
		}
		r, err = newAccuracy(name, args[0])
	case "calc":
		if err = checkArgs(1); err != nil {
			return r, err
		}
		r, err = newCalc(name, args[0])
	case "count":
		if err = checkArgs(1); err != nil {
			return r, err
		}
		r, err = newCount(name, args[0])
	case "percent":
		if err = checkArgs(1); err != nil {
			return r, err
		}
		r, err = newPercent(name, args[0])
	case "sum":
		if err = checkArgs(1); err != nil {
			return r, err
		}
		r, err = newSum(name, args[0])
	case "goalsscore":
		if err = checkArgs(0); err != nil {
			return r, err
		}
		r, err = newGoalsScore(name)
	default:
		err = fmt.Errorf("Unrecognized aggregator: %s", aggType)
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Can't make aggregator: %s", err))
	}
	return r, err
}

func MustNew(name string, aggType string, args ...string) Aggregator {
	a, err := New(name, aggType, args...)
	if err != nil {
		panic(err)
	}
	return a
}

// Get the results of each Aggregator and return
// as a map with the aggregator name as the key
func AggregatorsToMap(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
	stopNames ...string,
) (map[string]*dlit.Literal, error) {
	r := make(map[string]*dlit.Literal, len(aggregators))
	numRecordsL := dlit.MustNew(numRecords)
	r["numRecords"] = numRecordsL
	for _, aggregator := range aggregators {
		for _, stopName := range stopNames {
			if stopName == aggregator.GetName() {
				return r, nil
			}
		}
		l := aggregator.GetResult(aggregators, goals, numRecords)
		if err := l.Err(); err != nil {
			return r, err
		}
		r[aggregator.GetName()] = l
	}
	return r, nil
}
