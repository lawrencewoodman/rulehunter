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

// GEFVI represents a rule determening if field >= intValue
type GEFVI struct {
	field string
	value int64
}

func NewGEFVI(field string, value int64) TweakableRule {
	return &GEFVI{field: field, value: value}
}

func (r *GEFVI) String() string {
	return fmt.Sprintf("%s >= %d", r.field, r.value)
}

func (r *GEFVI) GetTweakableParts() (string, string, string) {
	return r.field, ">=", fmt.Sprintf("%d", r.value)
}

func (r *GEFVI) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		return lhInt >= r.value, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *GEFVI) CloneWithValue(newValue interface{}) TweakableRule {
	f, ok := newValue.(int64)
	if ok {
		return NewGEFVI(r.field, f)
	}
	panic(fmt.Sprintf(
		"can't clone with newValue: %v of type %T, need type int64",
		newValue, newValue,
	))
}
