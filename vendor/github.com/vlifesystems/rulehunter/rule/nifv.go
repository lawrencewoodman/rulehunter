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

// NiFV represents a rule determening if field is equal to
// any of the supplied values
type NiFV struct {
	field  string
	values []*dlit.Literal
}

func NewNiFV(field string, values []*dlit.Literal) Rule {
	if len(values) == 0 {
		panic("NewNiFV: Must contain at least one value")
	}
	return &NiFV{field: field, values: values}
}

func makeNiFVString(field string, values []*dlit.Literal) string {
	return "ni(" + field + "," + commaJoinValues(values) + ")"
}

func (r *NiFV) String() string {
	return makeNiFVString(r.field, r.values)
}

func (r *NiFV) GetInNiParts() (bool, string, string) {
	return true, "ni", r.field
}

func (r *NiFV) IsTrue(record ddataset.Record) (bool, error) {
	needle, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	if needle.Err() != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}
	for _, v := range r.values {
		if needle.String() == v.String() {
			return false, nil
		}
	}
	return true, nil
}
