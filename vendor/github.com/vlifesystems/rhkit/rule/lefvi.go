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
	"github.com/lawrencewoodman/dlit"
)

// LEFVI represents a rule determining if field <= intValue
type LEFVI struct {
	field string
	value int64
}

func NewLEFVI(field string, value int64) *LEFVI {
	return &LEFVI{field: field, value: value}
}

func (r *LEFVI) String() string {
	return fmt.Sprintf("%s <= %d", r.field, r.value)
}

func (r *LEFVI) GetValue() int64 {
	return r.value
}

func (r *LEFVI) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		return lhInt <= r.value, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *LEFVI) Tweak(
	min *dlit.Literal,
	max *dlit.Literal,
	maxDP int,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	minInt, _ := min.Int()
	maxInt, _ := max.Int()
	step := (maxInt - minInt) / (10 * int64(stage))
	low := r.value - step
	high := r.value + step
	interStep := (high - low) / 20
	if interStep < 1 {
		interStep = 1
	}
	for n := low; n <= high; n += interStep {
		if n != r.value && n != low && n != high && n >= minInt && n <= maxInt {
			r := NewLEFVI(r.field, n)
			rules = append(rules, r)
		}
	}
	return rules
}

func (r *LEFVI) GetFields() []string {
	return []string{r.field}
}

func (r *LEFVI) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *LEFVI:
		oField := x.GetFields()[0]
		return r.field == oField
	}
	return false
}
