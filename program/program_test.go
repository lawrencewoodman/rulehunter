package program

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/progresstest"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

func TestProcessFile(t *testing.T) {
	wantEntries := []testhelpers.Entry{
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 3: did not find expected key"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.json"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.yaml"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.yaml"},
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: debt.jso, invalid extension: .jso"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt2.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt2.json"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt_when_nothasrun.yaml"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt_when_nothasrun.yaml"},
		{Level: testhelpers.Error,
			Msg: "Error processing experiment: debt_invalid_when.yaml, invalid expression: never (variable doesn't exist: never)"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt_invalid_goal.yaml"},
		{Level: testhelpers.Error,
			Msg: "Error processing experiment: debt_invalid_goal.yaml, Couldn't assess rules: invalid expression: dummy > 0 (variable doesn't exist: dummy)"},
	}
	wantPMExperiments := []*progress.Experiment{
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
			Title:    "What is most likely to indicate success",
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
			Filename: "debt_when_hasrun.yaml",
			Tags:     []string{},
			Status: &progress.Status{
				Stamp:   time.Now(),
				Msg:     "Waiting to be processed",
				Percent: 0.0,
				State:   progress.Waiting,
			},
		},
		&progress.Experiment{
			Title:    "What is most likely to indicate success",
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
				Msg:     "Error loading experiment: yaml: line 3: did not find expected key",
				Percent: 0.0,
				State:   progress.Error,
			},
		},
	}

	cfgDir := testhelpers.BuildConfigDirs(t, true)
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
		filepath.Join("fixtures", "0debt_broken.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.jso"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt2.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_when_hasrun.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_when_nothasrun.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_invalid_when.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_invalid_goal.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
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

	p := New(cfg, pm, l, quit)

	for _, f := range files {
		if err := p.ProcessFile(f); err != nil {
			t.Fatalf("ProcessFile(%s): %s", f.Name(), err)
		}
		time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	}

	if !reflect.DeepEqual(l.GetEntries(), wantEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v", l.GetEntries(), wantEntries)
	}

	got := pm.GetExperiments()
	err = progresstest.CheckExperimentsMatch(got, wantPMExperiments)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestStart(t *testing.T) {
	wantEntries := []testhelpers.Entry{
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 3: did not find expected key"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.json"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.yaml"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.yaml"},
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: debt.jso, invalid extension: .jso"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt2.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt2.json"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt_when_nothasrun.yaml"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt_when_nothasrun.yaml"},
		{Level: testhelpers.Error,
			Msg: "Error processing experiment: debt_invalid_when.yaml, invalid expression: never (variable doesn't exist: never)"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt_invalid_goal.yaml"},
		{Level: testhelpers.Error,
			Msg: "Error processing experiment: debt_invalid_goal.yaml, Couldn't assess rules: invalid expression: dummy > 0 (variable doesn't exist: dummy)"},
	}
	wantPMExperiments := []*progress.Experiment{
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
			Title:    "What is most likely to indicate success",
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
			Filename: "debt_when_hasrun.yaml",
			Tags:     []string{},
			Status: &progress.Status{
				Stamp:   time.Now(),
				Msg:     "Waiting to be processed",
				Percent: 0.0,
				State:   progress.Waiting,
			},
		},
		&progress.Experiment{
			Title:    "What is most likely to indicate success",
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
				Msg:     "Error loading experiment: yaml: line 3: did not find expected key",
				Percent: 0.0,
				State:   progress.Error,
			},
		},
	}

	wantFiles := []string{
		"debt.json",
		"debt.yaml",
		"debt2.json",
		"debt_when_nothasrun.yaml",
	}

	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	cfg := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumRecords:     100,
		MaxNumProcesses:   4,
		MaxNumReportRules: 100,
	}

	l := testhelpers.NewLogger()
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
	defer p.Stop(svc)

	if !testing.Short() {
		time.Sleep(4 * time.Second)
	}

	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "0debt_broken.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.jso"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt2.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_when_hasrun.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_when_nothasrun.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_invalid_when.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_invalid_goal.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)

	files := []string{}
	timeoutC := time.NewTimer(10 * time.Second).C
	tickerC := time.NewTicker(400 * time.Millisecond).C
	for {
		select {
		case <-tickerC:
			files = testhelpers.GetFilesInDir(
				t,
				filepath.Join(cfgDir, "build", "reports"),
			)
			if reflect.DeepEqual(files, wantFiles) {
				return
			}
		case <-timeoutC:
			t.Errorf("didn't generate correct files within time period, got: %v, want: %v",
				files, wantFiles)
			return
		}
	}

	if !reflect.DeepEqual(l.GetEntries(), wantEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v", l.GetEntries(), wantEntries)
	}

	got := pm.GetExperiments()
	err = progresstest.CheckExperimentsMatch(got, wantPMExperiments)
	if err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}