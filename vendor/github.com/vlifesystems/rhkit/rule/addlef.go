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
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

// AddLEF represents a rule determining if fieldA + fieldB <= floatValue
type AddLEF struct {
	fieldA string
	fieldB string
	value  *dlit.Literal
}

func NewAddLEF(fieldA string, fieldB string, value *dlit.Literal) *AddLEF {
	return &AddLEF{fieldA: fieldA, fieldB: fieldB, value: value}
}

func (r *AddLEF) String() string {
	return r.fieldA + " + " + r.fieldB + " <= " + r.value.String()
}

func (r *AddLEF) GetFields() []string {
	return []string{r.fieldA, r.fieldB}
}

// IsTrue returns whether the rule is true for this record.
// This rule relies on making shure that the two fields when
// added will not overflow, so this must have been checked
// before hand by looking at their max/min in the input description.
func (r *AddLEF) IsTrue(record ddataset.Record) (bool, error) {
	vA, ok := record[r.fieldA]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	vB, ok := record[r.fieldB]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	vAInt, vAIsInt := vA.Int()
	if vAIsInt {
		vBInt, vBIsInt := vB.Int()
		if vBIsInt {
			if i, ok := r.value.Int(); ok {
				return vAInt+vBInt <= i, nil
			}
		}
	}

	vAFloat, vAIsFloat := vA.Float()
	vBFloat, vBIsFloat := vB.Float()
	valueFloat, valueIsFloat := r.value.Float()
	if !vAIsFloat || !vBIsFloat || !valueIsFloat {
		return false, IncompatibleTypesRuleError{Rule: r}
	}

	return vAFloat+vBFloat <= valueFloat, nil
}

func (r *AddLEF) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	vars := map[string]*dlit.Literal{
		"aMin": inputDescription.Fields[r.fieldA].Min,
		"bMin": inputDescription.Fields[r.fieldB].Min,
		"aMax": inputDescription.Fields[r.fieldA].Max,
		"bMax": inputDescription.Fields[r.fieldB].Max,
	}
	maxDP := inputDescription.Fields[r.fieldA].MaxDP
	bMaxDP := inputDescription.Fields[r.fieldB].MaxDP
	if bMaxDP > maxDP {
		maxDP = bMaxDP
	}
	rules := make([]Rule, 0)
	min := dexpr.Eval("aMin + bMin", dexprfuncs.CallFuncs, vars)
	max := dexpr.Eval("aMax + bMax", dexprfuncs.CallFuncs, vars)
	points := generateTweakPoints(r.value, min, max, maxDP, stage)
	for _, p := range points {
		r := NewAddLEF(r.fieldA, r.fieldB, p)
		rules = append(rules, r)
	}
	return rules
}

func (r *AddLEF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *AddLEF:
		oFields := x.GetFields()
		if r.fieldA == oFields[0] && r.fieldB == oFields[1] {
			return true
		}
	}
	return false
}
