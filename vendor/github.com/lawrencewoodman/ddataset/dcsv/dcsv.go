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
	}
}

// Open creates a connection to the Dataset
func (c *DCSV) Open() (ddataset.Conn, error) {
	f, r, err := makeCsvReader(c.filename, c.separator, c.hasHeader)
	if err != nil {
		return nil, err
	}
	r.Comma = c.separator

	return &DCSVConn{
		dataset:       c,
		file:          f,
		reader:        r,
		currentRecord: make(ddataset.Record, c.numFields),
		err:           nil,
	}, nil
}

// Fields returns the field names used by the Dataset
func (c *DCSV) Fields() []string {
	return c.fieldNames
}

// Next returns whether there is a Record to be Read
func (cc *DCSVConn) Next() bool {
	if cc.err != nil {
		return false
	}
	if cc.reader == nil {
		cc.err = ddataset.ErrConnClosed
		return false
	}
	row, err := cc.reader.Read()
	if err == io.EOF {
		return false
	} else if err != nil {
		cc.Close()
		cc.err = err
		return false
	}
	if err := cc.makeRowCurrentRecord(row); err != nil {
		cc.Close()
		cc.err = err
		return false
	}
	return true
}

// Err returns any errors from the connection
func (cc *DCSVConn) Err() error {
	return cc.err
}

// Read returns the current Record
func (cc *DCSVConn) Read() ddataset.Record {
	return cc.currentRecord
}

// Close closes the connection
func (cc *DCSVConn) Close() error {
	err := cc.file.Close()
	cc.file = nil
	cc.reader = nil
	return err
}

func (cc *DCSVConn) numFields() int {
	return cc.dataset.numFields
}

func (cc *DCSVConn) makeRowCurrentRecord(row []string) error {
	fieldNames := cc.dataset.Fields()
	if len(row) != cc.dataset.numFields {
		cc.err = ddataset.ErrWrongNumFields
		cc.Close()
		return cc.err
	}
	for i, field := range row {
		cc.currentRecord[fieldNames[i]] = dlit.NewString(field)
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
