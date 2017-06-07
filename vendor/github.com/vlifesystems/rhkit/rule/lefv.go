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
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/fieldtype"
)

// LEFV represents a rule determining if field <= value
type LEFV struct {
	field string
	value *dlit.Literal
}

func init() {
	registerGenerator("LEFV", generateLEFV)
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
	points := generateTweakPoints(
		r.value,
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	rules := make([]Rule, len(points))
	for i, p := range points {
		rules[i] = NewLEFV(r.field, p)
	}
	return rules
}

func (r *LEFV) Fields() []string {
	return []string{r.field}
}

func (r *LEFV) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *LEFV:
		oField := x.Fields()[0]
		return r.field == oField
	}
	return false
}

func (r *LEFV) DPReduce() []Rule {
	return roundRules(r.value, func(p *dlit.Literal) Rule {
		return NewLEFV(r.field, p)
	})
}

func generateLEFV(
	inputDescription *description.Description,
	ruleFields []string,
	complexity int,
	field string,
) []Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.Number {
		return []Rule{}
	}
	points := internal.GeneratePoints(fd.Min, fd.Max, fd.MaxDP)
	rules := make([]Rule, len(points))
	for i, p := range points {
		rules[i] = NewLEFV(field, p)
	}
	return rules
}
