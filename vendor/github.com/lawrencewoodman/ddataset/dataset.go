/*
 * A Go package to describe a dynamic Dataset interface
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package ddataset describes a dynamic Dataset interface
package ddataset

import (
	"errors"
	"github.com/lawrencewoodman/dlit"
)

// ErrConnClosed indicates a connection has been closed
var ErrConnClosed = errors.New("connection has been closed")

// ErrWrongNumFields indicates that the wrong number of field names has
// been given for the Dataset
var ErrWrongNumFields = errors.New("wrong number of field names for dataset")

// Dataset provides access to a data source
type Dataset interface {
	// Open creates a connection to the Dataset
	Open() (Conn, error)
	// GetFieldNames returns the field names used by the Dataset
	GetFieldNames() []string
}

// Conn represents a connection to a Dataset
type Conn interface {
	// Next returns whether there is a Record to be Read
	Next() bool
	// Err returns any errors from the connection
	Err() error
	// Read returns the current Record
	Read() Record
	// Close closes the connection
	Close() error
}

// Record represents a single record/row from the Dataset
type Record map[string]*dlit.Literal

// Clone creates a copy of the Record.  This is important where you might
// want to store a record for later use or make use of two or more different
// records at the same time.
func (r Record) Clone() Record {
	ret := make(Record, len(r))
	for k, v := range r {
		ret[k] = v
	}
	return ret
}
