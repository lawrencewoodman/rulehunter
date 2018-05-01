// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// +build nosqlite3
// TODO: Remove this file and all references to nosqlite3
// TODO: once sure it is no longer needed

package experiment

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var errDatabaseNotOpen = errors.New("connection to database not open")

type sqlHandler struct {
	driverName     string
	dataSourceName string
	query          string
	db             *sql.DB
	openConn       int
	sync.Mutex
}

func newSQLHandler(
	driverName string,
	dataSourceName string,
	query string,
) (*sqlHandler, error) {
	if driverName == "sqlite3" {
		return nil, fmt.Errorf("invalid driverName: sqlite3, this is temporarily disabled in this release")
	}
	validSQLDriverNames := []string{"sqlite3", "mysql", "mssql", "postgres"}
	for _, name := range validSQLDriverNames {
		if name == driverName {
			return &sqlHandler{
				driverName:     driverName,
				dataSourceName: dataSourceName,
				query:          query,
				db:             nil,
				openConn:       0,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid driverName: %s", driverName)
}

func (s *sqlHandler) Open() error {
	s.Lock()
	defer s.Unlock()
	if s.openConn == 0 {
		if s.driverName == "sqlite3" && !fileExists(s.dataSourceName) {
			return fmt.Errorf("database doesn't exist: %s", s.dataSourceName)
		}
		db, err := sql.Open(s.driverName, s.dataSourceName)
		s.db = db
		if err != nil {
			return err
		}
	}
	s.openConn++
	return nil
}

func (s *sqlHandler) Close() error {
	s.Lock()
	defer s.Unlock()
	if s.openConn >= 1 {
		s.openConn--
		if s.openConn == 0 {
			return s.db.Close()
		}
	}
	return nil
}

func (s *sqlHandler) Rows() (*sql.Rows, error) {
	s.Lock()
	if s.openConn < 1 {
		s.Unlock()
		return nil, errDatabaseNotOpen
	}
	s.Unlock()
	rows, err := s.db.Query(s.query)
	if err != nil {
		s.Close()
	}
	return rows, err
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}
