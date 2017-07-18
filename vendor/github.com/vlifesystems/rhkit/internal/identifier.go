// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package internal

import "regexp"

var validIdentifierRegexp = regexp.MustCompile("^[a-zA-Z]([0-9a-zA-Z_])*$")

// Returns whether the string can be used as an identifier for a field name or
// for an aggregator
func IsIdentifierValid(identifier string) bool {
	return validIdentifierRegexp.MatchString(identifier)
}
