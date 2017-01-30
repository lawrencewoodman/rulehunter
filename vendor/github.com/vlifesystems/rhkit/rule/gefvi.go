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

// GEFVI represents a rule determening if field >= intValue
type GEFVI struct {
	field string
	value int64
}

func NewGEFVI(field string, value int64) *GEFVI {
	return &GEFVI{field: field, value: value}
}

func (r *GEFVI) String() string {
	return fmt.Sprintf("%s >= %d", r.field, r.value)
}

func (r *GEFVI) GetValue() int64 {
	return r.value
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

func (r *GEFVI) Tweak(
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
			r := NewGEFVI(r.field, n)
			rules = append(rules, r)
		}
	}
	return rules
}

func (r *GEFVI) GetFields() []string {
	return []string{r.field}
}
