/*
 * A Go package to copy a Dataset
 *
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dcopy copies a Dataset so that you can work consistently on
// the same Dataset.  This is important where a database is likely to be
// updated while you are working on it.  The copy of the database is stored
// in an sqlite3 database located in a temporary directory.
package dcopy

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/internal"
)

// DCopy represents a copy of a Dataset
type DCopy struct {
	dataset    ddataset.Dataset
	tmpDir     string
	isReleased bool
	numRecords int64
}

// DCopyConn represents a connection to a DCopy Dataset
type DCopyConn struct {
	conn ddataset.Conn
	err  error
}

// New creates a new DCopy Dataset which will be a copy of the Dataset
// supplied at the time it is run. Please note that this creates a file
// on the disk containing a copy of the supplied Dataset.  The copy is
// created in a sub-directory of tmpDir.  If tmpDir is the empty string,
// then it uses the default system temporary directory.
func New(dataset ddataset.Dataset, tmpDir string) (ddataset.Dataset, error) {
	tmpDir, err := ioutil.TempDir(tmpDir, "dcopy")
	if err != nil {
		return nil, err
	}
	copyFilename := filepath.Join(tmpDir, "copy.csv")
	f, err := os.Create(copyFilename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	w := csv.NewWriter(f)

	conn, err := dataset.Open()
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}
	defer conn.Close()

	strRecord := make([]string, len(dataset.Fields()))
	for conn.Next() {
		record := conn.Read()
		for i, f := range dataset.Fields() {
			strRecord[i] = record[f].String()
		}
		if err := w.Write(strRecord); err != nil {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("error writing record to csv copy: %s", err)
		}
	}

	if err := conn.Err(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	return &DCopy{
		dataset:    dcsv.New(copyFilename, false, ',', dataset.Fields()),
		tmpDir:     tmpDir,
		isReleased: false,
		numRecords: -1,
	}, nil
}

// Open creates a connection to the Dataset
func (d *DCopy) Open() (ddataset.Conn, error) {
	if d.isReleased {
		return nil, ddataset.ErrReleased
	}
	conn, err := d.dataset.Open()
	if err != nil {
		return nil, err
	}
	return &DCopyConn{
		conn: conn,
		err:  nil,
	}, nil
}

// Fields returns the field names used by the Dataset
func (d *DCopy) Fields() []string {
	if d.isReleased {
		return []string{}
	}
	return d.dataset.Fields()
}

// NumRecords returns the number of records in the Dataset.  If there is
// a problem getting the number of records it returns -1.
func (d *DCopy) NumRecords() int64 {
	if d.numRecords != -1 {
		return d.numRecords
	}
	d.numRecords = internal.CountNumRecords(d)
	return d.numRecords
}

// Release releases any resources associated with the Dataset d,
// rendering it unusable in the future.  In this case it deletes
// the temporary copy of the Dataset.
func (d *DCopy) Release() error {
	if !d.isReleased {
		err := os.RemoveAll(d.tmpDir)
		if err == nil {
			d.isReleased = true
		}
		return err
	}
	return ddataset.ErrReleased
}

// Next returns whether there is a Record to be Read
func (c *DCopyConn) Next() bool {
	return c.conn.Next()
}

// Err returns any errors from the connection
func (c *DCopyConn) Err() error {
	return c.conn.Err()
}

// Read returns the current Record
func (c *DCopyConn) Read() ddataset.Record {
	return c.conn.Read()
}

// Close closes the connection and deletes the copy
func (c *DCopyConn) Close() error {
	return c.conn.Close()
}
