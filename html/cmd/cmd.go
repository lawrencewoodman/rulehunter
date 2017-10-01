// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package cmd

type Cmd int

const (
	All Cmd = iota
	Flush
	Progress
	Reports
)

func (c Cmd) String() string {
	switch c {
	case All:
		return "all"
	case Flush:
		return "flush"
	case Progress:
		return "progress"
	case Reports:
		return "reports"
	}
	return "error"
}
