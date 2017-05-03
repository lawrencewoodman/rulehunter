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

// OutsideFVF represents a rule determining if:
// field <= intValue || field >= intValue
type OutsideFVF struct {
	field string
	low   float64
	high  float64
}

func NewOutsideFVF(
	field string,
	low float64,
	high float64,
) (*OutsideFVF, error) {
	if high <= low {
		msg := "can't create Outside rule where high: " +
			strconv.FormatFloat(high, 'f', -1, 64) + " <= low: " +
			strconv.FormatFloat(low, 'f', -1, 64)
		return nil, errors.New(msg)
	}
	return &OutsideFVF{field: field, low: low, high: high}, nil
}

func MustNewOutsideFVF(field string, low float64, high float64) *OutsideFVF {
	r, err := NewOutsideFVF(field, low, high)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *OutsideFVF) GetHigh() float64 {
	return r.high
}

func (r *OutsideFVF) GetLow() float64 {
	return r.low
}

func (r *OutsideFVF) String() string {
	return r.field + " <= " + strconv.FormatFloat(r.low, 'f', -1, 64) +
		" || " + r.field + " >= " + strconv.FormatFloat(r.high, 'f', -1, 64)
}

func (r *OutsideFVF) IsTrue(record ddataset.Record) (bool, error) {
	value, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	valueFloat, valueIsFloat := value.Float()
	if valueIsFloat {
		return valueFloat <= r.low || valueFloat >= r.high, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *OutsideFVF) GetFields() []string {
	return []string{r.field}
}

func (r *OutsideFVF) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	pointsL := generateTweakPoints(
		dlit.MustNew(r.low),
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	pointsH := generateTweakPoints(
		dlit.MustNew(r.high),
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
				r := MustNewOutsideFVF(r.field, pLFloat, pHFloat)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func (r *OutsideFVF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *OutsideFVF:
		oField := x.GetFields()[0]
		return oField == r.field
	}
	return false
}
