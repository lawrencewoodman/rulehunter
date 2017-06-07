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
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"github.com/vlifesystems/rhkit/internal/fieldtype"
)

// OutsideFV represents a rule determining if:
// field <= lowValue || field >= highValue
type OutsideFV struct {
	field string
	low   *dlit.Literal
	high  *dlit.Literal
}

func init() {
	registerGenerator("OutsideFV", generateOutsideFV)
}

func NewOutsideFV(
	field string,
	low *dlit.Literal,
	high *dlit.Literal,
) (*OutsideFV, error) {
	vars := map[string]*dlit.Literal{
		"high": high,
		"low":  low,
	}
	ok, err := dexpr.EvalBool("high > low", dexprfuncs.CallFuncs, vars)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(
			"can't create Outside rule where high: %s <= low: %s", high, low,
		)
	}
	return &OutsideFV{field: field, low: low, high: high}, nil
}

func MustNewOutsideFV(
	field string,
	low *dlit.Literal,
	high *dlit.Literal,
) *OutsideFV {
	r, err := NewOutsideFV(field, low, high)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *OutsideFV) High() *dlit.Literal {
	return r.high
}

func (r *OutsideFV) Low() *dlit.Literal {
	return r.low
}

func (r *OutsideFV) String() string {
	return fmt.Sprintf("%s <= %s || %s >= %s", r.field, r.low, r.field, r.high)
}

func (r *OutsideFV) IsTrue(record ddataset.Record) (bool, error) {
	value, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	if vInt, vIsInt := value.Int(); vIsInt {
		if lowInt, lowIsInt := r.low.Int(); lowIsInt {
			if highInt, highIsInt := r.high.Int(); highIsInt {
				return vInt <= lowInt || vInt >= highInt, nil
			}
		}
	}

	if vFloat, vIsFloat := value.Float(); vIsFloat {
		if lowFloat, lowIsFloat := r.low.Float(); lowIsFloat {
			if highFloat, highIsFloat := r.high.Float(); highIsFloat {
				return vFloat <= lowFloat || vFloat >= highFloat, nil
			}
		}
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *OutsideFV) Fields() []string {
	return []string{r.field}
}

func (r *OutsideFV) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	pointsL := generateTweakPoints(
		r.low,
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	pointsH := generateTweakPoints(
		r.high,
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
				r := MustNewOutsideFV(r.field, pL, pH)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func (r *OutsideFV) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *OutsideFV:
		oField := x.Fields()[0]
		return oField == r.field
	}
	return false
}

func generateOutsideFV(
	inputDescription *description.Description,
	ruleFields []string,
	complexity int,
	field string,
) []Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.Number {
		return []Rule{}
	}
	rulesMap := make(map[string]Rule)
	rules := make([]Rule, 0)
	points := internal.GeneratePoints(fd.Min, fd.Max, fd.MaxDP)
	isValidExpr := dexpr.MustNew("pH > pL", dexprfuncs.CallFuncs)

	for _, pL := range points {
		for _, pH := range points {
			vars := map[string]*dlit.Literal{
				"pL": pL,
				"pH": pH,
			}
			if ok, err := isValidExpr.EvalBool(vars); ok && err == nil {
				if r, err := NewOutsideFV(field, pL, pH); err == nil {
					if _, dup := rulesMap[r.String()]; !dup {
						rulesMap[r.String()] = r
						rules = append(rules, r)
					}
				}
			}
		}
	}
	return rules
}
