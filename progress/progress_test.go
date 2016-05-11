package progress

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetExperiments(t *testing.T) {
	/* This sorts in reverse order of date */
	expected := []*Experiment{
		&Experiment{
			Title:              "This is a jolly nice title",
			Categories:         []string{"test", "bank", "fred / ned"},
			Stamp:              mustParseTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Categories:         []string{},
			Stamp:              mustParseTime("2016-05-05T09:36:59.762587932+01:00"),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Categories:         []string{"test", "bank"},
			Stamp:              mustParseTime("2016-05-04T14:53:00.570347516+01:00"),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "",
			Categories:         []string{},
			Stamp:              mustParseTime("2016-05-04T14:52:08.993750731+01:00"),
			ExperimentFilename: "bank-bad.json",
			Msg:                "Couldn't load experiment file: open csv/bank-tiny.cs: no such file or directory",
			Status:             Failure,
		},
	}

	tempDir, err := ioutil.TempDir("", "progress_test")
	if err != nil {
		t.Errorf("TempDir() err: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)

	err = copyFile(filepath.Join("fixtures", "progress.json"), tempDir)
	if err != nil {
		t.Errorf("copyFile() err: %s", err)
		return
	}
	pm := NewMonitor(tempDir)
	got, err := pm.GetExperiments()
	if err != nil {
		t.Errorf("GetExperiments() err: %s", err)
		return
	}
	experimentsMatch, matchMsg := doExperimentsMatch(got, expected)
	if !experimentsMatch {
		t.Errorf("doExperimentsMatch() msg: %s", matchMsg)
	}
}

func TestGetExperiments_notExists(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "progress_test")
	if err != nil {
		t.Errorf("TempDir() err: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)

	pm := NewMonitor(tempDir)
	experiments, err := pm.GetExperiments()
	if err != nil {
		t.Errorf("GetExperiments() err: %s", err)
	}
	if len(experiments) != 0 {
		t.Errorf("GetExperiments() expected 0 experiments got: %d",
			len(experiments))
	}
}

func TestAddExperiment_experiment_exists(t *testing.T) {
	expected := []*Experiment{
		&Experiment{
			Title:              "",
			Categories:         []string{},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Categories:         []string{"test", "bank", "fred / ned"},
			Stamp:              mustParseTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Categories:         []string{},
			Stamp:              mustParseTime("2016-05-05T09:36:59.762587932+01:00"),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "",
			Categories:         []string{},
			Stamp:              mustParseTime("2016-05-04T14:52:08.993750731+01:00"),
			ExperimentFilename: "bank-bad.json",
			Msg:                "Couldn't load experiment file: open csv/bank-tiny.cs: no such file or directory",
			Status:             Failure,
		},
	}

	tempDir, err := ioutil.TempDir("", "progress_test")
	if err != nil {
		t.Errorf("TempDir() err: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)

	err = copyFile(filepath.Join("fixtures", "progress.json"), tempDir)
	if err != nil {
		t.Errorf("copyFile() err: %s", err)
		return
	}
	pm := NewMonitor(tempDir)
	if err := pm.AddExperiment("bank-divorced.json"); err != nil {
		t.Errorf("AddExperiment() err: %s", err)
		return
	}
	got, err := pm.GetExperiments()
	if err != nil {
		t.Errorf("GetExperiments() err: %s", err)
		return
	}
	experimentsMatch, matchMsg := doExperimentsMatch(got, expected)
	if !experimentsMatch {
		t.Errorf("doExperimentsMatch() msg: %s", matchMsg)
	}
}

func TestAddExperiment_experiment_doesnt_exist(t *testing.T) {
	expected := []*Experiment{
		&Experiment{
			Title:              "",
			Categories:         []string{},
			Stamp:              time.Now(),
			ExperimentFilename: "bank-credit.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "This is a jolly nice title",
			Categories:         []string{"test", "bank", "fred / ned"},
			Stamp:              mustParseTime("2016-05-05T09:37:58.220312223+01:00"),
			ExperimentFilename: "bank-tiny.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Categories:         []string{},
			Stamp:              mustParseTime("2016-05-05T09:36:59.762587932+01:00"),
			ExperimentFilename: "bank-full-divorced.json",
			Msg:                "Waiting to be processed",
			Status:             Waiting,
		},
		&Experiment{
			Title:              "Who is more likely to be divorced",
			Categories:         []string{"test", "bank"},
			Stamp:              mustParseTime("2016-05-04T14:53:00.570347516+01:00"),
			ExperimentFilename: "bank-divorced.json",
			Msg:                "Finished processing successfully",
			Status:             Success,
		},
		&Experiment{
			Title:              "",
			Categories:         []string{},
			Stamp:              mustParseTime("2016-05-04T14:52:08.993750731+01:00"),
			ExperimentFilename: "bank-bad.json",
			Msg:                "Couldn't load experiment file: open csv/bank-tiny.cs: no such file or directory",
			Status:             Failure,
		},
	}

	tempDir, err := ioutil.TempDir("", "progress_test")
	if err != nil {
		t.Errorf("TempDir() err: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)

	err = copyFile(filepath.Join("fixtures", "progress.json"), tempDir)
	if err != nil {
		t.Errorf("copyFile() err: %s", err)
		return
	}
	pm := NewMonitor(tempDir)
	if err := pm.AddExperiment("bank-credit.json"); err != nil {
		t.Errorf("AddExperiment() err: %s", err)
		return
	}
	got, err := pm.GetExperiments()
	if err != nil {
		t.Errorf("GetExperiments() err: %s", err)
		return
	}
	experimentsMatch, matchMsg := doExperimentsMatch(got, expected)
	if !experimentsMatch {
		t.Errorf("doExperimentsMatch() msg: %s", matchMsg)
	}
}

/**************************************
 *   Helper functions
 **************************************/

func copyFile(srcFilename, dstDir string) error {
	contents, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcFilename)
	if err != nil {
		return err
	}
	mode := info.Mode()
	dstFilename := filepath.Join(dstDir, filepath.Base(srcFilename))
	return ioutil.WriteFile(dstFilename, contents, mode)
}

func doExperimentsMatch(
	experiments1, experiments2 []*Experiment,
) (bool, string) {
	if len(experiments1) != len(experiments2) {
		return false, "Lengths of experiments don't match"
	}
	for i, e := range experiments1 {
		if match, msg := doesExperimentMatch(e, experiments2[i]); !match {
			return false, msg
		}
	}
	return true, ""
}

func doesExperimentMatch(e1, e2 *Experiment) (bool, string) {
	if e1.Title != e2.Title {
		return false,
			fmt.Sprintf("Title doesn't match: %s != %s", e1.Title, e2.Title)
	}
	if e1.ExperimentFilename != e2.ExperimentFilename {
		return false, "ExperimentFilename doesn't match"
	}
	if e1.Msg != e2.Msg {
		return false, "Msg doesn't match"
	}
	if e1.Status != e2.Status {
		return false, "Status doesn't match"
	}
	if !timesClose(e1.Stamp, e2.Stamp, 1) {
		return false, "Stamp not close in time"
	}
	if len(e1.Categories) != len(e2.Categories) {
		return false, "Categories doesn't match"
	}
	for i, c := range e1.Categories {
		if c != e2.Categories[i] {
			return false, "Categories doesn't match"
		}
	}
	return true, ""
}

func timesClose(t1, t2 time.Time, maxSecondsDiff int) bool {
	diff := t1.Sub(t2)
	secondsDiff := math.Abs(diff.Seconds())
	return secondsDiff <= float64(maxSecondsDiff)
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Sprintf("Can't parse time: %s", err))
	}
	return t
}
