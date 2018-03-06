package html

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/vlifesystems/rulehunter/report"
)

func TestGenReportURLDir(t *testing.T) {
	cases := []struct {
		mode     report.ModeKind
		category string
		title    string
		want     string
	}{
		{mode: report.Train,
			category: "",
			title:    "This could be very interesting",
			want:     "reports/nocategory/this-could-be-very-interesting/train/",
		},
		{mode: report.Train,
			category: "acme or emca",
			title:    "This could be very interesting",
			want:     "reports/category/acme-or-emca/this-could-be-very-interesting/train/",
		},
		{mode: report.Test,
			category: "",
			title:    "This could be very interesting",
			want:     "reports/nocategory/this-could-be-very-interesting/test/",
		},
		{mode: report.Test,
			category: "acme or emca",
			title:    "This could be very interesting",
			want:     "reports/category/acme-or-emca/this-could-be-very-interesting/test/",
		},
	}
	for _, c := range cases {
		got := genReportURLDir(c.mode, c.category, c.title)
		if got != c.want {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.category, c.title, got, c.want)
		}
	}
}

func TestEscapeString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"This is a TITLE",
			"this-is-a-title"},
		{"  hello how are % you423 33  today __ --",
			"hello-how-are-you423-33-today"},
		{"--  hello how are %^& you423 33  today __ --",
			"hello-how-are-you423-33-today"},
		{"hello((_ how are % you423 33  today",
			"hello-how-are-you423-33-today"},
		{"This is it's TITLE",
			"this-is-its-title"},
		{"", ""},
	}
	for _, c := range cases {
		got := escapeString(c.in)
		if got != c.want {
			t.Errorf("escapeString(%s) got: %s, want: %s", c.in, got, c.want)
		}
	}
}

func TestCreatePageErrorError(t *testing.T) {
	err := CreatePageError{
		Filename: "/tmp/somefilename.html",
		Op:       "execute",
		Err:      errors.New("can't write to file"),
	}
	want := "can't create html page for filename: /tmp/somefilename.html, can't write to file (execute)"
	got := err.Error()
	if got != want {
		t.Errorf("Error - got: %s, want: %s", got, want)
	}
}

/***********************************************
 *   Helper Functions
 ***********************************************/

func checkFilesExist(files []string) error {
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("file doesn't exist: %s", f)
		}
	}
	return nil
}

func removeFiles(files []string) error {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
