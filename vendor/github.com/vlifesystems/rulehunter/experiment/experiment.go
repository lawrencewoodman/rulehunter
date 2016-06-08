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

// Package experiment handles initialization and validation of experiment
package experiment

import (
	"errors"
	"fmt"
	"github.com/vlifesystems/rulehunter/aggregators"
	"github.com/vlifesystems/rulehunter/dataset"
	"github.com/vlifesystems/rulehunter/goal"
	"github.com/vlifesystems/rulehunter/internal"
)

type ExperimentDesc struct {
	Title         string
	Dataset       dataset.Dataset
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
	Dataset           dataset.Dataset
	ExcludeFieldNames []string
	Aggregators       []aggregators.Aggregator
	Goals             []*goal.Goal
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

// Create a new Experiment from the description
func New(e *ExperimentDesc) (*Experiment, error) {
	var goals []*goal.Goal
	var aggregators []aggregators.Aggregator
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
		Dataset:           e.Dataset,
		ExcludeFieldNames: e.ExcludeFields,
		Aggregators:       aggregators,
		Goals:             goals,
		SortOrder:         sortOrder,
	}, nil
}

func checkExperimentDescValid(e *ExperimentDesc) error {
	if err := checkSortDescsValid(e); err != nil {
		return err
	}

	if err := checkExcludeFieldsValid(e); err != nil {
		return err
	}

	if err := checkAggregatorsValid(e); err != nil {
		return err
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
			sortName != "goalsScore" {
			return fmt.Errorf("Invalid sort field: %s", sortName)
		}
	}
	return nil
}

func checkExcludeFieldsValid(e *ExperimentDesc) error {
	fieldNames := e.Dataset.GetFieldNames()
	for _, excludeField := range e.ExcludeFields {
		if !internal.IsIdentifierValid(excludeField) {
			return fmt.Errorf("Invalid exclude field: %s", excludeField)
		}
		if !isStringInSlice(excludeField, fieldNames) {
			return fmt.Errorf("Invalid exclude field: %s", excludeField)
		}
	}
	return nil
}

func isStringInSlice(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func checkAggregatorsValid(e *ExperimentDesc) error {
	fieldNames := e.Dataset.GetFieldNames()
	for _, aggregator := range e.Aggregators {
		if !internal.IsIdentifierValid(aggregator.Name) {
			return fmt.Errorf("Invalid aggregator name: %s", aggregator.Name)
		}
		if isStringInSlice(aggregator.Name, fieldNames) {
			return fmt.Errorf("Aggregator name clashes with field name: %s",
				aggregator.Name)
		}
		if aggregator.Name == "percentMatches" ||
			aggregator.Name == "numMatches" ||
			aggregator.Name == "goalsScore" {
			return fmt.Errorf("Aggregator name reserved: %s", aggregator.Name)
		}
	}
	return nil
}

func makeGoal(expr string) (*goal.Goal, error) {
	r, err := goal.New(expr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't make goal: %s", err))
	}
	return r, nil
}

func makeGoals(exprs []string) ([]*goal.Goal, error) {
	var err error
	r := make([]*goal.Goal, len(exprs))
	for i, s := range exprs {
		r[i], err = makeGoal(s)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func makeAggregators(
	eAggregators []*AggregatorDesc,
) ([]aggregators.Aggregator, error) {
	var err error
	r := make([]aggregators.Aggregator, len(eAggregators))
	for i, ea := range eAggregators {
		r[i], err = aggregators.New(ea.Name, ea.Function, ea.Arg)
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
