package progress

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/vlifesystems/rulehuntersrv/html/cmd"
	"github.com/vlifesystems/rulehuntersrv/internal/testhelpers"
)

func TestGetExperiments(t *testing.T) {
	/* This sorts in reverse order of date */
	expected := []*Experiment{
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
			Tags:               []string{},
			Stamp:              mustNewTime("2016-05-05T09:36:59.762587932+01:00"),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Tags:               []string{"test", "bank"},
			Stamp:              mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              mustNewTime("2016-05-04T14:52:08.993750731+01:00"),
			ExperimentFilename: "bank-bad.json",
			Msg:                "Couldn't load experiment file: open csv/bank-tiny.cs: no such file or directory",
			Status:             Failure,
		},
	}

	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tempDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := newHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.run()
	pm, err := NewMonitor(tempDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %s", err)
	}
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, expected); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestGetExperiments_notExists(t *testing.T) {
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := newHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.run()
	pm, err := NewMonitor(tempDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %s", err)
	}
	experiments := pm.GetExperiments()
	if len(experiments) != 0 {
		t.Errorf("GetExperiments() expected 0 experiments got: %d",
			len(experiments))
	}
}

func TestAddExperiment_experiment_exists(t *testing.T) {
	expected := []*Experiment{
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
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              mustNewTime("2016-05-04T14:52:08.993750731+01:00"),
			ExperimentFilename: "bank-bad.json",
			Msg:                "Couldn't load experiment file: open csv/bank-tiny.cs: no such file or directory",
			Status:             Failure,
		},
	}

	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tempDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := newHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.run()
	pm, err := NewMonitor(tempDir, htmlCmds)
	if err != nil {
		t.Errorf("NewMonitor() err: %s", err)
	}
	if err := pm.AddExperiment("bank-divorced.json"); err != nil {
		t.Fatalf("AddExperiment() err: %s", err)
	}
	if err := pm.AddExperiment("bank-full-divorced.json"); err != nil {
		t.Fatalf("AddExperiment() err: %s", err)
	}
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, expected); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
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
		&Experiment{
			Title:              "",
			Tags:               []string{},
			Stamp:              mustNewTime("2016-05-04T14:52:08.993750731+01:00"),
			ExperimentFilename: "bank-bad.json",
			Msg:                "Couldn't load experiment file: open csv/bank-tiny.cs: no such file or directory",
			Status:             Failure,
		},
	}
	wantHtmlCmdsReceived := []cmd.Cmd{cmd.Progress, cmd.Progress, cmd.Reports}

	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tempDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := newHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.run()
	pm, err := NewMonitor(tempDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	epr, err := NewExperimentProgressReporter(pm, "bank-full-divorced.json")
	if err != nil {
		t.Fatalf("NewExperimentProgressReporter(pm, \"bank-full-divorced.json\") err: %s", err)
	}

	epr.ReportSuccess()
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, wantExperiments); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
	sleep(1)
	close(htmlCmds)
	htmlCmdsReceived := cmdMonitor.getCmdsReceived()
	if !reflect.DeepEqual(htmlCmdsReceived, wantHtmlCmdsReceived) {
		t.Errorf("getCmdsRecevied() received commands - got: %s, want: %s",
			htmlCmdsReceived, wantHtmlCmdsReceived)
	}
}

func TestGetFinishStamp(t *testing.T) {
	cases := []struct {
		filename       string
		wantIsFinished bool
		wantStamp      time.Time
	}{
		{"bank-bad.json",
			true,
			mustNewTime("2016-05-04T14:52:08.993750731+01:00"),
		},
		{"bank-divorced.json",
			true,
			mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
		},
		{"bank-full-divorced.json", false, time.Now()},
		{"nothing", false, time.Now()},
		{"bank-tiny.json",
			true,
			mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
		},
	}
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tempDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := newHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.run()
	pm, err := NewMonitor(tempDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %s", err)
	}

	for _, c := range cases {
		gotIsFinished, gotStamp := pm.GetFinishStamp(c.filename)
		if gotIsFinished != c.wantIsFinished {
			t.Errorf("GetFinishStamp(%s) gotIsFinished: %t, wantIsFinished: %t",
				c.filename, gotIsFinished, c.wantIsFinished)
		}
		if gotIsFinished && !gotStamp.Equal(c.wantStamp) {
			t.Errorf("GetFinishStamp(%s) gotStamp: %v, wantStamp: %v",
				c.filename, gotStamp, c.wantStamp)
		}
	}
}

/**************************************
 *   Helper functions
 **************************************/

func mustNewTime(stamp string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		panic(err)
	}
	return t
}

func sleep(s int) {
	sleepInSeconds := time.Duration(s)
	time.Sleep(sleepInSeconds * time.Second)
}

type htmlCmdMonitor struct {
	htmlCmds     chan cmd.Cmd
	cmdsReceived []cmd.Cmd
}

func newHtmlCmdMonitor(cmds chan cmd.Cmd) *htmlCmdMonitor {
	return &htmlCmdMonitor{cmds, []cmd.Cmd{}}
}

func (h *htmlCmdMonitor) run() {
	for c := range h.htmlCmds {
		h.cmdsReceived = append(h.cmdsReceived, c)
	}
}

func (h *htmlCmdMonitor) getCmdsReceived() []cmd.Cmd {
	return h.cmdsReceived
}

func checkExperimentsMatch(experiments1, experiments2 []*Experiment) error {
	if len(experiments1) != len(experiments2) {
		return errors.New("Lengths of experiments don't match")
	}
	for i, e := range experiments1 {
		if err := checkExperimentMatch(e, experiments2[i]); err != nil {
			return err
		}
	}
	return nil
}

func checkExperimentMatch(e1, e2 *Experiment) error {
	if e1.Title != e2.Title {
		return fmt.Errorf("Title doesn't match: %s != %s", e1.Title, e2.Title)
	}
	if e1.ExperimentFilename != e2.ExperimentFilename {
		return errors.New("ExperimentFilename doesn't match")
	}
	if e1.Msg != e2.Msg {
		return errors.New("Msg doesn't match")
	}
	if e1.Status != e2.Status {
		return errors.New("Status doesn't match")
	}
	if !timesClose(e1.Stamp, e2.Stamp, 1) {
		return errors.New("Stamp not close in time")
	}
	if len(e1.Tags) != len(e2.Tags) {
		return errors.New("Tags doesn't match")
	}
	for i, t := range e1.Tags {
		if t != e2.Tags[i] {
			return errors.New("Tags doesn't match")
		}
	}
	return nil
}

func timesClose(t1, t2 time.Time, maxSecondsDiff int) bool {
	diff := t1.Sub(t2)
	secondsDiff := math.Abs(diff.Seconds())
	return secondsDiff <= float64(maxSecondsDiff)
}
