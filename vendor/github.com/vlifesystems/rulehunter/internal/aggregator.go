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
package internal

import (
	"errors"
	"github.com/lawrencewoodman/dlit"
)

type Aggregator interface {
	CloneNew() Aggregator
	GetName() string
	GetArg() string
	GetResult([]Aggregator, []*Goal, int64) *dlit.Literal
	NextRecord(map[string]*dlit.Literal, bool) error
	IsEqual(Aggregator) bool
}

// TODO: Make the thisName optional
// TODO: Test this
func AggregatorsToMap(
	aggregators []Aggregator,
	goals []*Goal,
	numRecords int64,
	thisName string,
) (map[string]*dlit.Literal, error) {
	r := make(map[string]*dlit.Literal, len(aggregators))
	numRecordsL := dlit.MustNew(numRecords)
	r["numRecords"] = numRecordsL
	for _, aggregator := range aggregators {
		if thisName == aggregator.GetName() {
			break
		}
		l := aggregator.GetResult(aggregators, goals, numRecords)
		if l.IsError() {
			return r, errors.New(l.String())
		}
		r[aggregator.GetName()] = l
	}
	return r, nil
}
