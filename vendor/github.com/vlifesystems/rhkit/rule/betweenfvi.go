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
	min *dlit.Literal,
	max *dlit.Literal,
	maxDP int,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	fdMinInt, _ := min.Int()
	fdMaxInt, _ := max.Int()
	step := (fdMaxInt - fdMinInt) / (10 * int64(stage))
	minLow := r.min - step
	minHigh := r.min + step
	minInterStep := (minHigh - minLow) / 20
	if minInterStep < 1 {
		minInterStep = 1
	}
	maxLow := r.max - step
	maxHigh := r.max + step
	maxInterStep := (maxHigh - maxLow) / 20
	if maxInterStep < 1 {
		maxInterStep = 1
	}
	for minN := minLow; minN <= minHigh; minN += minInterStep {
		for maxN := maxLow; maxN <= maxHigh; maxN += maxInterStep {
			if (minN != r.min || maxN != r.max) &&
				minN != minLow &&
				minN != minHigh &&
				minN >= fdMinInt &&
				minN <= fdMaxInt &&
				maxN != minLow &&
				maxN != minHigh &&
				maxN >= fdMinInt &&
				maxN <= fdMaxInt &&
				minN < maxN {
				r := MustNewBetweenFVI(r.field, minN, maxN)
				rules = append(rules, r)
			}
		}
	}
	return rules
}
