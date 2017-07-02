package progress

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestAddExperiment_experiment_exists(t *testing.T) {
	expected := []*Experiment{
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-married.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Tags:               []string{"test", "bank"},
			Stamp:              mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}

	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := NewMonitor(tmpDir, htmlCmds)
	if err != nil {
		t.Errorf("NewMonitor() err: %s", err)
	}
	if err := pm.AddExperiment("bank-divorced.json"); err != nil {
		t.Fatalf("AddExperiment() err: %s", err)
	}
	if err := pm.AddExperiment("bank-full-divorced.json"); err != nil {
		t.Fatalf("AddExperiment() err: %s", err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := pm.AddExperiment("bank-married.json"); err != nil {
		t.Fatalf("AddExperiment() err: %s", err)
	}
	epr, err := NewExperimentReporter(pm, "bank-married.json")
	if err != nil {
		t.Fatalf("NewExperimentReporter: %s", err)
	}
	epr.ReportProgress("something is happening", 0)
	if err := pm.AddExperiment("bank-married.json"); err != nil {
		t.Fatalf("AddExperiment() err: %s", err)
	}
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, expected); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestUpdateDetails(t *testing.T) {
	wantExperiments := []*Experiment{
		&Experiment{
			Title:              "this is my title",
			Tags:               []string{"big", "little"},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Tags:               []string{"test", "bank"},
			Stamp:              mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}

	wantHtmlCmdsReceived := []cmd.Cmd{cmd.Progress, cmd.Progress}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := NewMonitor(tmpDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	epr, err := NewExperimentReporter(pm, "bank-full-divorced.json")
	if err != nil {
		t.Fatalf("NewExperimentReporter(pm, \"bank-full-divorced.json\") err: %s", err)
	}
	err = epr.UpdateDetails("this is my title", []string{"big", "little"})
	if err != nil {
		t.Fatalf("UpdateDetails: %s", err)
	}

	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, wantExperiments); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
	time.Sleep(1 * time.Second)
	close(htmlCmds)
	htmlCmdsReceived := cmdMonitor.GetCmdsReceived()
	if !reflect.DeepEqual(htmlCmdsReceived, wantHtmlCmdsReceived) {
		t.Errorf("GetCmdsRecevied() received commands - got: %s, want: %s",
			htmlCmdsReceived, wantHtmlCmdsReceived)
	}
}

func TestReportSuccess(t *testing.T) {
	wantExperiments := []*Experiment{
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Tags:               []string{"test", "bank"},
			Stamp:              mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}
	cases := []struct {
		run                  int
		wantHtmlCmdsReceived []cmd.Cmd
	}{
		{run: 0,
			wantHtmlCmdsReceived: []cmd.Cmd{cmd.Progress, cmd.Progress, cmd.Reports},
		},
		{run: 1,
			wantHtmlCmdsReceived: []cmd.Cmd{},
		},
	}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	for _, c := range cases {
		htmlCmds := make(chan cmd.Cmd)
		cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
		go cmdMonitor.Run()
		pm, err := NewMonitor(tmpDir, htmlCmds)
		if err != nil {
			t.Fatalf("NewMonitor() err: %v", err)
		}
		if c.run == 0 {
			epr, err := NewExperimentReporter(pm, "bank-full-divorced.json")
			if err != nil {
				t.Fatalf("NewExperimentReporter(pm, \"bank-full-divorced.json\") err: %s", err)
			}
			epr.ReportSuccess()
		}

		got := pm.GetExperiments()
		if err := checkExperimentsMatch(got, wantExperiments); err != nil {
			t.Errorf("checkExperimentsMatch() err: %s", err)
		}
		time.Sleep(1 * time.Second)
		close(htmlCmds)
		htmlCmdsReceived := cmdMonitor.GetCmdsReceived()
		if !reflect.DeepEqual(htmlCmdsReceived, c.wantHtmlCmdsReceived) {
			t.Errorf("GetCmdsRecevied() received commands - got: %s, want: %s",
				htmlCmdsReceived, c.wantHtmlCmdsReceived)
		}
	}
}

func TestReportInfo(t *testing.T) {
	wantExperimentsMemory := []*Experiment{
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Assessing rules",
			Percent:            float64(0.24),
			Status:             Processing,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Tags:               []string{"test", "bank"},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Describing dataset",
			Status:             Processing,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}
	wantExperimentsFile := []*Experiment{
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}
	cases := []struct {
		run             int
		wantExperiments []*Experiment
	}{
		{run: 0,
			wantExperiments: wantExperimentsMemory,
		},
		{run: 1,
			wantExperiments: wantExperimentsFile,
		},
	}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	for _, c := range cases {
		htmlCmds := make(chan cmd.Cmd)
		cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
		go cmdMonitor.Run()
		pm, err := NewMonitor(tmpDir, htmlCmds)
		if err != nil {
			t.Fatalf("NewMonitor() err: %v", err)
		}
		if c.run == 0 {
			epr1, err := NewExperimentReporter(pm, "bank-divorced.json")
			if err != nil {
				t.Fatalf("NewExperimentReporter(pm, \"bank-divorced.json\") err: %s", err)
			}

			epr2, err := NewExperimentReporter(pm, "bank-full-divorced.json")
			if err != nil {
				t.Fatalf("NewExperimentReporter(pm, \"bank-full-divorced.json\") err: %s", err)
			}

			epr1.ReportProgress("Describing dataset", 0)
			time.Sleep(time.Second)
			epr2.ReportProgress("Tweaking rules", 0)
			epr2.ReportProgress("Assessing rules", 0.24)
		}
		got := pm.GetExperiments()
		if err := checkExperimentsMatch(got, c.wantExperiments); err != nil {
			t.Errorf("checkExperimentsMatch() err: %s", err)
		}
		time.Sleep(1 * time.Second)
		close(htmlCmds)
	}
}

func TestReportError(t *testing.T) {
	wantExperimentsMemory := []*Experiment{
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Tags:               []string{"test", "bank"},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Couldn't load experiment file: open csv/bank-divorced.cs: no such file or directory",
			Status:             Failure,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}
	wantExperimentsFile := []*Experiment{
		&Experiment{
			Title:              "This is a jolly nice title",
			Tags:               []string{"test", "bank", "fred / ned"},
			Stamp:              mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
	}
	cases := []struct {
		run             int
		wantExperiments []*Experiment
	}{
		{run: 0,
			wantExperiments: wantExperimentsMemory,
		},
		{run: 1,
			wantExperiments: wantExperimentsFile,
		},
	}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	for _, c := range cases {
		htmlCmds := make(chan cmd.Cmd)
		cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
		go cmdMonitor.Run()
		pm, err := NewMonitor(tmpDir, htmlCmds)
		if err != nil {
			t.Fatalf("NewMonitor() err: %v", err)
		}
		if c.run == 0 {
			epr, err := NewExperimentReporter(pm, "bank-divorced.json")
			if err != nil {
				t.Fatalf("NewExperimentReporter(pm, \"bank-divorced.json\") err: %s", err)
			}

			epr.ReportError(errors.New("Couldn't load experiment file: open csv/bank-divorced.cs: no such file or directory"))
		}
		got := pm.GetExperiments()
		if err := checkExperimentsMatch(got, c.wantExperiments); err != nil {
			t.Errorf("checkExperimentsMatch() err: %s", err)
		}
		time.Sleep(1 * time.Second)
		close(htmlCmds)
	}
}
