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
	"github.com/lawrencewoodman/ddataset"
)

// NEFF represents a rule determining if fieldA != fieldB
type NEFF struct {
	fieldA string
	fieldB string
}

func NewNEFF(fieldA, fieldB string) Rule {
	return &NEFF{fieldA: fieldA, fieldB: fieldB}
}

func (r *NEFF) String() string {
	return r.fieldA + " != " + r.fieldB
}

func (r *NEFF) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.fieldA]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}
	rh, ok := record[r.fieldB]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return lhInt != rhInt, nil
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return lhFloat != rhFloat, nil
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	lhErr := lh.Err()
	rhErr := rh.Err()
	if lhErr != nil || rhErr != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}

	return lh.String() != rh.String(), nil
}

func (r *NEFF) GetFields() []string {
	return []string{r.fieldA, r.fieldB}
}
