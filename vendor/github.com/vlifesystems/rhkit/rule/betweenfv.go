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
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

// BetweenFV represents a rule determining if:
// field >= minValue && field <= maxValue
type BetweenFV struct {
	field string
	min   *dlit.Literal
	max   *dlit.Literal
}

func NewBetweenFV(
	field string,
	min *dlit.Literal,
	max *dlit.Literal,
) (*BetweenFV, error) {
	vars := map[string]*dlit.Literal{
		"max": max,
		"min": min,
	}
	ok, err := dexpr.EvalBool("max > min", dexprfuncs.CallFuncs, vars)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil,
			fmt.Errorf("can't create Between rule where max: %s <= min: %s", max, min)
	}
	return &BetweenFV{field: field, min: min, max: max}, nil
}

func MustNewBetweenFV(
	field string,
	min *dlit.Literal,
	max *dlit.Literal,
) *BetweenFV {
	r, err := NewBetweenFV(field, min, max)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *BetweenFV) Min() *dlit.Literal {
	return r.min
}

func (r *BetweenFV) Max() *dlit.Literal {
	return r.max
}

func (r *BetweenFV) String() string {
	return fmt.Sprintf("%s >= %s && %s <= %s", r.field, r.min, r.field, r.max)
}

func (r *BetweenFV) IsTrue(record ddataset.Record) (bool, error) {
	value, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	if vInt, vIsInt := value.Int(); vIsInt {
		if minInt, minIsInt := r.min.Int(); minIsInt {
			if maxInt, maxIsInt := r.max.Int(); maxIsInt {
				return vInt >= minInt && vInt <= maxInt, nil
			}
		}
	}

	if vFloat, vIsFloat := value.Float(); vIsFloat {
		if minFloat, minIsFloat := r.min.Float(); minIsFloat {
			if maxFloat, maxIsFloat := r.max.Float(); maxIsFloat {
				return vFloat >= minFloat && vFloat <= maxFloat, nil
			}
		}
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *BetweenFV) Fields() []string {
	return []string{r.field}
}

func (r *BetweenFV) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	pointsL := generateTweakPoints(
		r.min,
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	pointsH := generateTweakPoints(
		r.max,
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	isValidExpr := dexpr.MustNew("pH > pL", dexprfuncs.CallFuncs)
	for _, pL := range pointsL {
		for _, pH := range pointsH {
			vars := map[string]*dlit.Literal{
				"pL": pL,
				"pH": pH,
			}
			if ok, err := isValidExpr.EvalBool(vars); ok && err == nil {
				r := MustNewBetweenFV(r.field, pL, pH)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func (r *BetweenFV) Overlaps(o Rule) bool {
	rangeOverlaps := dexpr.MustNew(
		"((oMin >= min && oMin <= max) || (oMax >= min && oMax <= max))",
		dexprfuncs.CallFuncs,
	)
	switch x := o.(type) {
	case *BetweenFV:
		vars := map[string]*dlit.Literal{
			"oMin": x.Min(),
			"oMax": x.Max(),
			"min":  r.min,
			"max":  r.max,
		}
		oField := x.Fields()[0]
		overlap, err := rangeOverlaps.EvalBool(vars)
		if err != nil {
			panic(err)
		}
		return oField == r.field && overlap
	}
	return false
}
