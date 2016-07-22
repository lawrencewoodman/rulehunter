/*
 * A Go package to handles access to a truncated subset of a Dataset
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dtruncate truncates a Dataset
package dtruncate

import (
	"github.com/lawrencewoodman/ddataset"
)

// DTruncate represents a truncated Dataset
type DTruncate struct {
	dataset    ddataset.Dataset
	numRecords int
}

// DTruncateConn represents a connection to a DTruncate Dataset
type DTruncateConn struct {
	dataset   *DTruncate
	conn      ddataset.Conn
	recordNum int
	err       error
}

// New creates a new DTruncate Dataset
func New(dataset ddataset.Dataset, numRecords int) ddataset.Dataset {
	return &DTruncate{
		dataset:    dataset,
		numRecords: numRecords,
	}
}

// Open creates a connection to the Dataset
func (r *DTruncate) Open() (ddataset.Conn, error) {
	conn, err := r.dataset.Open()
	if err != nil {
		return nil, err
	}
	return &DTruncateConn{
		dataset:   r,
		conn:      conn,
		recordNum: 0,
		err:       nil,
	}, nil
}

// GetFieldNames returns the field names used by the Dataset
func (r *DTruncate) GetFieldNames() []string {
	return r.dataset.GetFieldNames()
}

// Next returns whether there is a Record to be Read
func (rc *DTruncateConn) Next() bool {
	if rc.conn.Err() != nil {
		return false
	}
	if rc.recordNum < rc.dataset.numRecords {
		rc.recordNum++
		return rc.conn.Next()
	}
	return false
}

// Err returns any errors from the connection
func (rc *DTruncateConn) Err() error {
	return rc.conn.Err()
}

// Read returns the current Record
func (rc *DTruncateConn) Read() ddataset.Record {
	return rc.conn.Read()
}

// Close closes the connection
func (rc *DTruncateConn) Close() error {
	return rc.conn.Close()
}
