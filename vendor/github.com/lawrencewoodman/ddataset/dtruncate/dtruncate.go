/*
 * A Go package to handle access to a truncated subset of a Dataset
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dtruncate truncates a Dataset
package dtruncate

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/internal"
)

// DTruncate represents a truncated Dataset
type DTruncate struct {
	dataset    ddataset.Dataset
	numRecords int64
	isReleased bool
}

// DTruncateConn represents a connection to a DTruncate Dataset
type DTruncateConn struct {
	dataset   *DTruncate
	conn      ddataset.Conn
	recordNum int64
	err       error
}

// New creates a new DTruncate Dataset
func New(dataset ddataset.Dataset, numRecords int64) ddataset.Dataset {
	return &DTruncate{
		dataset:    dataset,
		numRecords: numRecords,
		isReleased: false,
	}
}

// Open creates a connection to the Dataset
func (d *DTruncate) Open() (ddataset.Conn, error) {
	if d.isReleased {
		return nil, ddataset.ErrReleased
	}
	conn, err := d.dataset.Open()
	if err != nil {
		return nil, err
	}
	return &DTruncateConn{
		dataset:   d,
		conn:      conn,
		recordNum: 0,
		err:       nil,
	}, nil
}

// Fields returns the field names used by the Dataset
func (d *DTruncate) Fields() []string {
	return d.dataset.Fields()
}

// NumRecords returns the number of records in the Dataset.  If there is
// a problem getting the number of records it returns -1.  NOTE: The returned
// value can change if the underlying Dataset changes.
func (d *DTruncate) NumRecords() int64 {
	return internal.CountNumRecords(d)
}

// Release releases any resources associated with the Dataset d,
// rendering it unusable in the future.
func (d *DTruncate) Release() error {
	if !d.isReleased {
		d.isReleased = true
		return nil
	}
	return ddataset.ErrReleased
}

// Next returns whether there is a Record to be Read
func (c *DTruncateConn) Next() bool {
	if c.conn.Err() != nil {
		return false
	}
	if c.recordNum < c.dataset.numRecords {
		c.recordNum++
		return c.conn.Next()
	}
	return false
}

// Err returns any errors from the connection
func (c *DTruncateConn) Err() error {
	return c.conn.Err()
}

// Read returns the current Record
func (c *DTruncateConn) Read() ddataset.Record {
	return c.conn.Read()
}

// Close closes the connection
func (c *DTruncateConn) Close() error {
	return c.conn.Close()
}
