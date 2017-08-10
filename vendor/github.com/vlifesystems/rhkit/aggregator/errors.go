// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregator

import (
	"errors"
	"fmt"
)

// DescError indicates that there is a problem with the description
// of the aggregator
type DescError struct {
	Name string
	Kind string
	Err  error
}

var (
	ErrInvalidNumArgs   = errors.New("invalid number of arguments")
	ErrUnregisteredKind = errors.New("unregistered kind")
	ErrInvalidName      = errors.New("invalid name")
	ErrNameClash        = errors.New("name clashes with field name")
	ErrNameReserved     = errors.New("name reserved")
)

func (e DescError) Error() string {
	return fmt.Sprintf(
		"problem with aggregator description - name: %s, kind: %s (%s)",
		e.Name, e.Kind, e.Err,
	)
}
