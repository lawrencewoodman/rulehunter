package html

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGenReportFilename(t *testing.T) {
	cases := []struct {
		stamp        time.Time
		title        string
		wantFilename string
	}{
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			title: "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"this-could-be-very-interesting",
				"index.html",
			)},
	}
	for _, c := range cases {
		got := genReportFilename(c.stamp, c.title)
		if got != c.wantFilename {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.stamp, c.title, got, c.wantFilename)
		}
	}
}
