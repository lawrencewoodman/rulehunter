/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
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

package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
)

// EQFVI represents a rule determining if field == intValue
type EQFVI struct {
	field string
	value int64
}

func NewEQFVI(field string, value int64) Rule {
	return &EQFVI{field: field, value: value}
}

func (r *EQFVI) String() string {
	return fmt.Sprintf("%s == %d", r.field, r.value)
}

func (r *EQFVI) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		return lhInt == r.value, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *EQFVI) GetFields() []string {
	return []string{r.field}
}
