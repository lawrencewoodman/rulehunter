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

// BetweenFVI represents a rule determining if:
// field >= intValue && field <= intValue
type BetweenFVI struct {
	field string
	min   int64
	max   int64
}

func NewBetweenFVI(
	field string,
	min int64,
	max int64,
) (*BetweenFVI, error) {
	if max <= min {
		return nil,
			fmt.Errorf("can't create Between rule where max: %d <= min: %d", max, min)
	}
	return &BetweenFVI{field: field, min: min, max: max}, nil
}

func MustNewBetweenFVI(field string, min int64, max int64) *BetweenFVI {
	r, err := NewBetweenFVI(field, min, max)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *BetweenFVI) GetMin() int64 {
	return r.min
}

func (r *BetweenFVI) GetMax() int64 {
	return r.max
}

func (r *BetweenFVI) String() string {
	return fmt.Sprintf("%s >= %d && %s <= %d", r.field, r.min, r.field, r.max)
}

func (r *BetweenFVI) IsTrue(record ddataset.Record) (bool, error) {
	value, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	valueInt, valueIsInt := value.Int()
	if valueIsInt {
		return valueInt >= r.min && valueInt <= r.max, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *BetweenFVI) GetFields() []string {
	return []string{r.field}
}

func (r *BetweenFVI) Tweak(
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
			pLInt, pLIsInt := pL.Int()
			pHInt, pHIsInt := pH.Int()
			isValid, err := isValidExpr.EvalBool(vars)
			if pLIsInt && pHIsInt && err == nil && isValid {
				r := MustNewBetweenFVI(r.field, pLInt, pHInt)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func (r *BetweenFVI) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *BetweenFVI:
		oMin := x.GetMin()
		oMax := x.GetMax()
		oField := x.GetFields()[0]
		return oField == r.field &&
			((oMin >= r.min && oMin <= r.max) || (oMax >= r.min && oMax <= r.max))
	}
	return false
}
