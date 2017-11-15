package html

import (
	"reflect"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/report"
)

func TestSortTplReportsByDate(t *testing.T) {
	reports := []*TplReport{
		newTplReport(
			report.Train,
			"title A",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing a",
			"category/testing-a",
			"titlea",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		),
		newTplReport(
			report.Train,
			"title B",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing b",
			"category/testing-b",
			"titleb",
			time.Date(2009, time.November, 10, 24, 0, 0, 0, time.UTC),
		),
		newTplReport(
			report.Test,
			"title C",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing c",
			"category/testing-c",
			"titlec",
			time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC),
		),
		newTplReport(
			report.Test,
			"title D",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing d",
			"category/testing-d",
			"titled",
			time.Date(2009, time.November, 10, 26, 0, 0, 0, time.UTC),
		),
	}
	wantReports := []*TplReport{
		newTplReport(
			report.Test,
			"title D",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing d",
			"category/testing-d",
			"titled",
			time.Date(2009, time.November, 10, 26, 0, 0, 0, time.UTC),
		),
		newTplReport(
			report.Train,
			"title B",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing b",
			"category/testing-b",
			"titleb",
			time.Date(2009, time.November, 10, 24, 0, 0, 0, time.UTC),
		),
		newTplReport(
			report.Train,
			"title A",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing a",
			"category/testing-a",
			"titlea",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		),
		newTplReport(
			report.Test,
			"title C",
			map[string]string{
				"bank": "/tag/bank/",
				"test": "/tag/test",
			},
			"testing c",
			"category/testing-c",
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
