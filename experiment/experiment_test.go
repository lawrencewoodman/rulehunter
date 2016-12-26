package experiment

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rhkit/aggregators"
	"github.com/vlifesystems/rhkit/experiment"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/progresstest"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"syscall"
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
				RuleFieldNames: []string{"group", "district", "height"},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("goodFlowMcc", "mcc", "flow > 60"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"goodFlowMcc", experiment.DESCENDING},
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
				RuleFieldNames: []string{"group", "district", "height"},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("goodFlowMcc", "mcc", "flow > 60"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"goodFlowMcc", experiment.DESCENDING},
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
				RuleFieldNames: []string{
					"name",
					"balance",
					"numCards",
					"martialStatus",
					"tertiaryEducated",
				},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("helpedMcc", "mcc", "success"),
				},
				Goals: []*goal.Goal{goal.MustNew("helpedMcc > 0")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"helpedMcc", experiment.DESCENDING},
					experiment.SortField{"numMatches", experiment.DESCENDING},
				},
			},
			[]string{"debt"},
			dexpr.MustNew("!hasRunToday || sinceLastRunHours > 2"),
		},
		{-1, filepath.Join("fixtures", "flow.yaml"),
			&experiment.Experiment{
				Title: "What would indicate good flow?",
				Dataset: dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
				RuleFieldNames: []string{"group", "district", "height"},
				Aggregators: []aggregators.AggregatorSpec{
					aggregators.MustNew("goodFlowMcc", "mcc", "flow > 60"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"goodFlowMcc", experiment.DESCENDING},
					experiment.SortField{"numMatches", experiment.DESCENDING},
				},
			},
			[]string{"test", "fred / ned"},
			dexpr.MustNew("!hasRun"),
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
			errors.New("Experiment field: sql, has invalid driverName: bob")},
		{filepath.Join("fixtures", "flow_no_sql_datasourcename.json"),
			errors.New("Experiment field missing: sql > dataSourceName")},
		{filepath.Join("fixtures", "flow_no_sql_query.json"),
			errors.New("Experiment field missing: sql > query")},
		{filepath.Join("fixtures", "flow_invalid_when.json"),
			InvalidWhenExprError("has(twolegs")},
		{filepath.Join("fixtures", "flow_invalid.json"),
			errors.New("invalid character '\\n' in string literal")},
		{filepath.Join("fixtures", "flow.bob"),
			InvalidExtError(".bob")},
		{filepath.Join("fixtures", "flow_nonexistant.json"),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "flow_nonexistant.json"),
				syscall.ENOENT,
			},
		},
		{filepath.Join("fixtures", "flow_nonexistant.yaml"),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "flow_nonexistant.yaml"),
				syscall.ENOENT,
			},
		},
		{filepath.Join("fixtures", "flow_invalid.yaml"),
			errors.New("yaml: line 3: did not find expected key"),
		},
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

func TestInvalidExtErrorError(t *testing.T) {
	ext := ".exe"
	err := InvalidExtError(ext)
	want := "invalid extension: " + ext
	if got := err.Error(); got != want {
		t.Errorf("Error() got: %v, want: %v", got, want)
	}
}

func TestShouldProcess(t *testing.T) {
	cases := []struct {
		file fileinfo.FileInfo
		when string
		want bool
	}{
		{file: testhelpers.NewFileInfo("bank-divorced.json", time.Now()),
			when: "!hasRun",
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-divorced.json",
			testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00")),
			when: "!hasRun",
			want: false,
		},
		{file: testhelpers.NewFileInfo("bank-tiny.json", time.Now()),
			when: "!hasRun",
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-tiny.json",
			testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00")),
			when: "!hasRun",
			want: false,
		},
		{file: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			when: "!hasRun",
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			when: "!hasRun",
			want: true,
		},
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
		got, err := shouldProcess(pm, c.file, whenExpr)
		if err != nil {
			t.Errorf("shouldProcess(pm, %v, %v) err: %s", c.file, c.when, err)
			continue
		}
		if got != c.want {
			t.Errorf("shouldProcess(pm, %v, %v) got: %t, want: %t",
				c.file, c.when, got, c.want)
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
	file := testhelpers.NewFileInfo("flow.json", time.Now())
	wantLogEntries := []testhelpers.Entry{
		{Level: testhelpers.Info,
			Msg: "Processing experiment: flow.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: flow.json"},
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

	quit := quitter.New()
	defer quit.Quit()
	l := testhelpers.NewLogger()
	go l.Run(quit)
	htmlCmds := make(chan cmd.Cmd, 100)
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
	if err := Process(file, cfg, l, pm); err != nil {
		t.Fatalf("Process: err: %v", err)
	}

	timeoutC := time.NewTimer(5 * time.Second).C
	tickerC := time.NewTicker(400 * time.Millisecond).C
	quitSelect := false
	for !quitSelect {
		select {
		case <-tickerC:
			if reflect.DeepEqual(l.GetEntries(), wantLogEntries) {
				quitSelect = true
			}
		case <-timeoutC:
			t.Errorf("l.GetEntries() got: %v, want: %v",
				l.GetEntries(), wantLogEntries)
			quitSelect = true
		}
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

func TestProcess_multiProcesses(t *testing.T) {
	var singleCPUTime int64
	maxNumProcesses := runtime.NumCPU()
	if maxNumProcesses < 2 {
		t.Skip("This test isn't implemented on single cpu systems.")
	}
	if testing.Short() {
		t.Skip("This test is skipped in short mode.")
	}

	t.Logf("Testing with %d processes.", maxNumProcesses)
	for numProcesses := 1; numProcesses <= maxNumProcesses; numProcesses++ {
		cfgDir := testhelpers.BuildConfigDirs(t)
		defer os.RemoveAll(cfgDir)
		cfg := &config.Config{
			ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
			WWWDir:            filepath.Join(cfgDir, "www"),
			BuildDir:          filepath.Join(cfgDir, "build"),
			MaxNumProcesses:   numProcesses,
			MaxNumReportRules: 100,
		}
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", "flow_big.yaml"),
			cfg.ExperimentsDir,
		)
		file := testhelpers.NewFileInfo("flow_big.yaml", time.Now())
		quit := quitter.New()
		defer quit.Quit()
		l := testhelpers.NewLogger()
		go l.Run(quit)
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
		start := time.Now()
		if err := Process(file, cfg, l, pm); err != nil {
			t.Fatalf("Process: err: %v", err)
		}
		elapsed := time.Since(start).Nanoseconds()
		if numProcesses == 1 {
			singleCPUTime = elapsed
		} else {
			if elapsed >= singleCPUTime {
				t.Errorf("Process was slower with %d processes than with 1",
					numProcesses)
			}
		}
	}
}

func TestProcess_errors(t *testing.T) {
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
		filepath.Join("fixtures", "flow_div_zero.yaml"),
		cfg.ExperimentsDir,
	)
	file := testhelpers.NewFileInfo("flow_div_zero.yaml", time.Now())
	wantLogEntries := []testhelpers.Entry{
		{Level: testhelpers.Info,
			Msg: "Processing experiment: flow_div_zero.yaml",
		},
		{Level: testhelpers.Error,
			Msg: "Failed processing experiment: flow_div_zero.yaml - invalid expression: numMatches / 0 (divide by zero)",
		},
	}
	wantPMExperiments := []*progress.Experiment{
		&progress.Experiment{
			Title:              "What would indicate good flow?",
			Tags:               []string{"test", "fred / ned"},
			Stamp:              time.Now(),
			ExperimentFilename: "flow_div_zero.yaml",
			Msg:                "Couldn't assess rules: invalid expression: numMatches / 0 (divide by zero)",
			Status:             progress.Failure,
		},
	}

	quit := quitter.New()
	defer quit.Quit()
	l := testhelpers.NewLogger()
	go l.Run(quit)
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
	wantErr := "Couldn't assess rules: invalid expression: numMatches / 0 (divide by zero)"
	if err := Process(file, cfg, l, pm); err.Error() != wantErr {
		t.Fatalf("Process: got err: %v, wantErr: %v", err, wantErr)
	}

	timeoutC := time.NewTimer(5 * time.Second).C
	tickerC := time.NewTicker(400 * time.Millisecond).C
	quitSelect := false
	for !quitSelect {
		select {
		case <-tickerC:
			if reflect.DeepEqual(l.GetEntries(), wantLogEntries) {
				quitSelect = true
			}
		case <-timeoutC:
			t.Errorf("l.GetEntries() got: %v, want: %v",
				l.GetEntries(), wantLogEntries)
			quitSelect = true
		}
	}

	err = progresstest.CheckExperimentsMatch(
		pm.GetExperiments(),
		wantPMExperiments,
	)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}

	/*
	 *	Check html commands received == {Progress,Progress,Progress,...}
	 */
	htmlCmdsReceived := cmdMonitor.GetCmdsReceived()
	numCmds := len(htmlCmdsReceived)
	if numCmds < 2 {
		t.Errorf("GetCmdsRecevied() received less than 2 commands")
	}
	for _, c := range htmlCmdsReceived {
		if c != cmd.Progress {
			t.Errorf(
				"GetCmdsRecevied() commands not all equal to Progress, found: %s",
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

func TestMakeDataset_err(t *testing.T) {
	e := &experimentFile{
		Dataset:    "sql",
		FieldNames: []string{},
		Sql: &sqlDesc{
			DriverName:     "mysql",
			DataSourceName: "invalid:invalid@tcp(127.0.0.1:9999)/master",
			Query:          "select * from invalid",
		},
	}
	ds, err := makeDataset(e)
	if err != nil {
		t.Fatalf("makeDataset(%v) err: %v", e, err)
	}
	wantErrRegexp :=
		regexp.MustCompile("^dial tcp 127.0.0.1:9999.*?connection.*?refused.*$")
	_, err = ds.Open()
	if !wantErrRegexp.MatchString(err.Error()) {
		t.Fatalf("ds.Open() gotErr: %v, wantErr: %v", err, wantErrRegexp)
	}
}

/*************************
       Benchmarks
*************************/
func BenchmarkProgress(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		cfgDir := testhelpers.BuildConfigDirs(b)
		defer os.RemoveAll(cfgDir)
		cfg := &config.Config{
			ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
			WWWDir:            filepath.Join(cfgDir, "www"),
			BuildDir:          filepath.Join(cfgDir, "build"),
			MaxNumProcesses:   4,
			MaxNumReportRules: 100,
		}
		testhelpers.CopyFile(
			b,
			filepath.Join("fixtures", "flow_big.yaml"),
			cfg.ExperimentsDir,
		)
		file := testhelpers.NewFileInfo("flow_big.yaml", time.Now())
		quit := quitter.New()
		defer quit.Quit()
		l := testhelpers.NewLogger()
		go l.Run(quit)
		htmlCmds := make(chan cmd.Cmd)
		defer close(htmlCmds)
		cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
		go cmdMonitor.Run()
		pm, err := progress.NewMonitor(
			filepath.Join(cfg.BuildDir, "progress"),
			htmlCmds,
		)
		if err != nil {
			b.Fatalf("progress.NewMonitor: err: %v", err)
		}
		b.StartTimer()
		if err := Process(file, cfg, l, pm); err != nil {
			b.Fatalf("Process: err: %v", err)
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
	if !areStringArraysEqual(e1.RuleFieldNames, e2.RuleFieldNames) {
		return errors.New("RuleFieldNames don't match")
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
