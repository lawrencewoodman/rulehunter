// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package description

import "fmt"

type FieldType int

const (
	Unknown FieldType = iota
	Ignore
	Number
	String
)

// NewFieldType creates a new FieldType and will panic if an unsupported type is given
func NewFieldType(s string) FieldType {
	switch s {
	case "Unknown":
		return Unknown
	case "Ignore":
		return Ignore
	case "Number":
		return Number
	case "String":
		return String
	}
	panic(fmt.Sprintf("unsupported type: %s", s))
}

// String returns the string representation of the FieldType
func (ft FieldType) String() string {
	switch ft {
	case Unknown:
		return "Unknown"
	case Ignore:
		return "Ignore"
	case Number:
		return "Number"
	case String:
		return "String"
	}
	panic(fmt.Sprintf("unsupported type: %d", ft))
}
