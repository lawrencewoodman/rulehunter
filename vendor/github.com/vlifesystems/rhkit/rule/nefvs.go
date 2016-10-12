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
)

// NEFVS represents a rule determening if field != stringValue
type NEFVS struct {
	field string
	value string
}

func NewNEFVS(field string, value string) Rule {
	return &NEFVS{field: field, value: value}
}

func (r *NEFVS) String() string {
	return fmt.Sprintf("%s != \"%s\"", r.field, r.value)
}

func (r *NEFVS) GetInNiParts() (bool, string, string) {
	return false, "", ""
}

func (r *NEFVS) IsTrue(record ddataset.Record) (bool, error) {
	lh, ok := record[r.field]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	if err := lh.Err(); err != nil {
		return false, IncompatibleTypesRuleError{Rule: r}
	}
	return r.value != lh.String(), nil
}
