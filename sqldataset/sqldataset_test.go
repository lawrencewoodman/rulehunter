// TODO: Test NULL handling of null values in database
// TODO: Test Next, Err for errors - using a mock database
// TODO: Test Open most fully for errors - using a mock database

package sqldataset

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/dataset"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	filename := filepath.Join("fixtures", "users.db")
	tableName := "userinfo"
	fieldNames := []string{"uid", "username", "dept", "started"}
	_, err := New("sqlite3", filename, tableName, fieldNames)
	if err != nil {
		t.Errorf("New() err: %s", err)
	}
}

func TestNew_errors(t *testing.T) {
	cases := []struct {
		driverName string
		filename   string
		tableName  string
		fieldNames []string
		wantErr    error
	}{
		{driverName: "sqlite3",
			filename:   filepath.Join("fixtures", "users_one_col.db"),
			tableName:  "userinfo",
			fieldNames: []string{"username"},
			wantErr: errors.New(
				"must specify at least two field names",
			),
		},
		{driverName: "sqlite3",
			filename:   filepath.Join("fixtures", "users.db"),
			tableName:  "userinfo",
			fieldNames: []string{"username", "d-ept", "started"},
			wantErr: errors.New(
				"invalid field name: d-ept",
			),
		},
		{driverName: "bob",
			filename:   filepath.Join("fixtures", "users.db"),
			tableName:  "userinfo",
			fieldNames: []string{"username", "dept", "started"},
			wantErr: errors.New(
				"unrecognized driver name: bob",
			),
		},
	}
	for _, c := range cases {
		_, err := New(c.driverName, c.filename, c.tableName, c.fieldNames)
		if err.Error() != c.wantErr.Error() {
			t.Errorf("Open() filename: %s, wantErr: %s, got err: %s",
				c.filename, c.wantErr, err)
		}
	}
}

func TestOpen(t *testing.T) {
	filename := filepath.Join("..", "fixtures", "users.db")
	tableName := "userinfo"
	fieldNames := []string{"uid", "username", "dept", "started"}
	ds, err := New("sqlite3", filename, tableName, fieldNames)
	if err != nil {
		t.Errorf("New() err: %s", err)
	}
	conn, err := ds.Open()
	if err != nil {
		t.Errorf("Open() err: %s", err)
	}
	conn.Close()
}

func TestOpen_errors(t *testing.T) {
	cases := []struct {
		filename   string
		tableName  string
		fieldNames []string
		wantErr    error
	}{
		{filename: filepath.Join("fixtures", "missing.db"),
			tableName:  "userinfo",
			fieldNames: []string{"uid", "username", "dept", "started"},
			wantErr: fmt.Errorf("database doesn't exist: %s",
				filepath.Join("fixtures", "missing.db")),
		},
		{filename: filepath.Join("..", "fixtures", "users.db"),
			tableName:  "missing",
			fieldNames: []string{"uid", "username", "dept", "started"},
			wantErr:    errors.New("table name doesn't exist: missing"),
		},
		{filename: filepath.Join("..", "fixtures", "users.db"),
			tableName:  "userinfo",
			fieldNames: []string{"username", "dept", "started"},
			wantErr: errors.New(
				"number of field names doesn't match number of columns in table",
			),
		},
	}
	for _, c := range cases {
		ds, err := New("sqlite3", c.filename, c.tableName, c.fieldNames)
		if err != nil {
			t.Errorf("New(%s, %s, %s) err: %s",
				c.filename, c.tableName, c.fieldNames, err)
		}
		if _, err := ds.Open(); err.Error() != c.wantErr.Error() {
			t.Errorf("Open() filename: %s, wantErr: %s, got err: %s",
				c.filename, c.wantErr, err)
		}
	}
}

func TestGetFieldNames(t *testing.T) {
	filename := filepath.Join("fixtures", "users.db")
	tableName := "userinfo"
	fieldNames := []string{"uid", "username", "dept", "started"}
	ds, err := New("sqlite3", filename, tableName, fieldNames)
	if err != nil {
		t.Errorf("New() err: %s", err)
	}
	got := ds.GetFieldNames()
	if !reflect.DeepEqual(got, fieldNames) {
		t.Errorf("GetFieldNames() - got: %s, want: %s", got, fieldNames)
	}
}

func TestNext(t *testing.T) {
	wantNumRecords := 4
	filename := filepath.Join("..", "fixtures", "users.db")
	tableName := "userinfo"
	fieldNames := []string{"uid", "username", "dept", "started"}
	ds, err := New("sqlite3", filename, tableName, fieldNames)
	if err != nil {
		t.Errorf("New() err: %s", err)
	}
	conn, err := ds.Open()
	if err != nil {
		t.Errorf("Open() - filename: %s, err: %s", filename, err)
	}
	defer conn.Close()
	numRecords := 0
	for conn.Next() {
		numRecords++
	}
	if conn.Next() {
		t.Errorf("conn.Next() - Return true, despite having finished")
	}
	if numRecords != wantNumRecords {
		t.Errorf("conn.Next() - wantNumRecords: %d, gotNumRecords: %d",
			wantNumRecords, numRecords)
	}
}

func TestRead(t *testing.T) {
	filename := filepath.Join("..", "fixtures", "users.db")
	tableName := "userinfo"
	fieldNames := []string{"uid", "name", "dpt", "startDate"}
	wantRecords := []dataset.Record{
		dataset.Record{
			"uid":       dlit.MustNew(1),
			"name":      dlit.MustNew("Fred Wilkins"),
			"dpt":       dlit.MustNew("Logistics"),
			"startDate": dlit.MustNew("2013-10-05 10:00:00"),
		},
		dataset.Record{
			"uid":       dlit.MustNew(2),
			"name":      dlit.MustNew("Bob Field"),
			"dpt":       dlit.MustNew("Logistics"),
			"startDate": dlit.MustNew("2013-05-05 10:00:00"),
		},
		dataset.Record{
			"uid":       dlit.MustNew(3),
			"name":      dlit.MustNew("Ned James"),
			"dpt":       dlit.MustNew("Shipping"),
			"startDate": dlit.MustNew("2012-05-05 10:00:00"),
		},
		dataset.Record{
			"uid":       dlit.MustNew(4),
			"name":      dlit.MustNew("Mary Terence"),
			"dpt":       dlit.MustNew("Shipping"),
			"startDate": dlit.MustNew("2011-05-05 10:00:00"),
		},
	}

	ds, err := New("sqlite3", filename, tableName, fieldNames)
	if err != nil {
		t.Errorf("New() err: %s", err)
		return
	}
	conn, err := ds.Open()
	if err != nil {
		t.Errorf("Open() - filename: %s, err: %s", filename, err)
		return
	}
	defer conn.Close()

	for _, wantRecord := range wantRecords {
		if !conn.Next() {
			t.Errorf("Next() - return false early")
			return
		}
		record := conn.Read()
		if !matchRecords(record, wantRecord) {
			t.Errorf("Read() got: %s, want: %s", record, wantRecord)
		}
		if err := conn.Err(); err != nil {
			t.Errorf("Err() err: %s", err)
		}
	}
}

/*************************
 *  Benchmarks
 *************************/
func BenchmarkNext(b *testing.B) {
	filename := filepath.Join("..", "fixtures", "debt.db")
	tableName := "people"
	fieldNames := []string{
		"name",
		"balance",
		"numCards",
		"martialStatus",
		"tertiaryEducated",
		"success",
	}
	ds, err := New("sqlite3", filename, tableName, fieldNames)
	if err != nil {
		b.Errorf("New() - filename: %s, err: %s", filename, err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		conn, err := ds.Open()
		if err != nil {
			b.Errorf("Open() - filename: %s, err: %s", filename, err)
		}
		b.StartTimer()
		for conn.Next() {
		}
	}
}

/*************************
 *   Helper functions
 *************************/

func matchRecords(r1 dataset.Record, r2 dataset.Record) bool {
	if len(r1) != len(r2) {
		return false
	}
	for fieldName, value := range r1 {
		if value.String() != r2[fieldName].String() {
			return false
		}
	}
	return true
}

func copyFile(srcFilename, dstDir string) error {
	contents, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcFilename)
	if err != nil {
		return err
	}
	mode := info.Mode()
	dstFilename := filepath.Join(dstDir, filepath.Base(srcFilename))
	return ioutil.WriteFile(dstFilename, contents, mode)
}
