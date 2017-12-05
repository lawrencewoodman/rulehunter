/*
 * A Go package to handles access to a CSV file as Dataset
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dcsv handles access to a CSV file as Dataset
package dcsv

import (
	"encoding/csv"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
	"io"
	"os"
)

// DCSV represents a CSV file Dataset
type DCSV struct {
	filename   string
	fieldNames []string
	hasHeader  bool
	separator  rune
	numFields  int
	isReleased bool
}

// DCSVConn represents a connection to a DCSV Dataset
type DCSVConn struct {
	dataset       *DCSV
	file          *os.File
	reader        *csv.Reader
	currentRecord ddataset.Record
	err           error
}

// New creates a new DCSV Dataset
func New(
	filename string,
	hasHeader bool,
	separator rune,
	fieldNames []string,
) ddataset.Dataset {
	return &DCSV{
		filename:   filename,
		fieldNames: fieldNames,
		hasHeader:  hasHeader,
		separator:  separator,
		numFields:  len(fieldNames),
		isReleased: false,
	}
}

// Open creates a connection to the Dataset
func (d *DCSV) Open() (ddataset.Conn, error) {
	if d.isReleased {
		return nil, ddataset.ErrReleased
	}
	f, r, err := makeCsvReader(d.filename, d.separator, d.hasHeader)
	if err != nil {
		return nil, err
	}
	r.Comma = d.separator

	return &DCSVConn{
		dataset:       d,
		file:          f,
		reader:        r,
		currentRecord: make(ddataset.Record, d.numFields),
		err:           nil,
	}, nil
}

// Fields returns the field names used by the Dataset
func (d *DCSV) Fields() []string {
	return d.fieldNames
}

// Release releases any resources associated with the Dataset d,
// rendering it unusable in the future.
func (d *DCSV) Release() error {
	if !d.isReleased {
		d.isReleased = true
		return nil
	}
	return ddataset.ErrReleased
}

// Next returns whether there is a Record to be Read
func (c *DCSVConn) Next() bool {
	if c.err != nil {
		return false
	}
	if c.reader == nil {
		c.err = ddataset.ErrConnClosed
		return false
	}
	row, err := c.reader.Read()
	if err == io.EOF {
		return false
	} else if err != nil {
		c.Close()
		c.err = err
		return false
	}
	if err := c.makeRowCurrentRecord(row); err != nil {
		c.Close()
		c.err = err
		return false
	}
	return true
}

// Err returns any errors from the connection
func (c *DCSVConn) Err() error {
	return c.err
}

// Read returns the current Record
func (c *DCSVConn) Read() ddataset.Record {
	return c.currentRecord
}

// Close closes the connection
func (c *DCSVConn) Close() error {
	err := c.file.Close()
	c.file = nil
	c.reader = nil
	return err
}

func (c *DCSVConn) numFields() int {
	return c.dataset.numFields
}

func (c *DCSVConn) makeRowCurrentRecord(row []string) error {
	fieldNames := c.dataset.Fields()
	if len(row) != c.dataset.numFields {
		c.err = ddataset.ErrWrongNumFields
		c.Close()
		return c.err
	}
	for i, field := range row {
		c.currentRecord[fieldNames[i]] = dlit.NewString(field)
	}
	return nil
}

func makeCsvReader(
	filename string,
	separator rune,
	hasHeader bool,
) (*os.File, *csv.Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	r := csv.NewReader(f)
	r.Comma = separator
	if hasHeader {
		_, err := r.Read()
		if err != nil {
			return nil, nil, err
		}
	}
	return f, r, err
}
