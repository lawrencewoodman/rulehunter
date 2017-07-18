// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import "errors"

var ErrNoRuleFieldsSpecified = errors.New("no rule fields specified")

type InvalidSortDirectionError struct {
	aggregatorName string
	direction      string
}

func (e *InvalidSortDirectionError) Error() string {
	return "invalid sort direction: " + e.direction +
		", for field: " + e.aggregatorName
}

type InvalidSortFieldError string

func (e InvalidSortFieldError) Error() string {
	return "invalid sort field: " + string(e)
}

type InvalidRuleFieldError string

func (e InvalidRuleFieldError) Error() string {
	return "invalid rule field: " + string(e)
}

type InvalidAggregatorNameError string

func (e InvalidAggregatorNameError) Error() string {
	return "invalid aggregator name: " + string(e)
}

type AggregatorNameClashError string

func (e AggregatorNameClashError) Error() string {
	return "aggregator name clashes with field name: " + string(e)
}

type AggregatorNameReservedError string

func (e AggregatorNameReservedError) Error() string {
	return "aggregator name reserved: " + string(e)
}
