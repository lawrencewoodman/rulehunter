// Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package internal

import "github.com/lawrencewoodman/ddataset"

// CountNumRecords counts the number of records in the Dataset and returns
// that if successful, otherwise it returns -1.
func CountNumRecords(d ddataset.Dataset) int64 {
	c, err := d.Open()
	if err != nil {
		return -1
	}
	defer c.Close()
	numRecords := int64(0)
	for c.Next() {
		numRecords++
	}
	if c.Err() != nil {
		return -1
	}
	return numRecords
}
