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
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"strconv"
)

// LEFVF represents a rule determening if fieldA <= floatValue
type LEFVF struct {
	field string
	value float64
}

func NewLEFVF(field string, value float64) *LEFVF {
	return &LEFVF{field: field, value: value}
}

func (r *LEFVF) String() string {
	return r.field + " <= " + strconv.FormatFloat(r.value, 'f', -1, 64)
}

func (r *LEFVF) GetValue() float64 {
	return r.value
}

func (r *LEFVF) GetFields() []string {
	return []string{r.field}
}

func (r *LEFVF) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		return lhFloat <= r.value, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *LEFVF) Tweak(
	min *dlit.Literal,
	max *dlit.Literal,
	maxDP int,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	minFloat, _ := min.Float()
	maxFloat, _ := max.Float()
	step := (maxFloat - minFloat) / (10 * float64(stage))
	low := r.value - step
	high := r.value + step
	interStep := (high - low) / 20
	for n := low; n <= high; n += interStep {
		v := truncateFloat(n, maxDP)
		if v != r.value && v != low && v != high && v >= minFloat && v <= maxFloat {
			r := NewLEFVF(r.field, truncateFloat(n, maxDP))
			rules = append(rules, r)
		}
	}
	return rules
}
