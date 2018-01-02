// Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package internal

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type sqlite3Handler struct {
	filename  string
	tableName string
	cacheMB   int
	db        *sql.DB
	query     string
	openConn  int
	sync.Mutex
}

func NewSqlite3Handler(
	filename,
	tableName string,
	cacheMB int,
) *sqlite3Handler {

	if cacheMB < 0 {
		cacheMB = 0
	}

	return &sqlite3Handler{
		filename:  filename,
		tableName: tableName,
		cacheMB:   cacheMB,
		db:        nil,
		openConn:  0,
	}
}

func (d *sqlite3Handler) Open() error {
	d.Lock()
	defer d.Unlock()
	d.openConn++
	if d.openConn == 1 {
		if !fileExists(d.filename) {
			return fmt.Errorf("database doesn't exist: %s", d.filename)
		}
		db, err := sql.Open("sqlite3", d.filename)
		d.db = db
		if err != nil {
			return err
		}

		sqlPragmaStmt := fmt.Sprintf("PRAGMA CACHE_SIZE = -%d000;", d.cacheMB)
		if _, err := db.Exec(sqlPragmaStmt); err != nil {
			return err
		}
		return nil
	}
	d.openConn++
	return nil
}

func (d *sqlite3Handler) Close() error {
	d.Lock()
	defer d.Unlock()
	if d.openConn >= 1 {
		d.openConn--
		if d.openConn == 0 {
			return d.db.Close()
		}
	}
	return nil
}

func (d *sqlite3Handler) Rows() (*sql.Rows, error) {
	if err := d.checkTableExists(d.tableName); err != nil {
		d.Close()
		return nil, err
	}
	rows, err := d.db.Query(fmt.Sprintf("SELECT * FROM \"%s\"", d.tableName))
	if err != nil {
		d.Close()
	}
	return rows, err
}

// checkTableExists returns error if table doesn't exist in database
func (d *sqlite3Handler) checkTableExists(tableName string) error {
	var rowTableName string
	var rows *sql.Rows
	var err error
	tableNames := make([]string, 0)

	rows, err = d.db.Query("select name from sqlite_master where type='table'")
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

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func inStringsSlice(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
