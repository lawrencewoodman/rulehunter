/*
	Copyright (C) 2017 vLife Systems Ltd <http://vlifesystems.com>
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
