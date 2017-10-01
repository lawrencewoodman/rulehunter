// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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
