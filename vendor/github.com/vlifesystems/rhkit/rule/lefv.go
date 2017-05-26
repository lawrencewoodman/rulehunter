/*
	Copyright (C) 2017 vLife Systems Ltd <http://vlifesystems.com>
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
	"github.com/vlifesystems/rhkit/description"
)

// LEFV represents a rule determining if field <= value
type LEFV struct {
	field string
	value *dlit.Literal
}

func NewLEFV(field string, value *dlit.Literal) *LEFV {
	return &LEFV{field: field, value: value}
}

func (r *LEFV) String() string {
	return fmt.Sprintf("%s <= %s", r.field, r.value)
}

func (r *LEFV) Value() *dlit.Literal {
	return r.value
}

func (r *LEFV) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	if lhInt, lhIsInt := lh.Int(); lhIsInt {
		if v, ok := r.value.Int(); ok {
			return lhInt <= v, nil
		}
	}

	if lhFloat, lhIsFloat := lh.Float(); lhIsFloat {
		if v, ok := r.value.Float(); ok {
			return lhFloat <= v, nil
		}
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *LEFV) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	points := generateTweakPoints(
		r.value,
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	for _, p := range points {
		r := NewLEFV(r.field, p)
		rules = append(rules, r)
	}
	return rules
}

func (r *LEFV) GetFields() []string {
	return []string{r.field}
}

func (r *LEFV) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *LEFV:
		oField := x.GetFields()[0]
		return r.field == oField
	}
	return false
}

func (r *LEFV) DPReduce() []Rule {
	return roundRules(r.value, func(p *dlit.Literal) Rule {
		return NewLEFV(r.field, p)
	})
}
