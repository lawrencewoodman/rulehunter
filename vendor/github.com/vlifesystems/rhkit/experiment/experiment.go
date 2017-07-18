// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package experiment handles initialization and validation of experiment
package experiment

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/vlifesystems/rhkit/aggregators"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/rule"
)

type ExperimentDesc struct {
	Title          string
	Dataset        ddataset.Dataset
	RuleFields     []string
	RuleComplexity rule.Complexity
	Aggregators    []*AggregatorDesc
	Goals          []string
	SortOrder      []*SortDesc
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
	Title          string
	Dataset        ddataset.Dataset
	RuleFields     []string
	RuleComplexity rule.Complexity
	Aggregators    []aggregators.AggregatorSpec
	Goals          []*goal.Goal
	SortOrder      []SortField
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
	if err := checkExperimentDescValid(e); err != nil {
		return nil, err
	}
	goals, err := makeGoals(e.Goals)
	if err != nil {
		return nil, err
	}
	aggregators, err := makeAggregators(e.Aggregators)
	if err != nil {
		return nil, err
	}

	sortOrder, err := makeSortOrder(e.SortOrder)
	if err != nil {
		return nil, err
	}

	return &Experiment{
		Title:          e.Title,
		Dataset:        e.Dataset,
		RuleFields:     e.RuleFields,
		RuleComplexity: e.RuleComplexity,
		Aggregators:    aggregators,
		Goals:          goals,
		SortOrder:      sortOrder,
	}, nil
}

func checkExperimentDescValid(e *ExperimentDesc) error {
	if err := checkSortDescsValid(e); err != nil {
		return err
	}
	if err := checkRuleFieldsValid(e); err != nil {
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
			return &InvalidSortDirectionError{
				sortDesc.AggregatorName,
				sortDesc.Direction,
			}
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
			return InvalidSortFieldError(sortName)
		}
	}
	return nil
}

func checkRuleFieldsValid(e *ExperimentDesc) error {
	if len(e.RuleFields) == 0 {
		return ErrNoRuleFieldsSpecified
	}
	fieldNames := e.Dataset.Fields()
	for _, ruleField := range e.RuleFields {
		if !internal.IsIdentifierValid(ruleField) {
			return InvalidRuleFieldError(ruleField)
		}
		if !isStringInSlice(ruleField, fieldNames) {
			return InvalidRuleFieldError(ruleField)
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
	fieldNames := e.Dataset.Fields()
	for _, aggregator := range e.Aggregators {
		if !internal.IsIdentifierValid(aggregator.Name) {
			return InvalidAggregatorNameError(aggregator.Name)
		}
		if isStringInSlice(aggregator.Name, fieldNames) {
			return AggregatorNameClashError(aggregator.Name)
		}
		if aggregator.Name == "percentMatches" ||
			aggregator.Name == "numMatches" ||
			aggregator.Name == "goalsScore" {
			return AggregatorNameReservedError(aggregator.Name)
		}
	}
	return nil
}

func makeGoals(exprs []string) ([]*goal.Goal, error) {
	var err error
	r := make([]*goal.Goal, len(exprs))
	for i, expr := range exprs {
		r[i], err = goal.New(expr)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func makeAggregators(
	eAggregators []*AggregatorDesc,
) ([]aggregators.AggregatorSpec, error) {
	var err error
	r := make([]aggregators.AggregatorSpec, len(eAggregators))
	for i, ea := range eAggregators {
		r[i], err = aggregators.New(ea.Name, ea.Function, ea.Arg)
		if err != nil {
			return r, err
		}
	}
	return addDefaultAggregators(r), nil
}

func addDefaultAggregators(
	aggregatorSpecs []aggregators.AggregatorSpec,
) []aggregators.AggregatorSpec {
	newAggregatorSpecs := make([]aggregators.AggregatorSpec, 2)
	newAggregatorSpecs[0] = aggregators.MustNew("numMatches", "count", "true()")
	newAggregatorSpecs[1] = aggregators.MustNew(
		"percentMatches",
		"calc",
		"roundto(100.0 * numMatches / numRecords, 2)",
	)
	goalsScoreAggregatorSpec := aggregators.MustNew("goalsScore", "goalsscore")
	newAggregatorSpecs = append(newAggregatorSpecs, aggregatorSpecs...)
	newAggregatorSpecs = append(newAggregatorSpecs, goalsScoreAggregatorSpec)
	return newAggregatorSpecs
}

func makeSortOrder(eSortOrder []*SortDesc) ([]SortField, error) {
	r := make([]SortField, len(eSortOrder))
	for i, eSortField := range eSortOrder {
		field := eSortField.AggregatorName
		direction := eSortField.Direction
		// TODO: Make case insensitive
		if direction == "ascending" {
			r[i] = SortField{field, ASCENDING}
		} else {
			r[i] = SortField{field, DESCENDING}
		}
	}
	return r, nil
}
