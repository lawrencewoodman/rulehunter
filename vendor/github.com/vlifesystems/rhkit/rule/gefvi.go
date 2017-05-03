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
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
)

// GEFVI represents a rule determining if field >= intValue
type GEFVI struct {
	field string
	value int64
}

func NewGEFVI(field string, value int64) *GEFVI {
	return &GEFVI{field: field, value: value}
}

func (r *GEFVI) String() string {
	return fmt.Sprintf("%s >= %d", r.field, r.value)
}

func (r *GEFVI) GetValue() int64 {
	return r.value
}

func (r *GEFVI) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		return lhInt >= r.value, nil
	}

	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *GEFVI) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	rules := make([]Rule, 0)
	points := generateTweakPoints(
		dlit.MustNew(r.value),
		inputDescription.Fields[r.field].Min,
		inputDescription.Fields[r.field].Max,
		inputDescription.Fields[r.field].MaxDP,
		stage,
	)
	for _, p := range points {
		pInt, pIsInt := p.Int()
		if !pIsInt {
			continue
		}
		r := NewGEFVI(r.field, pInt)
		rules = append(rules, r)
	}
	return rules
}

func (r *GEFVI) GetFields() []string {
	return []string{r.field}
}

func (r *GEFVI) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *GEFVI:
		oField := x.GetFields()[0]
		return r.field == oField
	}
	return false
}
