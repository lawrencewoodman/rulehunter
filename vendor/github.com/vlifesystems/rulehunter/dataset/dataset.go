/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

// Package dataset describes the Dataset interface
package dataset

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/internal"
)

type Dataset interface {
	Open() (Conn, error)
	GetFieldNames() []string
}

type Conn interface {
	Next() bool
	Err() error
	Read() Record
	Close() error
}

type Record map[string]*dlit.Literal

// Check that the field names are valid identifiers
func CheckFieldNamesValid(fieldNames []string) error {
	if len(fieldNames) < 2 {
		return fmt.Errorf("must specify at least two field names")
	}
	for _, field := range fieldNames {
		if !internal.IsIdentifierValid(field) {
			return fmt.Errorf("invalid field name: %s", field)
		}
	}
	return nil
}
