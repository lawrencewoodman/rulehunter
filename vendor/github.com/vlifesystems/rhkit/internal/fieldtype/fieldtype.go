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

package fieldtype

import "fmt"

type FieldType int

const (
	Unknown FieldType = iota
	Ignore
	Int
	Float
	String
)

func New(s string) FieldType {
	switch s {
	case "Unknown":
		return Unknown
	case "Ignore":
		return Ignore
	case "Int":
		return Int
	case "Float":
		return Float
	case "String":
		return String
	}
	panic(fmt.Sprintf("Unsupported type: %s", s))
}

func (ft FieldType) String() string {
	switch ft {
	case Unknown:
		return "Unknown"
	case Ignore:
		return "Ignore"
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	}
	panic(fmt.Sprintf("Unsupported type: %d", ft))
}
