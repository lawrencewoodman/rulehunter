/*
 * A Go package to handles access to a an Sql database as a Dataset
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dsql handles access to an SQL database as a Dataset
package dsql

import (
	"database/sql"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
)

// DSQL represents a SQL database Dataset
type DSQL struct {
	dbHandler  DBHandler
	openConn   int
	fieldNames []string
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
	}
}

// Open creates a connection to the Dataset
func (s *DSQL) Open() (ddataset.Conn, error) {
	if err := s.dbHandler.Open(); err != nil {
		return nil, err
	}
	rows, err := s.dbHandler.Rows()
	if err != nil {
		s.dbHandler.Close()
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		s.dbHandler.Close()
		return nil, err
	}
	numColumns := len(columns)
	if err := checkTableValid(s.fieldNames, numColumns); err != nil {
		s.dbHandler.Close()
		return nil, err
	}
	row := make([]sql.NullString, numColumns)
	rowPtrs := make([]interface{}, numColumns)
	for i := range s.fieldNames {
		rowPtrs[i] = &row[i]
	}

	return &DSQLConn{
		dataset:       s,
		rows:          rows,
		row:           row,
		rowPtrs:       rowPtrs,
		currentRecord: make(ddataset.Record, numColumns),
		err:           nil,
	}, nil
}

// GetFieldNames returns the field names used by the Dataset
func (s *DSQL) GetFieldNames() []string {
	return s.fieldNames
}

// Next returns whether there is a Record to be Read
func (sc *DSQLConn) Next() bool {
	if sc.err != nil {
		return false
	}
	if sc.rows.Next() {
		if err := sc.rows.Scan(sc.rowPtrs...); err != nil {
			sc.Close()
			sc.err = err
			return false
		}
		sc.makeRowCurrentRecord()
		return true
	}
	if err := sc.rows.Err(); err != nil {
		sc.Close()
		sc.err = err
		return false
	}
	return false
}

// Err returns any errors from the connection
func (sc *DSQLConn) Err() error {
	return sc.err
}

// Read returns the current Record
func (sc *DSQLConn) Read() ddataset.Record {
	return sc.currentRecord
}

// Close closes the connection
func (sc *DSQLConn) Close() error {
	return sc.dataset.dbHandler.Close()
}

func (sc *DSQLConn) makeRowCurrentRecord() {
	for i, v := range sc.row {
		sc.currentRecord[sc.dataset.fieldNames[i]] = dlit.NewString(v.String)
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
