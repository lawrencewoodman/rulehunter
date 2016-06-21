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

package experiment

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
)

type sqlHandler struct {
	driverName     string
	dataSourceName string
	tableName      string
	db             *sql.DB
	openConn       int
	sync.Mutex
}

func newSQLHandler(driverName, dataSourceName, tableName string) *sqlHandler {
	return &sqlHandler{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		tableName:      tableName,
		db:             nil,
		openConn:       0,
	}
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
		return err
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
	if err := s.checkTableExists(s.tableName); err != nil {
		s.Close()
		return nil, err
	}
	rows, err := s.db.Query(fmt.Sprintf("SELECT * FROM \"%s\"", s.tableName))
	if err != nil {
		s.Close()
	}
	return rows, err
}

// checkTableExists returns error if table doesn't exist in database
func (s *sqlHandler) checkTableExists(tableName string) error {
	var rowTableName string
	var rows *sql.Rows
	var err error
	tableNames := make([]string, 0)

	rows, err = s.db.Query("select name from sqlite_master where type='table'")
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
