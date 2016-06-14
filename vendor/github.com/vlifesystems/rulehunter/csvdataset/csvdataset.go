/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

// Package csvdataset handles access to a CSV dataset
package csvdataset

import (
	"encoding/csv"
	"errors"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/dataset"
	"io"
	"os"
)

type CsvDataset struct {
	fieldNames    []string
	numFields     int
	filename      string
	separator     rune
	skipFirstLine bool
}

type CsvDatasetConn struct {
	dataset       *CsvDataset
	file          *os.File
	reader        *csv.Reader
	currentRecord dataset.Record
	err           error
}

func New(
	fieldNames []string,
	filename string,
	separator rune,
	skipFirstLine bool,
) (dataset.Dataset, error) {
	if err := dataset.CheckFieldNamesValid(fieldNames); err != nil {
		return nil, err
	}

	return &CsvDataset{
		fieldNames:    fieldNames,
		numFields:     len(fieldNames),
		filename:      filename,
		separator:     separator,
		skipFirstLine: skipFirstLine,
	}, nil
}

func (c *CsvDataset) Open() (dataset.Conn, error) {
	f, r, err := makeCsvReader(c.filename, c.separator, c.skipFirstLine)
	if err != nil {
		return nil, err
	}
	r.Comma = c.separator

	return &CsvDatasetConn{
		dataset:       c,
		file:          f,
		reader:        r,
		currentRecord: make(dataset.Record, c.numFields),
		err:           nil,
	}, nil
}

func (c *CsvDataset) GetFieldNames() []string {
	return c.fieldNames
}

func (cc *CsvDatasetConn) Next() bool {
	if cc.err != nil {
		return false
	}
	if cc.reader == nil {
		cc.err = errors.New("connection has been closed")
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

func (cc *CsvDatasetConn) Err() error {
	return cc.err
}

func (cc *CsvDatasetConn) Read() dataset.Record {
	return cc.currentRecord
}

func (cc *CsvDatasetConn) Close() error {
	err := cc.file.Close()
	cc.file = nil
	cc.reader = nil
	return err
}

func (cc *CsvDatasetConn) getNumFields() int {
	return cc.dataset.numFields
}

func (cc *CsvDatasetConn) getFieldNames() []string {
	return cc.dataset.fieldNames
}

func (cc *CsvDatasetConn) makeRowCurrentRecord(row []string) error {
	fieldNames := cc.dataset.GetFieldNames()
	if len(row) != cc.getNumFields() {
		// TODO: Create specific error type for this
		cc.err = errors.New("wrong number of field names for dataset")
		cc.Close()
		return cc.err
	}
	for i, field := range row {
		l, err := dlit.New(field)
		if err != nil {
			cc.Close()
			cc.err = err
			return err
		}
		cc.currentRecord[fieldNames[i]] = l
	}
	return nil
}

func makeCsvReader(
	filename string,
	separator rune,
	skipFirstLine bool,
) (*os.File, *csv.Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	r := csv.NewReader(f)
	r.Comma = separator
	if skipFirstLine {
		_, err := r.Read()
		if err != nil {
			return nil, nil, err
		}
	}
	return f, r, err
}
