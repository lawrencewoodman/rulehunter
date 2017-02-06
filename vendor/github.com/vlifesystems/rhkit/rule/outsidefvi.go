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
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
)

// OutsideFVI represents a rule determining if:
// field <= intValue || field >= intValue
type OutsideFVI struct {
	field string
	low   int64
	high  int64
}

func NewOutsideFVI(
	field string,
	low int64,
	high int64,
) (*OutsideFVI, error) {
	if high <= low {
		return nil,
			fmt.Errorf("can't create Outside rule where high: %d <= low: %d",
				high, low)
	}
	return &OutsideFVI{field: field, low: low, high: high}, nil
}

func MustNewOutsideFVI(field string, low int64, high int64) *OutsideFVI {
	r, err := NewOutsideFVI(field, low, high)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *OutsideFVI) GetHigh() int64 {
	return r.high
}

func (r *OutsideFVI) GetLow() int64 {
	return r.low
}

func (r *OutsideFVI) String() string {
	return fmt.Sprintf("%s <= %d || %s >= %d", r.field, r.low, r.field, r.high)
}

func (r *OutsideFVI) IsTrue(record ddataset.Record) (bool, error) {
	value, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	valueInt, valueIsInt := value.Int()
	if valueIsInt {
		return valueInt <= r.low || valueInt >= r.high, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *OutsideFVI) GetFields() []string {
	return []string{r.field}
}

func (r *OutsideFVI) Tweak(
	min *dlit.Literal,
	max *dlit.Literal,
	maxDP int,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	fdMinInt, _ := min.Int()
	fdMaxInt, _ := max.Int()
	step := (fdMaxInt - fdMinInt) / (10 * int64(stage))
	lowLow := r.low - step
	lowHigh := r.low + step
	minInterStep := (lowHigh - lowLow) / 20
	if minInterStep < 1 {
		minInterStep = 1
	}
	highLow := r.high - step
	highHigh := r.high + step
	maxInterStep := (highHigh - highLow) / 20
	if maxInterStep < 1 {
		maxInterStep = 1
	}
	for minN := lowLow; minN <= lowHigh; minN += minInterStep {
		for maxN := highLow; maxN <= highHigh; maxN += maxInterStep {
			if (minN != r.low || maxN != r.high) &&
				minN != lowLow &&
				minN != lowHigh &&
				minN >= fdMinInt &&
				minN <= fdMaxInt &&
				maxN != lowLow &&
				maxN != lowHigh &&
				maxN >= fdMinInt &&
				maxN <= fdMaxInt &&
				minN < maxN {
				r := MustNewOutsideFVI(r.field, minN, maxN)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func (r *OutsideFVI) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *OutsideFVI:
		oField := x.GetFields()[0]
		return oField == r.field
	}
	return false
}
