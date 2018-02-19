package experiment

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rhkit/aggregator"
	"github.com/vlifesystems/rhkit/assessment"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/rule"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal"
	"github.com/vlifesystems/rulehunter/internal/progresstest"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

func TestLoad(t *testing.T) {
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	funcs := map[string]dexpr.CallFun{}
	cases := []struct {
		cfg  *config.Config
		file fileinfo.FileInfo
		want *Experiment
	}{
		{cfg: &config.Config{
			MaxNumRecords: -1,
			BuildDir:      filepath.Join(tmpDir, "build"),
		},
			file: testhelpers.NewFileInfo(
				filepath.Join("fixtures", "flow.json"),
				time.Now(),
			),
			want: &Experiment{
				Title: "What would indicate good flow?",
				TrainDataset: dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
				RuleGeneration: ruleGeneration{
					fields:     []string{"group", "district", "height"},
					arithmetic: true,
				},
				Aggregators: []aggregator.Spec{
					aggregator.MustNew("numMatches", "count", "true()"),
					aggregator.MustNew(
						"percentMatches",
						"calc",
						"roundto(100.0 * numMatches / numRecords, 2)",
					),
					aggregator.MustNew("goodFlowMcc", "mcc", "flow > 60"),
					aggregator.MustNew("goalsScore", "goalsscore"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []assessment.SortOrder{
					assessment.SortOrder{"goodFlowMcc", assessment.DESCENDING},
					assessment.SortOrder{"numMatches", assessment.DESCENDING},
				},
				Tags:     []string{"test", "fred / ned"},
				Category: "testing",
				When:     dexpr.MustNew("!hasRun", funcs),
				Rules: []rule.Rule{
					mustNewDynamicRule("flow > 20"),
					mustNewDynamicRule("flow < 60"),
					mustNewDynamicRule("height > 67"),
					mustNewDynamicRule("height >= 129"),
					mustNewDynamicRule("group == \"a\""),
					mustNewDynamicRule("flow <= 9.42"),
					mustNewDynamicRule("district != \"northcal\" && group == \"b\""),
				},
			},
		},
		{cfg: &config.Config{
			MaxNumRecords: 4,
			BuildDir:      filepath.Join(tmpDir, "build"),
		},
			file: testhelpers.NewFileInfo(
				filepath.Join("fixtures", "flow.json"),
				time.Now(),
			),
			want: &Experiment{
				Title: "What would indicate good flow?",
				TrainDataset: dtruncate.New(
					dcsv.New(
						filepath.Join("fixtures", "flow.csv"),
						true,
						rune(','),
						[]string{"group", "district", "height", "flow"},
					),
					4,
				),
				RuleGeneration: ruleGeneration{
					fields:     []string{"group", "district", "height"},
					arithmetic: true,
				},
				Aggregators: []aggregator.Spec{
					aggregator.MustNew("numMatches", "count", "true()"),
					aggregator.MustNew(
						"percentMatches",
						"calc",
						"roundto(100.0 * numMatches / numRecords, 2)",
					),
					aggregator.MustNew("goodFlowMcc", "mcc", "flow > 60"),
					aggregator.MustNew("goalsScore", "goalsscore"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []assessment.SortOrder{
					assessment.SortOrder{"goodFlowMcc", assessment.DESCENDING},
					assessment.SortOrder{"numMatches", assessment.DESCENDING},
				},
				Tags:     []string{"test", "fred / ned"},
				Category: "testing",
				When:     dexpr.MustNew("!hasRun", funcs),
				Rules: []rule.Rule{
					mustNewDynamicRule("flow > 20"),
					mustNewDynamicRule("flow < 60"),
					mustNewDynamicRule("height > 67"),
					mustNewDynamicRule("height >= 129"),
					mustNewDynamicRule("group == \"a\""),
					mustNewDynamicRule("flow <= 9.42"),
					mustNewDynamicRule("district != \"northcal\" && group == \"b\""),
				},
			},
		},
		{cfg: &config.Config{
			MaxNumRecords: -1,
			BuildDir:      filepath.Join(tmpDir, "build"),
		},
			file: testhelpers.NewFileInfo(
				filepath.Join("fixtures", "flow_no_traindataset.json"),
				time.Now(),
			),
			want: &Experiment{
				Title: "What would indicate good flow?",
				TestDataset: dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
				RuleGeneration: ruleGeneration{
					fields:     []string{"group", "district", "height"},
					arithmetic: true,
				},
				Aggregators: []aggregator.Spec{
					aggregator.MustNew("numMatches", "count", "true()"),
					aggregator.MustNew(
						"percentMatches",
						"calc",
						"roundto(100.0 * numMatches / numRecords, 2)",
					),
					aggregator.MustNew("goodFlowMcc", "mcc", "flow > 60"),
					aggregator.MustNew("goalsScore", "goalsscore"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []assessment.SortOrder{
					assessment.SortOrder{"goodFlowMcc", assessment.DESCENDING},
					assessment.SortOrder{"numMatches", assessment.DESCENDING},
				},
				Tags:     []string{"test", "fred / ned"},
				Category: "testing",
				When:     dexpr.MustNew("!hasRun", funcs),
				Rules: []rule.Rule{
					mustNewDynamicRule("height > 67"),
					mustNewDynamicRule("height >= 129"),
					mustNewDynamicRule("group == \"a\""),
					mustNewDynamicRule("flow <= 9.42"),
					mustNewDynamicRule("district != \"northcal\" && group == \"b\""),
				},
			},
		},
		{cfg: &config.Config{
			MaxNumRecords: -1,
			BuildDir:      filepath.Join(tmpDir, "build"),
		},
			file: testhelpers.NewFileInfo(
				filepath.Join("fixtures", "debt.json"),
				time.Now(),
			),
			want: &Experiment{
				Title: "What would predict people being helped to be debt free?",
				TrainDataset: dcsv.New(
					filepath.Join("fixtures", "debt.csv"),
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
				RuleGeneration: ruleGeneration{
					fields: []string{
						"name",
						"balance",
						"numCards",
						"martialStatus",
						"tertiaryEducated",
					},
					arithmetic: false,
				},
				Aggregators: []aggregator.Spec{
					aggregator.MustNew("numMatches", "count", "true()"),
					aggregator.MustNew(
						"percentMatches",
						"calc",
						"roundto(100.0 * numMatches / numRecords, 2)",
					),
					aggregator.MustNew("helpedMcc", "mcc", "success"),
					aggregator.MustNew("goalsScore", "goalsscore"),
				},
				Goals: []*goal.Goal{goal.MustNew("helpedMcc > 0")},
				SortOrder: []assessment.SortOrder{
					assessment.SortOrder{"helpedMcc", assessment.DESCENDING},
					assessment.SortOrder{"numMatches", assessment.DESCENDING},
				},
				Tags: []string{"debt"},
				When: dexpr.MustNew(
					"!hasRunToday || sinceLastRunHours > 2",
					funcs,
				),
			},
		},
		{cfg: &config.Config{
			MaxNumRecords: -1,
			BuildDir:      filepath.Join(tmpDir, "build"),
		},
			file: testhelpers.NewFileInfo(
				filepath.Join("fixtures", "flow.yaml"),
				time.Now(),
			),
			want: &Experiment{
				Title: "What would indicate good flow?",
				TrainDataset: dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"group", "district", "height", "flow"},
				),
				RuleGeneration: ruleGeneration{
					fields:     []string{"group", "district", "height"},
					arithmetic: false,
				},
				Aggregators: []aggregator.Spec{
					aggregator.MustNew("numMatches", "count", "true()"),
					aggregator.MustNew(
						"percentMatches",
						"calc",
						"roundto(100.0 * numMatches / numRecords, 2)",
					),
					aggregator.MustNew("goodFlowMcc", "mcc", "flow > 60"),
					aggregator.MustNew("goalsScore", "goalsscore"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowMcc > 0")},
				SortOrder: []assessment.SortOrder{
					assessment.SortOrder{"goodFlowMcc", assessment.DESCENDING},
					assessment.SortOrder{"numMatches", assessment.DESCENDING},
				},
				Tags: []string{"test", "fred / ned"},
				When: dexpr.MustNew("!hasRun", funcs),
			},
		},
	}
	for _, c := range cases {
		gotExperiment, err := Load(c.cfg, c.file)
		if err != nil {
			t.Errorf("load(%s) err: %s", c.file, err)
			continue
		}
		if err := checkExperimentMatch(gotExperiment, c.want); err != nil {
			t.Errorf("load(%s) experiments don't match: %s\n"+
				"gotExperiment: %s, wantExperiment: %s",
				c.file, err, gotExperiment, c.want)
		}
	}
}

func TestLoad_error(t *testing.T) {
	cases := []struct {
		file    fileinfo.FileInfo
		wantErr error
	}{
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_title.json"),
			time.Now(),
		),
			errors.New("Experiment field missing: title")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_traindataset_or_testdataset.json"),
			time.Now(),
		),
			errors.New("Experiment field missing either: trainDataset or testDataset")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_csv_sql.json"),
			time.Now(),
		),
			errors.New("Experiment field: trainDataset, has no csv or sql field")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_csv_filename.json"),
			time.Now(),
		),
			errors.New("Experiment field missing: trainDataset > csv > filename")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_csv_separator.json"),
			time.Now(),
		),
			errors.New("Experiment field missing: trainDataset > csv > separator")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_csv_and_sql.yaml"),
			time.Now(),
		),
			errors.New(
				"Experiment field: trainDataset, can't specify csv and sql source",
			),
		},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_sql_drivername.json"),
			time.Now(),
		),
			errors.New("Experiment field missing: trainDataset > sql > driverName")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_invalid_sql_drivername.json"),
			time.Now(),
		),
			errors.New("Experiment field: trainDataset > sql, has invalid driverName: bob")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_sql_datasourcename.json"),
			time.Now(),
		),
			errors.New("Experiment field missing: trainDataset > sql > dataSourceName")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_no_sql_query.json"),
			time.Now(),
		),
			errors.New("Experiment field missing: trainDataset > sql > query")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_invalid_when.json"),
			time.Now(),
		),
			InvalidWhenExprError("has(twolegs")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_invalid.json"),
			time.Now(),
		),
			errors.New("invalid character '\\n' in string literal")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow.bob"),
			time.Now(),
		),
			InvalidExtError(".bob")},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_nonexistant.json"),
			time.Now(),
		),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "flow_nonexistant.json"),
				syscall.ENOENT,
			},
		},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_nonexistant.yaml"),
			time.Now(),
		),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "flow_nonexistant.yaml"),
				syscall.ENOENT,
			},
		},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_invalid.yaml"),
			time.Now(),
		),
			errors.New("yaml: line 3: did not find expected key"),
		},
		{testhelpers.NewFileInfo(
			filepath.Join("fixtures", "flow_invalid_rules.json"),
			time.Now(),
		),
			fmt.Errorf("rules: %s", rule.InvalidExprError{Expr: "flow < <= 9.42"}),
		},
	}
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cfg := &config.Config{
		MaxNumRecords: -1,
		BuildDir:      filepath.Join(tmpDir, "build"),
	}
	for _, c := range cases {
		_, err := Load(cfg, c.file)
		if err == nil {
			t.Errorf("load(%s) no error, wantErr:%s", c.file, c.wantErr)
			continue
		}
		if err.Error() != c.wantErr.Error() {
			t.Errorf("load(%s) gotErr: %s, wantErr:%s", c.file, err, c.wantErr)
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
	funcs := map[string]dexpr.CallFun{}
	cases := []struct {
		e          *Experiment
		isFinished bool
		stamp      time.Time
		want       bool
	}{
		{e: &Experiment{
			File: testhelpers.NewFileInfo("bank-divorced.json", time.Now()),
			When: dexpr.MustNew("!hasRun", funcs),
		},
			isFinished: true,
			stamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00"),
			want: true,
		},
		{e: &Experiment{
			File: testhelpers.NewFileInfo("bank-divorced.json",
				testhelpers.MustParse(time.RFC3339Nano,
					"2016-05-04T14:53:00.570347516+01:00")),
			When: dexpr.MustNew("!hasRun", funcs),
		},
			isFinished: true,
			stamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00"),
			want: false,
		},
		{e: &Experiment{
			File: testhelpers.NewFileInfo("bank-tiny.json", time.Now()),
			When: dexpr.MustNew("!hasRun", funcs),
		},
			isFinished: true,
			stamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00"),
			want: true,
		},
		{e: &Experiment{
			File: testhelpers.NewFileInfo("bank-tiny.json",
				testhelpers.MustParse(time.RFC3339Nano,
					"2016-05-05T09:37:58.220312223+01:00")),
			When: dexpr.MustNew("!hasRun", funcs),
		},
			isFinished: true,
			stamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00"),
			want: false,
		},
		{e: &Experiment{
			File: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			When: dexpr.MustNew("!hasRun", funcs),
		},
			isFinished: false,
			stamp:      time.Now(),
			want:       true,
		},
		{e: &Experiment{
			File: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			When: dexpr.MustNew("!hasRun", funcs),
		},
			isFinished: false,
			stamp:      time.Now(),
			want:       true,
		},
	}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("..", "progress", "fixtures", "progress.json"),
		tmpDir,
	)

	for i, c := range cases {
		got, err := c.e.ShouldProcess(c.isFinished, c.stamp)
		if err != nil {
			t.Errorf("(%d) shouldProcess: %s", i, err)
			continue
		}
		if got != c.want {
			t.Errorf("(%d) shouldProcess, got: %t, want: %t", i, got, c.want)
		}
	}
}

func TestProcess(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "flow.json"),
		cfg.ExperimentsDir,
	)
	file := testhelpers.NewFileInfo("flow.json", time.Now())
	wantPMExperiments := []*progress.Experiment{
		&progress.Experiment{
			Title:    "What would indicate good flow?",
			Filename: "flow.json",
			Tags:     []string{"test", "fred / ned"},
			Category: "testing",
			Status: &progress.Status{
				Stamp:   time.Now(),
				Msg:     "Train > Assessing rules 5/5",
				Percent: 100,
				State:   progress.Processing,
			},
		},
	}

	quit := quitter.New()
	defer quit.Quit()
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
	e, err := Load(cfg, file)
	if err != nil {
		t.Fatalf("Load: %s", err)
	}
	err = pm.AddExperiment(file.Name(), e.Title, e.Tags, e.Category)
	if err != nil {
		t.Fatalf("AddExperiment: %s", err)
	}
	if err := e.Process(cfg, pm); err != nil {
		t.Fatalf("Process: %s", err)
	}

	err = progresstest.CheckExperimentsMatch(
		pm.GetExperiments(),
		wantPMExperiments,
		false,
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

func TestProcess_supplied_rules(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "flow.json"),
		cfg.ExperimentsDir,
	)
	file := testhelpers.NewFileInfo("flow.json", time.Now())

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
		t.Fatalf("progress.NewMonitor: %s", err)
	}
	e, err := Load(cfg, file)
	if err != nil {
		t.Fatalf("Load: %s", err)
	}
	err = pm.AddExperiment(file.Name(), e.Title, e.Tags, e.Category)
	if err != nil {
		t.Fatalf("AddExperiment: %s", err)
	}

	if err := e.Process(cfg, pm); err != nil {
		t.Fatalf("Process: %s", err)
	}

	flowBuildFullFilename := filepath.Join(
		cfgDir,
		"build",
		"reports",
		internal.MakeBuildFilename("train", e.Category, e.Title),
	)
	b, err := ioutil.ReadFile(flowBuildFullFilename)
	if err != nil {
		t.Fatalf("ReadFile: %s", err)
	}
	s := string(b)
	wantRules := []string{
		"height \\u003e 67",
		"flow \\u003e 20",
	}
	for _, wantRule := range wantRules {
		if !strings.Contains(s, wantRule) {
			t.Errorf("rule: %s, missing from: %s", wantRule, flowBuildFullFilename)
		}
	}

	// TODO: Test files generated
}

func TestProcess_multiProcesses(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip("This test isn't implemented on single cpu systems.")
	}
	if testing.Short() {
		t.Skip("This test is skipped in short mode.")
	}

	timeProcess := func(numProcesses int) (nanoseconds int64) {
		cfgDir := testhelpers.BuildConfigDirs(t, true)
		defer os.RemoveAll(cfgDir)
		cfg := &config.Config{
			ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
			WWWDir:          filepath.Join(cfgDir, "www"),
			BuildDir:        filepath.Join(cfgDir, "build"),
			MaxNumProcesses: numProcesses,
			MaxNumRecords:   5000,
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
			t.Fatalf("progress.NewMonitor: %s", err)
		}

		e, err := Load(cfg, file)
		if err != nil {
			t.Fatalf("Load: %s", err)
		}
		err = pm.AddExperiment(file.Name(), e.Title, e.Tags, e.Category)
		if err != nil {
			t.Fatalf("AddExperiment: %s", err)
		}

		start := time.Now()
		if err := e.Process(cfg, pm); err != nil {
			t.Fatalf("Process: %s", err)
		}
		return time.Since(start).Nanoseconds()
	}

	singleProcessTime := timeProcess(1)
	pass := false
	for attempts := 0; attempts < 5 && !pass; attempts++ {
		multiProcessTime := timeProcess(runtime.NumCPU())
		t.Logf("Tested with %d processes. Speed-up over single process: %0.2fx (attempt: %d)",
			runtime.NumCPU(),
			float64(singleProcessTime)/float64(multiProcessTime),
			attempts+1,
		)
		if multiProcessTime < singleProcessTime {
			pass = true
		} else {
			t.Logf("Process was slower with %d processes than with 1 (attempt: %d)",
				runtime.NumCPU(), attempts+1)
		}
	}
	if !pass {
		t.Errorf("Process was slower with %d processes than with 1 (after 5 attempts)",
			runtime.NumCPU())
	}
}

func TestProcess_errors(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	files := []string{"flow_div_zero_train.yaml",
		"flow_div_zero_test.yaml",
		"flow_invalid_aggregator.yaml",
	}

	for _, f := range files {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", f),
			cfg.ExperimentsDir,
		)
	}

	cases := []struct {
		file    testhelpers.FileInfo
		wantErr string
	}{
		{file: testhelpers.NewFileInfo("flow_div_zero_train.yaml", time.Now()),
			wantErr: "Couldn't assess rules: invalid expression: numMatches / 0 (divide by zero)",
		},
		{file: testhelpers.NewFileInfo("flow_div_zero_test.yaml", time.Now()),
			wantErr: "Couldn't assess rules: invalid rule: height / 0",
		},
		{file: testhelpers.NewFileInfo("flow_invalid_aggregator.yaml", time.Now()),
			wantErr: "Couldn't assess rules: invalid expression: bob > 60 (variable doesn't exist: bob)",
		},
	}
	wantPMExperiments := []*progress.Experiment{
		&progress.Experiment{
			Title:    "What would indicate good flow?",
			Filename: "flow_div_zero_train.yaml",
			Tags:     []string{"test", "fred / ned"},
			Category: "testing",
			Status: &progress.Status{
				Stamp: time.Now(),
				Msg:   "Train > Assessing rules 2/5",
				State: progress.Processing,
			},
		},
		&progress.Experiment{
			Title:    "What would indicate good flow?",
			Filename: "flow_div_zero_test.yaml",
			Tags:     []string{"test", "fred / ned"},
			Category: "testing",
			Status: &progress.Status{
				Stamp: time.Now(),
				Msg:   "Test > Assessing rules",
				State: progress.Processing,
			},
		},
		&progress.Experiment{
			Title:    "What would indicate good flow?",
			Filename: "flow_invalid_aggregator.yaml",
			Tags:     []string{"test", "fred / ned"},
			Category: "",
			Status: &progress.Status{
				Stamp: time.Now(),
				Msg:   "Train > Assessing rules 2/5",
				State: progress.Processing,
			},
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
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	for _, c := range cases {
		// Multiple tests for each because of channel problem that appeared
		// intermittently
		for i := 0; i < 100; i++ {
			e, err := Load(cfg, c.file)
			if err != nil {
				t.Fatalf("Load: %s", err)
			}
			err = pm.AddExperiment(c.file.Name(), e.Title, e.Tags, e.Category)
			if err != nil {
				t.Fatalf("AddExperiment: %s", err)
			}

			if err := e.Process(cfg, pm); err == nil || err.Error() != c.wantErr {
				t.Fatalf("Process: file: %s, gotErr: %s, wantErr: %s",
					c.file.Name(), err, c.wantErr)
			}
		}

	}
	err = progresstest.CheckExperimentsMatch(
		pm.GetExperiments(),
		wantPMExperiments,
		true,
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
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cases := []struct {
		desc           *datasetDesc
		dataSourceName string
		query          string
		fields         []string
		config         *config.Config
		want           ddataset.Dataset
	}{
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select * from flow",
			},
		},
			fields: []string{"grp", "district", "height", "flow"},
			config: &config.Config{
				MaxNumRecords: -1,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select * from flow",
			},
		},
			fields: []string{"grp", "district", "height", "flow"},
			config: &config.Config{
				MaxNumRecords: 1000,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select * from flow",
			},
		},
			fields: []string{"grp", "district", "height", "flow"},
			config: &config.Config{
				MaxNumRecords: 4,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dtruncate.New(
				dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"grp", "district", "height", "flow"},
				),
				4,
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select grp,district,flow from flow",
			},
		},
			fields: []string{"grp", "district", "flow"},
			config: &config.Config{
				MaxNumRecords: -1,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
	}
	for i, c := range cases {
		got, err := makeDataset("trainDataset", c.config, c.fields, c.desc)
		if err != nil {
			t.Errorf("(%d) makeDataset: %s", i, err)
		} else if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("(%d) checkDatasetsEqual: %s", i, err)
		}
	}
}

func TestMakeDataset_err(t *testing.T) {
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cases := []struct {
		experimentField   string
		fields            []string
		desc              *datasetDesc
		wantOpenErrRegexp *regexp.Regexp
	}{
		{experimentField: "trainDataset",
			fields: []string{},
			desc: &datasetDesc{
				SQL: &sqlDesc{
					DriverName:     "mysql",
					DataSourceName: "invalid:invalid@tcp(127.0.0.1:9999)/master",
					Query:          "select * from invalid",
				},
			},
			wantOpenErrRegexp: regexp.MustCompile("^dial tcp 127.0.0.1:9999.*?connection.*?refused.*$"),
		},
		{experimentField: "trainDataset",
			fields: []string{},
			desc: &datasetDesc{
				CSV: &csvDesc{
					Filename:  filepath.Join("fixtures", "nonexistant.csv"),
					HasHeader: false,
					Separator: ",",
				},
			},
			wantOpenErrRegexp: regexp.MustCompile(
				// Replace used because in Windows the backslash in the path is
				// altering the meaning of the regexp
				strings.Replace(
					fmt.Sprintf(
						"^%s$",
						&os.PathError{
							Op:   "open",
							Path: filepath.Join("fixtures", "nonexistant.csv"),
							Err:  syscall.ENOENT,
						},
					),
					"\\",
					"\\\\",
					-1,
				),
			),
		},
	}
	cfg := &config.Config{
		MaxNumRecords: -1,
		BuildDir:      filepath.Join(tmpDir, "build"),
	}
	for i, c := range cases {
		_, err := makeDataset(c.experimentField, cfg, c.fields, c.desc)
		if !c.wantOpenErrRegexp.MatchString(err.Error()) {
			t.Fatalf("(%d) makeDataset: %s", i, err)
		}
	}
}

/*************************
       Benchmarks
*************************/

func BenchmarkProcess_csv(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		cfgDir := testhelpers.BuildConfigDirs(b, true)
		defer os.RemoveAll(cfgDir)
		cfg := &config.Config{
			ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
			WWWDir:          filepath.Join(cfgDir, "www"),
			BuildDir:        filepath.Join(cfgDir, "build"),
			MaxNumProcesses: 4,
			MaxNumRecords:   10000,
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
		e, err := Load(cfg, file)
		if err != nil {
			b.Fatalf("Load: %s", err)
		}
		err = pm.AddExperiment(file.Name(), e.Title, e.Tags, e.Category)
		if err != nil {
			b.Fatalf("AddExperiment: %s", err)
		}
		b.StartTimer()
		if err := e.Process(cfg, pm); err != nil {
			b.Fatalf("Process: %s", err)
		}
	}
}

func BenchmarkProcess_sql(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		cfgDir := testhelpers.BuildConfigDirs(b, true)
		defer os.RemoveAll(cfgDir)
		cfg := &config.Config{
			ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
			WWWDir:          filepath.Join(cfgDir, "www"),
			BuildDir:        filepath.Join(cfgDir, "build"),
			MaxNumProcesses: 4,
			MaxNumRecords:   10000,
		}
		testhelpers.CopyFile(
			b,
			filepath.Join("fixtures", "flow_big_sql.yaml"),
			cfg.ExperimentsDir,
		)
		file := testhelpers.NewFileInfo("flow_big_sql.yaml", time.Now())
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
		e, err := Load(cfg, file)
		if err != nil {
			b.Fatalf("Load: %s", err)
		}
		err = pm.AddExperiment(file.Name(), e.Title, e.Tags, e.Category)
		if err != nil {
			b.Fatalf("AddExperiment: %s", err)
		}
		b.StartTimer()
		if err := e.Process(cfg, pm); err != nil {
			b.Fatalf("Process: %s", err)
		}
	}
}

/***********************
    Helper functions
************************/

func checkExperimentMatch(
	e1 *Experiment,
	e2 *Experiment,
) error {
	if e1.Title != e2.Title {
		return errors.New("Titles don't match")
	}
	if !areStringArraysEqual(e1.Tags, e2.Tags) {
		return errors.New("Tags don't match")
	}
	if e1.Category != e2.Category {
		return errors.New("Categories don't match")
	}
	if !areGenerationDescribersEqual(e1.RuleGeneration, e2.RuleGeneration) {
		return errors.New("RuleGeneration don't match")
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
	if e1.When.String() != e2.When.String() {
		return errors.New("Whens don't match")
	}
	if !areRulesEqual(e1.Rules, e2.Rules) {
		return errors.New("Rules don't match")
	}
	if err := checkDatasetsEqual(e1.TrainDataset, e2.TrainDataset); err != nil {
		return fmt.Errorf("trainDataset: %s", err)
	}
	if err := checkDatasetsEqual(e1.TestDataset, e2.TestDataset); err != nil {
		return fmt.Errorf("testDataset: %s", err)
	}
	return nil
}

func checkDatasetsEqual(ds1, ds2 ddataset.Dataset) error {
	if ds1 == nil && ds2 == nil {
		return nil
	}
	if (ds1 == nil && ds2 != nil) || (ds1 != nil && ds2 == nil) {
		return errors.New("one dataset is nil")
	}

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
			return errors.New("datasets don't finish at same point")
		}
		if !conn1Next {
			break
		}

		conn1Record := conn1.Read()
		conn2Record := conn2.Read()
		if !reflect.DeepEqual(conn1Record, conn2Record) {
			return fmt.Errorf("rows don't match %s != %s", conn1Record, conn2Record)
		}
	}
	if conn1.Err() != conn2.Err() {
		return errors.New("datasets final error doesn't match")
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
	a1 []aggregator.Spec,
	a2 []aggregator.Spec,
) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, a := range a1 {
		if reflect.TypeOf(a) != reflect.TypeOf(a2[i]) ||
			a.Name() != a2[i].Name() ||
			a.Arg() != a2[i].Arg() {
			return false
		}
	}
	return true
}

func areSortOrdersEqual(
	so1 []assessment.SortOrder,
	so2 []assessment.SortOrder,
) bool {
	if len(so1) != len(so2) {
		return false
	}
	for i, sf1 := range so1 {
		sf2 := so2[i]
		if sf1.Aggregator != sf2.Aggregator || sf1.Direction != sf2.Direction {
			return false
		}
	}
	return true
}

func areGenerationDescribersEqual(
	gd1 rule.GenerationDescriber,
	gd2 rule.GenerationDescriber,
) bool {
	if !areStringArraysEqual(gd1.Fields(), gd2.Fields()) {
		return false
	}
	return gd1.Arithmetic() == gd2.Arithmetic()
}

func areRulesEqual(
	rs1 []rule.Rule,
	rs2 []rule.Rule,
) bool {
	if len(rs1) != len(rs2) {
		return false
	}
	for i, r1 := range rs1 {
		if r1.String() != rs2[i].String() {
			return false
		}
	}
	return true
}

func mustNewDynamicRule(e string) rule.Rule {
	r, err := rule.NewDynamic(e)
	if err != nil {
		panic(err)
	}
	return r
}
