/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package version

import (
	"fmt"
)

// rulehunterVersion represents the Rulehunter build version.
type rulehunterVersion struct {
	// Major version
	Major int

	// Minor version
	Minor int

	// Increment this for bug releases
	Patch int

	// Suffix used in the Rulehunter version string.
	// It will be blank for release versions.
	Suffix string
}

// CurrentHugoVersion represents the current build version.
var currentRulehunterVersion = rulehunterVersion{
	Major:  0,
	Minor:  1,
	Patch:  0,
	Suffix: "-DEV",
}

func Version() string {
	return version(
		currentRulehunterVersion.Major,
		currentRulehunterVersion.Minor,
		currentRulehunterVersion.Patch,
		currentRulehunterVersion.Suffix,
	)
}

func version(major, minor, patch int, suffix string) string {
	if patch > 0 {
		return fmt.Sprintf("%d.%02d.%d%s", major, minor, patch, suffix)
	}
	return fmt.Sprintf("%d.%02d%s", major, minor, suffix)
}
