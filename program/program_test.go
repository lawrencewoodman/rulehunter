package program

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/internal"
	"github.com/vlifesystems/rulehunter/internal/progresstest"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

var wantPMExperiments = []*progress.Experiment{
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt_invalid_goal.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Couldn't assess rules: invalid expression: dummy > 0 (variable doesn't exist: dummy)",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt_invalid_when.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "invalid expression: never (variable doesn't exist: never)",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success (norun)",
		Filename: "debt_when_nothasrun.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success (2)",
		Filename: "debt2.json",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "",
		Filename: "debt.jso",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Error loading experiment: invalid extension: .jso",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt.yaml",
		Tags:     []string{"bank", "loan"},
		Category: "testing",
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt.json",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "",
		Filename: "0debt_broken.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Error loading experiment: yaml: line 4: did not find expected key",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
}

var wantValidNamePMExperiments = []*progress.Experiment{
	&progress.Experiment{
		Title:    "What is most likely to indicate success (norun)",
		Filename: "debt_when_nothasrun.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt_invalid_when.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "invalid expression: never (variable doesn't exist: never)",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt_invalid_goal.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Couldn't assess rules: invalid expression: dummy > 0 (variable doesn't exist: dummy)",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success (2)",
		Filename: "debt2.json",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt.yaml",
		Tags:     []string{"bank", "loan"},
		Category: "testing",
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "What is most likely to indicate success",
		Filename: "debt.json",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Finished processing successfully",
			Percent: 0.0,
			State:   progress.Success,
		},
	},
	&progress.Experiment{
		Title:    "",
		Filename: "0debt_broken.yaml",
		Tags:     []string{},
		Status: &progress.Status{
			Stamp:   time.Now(),
			Msg:     "Error loading experiment: yaml: line 4: did not find expected key",
			Percent: 0.0,
			State:   progress.Error,
		},
	},
}

// Sets the time for the experiments to Now()
func initExperimentsTime(experiments []*progress.Experiment) {
	for _, e := range experiments {
		e.Status.Stamp = time.Now()
	}
}

var wantLogEntries = []testhelpers.Entry{
	{Level: testhelpers.Error,
		Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 4: did not find expected key"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt.yaml, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt.yaml, mode: train"},
	{Level: testhelpers.Error,
		Msg: "Can't load experiment: debt.jso, invalid extension: .jso"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt2.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt2.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt_when_nothasrun.yaml, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt_when_nothasrun.yaml, mode: train"},
	{Level: testhelpers.Error,
		Msg: "Error processing experiment: debt_invalid_when.yaml, invalid expression: never (variable doesn't exist: never)"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt_invalid_goal.yaml, mode: train"},
	{Level: testhelpers.Error,
		Msg: "Error processing experiment: debt_invalid_goal.yaml, Couldn't assess rules: invalid expression: dummy > 0 (variable doesn't exist: dummy)"},
}

var wantValidNameLogEntries = []testhelpers.Entry{
	{Level: testhelpers.Error,
		Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 4: did not find expected key"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt.yaml, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt.yaml, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt2.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt2.json, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt_invalid_goal.yaml, mode: train"},
	{Level: testhelpers.Error,
		Msg: "Error processing experiment: debt_invalid_goal.yaml, Couldn't assess rules: invalid expression: dummy > 0 (variable doesn't exist: dummy)"},
	{Level: testhelpers.Error,
		Msg: "Error processing experiment: debt_invalid_when.yaml, invalid expression: never (variable doesn't exist: never)"},
	{Level: testhelpers.Info,
		Msg: "Processing experiment: debt_when_nothasrun.yaml, mode: train"},
	{Level: testhelpers.Info,
		Msg: "Successfully processed experiment: debt_when_nothasrun.yaml, mode: train"},
}

var wantReportFiles = []string{
	// "debt_when_nothasrun.yaml"
	internal.MakeBuildFilename(
		"train",
		"",
		"What is most likely to indicate success (norun)",
	),
	// "debt2.json"
	internal.MakeBuildFilename(
		"train",
		"",
		"What is most likely to indicate success (2)",
	),
	// "debt.yaml"
	internal.MakeBuildFilename(
		"train",
		"testing",
		"What is most likely to indicate success",
	),
	// "debt.json"
	internal.MakeBuildFilename(
		"train",
		"",
		"What is most likely to indicate success",
	),
}

var experimentFiles = []string{
	"0debt_broken.yaml",
	"debt.json",
	"debt.yaml",
	"debt.jso",
	"debt2.json",
	"debt_when_hasrun.yaml",
	"debt_when_nothasrun.yaml",
	"debt_invalid_when.yaml",
	"debt_invalid_goal.yaml",
}

func TestProcessFile(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	for _, f := range experimentFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", f),
			filepath.Join(cfgDir, "experiments"),
		)
	}
	files := []fileinfo.FileInfo{
		testhelpers.NewFileInfo("0debt_broken.yaml", time.Now()),
		testhelpers.NewFileInfo("debt.json", time.Now()),
		testhelpers.NewFileInfo("debt.yaml", time.Now()),
		testhelpers.NewFileInfo("debt.jso", time.Now()),
		testhelpers.NewFileInfo("debt2.json", time.Now()),
		testhelpers.NewFileInfo("debt_when_hasrun.yaml", time.Now()),
		testhelpers.NewFileInfo("debt_when_nothasrun.yaml", time.Now()),
		testhelpers.NewFileInfo("debt_invalid_when.yaml", time.Now()),
		testhelpers.NewFileInfo("debt_invalid_goal.yaml", time.Now()),
	}

	l := testhelpers.NewLogger()
	quit := quitter.New()
	defer quit.Quit()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	p := New(cfg, pm, l, quit)

	for _, f := range files {
		if err := p.ProcessFile(f, false); err != nil {
			t.Fatalf("ProcessFile(%s): %s", f.Name(), err)
		}
		time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	}

	gotReportFiles := testhelpers.GetFilesInDir(
		t,
		filepath.Join(cfgDir, "build", "reports"),
	)
	sort.Strings(gotReportFiles)
	sort.Strings(wantReportFiles)
	if !reflect.DeepEqual(gotReportFiles, wantReportFiles) {
		t.Errorf("GetFilesInDir - got: %v\n want: %v",
			gotReportFiles, wantReportFiles)
	}

	if !reflect.DeepEqual(l.GetEntries(), wantLogEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v",
			l.GetEntries(true), wantLogEntries)
	}

	got := pm.GetExperiments()
	initExperimentsTime(wantPMExperiments)
	err = progresstest.CheckExperimentsMatch(got, wantPMExperiments, false)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

// This test ensures that if the title changes for an experiment that
// has finished but is going to run again, changes in the progress log.
func TestProcessFile_title_change(t *testing.T) {
	var wantPMExperiments = []*progress.Experiment{
		&progress.Experiment{
			Title:    "What is most likely to indicate success",
			Filename: "debt.yaml",
			Tags:     []string{},
			Status: &progress.Status{
				Stamp:   time.Now(),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   progress.Success,
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

	l := testhelpers.NewLogger()
	quit := quitter.New()
	defer quit.Quit()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	p := New(cfg, pm, l, quit)

	file := testhelpers.NewFileInfo("debt.yaml", time.Now())
	experimentFiles := []string{
		filepath.Join("fixtures", "debt_when_nothasrun.yaml"),
		filepath.Join("fixtures", "debt_when_hasrun.yaml"),
	}

	for _, experimentFile := range experimentFiles {
		testhelpers.CopyFile(
			t,
			experimentFile,
			filepath.Join(cfgDir, "experiments"),
			"debt.yaml",
		)

		if err := p.ProcessFile(file, false); err != nil {
			t.Fatalf("ProcessFile: %s", err)
		}
		time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	}

	got := pm.GetExperiments()
	initExperimentsTime(wantPMExperiments)
	err = progresstest.CheckExperimentsMatch(got, wantPMExperiments, false)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestProcessFile_ignoreWhen(t *testing.T) {
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
		filepath.Join("fixtures", "debt_when_hasrun.yaml"),
		cfg.ExperimentsDir,
	)
	var wantReportFiles = []string{
		// "debt_when_hasrun.yaml"
		internal.MakeBuildFilename(
			"train",
			"",
			"What is most likely to indicate success",
		),
	}
	var wantLogEntries = []testhelpers.Entry{
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt_when_hasrun.yaml, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt_when_hasrun.yaml, mode: train"},
	}

	l := testhelpers.NewLogger()
	quit := quitter.New()
	defer quit.Quit()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	p := New(cfg, pm, l, quit)

	experimentFile :=
		testhelpers.NewFileInfo("debt_when_hasrun.yaml", time.Now())
	if err := p.ProcessFile(experimentFile, true); err != nil {
		t.Fatalf("ProcessFile(%s): %s", experimentFile.Name(), err)
	}
	time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */

	gotReportFiles := testhelpers.GetFilesInDir(
		t,
		filepath.Join(cfgDir, "build", "reports"),
	)
	sort.Strings(gotReportFiles)
	sort.Strings(wantReportFiles)
	if !reflect.DeepEqual(gotReportFiles, wantReportFiles) {
		t.Errorf("GetFilesInDir - got: %v\n want: %v",
			gotReportFiles, wantReportFiles)
	}

	if !reflect.DeepEqual(l.GetEntries(), wantLogEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v",
			l.GetEntries(true), wantLogEntries)
	}
}

func TestProcessFilename(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   10,
		MaxNumProcesses: 4,
	}
	for _, f := range experimentFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", f),
			cfg.ExperimentsDir,
		)
	}
	filenames := []string{
		"0debt_broken.yaml",
		"debt.json",
		"debt.yaml",
		"debt.jso",
		"debt2.json",
		"debt_when_hasrun.yaml",
		"debt_when_nothasrun.yaml",
		"debt_invalid_when.yaml",
		"debt_invalid_goal.yaml",
	}

	l := testhelpers.NewLogger()
	quit := quitter.New()
	defer quit.Quit()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	p := New(cfg, pm, l, quit)

	for _, f := range filenames {
		fullFilename := filepath.Join(cfg.ExperimentsDir, f)
		if err := p.ProcessFilename(fullFilename, false); err != nil {
			t.Fatalf("ProcessFilename(%s): %s", f, err)
		}
		time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	}

	gotReportFiles := testhelpers.GetFilesInDir(
		t,
		filepath.Join(cfgDir, "build", "reports"),
	)
	sort.Strings(gotReportFiles)
	sort.Strings(wantReportFiles)
	if !reflect.DeepEqual(gotReportFiles, wantReportFiles) {
		t.Errorf("GetFilesInDir - got: %v\n want: %v",
			gotReportFiles, wantReportFiles)
	}

	if !reflect.DeepEqual(l.GetEntries(), wantLogEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v",
			l.GetEntries(true), wantLogEntries)
	}

	got := pm.GetExperiments()
	initExperimentsTime(wantPMExperiments)
	err = progresstest.CheckExperimentsMatch(got, wantPMExperiments, false)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestProcessDir(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}
	for _, f := range experimentFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", f),
			filepath.Join(cfgDir, "experiments"),
		)
	}

	l := testhelpers.NewLogger()
	quit := quitter.New()
	defer quit.Quit()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	p := New(cfg, pm, l, quit)

	if err := p.ProcessDir(cfg.ExperimentsDir, false); err != nil {
		t.Fatalf("ProcessDir: %s", err)
	}

	gotReportFiles := testhelpers.GetFilesInDir(
		t,
		filepath.Join(cfgDir, "build", "reports"),
	)
	sort.Strings(gotReportFiles)
	sort.Strings(wantReportFiles)
	if !reflect.DeepEqual(gotReportFiles, wantReportFiles) {
		t.Errorf("GetFilesInDir - got: %v\n want: %v",
			gotReportFiles, wantReportFiles)
	}

	if !reflect.DeepEqual(l.GetEntries(), wantValidNameLogEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v",
			l.GetEntries(true), wantLogEntries)
	}

	got := pm.GetExperiments()
	initExperimentsTime(wantValidNamePMExperiments)
	err =
		progresstest.CheckExperimentsMatch(got, wantValidNamePMExperiments, false)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestStart(t *testing.T) {
	wantLogEntries := testhelpers.SortLogEntries(wantValidNameLogEntries)

	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:  filepath.Join(cfgDir, "experiments"),
		WWWDir:          filepath.Join(cfgDir, "www"),
		BuildDir:        filepath.Join(cfgDir, "build"),
		MaxNumRecords:   100,
		MaxNumProcesses: 4,
	}

	l := testhelpers.NewLogger()
	quit := quitter.New()
	defer quit.Quit()
	pm, err := progress.NewMonitor(
		filepath.Join(cfg.BuildDir, "progress"),
	)
	if err != nil {
		t.Fatalf("progress.NewMonitor: %s", err)
	}

	p := New(cfg, pm, l, quit)
	svcConfig := &service.Config{
		Name:        "rulehunter",
		DisplayName: "Rulehunter server",
		Description: "Rulehunter finds rules in data based on user defined goals.",
	}

	svc, err := service.New(p, svcConfig)
	if err != nil {
		t.Fatalf("service.New: %s", err)
	}

	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.csv"),
		filepath.Join(cfgDir, "datasets"),
	)
	p.Start(svc)
	defer func() {
		p.Stop(svc)
		for i := 0; i < 1000; i++ {
			if !p.Running() {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	if !testing.Short() {
		time.Sleep(4 * time.Second)
	}

	for _, f := range experimentFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", f),
			filepath.Join(cfgDir, "experiments"),
		)
	}

	files := []string{}
	gotLogEntries := []testhelpers.Entry{}
	gotPMExperiments := []*progress.Experiment{}

	timeoutC := time.NewTimer(20 * time.Second).C
	tickerC := time.NewTicker(400 * time.Millisecond).C
	sort.Strings(wantReportFiles)
	finishedWaiting := false
	for !finishedWaiting {
		select {
		case <-tickerC:
			files = testhelpers.GetFilesInDir(
				t,
				filepath.Join(cfgDir, "build", "reports"),
			)
			sort.Strings(files)
			if reflect.DeepEqual(files, wantReportFiles) {
				gotLogEntries = testhelpers.SortLogEntries(l.GetEntries(true))
				if reflect.DeepEqual(gotLogEntries, wantLogEntries) {
					gotPMExperiments = pm.GetExperiments()
					initExperimentsTime(wantValidNamePMExperiments)
					err = progresstest.CheckExperimentsMatch(
						gotPMExperiments,
						wantValidNamePMExperiments,
						false,
					)
					if err == nil {
						finishedWaiting = true
						break
					}
				}
			}
		case <-timeoutC:
			if !reflect.DeepEqual(files, wantReportFiles) {
				t.Errorf(
					"didn't generate correct files within time period, got: %v\n want: %v",
					files,
					wantReportFiles,
				)
				if !reflect.DeepEqual(gotLogEntries, wantLogEntries) {
					t.Errorf("GetEntries() got: %v\n want: %v",
						gotLogEntries, wantLogEntries)
				}
				err := progresstest.CheckExperimentsMatch(
					gotPMExperiments,
					wantValidNamePMExperiments,
					false,
				)
				if err != nil {
					t.Errorf("checkExperimentsMatch: %s", err)
				}
			}
			finishedWaiting = true
			break
		}
	}
}
