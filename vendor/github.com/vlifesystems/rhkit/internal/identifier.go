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

package internal

import "regexp"

var validIdentifierRegexp = regexp.MustCompile("^[a-zA-Z]([0-9a-zA-Z_])*$")

// Returns whether the string can be used as an identifier for a field name or
// for an aggregator
func IsIdentifierValid(identifier string) bool {
	return validIdentifierRegexp.MatchString(identifier)
}
