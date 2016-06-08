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

package internal

import "fmt"

type FieldType int

const (
	UNKNOWN FieldType = iota
	IGNORE
	INT
	FLOAT
	STRING
)

func (ft FieldType) String() string {
	switch ft {
	case UNKNOWN:
		return "Unknown"
	case IGNORE:
		return "Ignore"
	case INT:
		return "Int"
	case FLOAT:
		return "Float"
	case STRING:
		return "String"
	}
	panic(fmt.Sprintf("Unsupported type: %d", ft))
}
