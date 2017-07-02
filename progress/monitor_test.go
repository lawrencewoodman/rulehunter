package progress

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestExperimentString(t *testing.T) {
	cases := []struct {
		status StatusKind
		want   string
	}{
		{status: Waiting, want: "waiting"},
		{status: Processing, want: "processing"},
		{status: Success, want: "success"},
		{status: Failure, want: "failure"},
	}
	for _, c := range cases {
		got := c.status.String()
		if got != c.want {
			t.Errorf("String() got: %s, want: %s", got, c.want)
		}
	}
}

func TestNewMonitor_errors(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "progress_invalid.json"),
		tmpDir,
		"progress.json",
	)
	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()

	wantErr := errors.New("invalid character '[' after object key")
	_, gotErr := NewMonitor(tmpDir, htmlCmds)
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("NewMonitor: gotErr: %s, wantErr: %s", gotErr, wantErr)
	}
}

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
		t.Fatalf("NewMonitor() err: %s", err)
	}
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, expected); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestGetExperiments_notExists(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := NewMonitor(tmpDir, htmlCmds)
	if err != nil {
		t.Fatalf("NewMonitor() err: %s", err)
	}
	experiments := pm.GetExperiments()
	if len(experiments) != 0 {
		t.Errorf("GetExperiments() expected 0 experiments got: %d",
			len(experiments))
	}
}

func TestGetFinishStamp(t *testing.T) {
	cases := []struct {
		filename       string
		wantIsFinished bool
		wantStamp      time.Time
	}{
		{"bank-bad.json",
			false,
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
		{"bank-what.json",
			false,
			mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
		},
	}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "progress_processing.json"),
		tmpDir,
		"progress.json",
	)

	htmlCmds := make(chan cmd.Cmd)
	cmdMonitor := testhelpers.NewHtmlCmdMonitor(htmlCmds)
	go cmdMonitor.Run()
	pm, err := NewMonitor(tmpDir, htmlCmds)
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

func checkExperimentsMatch(
	experiments1 []*Experiment,
	experiments2 []*Experiment,
) error {
	if len(experiments1) != len(experiments2) {
		return fmt.Errorf("Lengths of experiments don't match: %d != %d",
			len(experiments1), len(experiments2))
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
		return fmt.Errorf("Title doesn't match: %s != %s", e1, e2)
	}
	if e1.ExperimentFilename != e2.ExperimentFilename {
		return fmt.Errorf("ExperimentFilename doesn't match: %s != %s",
			e1, e2)
	}
	if e1.Msg != e2.Msg {
		return fmt.Errorf("Msg doesn't match: %s != %s", e1, e2)
	}
	if e1.Percent != e2.Percent {
		return fmt.Errorf("Percent doesn't match: %s != %s", e1, e2)
	}
	if e1.Status != e2.Status {
		return errors.New("Status doesn't match")
	}
	if !timesClose(e1.Stamp, e2.Stamp, 5) {
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
