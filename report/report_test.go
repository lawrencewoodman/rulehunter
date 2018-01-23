package report

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/aggregator"
	rhkassessment "github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
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
	}}

func TestNew(t *testing.T) {
	aggregatorSpecs := []aggregator.Spec{
		aggregator.MustNew("numMatches", "count", "true()"),
		aggregator.MustNew(
			"percentMatches",
			"calc",
			"roundto(100.0 * numMatches / numRecords, 2)",
		),
		aggregator.MustNew("numIncomeGt2", "count", "income > 2"),
		aggregator.MustNew("goalsScore", "goalsscore"),
	}
	goals := []*goal.Goal{
		goal.MustNew("numIncomeGt2 == 1"),
		goal.MustNew("numIncomeGt2 == 2"),
	}

	assessment := rhkassessment.New(aggregatorSpecs, goals)
	assessment.RuleAssessments = []*rhkassessment.RuleAssessment{
		&rhkassessment.RuleAssessment{
			Rule: rule.NewEQFV("month", dlit.NewString("may")),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("2142"),
				"percentMatches": dlit.MustNew("242"),
				"numIncomeGt2":   dlit.MustNew("22"),
				"goalsScore":     dlit.MustNew(20.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", false},
			},
		},
		&rhkassessment.RuleAssessment{
			Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("3142"),
				"percentMatches": dlit.MustNew("342"),
				"numIncomeGt2":   dlit.MustNew("32"),
				"goalsScore":     dlit.MustNew(30.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", false},
			},
		},
		&rhkassessment.RuleAssessment{
			Rule: rule.NewTrue(),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("142"),
				"percentMatches": dlit.MustNew("42"),
				"numIncomeGt2":   dlit.MustNew("2"),
				"goalsScore":     dlit.MustNew(0.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
	}

	wantReport := &Report{
		Mode:               Train,
		Title:              "some title",
		Tags:               []string{"bank", "test / fred"},
		Category:           "testing",
		Stamp:              time.Now(),
		ExperimentFilename: "somename.yaml",
		NumRecords:         assessment.NumRecords,
		SortOrder: []rhkassessment.SortOrder{
			rhkassessment.SortOrder{
				Aggregator: "goalsScore",
				Direction:  rhkassessment.DESCENDING,
			},
			rhkassessment.SortOrder{
				Aggregator: "percentMatches",
				Direction:  rhkassessment.ASCENDING,
			},
		},
		Aggregators: []AggregatorDesc{
			AggregatorDesc{Name: "numMatches", Kind: "count", Arg: "true()"},
			AggregatorDesc{
				Name: "percentMatches",
				Kind: "calc",
				Arg:  "roundto(100.0 * numMatches / numRecords, 2)",
			},
			AggregatorDesc{Name: "numIncomeGt2", Kind: "count", Arg: "income > 2"},
			AggregatorDesc{Name: "goalsScore", Kind: "goalsscore", Arg: ""},
		},
		Description: testDescription,
		Assessments: []*Assessment{
			&Assessment{
				Rule: "rate >= 789.2",
				Aggregators: []*Aggregator{
					&Aggregator{
						Name:          "goalsScore",
						OriginalValue: "0.1",
						RuleValue:     "30.1",
						Difference:    "30",
					},
					&Aggregator{
						Name:          "numIncomeGt2",
						OriginalValue: "2",
						RuleValue:     "32",
						Difference:    "30",
					},
					&Aggregator{
						Name:          "numMatches",
						OriginalValue: "142",
						RuleValue:     "3142",
						Difference:    "3000",
					},
					&Aggregator{
						Name:          "percentMatches",
						OriginalValue: "42",
						RuleValue:     "342",
						Difference:    "300",
					},
				},
				Goals: []*Goal{
					&Goal{
						Expr:           "numIncomeGt2 == 1",
						OriginalPassed: false,
						RulePassed:     false,
					},
					&Goal{
						Expr:           "numIncomeGt2 == 2",
						OriginalPassed: true,
						RulePassed:     false,
					},
				},
			},
			&Assessment{
				Rule: "month == \"may\"",
				Aggregators: []*Aggregator{
					&Aggregator{
						Name:          "goalsScore",
						OriginalValue: "0.1",
						RuleValue:     "20.1",
						Difference:    "20",
					},
					&Aggregator{
						Name:          "numIncomeGt2",
						OriginalValue: "2",
						RuleValue:     "22",
						Difference:    "20",
					},
					&Aggregator{
						Name:          "numMatches",
						OriginalValue: "142",
						RuleValue:     "2142",
						Difference:    "2000",
					},
					&Aggregator{
						Name:          "percentMatches",
						OriginalValue: "42",
						RuleValue:     "242",
						Difference:    "200",
					},
				},
				Goals: []*Goal{
					&Goal{
						Expr:           "numIncomeGt2 == 1",
						OriginalPassed: false,
						RulePassed:     false,
					},
					&Goal{
						Expr:           "numIncomeGt2 == 2",
						OriginalPassed: true,
						RulePassed:     false,
					},
				},
			},
		},
	}
	got := New(
		wantReport.Mode,
		wantReport.Title,
		wantReport.Description,
		assessment,
		aggregatorSpecs,
		wantReport.SortOrder,
		wantReport.ExperimentFilename,
		wantReport.Tags,
		wantReport.Category,
	)
	if err := checkReportsMatch(got, wantReport); err != nil {
		t.Errorf("New: %s", err)
	}
}

func TestLoadJSON_errors(t *testing.T) {
	// File mode permission used as standard:
	// No special permission bits
	// User: Read, Write Execute
	// Group: Read
	// Other: None
	const modePerm = 0740
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	cfg := &config.Config{BuildDir: tmpDir}
	reportsDir := filepath.Join(tmpDir, "reports")
	if err := os.MkdirAll(reportsDir, modePerm); err != nil {
		t.Fatalf("MkdirAll: %s", err)
	}
	testhelpers.CopyFile(t, filepath.Join("fixtures", "empty.json"), reportsDir)

	cases := []struct {
		filename string
		wantErr  error
	}{
		{filename: "nonexistent.json",
			wantErr: &os.PathError{
				Op:   "open",
				Path: filepath.Join(reportsDir, "nonexistent.json"),
				Err:  syscall.ENOENT,
			},
		},
		{filename: "empty.json",
			wantErr: fmt.Errorf(
				"can't decode JSON file: %s, %s",
				filepath.Join(reportsDir, "empty.json"),
				io.EOF),
		},
	}
	for _, c := range cases {
		got, err := LoadJSON(cfg, c.filename)
		if got != nil {
			t.Errorf("LoadJSON: got: %v, want: nil", got)
		}
		if err == nil || err.Error() != c.wantErr.Error() {
			t.Errorf("LoadJSON: gotErr: %s, wantErr: %s", err, c.wantErr)
		}
	}
}

func TestWriteLoadJSON(t *testing.T) {
	// File mode permission used as standard:
	// No special permission bits
	// User: Read, Write Execute
	// Group: Read
	// Other: None
	const modePerm = 0740

	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	reportsDir := filepath.Join(tmpDir, "reports")
	if err := os.MkdirAll(reportsDir, modePerm); err != nil {
		t.Fatalf("MkdirAll: %s", err)
	}
	goals := []*goal.Goal{
		goal.MustNew("numIncomeGt2 == 1"),
		goal.MustNew("numIncomeGt2 == 2"),
	}
	aggregators := []aggregator.Spec{
		aggregator.MustNew("numMatches", "count", "true()"),
		aggregator.MustNew(
			"percentMatches",
			"calc",
			"roundto(100.0 * numMatches / numRecords, 2)",
		),
		aggregator.MustNew("numIncomeGt2", "count", "income > 2"),
		aggregator.MustNew("goalsScore", "goalsscore"),
	}
	assessment := rhkassessment.New(aggregators, goals)
	assessment.RuleAssessments = []*rhkassessment.RuleAssessment{
		&rhkassessment.RuleAssessment{
			Rule: rule.NewEQFV("month", dlit.NewString("may")),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("2142"),
				"percentMatches": dlit.MustNew("242"),
				"numIncomeGt2":   dlit.MustNew("22"),
				"goalsScore":     dlit.MustNew(20.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
		&rhkassessment.RuleAssessment{
			Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("3142"),
				"percentMatches": dlit.MustNew("342"),
				"numIncomeGt2":   dlit.MustNew("32"),
				"goalsScore":     dlit.MustNew(30.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
		&rhkassessment.RuleAssessment{
			Rule: rule.NewTrue(),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("142"),
				"percentMatches": dlit.MustNew("42"),
				"numIncomeGt2":   dlit.MustNew("2"),
				"goalsScore":     dlit.MustNew(0.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
	}

	title := "some title"
	sortOrder := []rhkassessment.SortOrder{
		rhkassessment.SortOrder{
			Aggregator: "goalsScore",
			Direction:  rhkassessment.DESCENDING,
		},
		rhkassessment.SortOrder{
			Aggregator: "percentMatches",
			Direction:  rhkassessment.ASCENDING,
		},
	}
	experimentFilename := "somename.yaml"
	tags := []string{"bank", "test / fred"}
	category := "testing"
	config := &config.Config{BuildDir: tmpDir}
	report := New(
		Train,
		title,
		testDescription,
		assessment,
		aggregators,
		sortOrder,
		experimentFilename,
		tags,
		category,
	)

	if err := report.WriteJSON(config); err != nil {
		t.Fatalf("WriteJSON: %s", err)
	}
	buildFilename := internal.MakeBuildFilename(Train.String(), category, title)
	loadedReport, err := LoadJSON(config, buildFilename)
	if err != nil {
		t.Fatalf("LoadJSON: %s", err)
	}
	if err := checkReportsMatch(report, loadedReport); err != nil {
		t.Errorf("Reports don't match: %s", err)
	}
}

func TestCalcTrueAggregatorDiff(t *testing.T) {
	trueAggregators := map[string]*dlit.Literal{
		"numMatches": dlit.MustNew(176),
		"profit":     dlit.MustNew(23),
		"bigNum":     dlit.MustNew(int64(math.MaxInt64)),
	}
	cases := []struct {
		name  string
		value *dlit.Literal
		want  string
	}{
		{name: "numMatches", value: dlit.MustNew(192), want: "16"},
		{name: "numMatches", value: dlit.MustNew(165), want: "-11"},
		{name: "bigNum",
			value: dlit.MustNew(int64(math.MinInt64)),
			want: dlit.MustNew(
				float64(math.MinInt64) - float64(math.MaxInt64),
			).String(),
		},
		{name: "bigNum",
			value: dlit.MustNew(errors.New("some error")),
			want:  "N/A",
		},
	}

	for _, c := range cases {
		got := calcTrueAggregatorDiff(trueAggregators, c.name, c.value)
		if got != c.want {
			t.Errorf("calcTrueAggregatorDifference(trueAggregators, %v, %v) got: %s, want: %s",
				c.name, c.value, got, c.want)
		}
	}
}

func TestLoadJSON_multiple_attempts(t *testing.T) {
	// File mode permission used as standard:
	// No special permission bits
	// User: Read, Write Execute
	// Group: Read
	// Other: None
	const modePerm = 0740

	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	reportsDir := filepath.Join(tmpDir, "reports")
	if err := os.MkdirAll(reportsDir, modePerm); err != nil {
		t.Fatalf("MkdirAll: %s", err)
	}
	aggregators := []aggregator.Spec{
		aggregator.MustNew("numMatches", "count", "true()"),
		aggregator.MustNew(
			"percentMatches",
			"calc",
			"roundto(100.0 * numMatches / numRecords, 2)",
		),
		aggregator.MustNew("numIncomeGt2", "count", "income > 2"),
		aggregator.MustNew("goalsScore", "goalsscore"),
	}
	goals := []*goal.Goal{
		goal.MustNew("numIncomeGt2 == 1"),
		goal.MustNew("numIncomeGt2 == 2"),
	}
	assessment := rhkassessment.New(aggregators, goals)
	assessment.RuleAssessments = []*rhkassessment.RuleAssessment{
		&rhkassessment.RuleAssessment{
			Rule: rule.NewEQFV("month", dlit.NewString("may")),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("2142"),
				"percentMatches": dlit.MustNew("242"),
				"numIncomeGt2":   dlit.MustNew("22"),
				"goalsScore":     dlit.MustNew(20.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
		&rhkassessment.RuleAssessment{
			Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("3142"),
				"percentMatches": dlit.MustNew("342"),
				"numIncomeGt2":   dlit.MustNew("32"),
				"goalsScore":     dlit.MustNew(30.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
		&rhkassessment.RuleAssessment{
			Rule: rule.NewTrue(),
			Aggregators: map[string]*dlit.Literal{
				"numMatches":     dlit.MustNew("142"),
				"percentMatches": dlit.MustNew("42"),
				"numIncomeGt2":   dlit.MustNew("2"),
				"goalsScore":     dlit.MustNew(0.1),
			},
			Goals: []*rhkassessment.GoalAssessment{
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
				&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
			},
		},
	}

	title := "some title"
	sortOrder := []rhkassessment.SortOrder{
		rhkassessment.SortOrder{
			Aggregator: "goalsScore",
			Direction:  rhkassessment.DESCENDING,
		},
		rhkassessment.SortOrder{
			Aggregator: "percentMatches",
			Direction:  rhkassessment.ASCENDING,
		},
	}
	experimentFilename := "somename.yaml"
	tags := []string{"bank", "test / fred"}
	category := "testing"
	config := &config.Config{BuildDir: tmpDir}
	report := New(
		Train,
		title,
		testDescription,
		assessment,
		aggregators,
		sortOrder,
		experimentFilename,
		tags,
		category,
	)

	buildFilename := internal.MakeBuildFilename(Train.String(), category, title)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "empty.json"),
		reportsDir,
		buildFilename,
	)
	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := report.WriteJSON(config); err != nil {
			t.Fatalf("WriteJSON: %s", err)
		}
	}()
	maxLoadAttempts := 5
	loadedReport, err := LoadJSON(config, buildFilename, maxLoadAttempts)
	if err != nil {
		t.Fatalf("LoadJSON: %s", err)
	}
	if err := checkReportsMatch(report, loadedReport); err != nil {
		t.Errorf("Reports don't match: %s", err)
	}
}

func TestGetSortedAggregatorNames(t *testing.T) {
	aggregators := map[string]*dlit.Literal{
		"numMatches": dlit.MustNew(176),
		"profit":     dlit.MustNew(23),
		"bigNum":     dlit.MustNew(int64(math.MaxInt64)),
	}
	want := []string{"bigNum", "numMatches", "profit"}
	got := getSortedAggregatorNames(aggregators)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getSortedAggregatorNames - got: %v, want: %v", got, want)
	}
}

func TestGetTrueRuleAssessment(t *testing.T) {
	assessment := &rhkassessment.Assessment{
		NumRecords: 20,
		RuleAssessments: []*rhkassessment.RuleAssessment{
			&rhkassessment.RuleAssessment{
				Rule: rule.NewEQFV("month", dlit.NewString("may")),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("2142"),
					"percentMatches": dlit.MustNew("242"),
					"numIncomeGt2":   dlit.MustNew("22"),
					"goalsScore":     dlit.MustNew(20.1),
				},
				Goals: []*rhkassessment.GoalAssessment{
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkassessment.RuleAssessment{
				Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("3142"),
					"percentMatches": dlit.MustNew("342"),
					"numIncomeGt2":   dlit.MustNew("32"),
					"goalsScore":     dlit.MustNew(30.1),
				},
				Goals: []*rhkassessment.GoalAssessment{
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkassessment.RuleAssessment{
				Rule: rule.NewTrue(),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("142"),
					"percentMatches": dlit.MustNew("42"),
					"numIncomeGt2":   dlit.MustNew("2"),
					"goalsScore":     dlit.MustNew(0.1),
				},
				Goals: []*rhkassessment.GoalAssessment{
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
		},
	}
	want := &rhkassessment.RuleAssessment{
		Rule: rule.NewTrue(),
		Aggregators: map[string]*dlit.Literal{
			"numMatches":     dlit.MustNew("142"),
			"percentMatches": dlit.MustNew("42"),
			"numIncomeGt2":   dlit.MustNew("2"),
			"goalsScore":     dlit.MustNew(0.1),
		},
		Goals: []*rhkassessment.GoalAssessment{
			&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
			&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
		},
	}

	got, err := getTrueRuleAssessment(assessment)
	if err != nil {
		t.Fatalf("getTrueRuleAssessment: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getTrueAggregators - got: %v, want: %v", got, want)
	}
}

func TestGetTrueRuleAssessment_error(t *testing.T) {
	assessment := &rhkassessment.Assessment{
		NumRecords: 20,
		RuleAssessments: []*rhkassessment.RuleAssessment{
			&rhkassessment.RuleAssessment{
				Rule: rule.NewEQFV("month", dlit.NewString("may")),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("2142"),
					"percentMatches": dlit.MustNew("242"),
					"numIncomeGt2":   dlit.MustNew("22"),
					"goalsScore":     dlit.MustNew(20.1),
				},
				Goals: []*rhkassessment.GoalAssessment{
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkassessment.RuleAssessment{
				Rule: rule.NewTrue(),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("142"),
					"percentMatches": dlit.MustNew("42"),
					"numIncomeGt2":   dlit.MustNew("2"),
					"goalsScore":     dlit.MustNew(0.1),
				},
				Goals: []*rhkassessment.GoalAssessment{
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkassessment.RuleAssessment{
				Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("3142"),
					"percentMatches": dlit.MustNew("342"),
					"numIncomeGt2":   dlit.MustNew("32"),
					"goalsScore":     dlit.MustNew(30.1),
				},
				Goals: []*rhkassessment.GoalAssessment{
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkassessment.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
		},
	}
	wantErr := errors.New("can't find true() rule")

	_, err := getTrueRuleAssessment(assessment)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("getTrueAggregators: err: %s, wantErr: %s", err, wantErr)
	}
}

/******************************
 *  Helper Functions
 ******************************/

func checkReportsMatch(r1, r2 *Report) error {
	if r1.Mode != r2.Mode {
		return fmt.Errorf("Modes don't match - %s != %s", r1.Mode, r2.Mode)
	}
	if r1.Title != r2.Title {
		return fmt.Errorf("Titles don't match - %s != %s", r1.Title, r2.Title)
	}
	if !reflect.DeepEqual(r1.Tags, r2.Tags) {
		return fmt.Errorf("Tags don't match - %v != %v", r1.Tags, r2.Tags)
	}
	if r1.Category != r2.Category {
		return fmt.Errorf("Categories don't match - %s != %s",
			r1.Category, r2.Category)
	}
	if math.Abs(r1.Stamp.Sub(r2.Stamp).Seconds()) > 1 {
		return fmt.Errorf("Stamps don't match - %s != %s", r1.Stamp, r2.Stamp)
	}
	if r1.ExperimentFilename != r2.ExperimentFilename {
		return fmt.Errorf("ExperimentFilenames don't match - %s != %s",
			r1.ExperimentFilename, r2.ExperimentFilename)
	}
	if r1.NumRecords != r2.NumRecords {
		return fmt.Errorf("NumRecords don't match - %d != %d",
			r1.NumRecords, r2.NumRecords)
	}
	if !reflect.DeepEqual(r1.SortOrder, r2.SortOrder) {
		return fmt.Errorf("SortOrder don't match - %v != %v",
			r1.SortOrder, r2.SortOrder)
	}
	if !reflect.DeepEqual(r1.Aggregators, r2.Aggregators) {
		return fmt.Errorf("Aggregators don't match - %v != %v",
			r1.Aggregators, r2.Aggregators)
	}
	if err := checkAssessmentsMatch(r1.Assessments, r2.Assessments); err != nil {
		return fmt.Errorf("Assessments don't match: %s - %v != %v",
			err, r1.Assessments, r2.Assessments)
	}
	return nil
}

func checkAssessmentsMatch(as1, as2 []*Assessment) error {
	if len(as1) != len(as2) {
		return fmt.Errorf("number of Assessments don't match %d != %d",
			len(as1), len(as2))
	}
	for i, assessment1 := range as1 {
		if assessment1.Rule != as2[i].Rule {
			return fmt.Errorf("assessment[%d] Rules don't match: %v != %v",
				i, assessment1.Rule, as2[i].Rule)
		}
		if !reflect.DeepEqual(assessment1.Aggregators, as2[i].Aggregators) {
			return fmt.Errorf("assessment[%d] Aggregators don't match: %v != %v",
				i, assessment1.Aggregators, as2[i].Aggregators)
		}
		if !reflect.DeepEqual(assessment1.Goals, as2[i].Goals) {
			return fmt.Errorf("assessment[%d] Goals don't match: %v != %v",
				i, assessment1.Goals, as2[i].Goals)
		}
	}
	return nil
}
