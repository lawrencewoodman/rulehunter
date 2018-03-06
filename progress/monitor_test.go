package progress

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/report"
)

func TestNewMonitor_errors(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "progress_invalid.json"),
		tmpDir,
		"progress.json",
	)

	wantErr := errors.New("invalid character '[' after object key")
	_, gotErr := NewMonitor(tmpDir)
	if gotErr == nil || gotErr.Error() != wantErr.Error() {
		t.Errorf("NewMonitor: gotErr: %s, wantErr: %s", gotErr, wantErr)
	}
}

func TestGetExperiments(t *testing.T) {
	/* This sorts in reverse order of date */
	expected := []*Experiment{
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp: mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:   "Finished processing successfully",
				State: Success,
			},
		},
		&Experiment{
			Title:    "Who is more likely to be divorced",
			Filename: "bank-divorced.json",
			Tags:     []string{"test", "bank"},
			Category: "contracts",
			Status: &Status{
				Stamp: mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
				Msg:   "Finished processing successfully",
				State: Success,
			},
		},
	}

	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	pm, err := NewMonitor(tmpDir)
	if err != nil {
		t.Fatalf("NewMonitor() err: %s", err)
	}
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, expected); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestGetExperiments_goroutines(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	// TODO: work out why this is commented
	//testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	pm, err := NewMonitor(tmpDir)
	if err != nil {
		t.Fatalf("NewMonitor: %s", err)
	}
	rand.Seed(47)
	numGoRoutines := 200
	wantNumExperiments := numGoRoutines / 2
	var wg sync.WaitGroup
	wg.Add(numGoRoutines)
	for i := 0; i < numGoRoutines/2; i++ {
		n := i
		go func() {
			time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond)
			pm.AddExperiment(fmt.Sprintf("experiment%d", n), "", []string{}, "")
			wg.Done()
		}()
		go func() {
			time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond)
			pm.GetExperiments()
			wg.Done()
		}()
	}
	wg.Wait()
	got := pm.GetExperiments()
	if len(got) != wantNumExperiments {
		t.Errorf("GetExperiments - len(got): %d, want: %d",
			len(got), wantNumExperiments)
	}
}

func TestGetExperiments_notExists(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)

	pm, err := NewMonitor(tmpDir)
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
			Title:    "Who is more likely to be married",
			Filename: "bank-married.json",
			Tags:     []string{"test", "bank"},
			Category: "contracts",
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Waiting to be processed",
				Percent: float64(0.0),
				State:   Waiting,
			},
		},
		&Experiment{
			Title:    "Who is more likely to be divorced (full)",
			Filename: "bank-full-divorced.json",
			Tags:     []string{"test", "bank", "full"},
			Category: "contracts",
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Waiting to be processed",
				Percent: float64(0.0),
				State:   Waiting,
			},
		},
		&Experiment{
			Title:    "Who is more likely to be divorced (normal)",
			Filename: "bank-divorced.json",
			Tags:     []string{"test", "bank", "normal"},
			Category: "contracts",
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Waiting to be processed",
				Percent: float64(0.0),
				State:   Waiting,
			},
		},
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: float64(0.0),
				State:   Success,
			},
		},
	}

	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	pm, err := NewMonitor(tmpDir)
	if err != nil {
		t.Errorf("NewMonitor() err: %s", err)
	}
	time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	err = pm.AddExperiment(
		"bank-divorced.json",
		"Who is more likely to be divorced (normal)",
		[]string{"test", "bank", "normal"},
		"contracts",
	)
	if err != nil {
		t.Fatalf("AddExperiment: %s", err)
	}
	time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	err = pm.AddExperiment(
		"bank-full-divorced.json",
		"Who is more likely to be divorced (full)",
		[]string{"test", "bank", "full"},
		"contracts",
	)
	time.Sleep(200 * time.Millisecond)
	err = pm.AddExperiment(
		"bank-married.json",
		"Who is more likely to be married",
		[]string{"test", "bank"},
		"contracts",
	)
	if err != nil {
		t.Fatalf("AddExperiment: %s", err)
	}
	time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	pm.ReportProgress(
		"bank-married.json",
		report.Train,
		"something is happening",
		0,
	)
	time.Sleep(100 * time.Millisecond) /* Windows time resolution is low */
	err = pm.AddExperiment(
		"bank-married.json",
		"Who is more likely to be married",
		[]string{"test", "bank"},
		"contracts",
	)
	if err != nil {
		t.Fatalf("AddExperiment: %s", err)
	}
	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, expected); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestReportSuccess(t *testing.T) {
	wantExperiments := []*Experiment{
		&Experiment{
			Title:    "",
			Filename: "bank-full-divorced.json",
			Tags:     []string{},
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
		},
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
		},
		&Experiment{
			Title:    "Who is more likely to be divorced",
			Filename: "bank-divorced.json",
			Tags:     []string{"test", "bank"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-04T14:53:00.570347516+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
		},
	}
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "progress.json"), tmpDir)

	filename := "bank-full-divorced.json"
	pm, err := NewMonitor(tmpDir)

	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	if err := pm.AddExperiment(filename, "", []string{}, ""); err != nil {
		t.Fatalf("AddExperiment: %s", err)
	}
	pm.ReportSuccess(filename)

	got := pm.GetExperiments()
	if err := checkExperimentsMatch(got, wantExperiments); err != nil {
		t.Errorf("checkExperimentsMatch() err: %s", err)
	}
}

func TestReportProgress(t *testing.T) {
	wantExperimentsMemory := []*Experiment{
		&Experiment{
			Title:    "A nice full title",
			Filename: "bank-full-divorced.json",
			Tags:     []string{"bank", "full"},
			Category: "contracts",
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Train > Assessing rules",
				Percent: 0.24,
				State:   Processing,
			},
		},
		&Experiment{
			Title:    "Who is more likely to be divorced",
			Filename: "bank-divorced.json",
			Tags:     []string{"test", "bank"},
			Category: "contracts",
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Train > Describing dataset",
				Percent: 0.0,
				State:   Processing,
			},
		},
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
		},
	}
	wantExperimentsFile := []*Experiment{
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
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
		pm, err := NewMonitor(tmpDir)
		if err != nil {
			t.Fatalf("NewMonitor() err: %v", err)
		}
		if c.run == 0 {
			err = pm.AddExperiment(
				"bank-full-divorced.json",
				"A nice full title",
				[]string{"bank", "full"},
				"contracts",
			)
			if err != nil {
				t.Fatalf("AddExperiment: %s", err)
			}

			pm.ReportProgress(
				"bank-divorced.json",
				report.Train,
				"Describing dataset",
				0,
			)
			time.Sleep(time.Second)
			pm.ReportProgress(
				"bank-full-divorced.json",
				report.Train,
				"Tweaking rules",
				0,
			)
			pm.ReportProgress(
				"bank-full-divorced.json",
				report.Train,
				"Assessing rules",
				0.24,
			)
		}
		got := pm.GetExperiments()
		if err := checkExperimentsMatch(got, c.wantExperiments); err != nil {
			t.Errorf("checkExperimentsMatch() err: %s", err)
		}
		time.Sleep(1 * time.Second)
	}
}

func TestReportError(t *testing.T) {
	wantExperimentsMemory := []*Experiment{
		&Experiment{
			Title:    "Who is more likely to be divorced",
			Filename: "bank-divorced.json",
			Tags:     []string{"test", "bank"},
			Category: "contracts",
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Couldn't load experiment file: open csv/bank-divorced.cs: no such file or directory",
				Percent: 0.0,
				State:   Error,
			},
		},
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
		},
	}
	wantExperimentsFile := []*Experiment{
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
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

	for i, c := range cases {
		pm, err := NewMonitor(tmpDir)
		if err != nil {
			t.Fatalf("NewMonitor() err: %v", err)
		}
		if c.run == 0 {
			pm.ReportError(
				"bank-divorced.json",
				errors.New("Couldn't load experiment file: open csv/bank-divorced.cs: no such file or directory"),
			)
		}
		got := pm.GetExperiments()
		if err := checkExperimentsMatch(got, c.wantExperiments); err != nil {
			t.Errorf("(%d) checkExperimentsMatch() err: %s", i, err)
		}
	}
}

func TestReportLoadError(t *testing.T) {
	wantExperimentsMemory := []*Experiment{
		&Experiment{
			Title:    "",
			Filename: "bank-divorced.json",
			Tags:     []string{},
			Status: &Status{
				Stamp:   time.Now(),
				Msg:     "Error loading experiment: open csv/bank-divorced.cs: no such file or directory",
				Percent: 0.0,
				State:   Error,
			},
		},
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
		},
	}
	wantExperimentsFile := []*Experiment{
		&Experiment{
			Title:    "This is a jolly nice title",
			Filename: "bank-tiny.json",
			Tags:     []string{"test", "bank", "fred / ned"},
			Category: "contracts",
			Status: &Status{
				Stamp:   mustNewTime("2016-05-05T09:37:58.220312223+01:00"),
				Msg:     "Finished processing successfully",
				Percent: 0.0,
				State:   Success,
			},
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
		pm, err := NewMonitor(tmpDir)
		if err != nil {
			t.Fatalf("NewMonitor() err: %v", err)
		}
		if c.run == 0 {
			pm.ReportLoadError(
				"bank-divorced.json",
				errors.New("open csv/bank-divorced.cs: no such file or directory"),
			)
		}
		got := pm.GetExperiments()
		if err := checkExperimentsMatch(got, c.wantExperiments); err != nil {
			t.Errorf("checkExperimentsMatch() err: %s", err)
		}
		time.Sleep(1 * time.Second)
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

	pm, err := NewMonitor(tmpDir)
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
	if e1.Filename != e2.Filename {
		return fmt.Errorf("Filename doesn't match: %s != %s", e1, e2)
	}
	if e1.Status.Msg != e2.Status.Msg {
		return fmt.Errorf("Status.Msg doesn't match: %s != %s",
			e1.Status.Msg, e2.Status.Msg)
	}
	if e1.Status.Percent != e2.Status.Percent {
		return errors.New("Status.Percent doesn't match")
	}
	if e1.Status.State != e2.Status.State {
		return fmt.Errorf("Status.State doesn't match: %s != %s (%s)",
			e1.Status.State, e2.Status.State, e1.Filename)
	}
	if !timesClose(e1.Status.Stamp, e2.Status.Stamp, 10) {
		return errors.New("Status.Stamp not close in time")
	}
	if len(e1.Tags) != len(e2.Tags) {
		return errors.New("Tags doesn't match")
	}
	for i, t := range e1.Tags {
		if t != e2.Tags[i] {
			return errors.New("Tags doesn't match")
		}
	}
	if e1.Category != e2.Category {
		return fmt.Errorf("Categories don't match: %s != %s",
			e1.Category, e2.Category)
	}
	return nil
}

func timesClose(t1, t2 time.Time, maxSecondsDiff int) bool {
	diff := t1.Sub(t2)
	secondsDiff := math.Abs(diff.Seconds())
	return secondsDiff <= float64(maxSecondsDiff)
}
