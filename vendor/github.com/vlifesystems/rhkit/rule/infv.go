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

package rule

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
)

// InFV represents a rule determening if field is equal to
// any of the supplied values when represented as a string
type InFV struct {
	field  string
	values []*dlit.Literal
}

func NewInFV(field string, values []*dlit.Literal) Rule {
	if len(values) == 0 {
		panic("NewInFV: Must contain at least one value")
	}
	return &InFV{field: field, values: values}
}

func makeInFVString(field string, values []*dlit.Literal) string {
	return "in(" + field + "," + commaJoinValues(values) + ")"
}

func (r *InFV) String() string {
	return makeInFVString(r.field, r.values)
}

func (r *InFV) GetInNiParts() (bool, string, string) {
	return true, "in", r.field
}

func (r *InFV) IsTrue(record ddataset.Record) (bool, error) {
	needle, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	if needle.Err() != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}
	for _, v := range r.values {
		if needle.String() == v.String() {
			return true, nil
		}
	}
	return false, nil
}