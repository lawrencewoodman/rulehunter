package html

import (
	"reflect"
	"testing"
	"time"
)

func TestSortTplReportsByDate(t *testing.T) {
	reports := []*TplReport{
		newTplReport(
			"title A",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titlea",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		),
		newTplReport(
			"title B",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titleb",
			time.Date(2009, time.November, 10, 24, 0, 0, 0, time.UTC),
		),
		newTplReport(
			"title C",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titlec",
			time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC),
		),
		newTplReport(
			"title D",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titled",
			time.Date(2009, time.November, 10, 26, 0, 0, 0, time.UTC),
		),
	}
	wantReports := []*TplReport{
		newTplReport(
			"title D",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titled",
			time.Date(2009, time.November, 10, 26, 0, 0, 0, time.UTC),
		),
		newTplReport(
			"title B",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titleb",
			time.Date(2009, time.November, 10, 24, 0, 0, 0, time.UTC),
		),
		newTplReport(
			"title A",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titlea",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		),
		newTplReport(
			"title C",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"titlec",
			time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC),
		),
	}
	sortTplReportsByDate(reports)
	if !reflect.DeepEqual(reports, wantReports) {
		t.Errorf("sortTplReportsByDate(reports) got: %s, want: %s",
			reports, wantReports)
	}
}
