/*
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
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
	"github.com/vlifesystems/rhkit/description"
	"strconv"
)

// LEFVF represents a rule determining if fieldA <= floatValue
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
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	points := generateTweakPoints(
		dlit.MustNew(r.value),
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	for _, p := range points {
		pFloat, pIsFloat := p.Float()
		if !pIsFloat {
			continue
		}
		r := NewLEFVF(r.field, pFloat)
		rules = append(rules, r)
	}
	return rules
}

func (r *LEFVF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *LEFVF:
		oField := x.GetFields()[0]
		return r.field == oField
	}
	return false
}
