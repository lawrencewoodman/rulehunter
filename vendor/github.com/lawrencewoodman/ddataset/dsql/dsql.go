/*
 * A Go package to handles access to a an Sql database as a Dataset
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dsql handles access to an SQL database as a Dataset
package dsql

import (
	"database/sql"
	"fmt"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/internal"
	"github.com/lawrencewoodman/dlit"
)

// DSQL represents a SQL database Dataset
type DSQL struct {
	dbHandler  DBHandler
	openConn   int
	fieldNames []string
	isReleased bool
}

// DSQLConn represents a connection to a DSQL Dataset
type DSQLConn struct {
	dataset       *DSQL
	rows          *sql.Rows
	row           []sql.NullString
	rowPtrs       []interface{}
	currentRecord ddataset.Record
	err           error
}

// DBHandler handles basic access to an Sql database
type DBHandler interface {
	// Open opens the database
	Open() error
	// Rows returns the rows for the database with each row having
	// the same number and same order of fields as those passed
	// to New.  However, the fields don't have to have the same names.
	Rows() (*sql.Rows, error)
	// Close closes the database
	Close() error
}

// New creates a new DSQL Dataset
func New(dbHandler DBHandler, fieldNames []string) ddataset.Dataset {
	return &DSQL{
		dbHandler:  dbHandler,
		openConn:   0,
		fieldNames: fieldNames,
		isReleased: false,
	}
}

// Open creates a connection to the Dataset
func (d *DSQL) Open() (ddataset.Conn, error) {
	if d.isReleased {
		return nil, ddataset.ErrReleased
	}
	if err := d.dbHandler.Open(); err != nil {
		return nil, err
	}
	rows, err := d.dbHandler.Rows()
	if err != nil {
		d.dbHandler.Close()
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		d.dbHandler.Close()
		return nil, err
	}
	numColumns := len(columns)
	if err := checkTableValid(d.fieldNames, numColumns); err != nil {
		d.dbHandler.Close()
		return nil, err
	}
	row := make([]sql.NullString, numColumns)
	rowPtrs := make([]interface{}, numColumns)
	for i := range d.fieldNames {
		rowPtrs[i] = &row[i]
	}

	return &DSQLConn{
		dataset:       d,
		rows:          rows,
		row:           row,
		rowPtrs:       rowPtrs,
		currentRecord: make(ddataset.Record, numColumns),
		err:           nil,
	}, nil
}

// Fields returns the field names used by the Dataset
func (d *DSQL) Fields() []string {
	return d.fieldNames
}

// NumRecords returns the number of records in the Dataset.  If there is
// a problem getting the number of records it returns -1. NOTE: The returned
// value can change if the underlying Dataset changes.
func (d *DSQL) NumRecords() int64 {
	return internal.CountNumRecords(d)
}

// Release releases any resources associated with the Dataset d,
// rendering it unusable in the future.
func (d *DSQL) Release() error {
	if !d.isReleased {
		d.isReleased = true
		return nil
	}
	return ddataset.ErrReleased
}

// Next returns whether there is a Record to be Read
func (c *DSQLConn) Next() bool {
	if c.err != nil {
		return false
	}
	if c.rows.Next() {
		if err := c.rows.Scan(c.rowPtrs...); err != nil {
			c.Close()
			c.err = err
			return false
		}
		c.makeRowCurrentRecord()
		return true
	}
	if err := c.rows.Err(); err != nil {
		c.Close()
		c.err = err
		return false
	}
	return false
}

// Err returns any errors from the connection
func (c *DSQLConn) Err() error {
	return c.err
}

// Read returns the current Record
func (c *DSQLConn) Read() ddataset.Record {
	return c.currentRecord
}

// Close closes the connection
func (c *DSQLConn) Close() error {
	return c.dataset.dbHandler.Close()
}

func (c *DSQLConn) makeRowCurrentRecord() {
	for i, v := range c.row {
		c.currentRecord[c.dataset.fieldNames[i]] = dlit.NewString(v.String)
	}
}

func checkTableValid(fieldNames []string, numColumns int) error {
	if len(fieldNames) != numColumns {
		return fmt.Errorf(
			"number of field names doesn't match number of columns in table",
		)
	}
	return nil
}
