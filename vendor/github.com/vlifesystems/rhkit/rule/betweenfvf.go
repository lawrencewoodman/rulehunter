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
	"errors"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
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
	min *dlit.Literal,
	max *dlit.Literal,
	maxDP int,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	fdMinFloat, _ := min.Float()
	fdMaxFloat, _ := max.Float()
	step := (fdMaxFloat - fdMinFloat) / (10.0 * float64(stage))
	minLow := r.min - step
	minHigh := r.min + step
	minFloaterStep := (minHigh - minLow) / 20.0
	if minFloaterStep < 1 {
		minFloaterStep = 1
	}
	maxLow := r.max - step
	maxHigh := r.max + step
	maxFloaterStep := (maxHigh - maxLow) / 20.0
	if maxFloaterStep < 1 {
		maxFloaterStep = 1
	}
	for minN := minLow; minN <= minHigh; minN += minFloaterStep {
		for maxN := maxLow; maxN <= maxHigh; maxN += maxFloaterStep {
			if (minN != r.min || maxN != r.max) &&
				minN != minLow &&
				minN != minHigh &&
				minN >= fdMinFloat &&
				minN <= fdMaxFloat &&
				maxN != minLow &&
				maxN != minHigh &&
				maxN >= fdMinFloat &&
				maxN <= fdMaxFloat &&
				minN < maxN {
				r := MustNewBetweenFVF(r.field, minN, maxN)
				rules = append(rules, r)
			}
		}
	}
	return rules
}
