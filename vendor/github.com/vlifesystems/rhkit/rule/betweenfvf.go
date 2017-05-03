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
	"errors"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"strconv"
)

// BetweenFVF represents a rule determining if:
// field >= intValue && field <= intValue
type BetweenFVF struct {
	field string
	min   float64
	max   float64
}

func NewBetweenFVF(
	field string,
	min float64,
	max float64,
) (*BetweenFVF, error) {
	if max <= min {
		msg := "can't create Between rule where max: " +
			strconv.FormatFloat(max, 'f', -1, 64) + " <= min: " +
			strconv.FormatFloat(min, 'f', -1, 64)
		return nil, errors.New(msg)
	}
	return &BetweenFVF{field: field, min: min, max: max}, nil
}

func MustNewBetweenFVF(field string, min float64, max float64) *BetweenFVF {
	r, err := NewBetweenFVF(field, min, max)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *BetweenFVF) GetMin() float64 {
	return r.min
}

func (r *BetweenFVF) GetMax() float64 {
	return r.max
}

func (r *BetweenFVF) String() string {
	return r.field + " >= " + strconv.FormatFloat(r.min, 'f', -1, 64) +
		" && " + r.field + " <= " + strconv.FormatFloat(r.max, 'f', -1, 64)
}

func (r *BetweenFVF) IsTrue(record ddataset.Record) (bool, error) {
	value, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	valueFloat, valueIsFloat := value.Float()
	if valueIsFloat {
		return valueFloat >= r.min && valueFloat <= r.max, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *BetweenFVF) GetFields() []string {
	return []string{r.field}
}

func (r *BetweenFVF) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	pointsL := generateTweakPoints(
		dlit.MustNew(r.min),
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	pointsH := generateTweakPoints(
		dlit.MustNew(r.max),
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
			pLFloat, pLIsFloat := pL.Float()
			pHFloat, pHIsFloat := pH.Float()
			isValid, err := isValidExpr.EvalBool(vars)
			if pLIsFloat && pHIsFloat && err == nil && isValid {
				r := MustNewBetweenFVF(r.field, pLFloat, pHFloat)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func (r *BetweenFVF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *BetweenFVF:
		oMin := x.GetMin()
		oMax := x.GetMax()
		oField := x.GetFields()[0]
		return oField == r.field &&
			((oMin >= r.min && oMin <= r.max) || (oMax >= r.min && oMax <= r.max))
	}
	return false
}
