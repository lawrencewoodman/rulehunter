/*
	rulehuntersrv - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

// Package to access a SQL database as a Dataset
package sqldataset

import (
	"database/sql"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vlifesystems/rulehunter/dataset"
	"os"
)

type SqlDataset struct {
	driverName     string
	dataSourceName string
	tableName      string
	fieldNames     []string
}

type SqlDatasetConn struct {
	dataset       *SqlDataset
	db            *sql.DB
	rows          *sql.Rows
	row           []sql.NullString
	rowPtrs       []interface{}
	currentRecord dataset.Record
	err           error
}

func New(
	driverName string,
	dataSourceName string,
	tableName string,
	fieldNames []string,
) (dataset.Dataset, error) {
	if err := dataset.CheckFieldNamesValid(fieldNames); err != nil {
		return nil, err
	}
	if driverName != "sqlite3" {
		return nil, fmt.Errorf("unrecognized driver name: %s", driverName)
	}
	return &SqlDataset{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		tableName:      tableName,
		fieldNames:     fieldNames,
	}, nil
}

func (s *SqlDataset) Open() (dataset.Conn, error) {
	if s.driverName == "sqlite3" {
		if !fileExists(s.dataSourceName) {
			return nil, fmt.Errorf("database doesn't exist: %s", s.dataSourceName)
		}
	}
	db, err := sql.Open(s.driverName, s.dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := checkTableExists(s.driverName, db, s.tableName); err != nil {
		db.Close()
		return nil, err
	}
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM \"%s\"", s.tableName))
	if err != nil {
		db.Close()
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		db.Close()
		return nil, err
	}
	numColumns := len(columns)
	if err := checkTableValid(s.fieldNames, numColumns); err != nil {
		db.Close()
		return nil, err
	}
	row := make([]sql.NullString, numColumns)
	rowPtrs := make([]interface{}, numColumns)
	for i, _ := range s.fieldNames {
		rowPtrs[i] = &row[i]
	}

	return &SqlDatasetConn{
		dataset:       s,
		db:            db,
		rows:          rows,
		row:           row,
		rowPtrs:       rowPtrs,
		currentRecord: make(dataset.Record, numColumns),
		err:           nil,
	}, nil
}

func (s *SqlDataset) GetFieldNames() []string {
	return s.fieldNames
}

func (sc *SqlDatasetConn) Next() bool {
	if sc.err != nil {
		return false
	}
	if sc.rows.Next() {
		if err := sc.rows.Scan(sc.rowPtrs...); err != nil {
			sc.Close()
			sc.err = err
			return false
		}
		if err := sc.makeRowCurrentRecord(); err != nil {
			sc.Close()
			sc.err = err
			return false
		}
		return true
	}
	if err := sc.rows.Err(); err != nil {
		sc.Close()
		sc.err = err
		return false
	}
	return false
}

func (sc *SqlDatasetConn) Err() error {
	return sc.err
}

func (sc *SqlDatasetConn) Read() dataset.Record {
	return sc.currentRecord
}

func (sc *SqlDatasetConn) Close() error {
	err := sc.db.Close()
	sc.db = nil
	return err
}

func (sc *SqlDatasetConn) makeRowCurrentRecord() error {
	var l *dlit.Literal
	var err error
	for i, v := range sc.row {
		if v.Valid {
			l = dlit.NewString(v.String)
		} else {
			l, err = dlit.New(nil)
			if err != nil {
				sc.Close()
				return err
			}
		}
		sc.currentRecord[sc.dataset.fieldNames[i]] = l
	}
	return nil
}

func checkTableValid(fieldNames []string, numColumns int) error {
	if len(fieldNames) < numColumns {
		return fmt.Errorf(
			"number of field names doesn't match number of columns in table",
		)
	}
	return nil
}

func checkTableExists(driverName string, db *sql.DB, tableName string) error {
	var rowTableName string
	var rows *sql.Rows
	var err error
	tableNames := make([]string, 0)

	if driverName == "sqlite3" {
		rows, err = db.Query("select name from sqlite_master where type='table'")
	} else {
		rows, err = db.Query("show tables")
	}
	if err != nil {
		return err
	}

	for rows.Next() {
		if err := rows.Scan(&rowTableName); err != nil {
			return err
		}
		tableNames = append(tableNames, rowTableName)
	}

	if !inStringsSlice(tableName, tableNames) {
		return fmt.Errorf("table name doesn't exist: %s", tableName)
	}
	return nil
}

func inStringsSlice(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}
