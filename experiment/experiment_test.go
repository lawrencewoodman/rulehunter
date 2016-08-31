package experiment

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rulehunter/aggregators"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/goal"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/html/cmd"
	"github.com/vlifesystems/rulehuntersrv/internal/progresstest"
	"github.com/vlifesystems/rulehuntersrv/internal/testhelpers"
	"github.com/vlifesystems/rulehuntersrv/logger"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/quitter"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestLoadExperiment(t *testing.T) {
	cases := []struct {
		cfgMaxNumRecords int
		filename         string
		wantExperiment   *experiment.Experiment
		wantTags         []string
		wantWhenExpr     *dexpr.Expr
	}{
		{-1, filepath.Join("fixtures", "flow.json"),
			&experiment.Experiment{
				Title: "What would indicate good flow?",
				Dataset: dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
				ExcludeFieldNames: []string{"flow"},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("goodFlowAccuracy", "accuracy", "flow > 60"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowAccuracy > 10")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"goodFlowAccuracy", experiment.DESCENDING},
					experiment.SortField{"numMatches", experiment.DESCENDING},
				},
			},
			[]string{"test", "fred / ned"},
			dexpr.MustNew("!hasRun"),
		},
		{4, filepath.Join("fixtures", "flow.json"),
			&experiment.Experiment{
				Title: "What would indicate good flow?",
				Dataset: dtruncate.New(
					dcsv.New(
						filepath.Join("fixtures", "flow.csv"),
						true,
						rune(','),
						[]string{"group", "district", "height", "flow"},
					),
					4,
				),
				ExcludeFieldNames: []string{"flow"},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("goodFlowAccuracy", "accuracy", "flow > 60"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowAccuracy > 10")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"goodFlowAccuracy", experiment.DESCENDING},
					experiment.SortField{"numMatches", experiment.DESCENDING},
				},
			},
			[]string{"test", "fred / ned"},
			dexpr.MustNew("!hasRun"),
		},
		{-1, filepath.Join("fixtures", "debt.json"),
			&experiment.Experiment{
				Title: "What would predict people being helped to be debt free?",
				Dataset: dcsv.New(
					filepath.Join("..", "fixtures", "debt.csv"),
					true,
					rune(','),
					[]string{
						"name",
						"balance",
						"numCards",
						"martialStatus",
						"tertiaryEducated",
						"success",
					},
				),
				ExcludeFieldNames: []string{"success"},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("helpedAccuracy", "accuracy", "success"),
				},
				Goals: []*goal.Goal{goal.MustNew("helpedAccuracy > 10")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"helpedAccuracy", experiment.DESCENDING},
					experiment.SortField{"numMatches", experiment.DESCENDING},
				},
			},
			[]string{"debt"},
			dexpr.MustNew("!hasRunToday || sinceLastRunHours > 2"),
		},
	}
	for _, c := range cases {
		gotExperiment, gotTags, gotWhenExpr, err :=
			loadExperiment(c.filename, c.cfgMaxNumRecords)
		if err != nil {
			t.Fatalf("loadExperiment(%s) err: %s", c.filename, err)
			return
		}
		err = checkExperimentMatch(gotExperiment, c.wantExperiment)
		if err != nil {
			t.Errorf("loadExperiment(%s) experiments don't match: %s\n"+
				"gotExperiment: %s, wantExperiment: %s",
				c.filename, err, gotExperiment, c.wantExperiment)
		}
		if gotWhenExpr.String() != c.wantWhenExpr.String() {
			t.Errorf("loadExperiment(%s) gotWhenExpr: %s, wantWhenExpr: %s",
				c.filename, gotWhenExpr, c.wantWhenExpr)
		}
		if !reflect.DeepEqual(gotTags, c.wantTags) {
			t.Errorf("loadExperiment(%s) gotTags: %s, wantTags: %s",
				c.filename, gotTags, c.wantTags)
		}
	}
}

func TestLoadExperiment_error(t *testing.T) {
	cases := []struct {
		filename string
		wantErr  error
	}{
		{filepath.Join("fixtures", "flow_no_title.json"),
			errors.New("Experiment field missing: title")},
		{filepath.Join("fixtures", "flow_no_dataset.json"),
			errors.New("Experiment field missing: dataset")},
		{filepath.Join("fixtures", "flow_invalid_dataset.json"),
			errors.New("Experiment field: dataset, has invalid type: llwyd")},
		{filepath.Join("fixtures", "flow_no_csv.json"),
			errors.New("Experiment field missing: csv")},
		{filepath.Join("fixtures", "flow_no_csv_filename.json"),
			errors.New("Experiment field missing: csv > filename")},
		{filepath.Join("fixtures", "flow_no_csv_separator.json"),
			errors.New("Experiment field missing: csv > separator")},
		{filepath.Join("fixtures", "flow_no_sql.json"),
			errors.New("Experiment field missing: sql")},
		{filepath.Join("fixtures", "flow_no_sql_drivername.json"),
			errors.New("Experiment field missing: sql > driverName")},
		{filepath.Join("fixtures", "flow_invalid_sql_drivername.json"),
			errors.New("Experiment has invalid sql > driverName")},
		{filepath.Join("fixtures", "flow_no_sql_datasourcename.json"),
			errors.New("Experiment field missing: sql > dataSourceName")},
		{filepath.Join("fixtures", "flow_no_sql_query.json"),
			errors.New("Experiment field missing: sql > query")},
		{filepath.Join("fixtures", "flow_invalid_when.json"),
			InvalidWhenExprError("has(twolegs")},
	}
	maxNumRecords := -1
	for _, c := range cases {
		_, _, _, err := loadExperiment(c.filename, maxNumRecords)
		if err == nil {
			t.Errorf("loadExperiment(%s) no error, wantErr:%s",
				c.filename, c.wantErr)
			continue
		}
		if err.Error() != c.wantErr.Error() {
			t.Errorf("loadExperiment(%s) gotErr: %s, wantErr:%s",
				c.filename, err, c.wantErr)
		}
	}
}

func TestInvalidWhenExprErrorError(t *testing.T) {
	e := InvalidWhenExprError("has)nothing")
	want := "When field invalid: has)nothing"
	if got := e.Error(); got != want {
		t.Errorf("Error() got: %v, want: %v", got, want)
	}
}

func TestShouldProcess(t *testing.T) {
	cases := []struct {
		filename string
		when     string
		want     bool
	}{
		{"bank-bad.json", "!hasRun", false},
		{"bank-divorced.json", "!hasRun", false},
		{"bank-tiny.json", "!hasRun", false},
		{"bank-full-divorced.json", "!hasRun", true},
	}
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("..", "progress", "fixtures", "progress.json"),
		tempDir,
	)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := progress.NewMonitor(tempDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}

	for _, c := range cases {
		whenExpr := dexpr.MustNew(c.when)
		got, err := shouldProcess(pm, c.filename, whenExpr)
		if err != nil {
			t.Errorf("shouldProcess(pm, %v, %v) err: %s", c.filename, c.when, err)
			continue
		}
		if got != c.want {
			t.Errorf("shouldProcess(pm, %v, %v) got: %t, want: %t",
				c.filename, c.when, got, c.want)
		}
	}
}

func TestProcess(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumRecords:     100,
		MaxNumProcesses:   4,
		MaxNumReportRules: 100,
	}
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "flow.json"),
		cfg.ExperimentsDir,
	)
	wantLogEntries := []logger.Entry{
		{Level: logger.Info, Msg: "Processing experiment: flow.json"},
		{Level: logger.Info, Msg: "Successfully processed experiment: flow.json"},
	}
	wantPMExperiments := []*progress.Experiment{
		&progress.Experiment{
			Title:              "What would indicate good flow?",
			Tags:               []string{"test", "fred / ned"},
			Stamp:              time.Now(),
			ExperimentFilename: "flow.json",
			Msg:                "Finished processing successfully",
			Status:             progress.Success,
		},
	}

	q := quitter.New()
	l := testhelpers.NewLogger()
	go l.Run(q)
	htmlCmds := make(chan cmd.Cmd)
	defer close(htmlCmds)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
		htmlCmds,
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: err: %v", err)
	}
	if err = Process("flow.json", cfg, l, pm); err != nil {
		t.Fatalf("Process: err: %v", err)
	}

	testStart := time.Now()
	gotCorrectLogEntries := false
	for !gotCorrectLogEntries && time.Since(testStart).Seconds() < 5 {
		if reflect.DeepEqual(l.GetEntries(), wantLogEntries) {
			gotCorrectLogEntries = true
		}
	}
	if !gotCorrectLogEntries {
		t.Errorf("l.GetEntries() got: %v, want: %v", l.GetEntries(), wantLogEntries)
	}

	err = progresstest.CheckExperimentsMatch(
		pm.GetExperiments(),
		wantPMExperiments,
	)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}

	/*
	 *	Check html commands received == {Progress,Progress,Progress,...,Reports}
	 */
	htmlCmdsReceived := cmdMonitor.GetCmdsReceived()
	numCmds := len(htmlCmdsReceived)
	if numCmds < 2 {
		t.Errorf("GetCmdsRecevied() received less than 2 commands")
	}
	lastCmd := htmlCmdsReceived[numCmds-1]
	if lastCmd != cmd.Reports {
		t.Errorf("GetCmdsRecevied() last command got: %s, want: %s",
			lastCmd, cmd.Reports)
	}
	for _, c := range htmlCmdsReceived[:numCmds-1] {
		if c != cmd.Progress {
			t.Errorf(
				"GetCmdsRecevied() rest of commands not all equal to Progress, found: %s",
				c,
			)
		}
	}
	// TODO: Test files generated
}

func TestMakeDataset(t *testing.T) {
	cases := []struct {
		dataSourceName string
		query          string
		fieldNames     []string
		want           ddataset.Dataset
	}{
		{dataSourceName: filepath.Join("fixtures", "flow.db"),
			query:      "select * from flow",
			fieldNames: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{dataSourceName: filepath.Join("fixtures", "flow.db"),
			query:      "select grp,district,flow from flow",
			fieldNames: []string{"grp", "district", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
	}
	for _, c := range cases {
		e := &experimentFile{
			Dataset:    "sql",
			FieldNames: c.fieldNames,
			Sql: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: c.dataSourceName,
				Query:          c.query,
			},
		}
		got, err := makeDataset(e)
		if err != nil {
			t.Errorf("makeDataset(%v) query: %s, err: %v",
				c.query, e, err)
			continue
		}
		if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("checkDatasetsEqual: query: %s, err: %v",
				c.query, err)
		}
	}
}

/***********************
   Helper functions
************************/

func checkExperimentMatch(
	e1 *experiment.Experiment,
	e2 *experiment.Experiment,
) error {
	if e1.Title != e2.Title {
		return errors.New("Titles don't match")
	}
	if !areStringArraysEqual(e1.ExcludeFieldNames, e2.ExcludeFieldNames) {
		return errors.New("ExcludeFieldNames don't match")
	}
	if !areGoalExpressionsEqual(e1.Goals, e2.Goals) {
		return errors.New("Goals don't match")
	}
	if !areAggregatorsEqual(e1.Aggregators, e2.Aggregators) {
		return errors.New("Aggregators don't match")
	}
	if !areSortOrdersEqual(e1.SortOrder, e2.SortOrder) {
		return errors.New("Sort Orders don't match")
	}
	return checkDatasetsEqual(e1.Dataset, e2.Dataset)
}

func checkDatasetsEqual(ds1, ds2 ddataset.Dataset) error {
	conn1, err := ds1.Open()
	if err != nil {
		return err
	}
	conn2, err := ds2.Open()
	if err != nil {
		return err
	}
	for {
		conn1Next := conn1.Next()
		conn2Next := conn2.Next()
		if conn1Next != conn2Next {
			return errors.New("Datasets don't finish at same point")
		}
		if !conn1Next {
			break
		}

		conn1Record := conn1.Read()
		conn2Record := conn2.Read()
		if !reflect.DeepEqual(conn1Record, conn2Record) {
			return fmt.Errorf("Rows don't match %s != %s", conn1Record, conn2Record)
		}
	}
	if conn1.Err() != conn2.Err() {
		return errors.New("Datasets final error doesn't match")
	}
	return nil
}

func areStringArraysEqual(a1 []string, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, e := range a1 {
		if e != a2[i] {
			return false
		}
	}
	return true
}

func areGoalExpressionsEqual(g1 []*goal.Goal, g2 []*goal.Goal) bool {
	if len(g1) != len(g2) {
		return false
	}
	for i, g := range g1 {
		if g.String() != g2[i].String() {
			return false
		}
	}
	return true

}

func areAggregatorsEqual(
	a1 []aggregators.AggregatorSpec,
	a2 []aggregators.AggregatorSpec,
) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, a := range a1 {
		if reflect.TypeOf(a) != reflect.TypeOf(a2[i]) ||
			a.GetName() != a2[i].GetName() ||
			a.GetArg() != a2[i].GetArg() {
			return false
		}
	}
	return true
}

func areSortOrdersEqual(
	so1 []experiment.SortField,
	so2 []experiment.SortField,
) bool {
	if len(so1) != len(so2) {
		return false
	}
	for i, sf1 := range so1 {
		sf2 := so2[i]
		if sf1.Field != sf2.Field || sf1.Direction != sf2.Direction {
			return false
		}
	}
	return true
}
