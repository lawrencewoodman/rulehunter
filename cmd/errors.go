// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package cmd

import (
	"fmt"
)

type errConfigLoad struct {
	filename string
	err      error
}

func (e errConfigLoad) Error() string {
	return fmt.Sprintf("couldn't load configuration file: %s: %s", e.filename, e.err)
}
