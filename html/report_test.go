package html

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/report"
)

var testDescription = &description.Description{
	map[string]*description.Field{
		"month": {description.String, nil, nil, 0,
			map[string]description.Value{
				"feb":  {dlit.MustNew("feb"), 3},
				"may":  {dlit.MustNew("may"), 2},
				"june": {dlit.MustNew("june"), 9},
			},
			3,
		},
		"rate": {
			description.Number,
			dlit.MustNew(0.3),
			dlit.MustNew(15.1),
			3,
			map[string]description.Value{
				"0.3":   {dlit.MustNew(0.3), 7},
				"7":     {dlit.MustNew(7), 2},
				"7.3":   {dlit.MustNew(7.3), 9},
				"9.278": {dlit.MustNew(9.278), 4},
			},
			4,
		},
		"method": {description.Ignore, nil, nil, 0,
			map[string]description.Value{}, -1},
	},
}

func TestGenerateReport_single_true_rule(t *testing.T) {
	report := &report.Report{
		Mode:               report.Train,
		Title:              "some title",
		Tags:               []string{"bank", "test / fred"},
		Category:           "testing",
		Stamp:              time.Now(),
		ExperimentFilename: "somename.yaml",
		NumRecords:         1000,
		SortOrder: []assessment.SortOrder{
			assessment.SortOrder{
				Aggregator: "goalsScore",
				Direction:  assessment.DESCENDING,
			},
			assessment.SortOrder{
				Aggregator: "percentMatches",
				Direction:  assessment.ASCENDING,
			},
		},
		Aggregators: []report.AggregatorDesc{
			report.AggregatorDesc{Name: "numMatches", Kind: "count", Arg: "true()"},
			report.AggregatorDesc{
				Name: "percentMatches",
				Kind: "calc",
				Arg:  "roundto(100.0 * numMatches / numRecords, 2)",
			},
			report.AggregatorDesc{Name: "numIncomeGt2", Kind: "count", Arg: "income > 2"},
			report.AggregatorDesc{Name: "goalsScore", Kind: "goalsscore", Arg: ""},
		},
		Description: testDescription,
		Assessments: []*report.Assessment{
			&report.Assessment{
				Rule: "true()",
				Aggregators: []*report.Aggregator{
					&report.Aggregator{
						Name:          "goalsScore",
						OriginalValue: "0.1",
						RuleValue:     "0.1",
						Difference:    "0",
					},
					&report.Aggregator{
						Name:          "numIncomeGt2",
						OriginalValue: "2",
						RuleValue:     "2",
						Difference:    "0",
					},
					&report.Aggregator{
						Name:          "numMatches",
						OriginalValue: "142",
						RuleValue:     "142",
						Difference:    "0",
					},
					&report.Aggregator{
						Name:          "percentMatches",
						OriginalValue: "42",
						RuleValue:     "42",
						Difference:    "0",
					},
				},
				Goals: []*report.Goal{
					&report.Goal{
						Expr:           "numIncomeGt2 == 1",
						OriginalPassed: false,
						RulePassed:     false,
					},
					&report.Goal{
						Expr:           "numIncomeGt2 == 2",
						OriginalPassed: true,
						RulePassed:     true,
					},
				},
			},
		},
	}

	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	wantReportURLDir := "reports/category/testing/some-title/train/"

	reportURLDir, err := generateReport(report, cfg)
	if err != nil {
		t.Fatalf("generateReport: %s", err)
	}
	if reportURLDir != wantReportURLDir {
		t.Errorf("generateReport - wantReportURLDir: %s, got: %s",
			wantReportURLDir, reportURLDir)
	}

	htmlFilename := filepath.Join(
		cfg.WWWDir,
		"reports",
		"category",
		"testing",
		"some-title",
		"train",
		"index.html",
	)
	// read the whole file at once
	b, err := ioutil.ReadFile(htmlFilename)
	if err != nil {
		t.Fatalf("ReadFile: %s", err)
	}
	s := string(b)
	wantText := "No rule found that improves on the original dataset"
	dontWantText := "Original Value"
	if !strings.Contains(s, wantText) {
		t.Errorf("html file: %s, doesn't contain text \"%s\"",
			htmlFilename, wantText)
	}
	if strings.Contains(s, dontWantText) {
		t.Errorf("html file: %s, contains text \"%s\"",
			htmlFilename, dontWantText)
	}
}

func TestGenerateReport_two_rules(t *testing.T) {
	report := &report.Report{
		Mode:               report.Train,
		Title:              "some title",
		Tags:               []string{"bank", "test / fred"},
		Category:           "testing",
		Stamp:              time.Now(),
		ExperimentFilename: "somename.yaml",
		NumRecords:         1000,
		SortOrder: []assessment.SortOrder{
			assessment.SortOrder{
				Aggregator: "goalsScore",
				Direction:  assessment.DESCENDING,
			},
			assessment.SortOrder{
				Aggregator: "percentMatches",
				Direction:  assessment.ASCENDING,
			},
		},
		Aggregators: []report.AggregatorDesc{
			report.AggregatorDesc{Name: "numMatches", Kind: "count", Arg: "true()"},
			report.AggregatorDesc{
				Name: "percentMatches",
				Kind: "calc",
				Arg:  "roundto(100.0 * numMatches / numRecords, 2)",
			},
			report.AggregatorDesc{Name: "numIncomeGt2", Kind: "count", Arg: "income > 2"},
			report.AggregatorDesc{Name: "goalsScore", Kind: "goalsscore", Arg: ""},
		},
		Description: testDescription,
		Assessments: []*report.Assessment{
			&report.Assessment{
				Rule: "rate >= 789.2",
				Aggregators: []*report.Aggregator{
					&report.Aggregator{
						Name:          "goalsScore",
						OriginalValue: "0.1",
						RuleValue:     "30.1",
						Difference:    "30",
					},
					&report.Aggregator{
						Name:          "numIncomeGt2",
						OriginalValue: "2",
						RuleValue:     "32",
						Difference:    "30",
					},
					&report.Aggregator{
						Name:          "numMatches",
						OriginalValue: "142",
						RuleValue:     "3142",
						Difference:    "3000",
					},
					&report.Aggregator{
						Name:          "percentMatches",
						OriginalValue: "42",
						RuleValue:     "342",
						Difference:    "300",
					},
				},
				Goals: []*report.Goal{
					&report.Goal{
						Expr:           "numIncomeGt2 == 1",
						OriginalPassed: false,
						RulePassed:     false,
					},
					&report.Goal{
						Expr:           "numIncomeGt2 == 2",
						OriginalPassed: true,
						RulePassed:     false,
					},
				},
			},
			&report.Assessment{
				Rule: "true()",
				Aggregators: []*report.Aggregator{
					&report.Aggregator{
						Name:          "goalsScore",
						OriginalValue: "0.1",
						RuleValue:     "0.1",
						Difference:    "0",
					},
					&report.Aggregator{
						Name:          "numIncomeGt2",
						OriginalValue: "2",
						RuleValue:     "2",
						Difference:    "0",
					},
					&report.Aggregator{
						Name:          "numMatches",
						OriginalValue: "142",
						RuleValue:     "142",
						Difference:    "0",
					},
					&report.Aggregator{
						Name:          "percentMatches",
						OriginalValue: "42",
						RuleValue:     "42",
						Difference:    "0",
					},
				},
				Goals: []*report.Goal{
					&report.Goal{
						Expr:           "numIncomeGt2 == 1",
						OriginalPassed: false,
						RulePassed:     false,
					},
					&report.Goal{
						Expr:           "numIncomeGt2 == 2",
						OriginalPassed: true,
						RulePassed:     true,
					},
				},
			},
		},
	}

	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	wantReportURLDir := "reports/category/testing/some-title/train/"

	reportURLDir, err := generateReport(report, cfg)
	if err != nil {
		t.Fatalf("generateReport: %s", err)
	}
	if reportURLDir != wantReportURLDir {
		t.Errorf("generateReport - wantReportURLDir: %s, got: %s",
			wantReportURLDir, reportURLDir)
	}

	htmlFilename := filepath.Join(
		cfg.WWWDir,
		"reports",
		"category",
		"testing",
		"some-title",
		"train",
		"index.html",
	)
	// read the whole file at once
	b, err := ioutil.ReadFile(htmlFilename)
	if err != nil {
		t.Fatalf("ReadFile: %s", err)
	}
	s := string(b)

	wantText := "Original Value"
	dontWantText := "No rule found that improves on the original dataset"
	if !strings.Contains(s, wantText) {
		t.Errorf("html file: %s, doesn't contain text \"%s\"",
			htmlFilename, wantText)
	}
	if strings.Contains(s, dontWantText) {
		t.Errorf("html file: %s, contains text \"%s\"",
			htmlFilename, dontWantText)
	}
}

func TestGenReportFilename(t *testing.T) {
	cases := []struct {
		stamp        time.Time
		mode         report.ModeKind
		category     string
		title        string
		wantFilename string
	}{
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			mode:     report.Train,
			category: "",
			title:    "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"nocategory",
				"this-could-be-very-interesting",
				"train",
				"index.html",
			)},
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			mode:     report.Train,
			category: "acme or emca",
			title:    "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"category",
				"acme-or-emca",
				"this-could-be-very-interesting",
				"train",
				"index.html",
			)},
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			mode:     report.Test,
			category: "",
			title:    "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"nocategory",
				"this-could-be-very-interesting",
				"test",
				"index.html",
			)},
		{stamp: time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			mode:     report.Test,
			category: "acme or emca",
			title:    "This could be very interesting",
			wantFilename: filepath.Join(
				"reports",
				"category",
				"acme-or-emca",
				"this-could-be-very-interesting",
				"test",
				"index.html",
			)},
	}
	for _, c := range cases {
		got := genReportFilename(c.mode, c.category, c.title)
		if got != c.wantFilename {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.stamp, c.title, got, c.wantFilename)
		}
	}
}
