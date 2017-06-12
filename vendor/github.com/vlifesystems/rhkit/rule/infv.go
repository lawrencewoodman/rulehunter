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
	"github.com/vlifesystems/rhkit/internal/fieldtype"
	"sort"
	"strings"
)

// InFV represents a rule determining if field is equal to
// any of the supplied values when represented as a string
type InFV struct {
	field  string
	values []*dlit.Literal
}

func init() {
	registerGenerator("InFV", generateInFV)
}

func NewInFV(field string, values []*dlit.Literal) *InFV {
	if len(values) == 0 {
		panic("NewInFV: Must contain at least one value")
	}
	return &InFV{field: field, values: values}
}

func makeInFVString(field string, values []*dlit.Literal) string {
	return "in(" + field + "," + commaJoinValues(values) + ")"
}

func (r *InFV) String() string {
	return makeInFVString(r.field, r.values)
}

func (r *InFV) Fields() []string {
	return []string{r.field}
}

func (r *InFV) Values() []*dlit.Literal {
	return r.values
}

func (r *InFV) IsTrue(record ddataset.Record) (bool, error) {
	needle, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	if needle.Err() != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}
	for _, v := range r.values {
		if needle.String() == v.String() {
			return true, nil
		}
	}
	return false, nil
}

func (r *InFV) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *InFV:
		oValues := x.Values()
		oField := x.Fields()[0]
		if r.field != oField {
			return false
		}
		for _, v := range r.values {
			for _, oV := range oValues {
				if v.String() == oV.String() {
					return true
				}
			}
		}
	}
	return false
}

func generateInFV(
	inputDescription *description.Description,
	ruleFields []string,
	complexity Complexity,
	field string,
) []Rule {
	extra := 0
	if len(ruleFields) == 2 {
		extra += 3
	}
	fd := inputDescription.Fields[field]
	numValues := len(fd.Values)
	if fd.Kind != fieldtype.String &&
		fd.Kind != fieldtype.Number ||
		numValues <= 3 || numValues > (12+extra) {
		return []Rule{}
	}
	rules := make([]Rule, 0)
	for i := 3; ; i++ {
		numOnBits := calcNumOnBits(i)
		if numOnBits >= numValues {
			break
		}
		if numOnBits >= 2 && numOnBits <= (5+extra) && numOnBits < (numValues-1) {
			compareValues := makeCompareValues(fd.Values, i)
			if len(compareValues) >= 2 {
				r := NewInFV(field, compareValues)
				rules = append(rules, r)
			}
		}
	}
	return rules
}

func makeCompareValues(
	values map[string]description.Value,
	i int,
) []*dlit.Literal {
	bStr := fmt.Sprintf("%b", i)
	numValues := len(values)
	lits := valuesToLiterals(values)
	j := numValues - 1
	compareValues := []*dlit.Literal{}
	for _, b := range reverseString(bStr) {
		if b == '1' {
			lit := lits[numValues-1-j]
			if values[lit.String()].Num < 2 {
				return []*dlit.Literal{}
			}
			compareValues = append(compareValues, lit)
		}
		j -= 1
	}
	return compareValues
}

func valuesToLiterals(values map[string]description.Value) []*dlit.Literal {
	lits := make([]*dlit.Literal, len(values))
	keys := make([]string, len(values))
	i := 0
	for k := range values {
		keys[i] = k
		i++
	}
	// The keys are sorted to make it easier to test because maps aren't ordered
	sort.Strings(keys)
	j := 0
	for _, k := range keys {
		lits[j] = values[k].Value
		j++
	}
	return lits
}

func calcNumOnBits(i int) int {
	bStr := fmt.Sprintf("%b", i)
	return strings.Count(bStr, "1")
}

func reverseString(s string) (r string) {
	for _, v := range s {
		r = string(v) + r
	}
	return
}
