package experiment

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"path/filepath"
	"testing"
)

func TestOpenRowsClose(t *testing.T) {
	checkNumRecords := func(handler *sqlHandler, wantNumRecords int) error {
		rows, err := handler.Rows()
		if err != nil {
			return err
		}
		gotNumRecords := 0
		for rows.Next() {
			gotNumRecords++
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if gotNumRecords != wantNumRecords {
			return fmt.Errorf("Next: gotNumRecords: %d, wantNumRecords: %d",
				gotNumRecords, wantNumRecords)
		}
		return nil
	}

	filename := filepath.Join("fixtures", "flow.db")
	h := newSQLHandler("sqlite3", filename, "select * from flow")

	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if err := h.Open(); err != nil {
		t.Fatalf("Open: err: %v", err)
	}
	if err := checkNumRecords(h, 9); err != nil {
		t.Fatalf("checkNumRecords: err: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if err := h.Open(); err != nil {
		t.Fatalf("Open: err: %v", err)
	}
	if err := checkNumRecords(h, 9); err != nil {
		t.Fatalf("checkNumRecords: err: %v", err)
	}
	if err := h.Open(); err != nil {
		t.Fatalf("Open: err: %v", err)
	}
	if err := checkNumRecords(h, 9); err != nil {
		t.Fatalf("checkNumRecords: err: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if err := checkNumRecords(h, 9); err != nil {
		t.Fatalf("checkNumRecords: err: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
}

func TestOpenRowsClose_errors(t *testing.T) {
	filename := filepath.Join("fixtures", "flow.db")
	h := newSQLHandler("sqlite3", filename, "select * from flow")

	if _, err := h.Rows(); err != errDatabaseNotOpen {
		t.Fatalf("Rows: gotErr: %v, want: %v", err, errDatabaseNotOpen)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if _, err := h.Rows(); err != errDatabaseNotOpen {
		t.Fatalf("Rows: gotErr: %v, want: %v", err, errDatabaseNotOpen)
	}
	if err := h.Open(); err != nil {
		t.Fatalf("Open: err: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if _, err := h.Rows(); err != errDatabaseNotOpen {
		t.Fatalf("Rows: gotErr: %v, want: %v", err, errDatabaseNotOpen)
	}
	if err := h.Open(); err != nil {
		t.Fatalf("Open: err: %v", err)
	}
	if err := h.Open(); err != nil {
		t.Fatalf("Open: err: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("Close: err: %v", err)
	}
	if _, err := h.Rows(); err != errDatabaseNotOpen {
		t.Fatalf("Rows: gotErr: %v, want: %v", err, errDatabaseNotOpen)
	}
}

/***********************
   Helper functions
************************/

func matchRecords(r1 ddataset.Record, r2 ddataset.Record) bool {
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
