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
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
)

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
) *sqlHandler {
	return &sqlHandler{
		driverName:     driverName,
		dataSourceName: dataSourceName,
		query:          query,
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
