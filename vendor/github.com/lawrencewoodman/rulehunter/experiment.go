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
package rulehunter

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/rulehunter/input"
	"github.com/lawrencewoodman/rulehunter/internal"
	"regexp"
)

type ExperimentDesc struct {
	Title         string
	Input         input.Input
	Fields        []string
	ExcludeFields []string
	Aggregators   []*AggregatorDesc
	Goals         []string
	SortOrder     []*SortDesc
}

type AggregatorDesc struct {
	Name     string
	Function string
	Arg      string
}

type SortDesc struct {
	AggregatorName string
	Direction      string
}

type Experiment struct {
	Title             string
	Input             input.Input
	FieldNames        []string
	ExcludeFieldNames []string
	Aggregators       []internal.Aggregator
	Goals             []*internal.Goal
	SortOrder         []SortField
}

type SortField struct {
	Field     string
	Direction direction
}

type direction int

const (
	ASCENDING direction = iota
	DESCENDING
)

func (d direction) String() string {
	if d == ASCENDING {
		return "ascending"
	}
	return "descending"
}

func MakeExperiment(e *ExperimentDesc) (*Experiment, error) {
	var goals []*internal.Goal
	var aggregators []internal.Aggregator
	var sortOrder []SortField
	var err error

	err = checkExperimentDescValid(e)
	if err != nil {
		return nil, err
	}
	goals, err = makeGoals(e.Goals)
	if err != nil {
		return nil, err
	}
	aggregators, err = makeAggregators(e.Aggregators)
	if err != nil {
		return nil, err
	}

	sortOrder, err = makeSortOrder(e.SortOrder)
	if err != nil {
		return nil, err
	}

	return &Experiment{
		Title:             e.Title,
		Input:             e.Input,
		FieldNames:        e.Fields,
		ExcludeFieldNames: e.ExcludeFields,
		Aggregators:       aggregators,
		Goals:             goals,
		SortOrder:         sortOrder,
	}, nil
}

func (e *Experiment) Close() error {
	return e.Input.Close()
}

func checkExperimentDescValid(e *ExperimentDesc) error {
	if len(e.Fields) < 2 {
		return errors.New("Must specify at least two field names")
	}
	err := checkSortDescsValid(e)
	if err != nil {
		return err
	}

	err = checkFieldsValid(e)
	if err != nil {
		return err
	}

	err = checkExcludeFieldsValid(e)
	if err != nil {
		return err
	}

	err = checkAggregatorsValid(e)
	if err != nil {
		return err
	}
	return nil
}

var validIdentifierRegexp = regexp.MustCompile("^[a-zA-z]([0-9a-zA-z_])*$")

func checkFieldsValid(e *ExperimentDesc) error {
	for _, field := range e.Fields {
		if !validIdentifierRegexp.MatchString(field) {
			return fmt.Errorf("Invalid field name: %s", field)
		}
	}
	return nil
}

func checkSortDescsValid(e *ExperimentDesc) error {
	for _, sortDesc := range e.SortOrder {
		if sortDesc.Direction != "ascending" && sortDesc.Direction != "descending" {
			return fmt.Errorf("Invalid sort direction: %s, for field: %s",
				sortDesc.Direction, sortDesc.AggregatorName)
		}
		sortName := sortDesc.AggregatorName
		nameFound := false
		for _, aggregator := range e.Aggregators {
			if aggregator.Name == sortName {
				nameFound = true
				break
			}
		}
		if !nameFound &&
			sortName != "percentMatches" &&
			sortName != "numMatches" &&
			sortName != "numGoalsPassed" {
			return fmt.Errorf("Invalid sort field: %s", sortName)
		}
	}
	return nil
}

func checkExcludeFieldsValid(e *ExperimentDesc) error {
	for _, excludeField := range e.ExcludeFields {
		found := false
		for _, field := range e.Fields {
			if excludeField == field {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Invalid exclude field: %s", excludeField)
		}
	}
	return nil
}

func checkAggregatorsValid(e *ExperimentDesc) error {
	for _, aggregator := range e.Aggregators {
		if !validIdentifierRegexp.MatchString(aggregator.Name) {
			return fmt.Errorf("Invalid aggregator name: %s", aggregator.Name)
		}
		nameClash := false
		for _, field := range e.Fields {
			if aggregator.Name == field {
				nameClash = true
				break
			}
		}
		if nameClash {
			return fmt.Errorf("Aggregator name clashes with field name: %s",
				aggregator.Name)
		} else if aggregator.Name == "percentMatches" ||
			aggregator.Name == "numMatches" ||
			aggregator.Name == "numGoalsPassed" {
			return fmt.Errorf("Aggregator name reserved: %s", aggregator.Name)
		}
	}
	return nil
}

func makeGoal(expr string) (*internal.Goal, error) {
	r, err := internal.NewGoal(expr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't make goal: %s", err))
	}
	return r, nil
}

func makeGoals(exprs []string) ([]*internal.Goal, error) {
	var err error
	r := make([]*internal.Goal, len(exprs))
	for i, s := range exprs {
		r[i], err = makeGoal(s)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func makeAggregator(name, aggType, arg string) (internal.Aggregator, error) {
	var r internal.Aggregator
	var err error
	switch aggType {
	case "accuracy":
		r, err = internal.NewAccuracyAggregator(name, arg)
		return r, err
	case "calc":
		r, err = internal.NewCalcAggregator(name, arg)
		return r, err
	case "count":
		r, err = internal.NewCountAggregator(name, arg)
		return r, err
	case "percent":
		r, err = internal.NewPercentAggregator(name, arg)
		return r, err
	case "sum":
		r, err = internal.NewSumAggregator(name, arg)
		return r, err
	default:
		err = errors.New("Unrecognized aggregator")
	}
	if err != nil {
		// TODO: Make custome error type
		err = errors.New(fmt.Sprintf("Can't make aggregator: %s", err))
	}
	return r, err
}

func makeAggregators(
	eAggregators []*AggregatorDesc,
) ([]internal.Aggregator, error) {
	var err error
	r := make([]internal.Aggregator, len(eAggregators))
	for i, ea := range eAggregators {
		r[i], err = makeAggregator(ea.Name, ea.Function, ea.Arg)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func makeSortOrder(eSortOrder []*SortDesc) ([]SortField, error) {
	r := make([]SortField, len(eSortOrder))
	for i, eSortField := range eSortOrder {
		field := eSortField.AggregatorName
		direction := eSortField.Direction
		// TODO: Make case insensitive
		if direction == "ascending" {
			r[i] = SortField{field, ASCENDING}
		} else if direction == "descending" {
			r[i] = SortField{field, DESCENDING}
		} else {
			err := errors.New(fmt.Sprintf("Invalid sort direction: %s, for field: %s",
				direction, field))
			return r, err
		}
	}
	return r, nil
}
