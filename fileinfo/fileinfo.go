// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package fileinfo

import "time"

// FileInfo represents a file name and modified time
type FileInfo interface {
	Name() string
	ModTime() time.Time
}

// IsEqual returns if two FileInfo objects are equal
func IsEqual(a, b FileInfo) bool {
	return a.Name() == b.Name() && a.ModTime().Equal(b.ModTime())
}
