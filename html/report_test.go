package html

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGenReportFilename(t *testing.T) {
	cases := []struct {
		stamp        time.Time
		category     string
		title        string
		wantFilename string
	}{
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			category: "",
			title:    "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"nocategory",
				"this-could-be-very-interesting",
				"index.html",
			)},
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			category: "acme or emca",
			title:    "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"category",
				"acme-or-emca",
				"this-could-be-very-interesting",
				"index.html",
			)},
	}
	for _, c := range cases {
		got := genReportFilename(c.category, c.title)
		if got != c.wantFilename {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.stamp, c.title, got, c.wantFilename)
		}
	}
}
