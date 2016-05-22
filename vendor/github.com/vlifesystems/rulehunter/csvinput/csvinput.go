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
package csvinput

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/input"
	"github.com/vlifesystems/rulehunter/internal"
	"io"
	"os"
)

type CsvInput struct {
	file          *os.File
	reader        *csv.Reader
	fieldNames    []string
	filename      string
	separator     rune
	skipFirstLine bool
	currentRecord []string
	err           error
}

func New(
	fieldNames []string,
	filename string,
	separator rune,
	skipFirstLine bool,
) (input.Input, error) {
	if err := checkFieldsValid(fieldNames); err != nil {
		return nil, err
	}
	f, r, err := makeCsvReader(filename, separator, skipFirstLine)
	if err != nil {
		return nil, err
	}
	r.Comma = separator
	return &CsvInput{
		file:          f,
		reader:        r,
		fieldNames:    fieldNames,
		filename:      filename,
		separator:     separator,
		skipFirstLine: skipFirstLine,
		currentRecord: []string{},
	}, nil
}

func (c *CsvInput) Clone() (input.Input, error) {
	newC, err :=
		New(c.fieldNames, c.filename, c.separator, c.skipFirstLine)
	return newC, err
}

func (c *CsvInput) Next() bool {
	if c.err != nil {
		return false
	}
	if c.reader == nil {
		c.err = errors.New("input has been closed")
		return false
	}
	record, err := c.reader.Read()
	if err == io.EOF {
		c.err = err
		return false
	} else if err != nil {
		c.Close()
		c.err = err
		return false
	}
	c.currentRecord = record
	return true
}

func (c *CsvInput) Err() error {
	if c.err == io.EOF {
		return nil
	}
	return c.err
}

func (c *CsvInput) Read() (map[string]*dlit.Literal, error) {
	recordLits := make(map[string]*dlit.Literal)
	if c.Err() != nil {
		return recordLits, c.Err()
	}
	if len(c.currentRecord) != len(c.fieldNames) {
		// TODO: Create specific error type for this
		c.err = errors.New("wrong number of field names for input")
		c.Close()
		return recordLits, c.err
	}
	for i, field := range c.currentRecord {
		l, err := dlit.New(field)
		if err != nil {
			c.Close()
			c.err = err
			return recordLits, err
		}
		recordLits[c.fieldNames[i]] = l
	}
	return recordLits, nil
}

func (c *CsvInput) Rewind() error {
	var err error
	if c.Err() != nil {
		return c.err
	}
	if c.reader == nil {
		c.err = errors.New("input has been closed")
		return c.err
	}
	if err := c.file.Close(); err != nil {
		c.err = err
		return err
	}
	c.file, c.reader, err =
		makeCsvReader(c.filename, c.separator, c.skipFirstLine)
	if err != nil {
		_ = c.Close
	}
	c.err = err
	return err
}

func (c *CsvInput) Close() error {
	err := c.file.Close()
	c.file = nil
	c.reader = nil
	return err
}

func (c *CsvInput) GetFieldNames() []string {
	return c.fieldNames
}

func makeCsvReader(filename string, separator rune,
	skipFirstLine bool) (*os.File, *csv.Reader, error) {
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

func checkFieldsValid(fieldNames []string) error {
	if len(fieldNames) < 2 {
		return fmt.Errorf("Must specify at least two field names")
	}
	for _, field := range fieldNames {
		if !internal.IsIdentifierValid(field) {
			return fmt.Errorf("Invalid field name: %s", field)
		}
	}
	return nil
}
