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
	"fmt"
	"github.com/lawrencewoodman/ddataset"
)

// EQFVS represents a rule determening if field == stringValue
type EQFVS struct {
	field string
	value string
}

func NewEQFVS(field string, value string) Rule {
	return &EQFVS{field: field, value: value}
}

func (r *EQFVS) String() string {
	return fmt.Sprintf("%s == \"%s\"", r.field, r.value)
}

func (r *EQFVS) GetInNiParts() (bool, string, string) {
	return false, "", ""
}

func (r *EQFVS) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	if err := lh.Err(); err != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}
	return r.value == lh.String(), nil
}
