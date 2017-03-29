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

// GEFVF represents a rule determining if field >= floatValue
type GEFVF struct {
	field string
	value float64
}

func NewGEFVF(field string, value float64) *GEFVF {
	return &GEFVF{field: field, value: value}
}

func (r *GEFVF) String() string {
	return r.field + " >= " + strconv.FormatFloat(r.value, 'f', -1, 64)
}

func (r *GEFVF) GetValue() float64 {
	return r.value
}

func (r *GEFVF) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		return lhFloat >= r.value, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *GEFVF) Tweak(
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
			r := NewGEFVF(r.field, v)
			rules = append(rules, r)
		}
	}
	return rules
}
func (r *GEFVF) GetFields() []string {
	return []string{r.field}
}

func (r *GEFVF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *GEFVF:
		oField := x.GetFields()[0]
		return r.field == oField
	}
	return false
}
