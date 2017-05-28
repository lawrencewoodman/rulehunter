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
)

// EQFV represents a rule determining if fieldA == value
type EQFV struct {
	field string
	value *dlit.Literal
}

func NewEQFV(field string, value *dlit.Literal) Rule {
	return &EQFV{field: field, value: value}
}

func (r *EQFV) String() string {
	_, vIsFloat := r.value.Float()
	_, vIsInt := r.value.Int()
	if vIsInt || vIsFloat {
		return fmt.Sprintf("%s == %s", r.field, r.value)
	}
	return fmt.Sprintf("%s == \"%s\"", r.field, r.value)
}

func (r *EQFV) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	if lhFloat, lhIsFloat := lh.Float(); lhIsFloat {
		if vFloat, vIsFloat := r.value.Float(); vIsFloat {
			return lhFloat == vFloat, nil
		}
	}
	if lhInt, lhIsInt := lh.Int(); lhIsInt {
		if vInt, vIsInt := r.value.Int(); vIsInt {
			return lhInt == vInt, nil
		}
	}
	if lh.Err() == nil && r.value.Err() == nil {
		return lh.String() == r.value.String(), nil
	}
	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *EQFV) Fields() []string {
	return []string{r.field}
}
