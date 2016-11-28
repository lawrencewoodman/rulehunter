package watcher

import (
	"errors"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

// Test initial filenames
func TestWatch_1(t *testing.T) {
	dir := "fixtures"
	filenames := make(chan string, 100)
	logger := testhelpers.NewLogger()
	quit := make(chan struct{})
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(dir, period, logger, quit, filenames)
	time.Sleep(100 * time.Millisecond)
	close(quit)

	wantFilenames := []string{"debt.json", "debt.yaml", "flow.yaml"}
	gotFilenames := []string{}
	for f := range filenames {
		gotFilenames = append(gotFilenames, f)
	}
	sort.Strings(gotFilenames)
	sort.Strings(wantFilenames)
	if !reflect.DeepEqual(gotFilenames, wantFilenames) {
		t.Errorf("Watch: gotFilenames: %v, wantFilenames: %v",
			gotFilenames, wantFilenames)
	}
	if logEntries := logger.GetEntries(); len(logEntries) != 0 {
		t.Errorf("Watch: gotLogEntries: %v, wanted: []", logEntries)
	}
}

// Test adding a filename to directory
func TestWatch_2(t *testing.T) {
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tempDir)

	filenames := make(chan string, 100)
	logger := testhelpers.NewLogger()
	quit := make(chan struct{})
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(tempDir, period, logger, quit, filenames)
	time.Sleep(100 * time.Millisecond)

	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.yaml"), tempDir)
	time.Sleep(100 * time.Millisecond)
	close(quit)

	wantFilenames := []string{"debt.json", "debt.yaml", "flow.yaml"}
	gotFilenames := []string{}
	for f := range filenames {
		gotFilenames = append(gotFilenames, f)
	}
	sort.Strings(gotFilenames)
	sort.Strings(wantFilenames)
	if !reflect.DeepEqual(gotFilenames, wantFilenames) {
		t.Errorf("Watch: gotFilenames: %v, wantFilenames: %v",
			gotFilenames, wantFilenames)
	}
	if logEntries := logger.GetEntries(); len(logEntries) != 0 {
		t.Errorf("Watch: gotLogEntries: %v, wanted: []", logEntries)
	}
}

// Test changing a file
func TestWatch_3(t *testing.T) {
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tempDir)

	filenames := make(chan string, 100)
	logger := testhelpers.NewLogger()
	quit := make(chan struct{})
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(tempDir, period, logger, quit, filenames)
	time.Sleep(100 * time.Millisecond)

	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tempDir)
	time.Sleep(100 * time.Millisecond)
	close(quit)

	wantFilenames := []string{"debt.json", "debt.json", "flow.yaml"}
	gotFilenames := []string{}
	for f := range filenames {
		gotFilenames = append(gotFilenames, f)
	}
	sort.Strings(gotFilenames)
	sort.Strings(wantFilenames)
	if !reflect.DeepEqual(gotFilenames, wantFilenames) {
		t.Errorf("Watch: gotFilenames: %v, wantFilenames: %v",
			gotFilenames, wantFilenames)
	}

	if logEntries := logger.GetEntries(); len(logEntries) != 0 {
		t.Errorf("Watch: gotLogEntries: %v, wanted: []", logEntries)
	}
}

func TestWatch_errors(t *testing.T) {
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	dir := filepath.Join(tempDir, "non")
	filenames := make(chan string, 100)
	logger := testhelpers.NewLogger()
	quit := make(chan struct{})
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(dir, period, logger, quit, filenames)
	time.Sleep(500 * time.Millisecond)
	close(quit)

	pe := &os.PathError{
		Op:   "open",
		Path: dir,
		Err:  errors.New("no such file or directory"),
	}
	wantFilenames := []string{}
	wantLogEntries := []testhelpers.Entry{
		testhelpers.Entry{
			Level: testhelpers.Error,
			Msg:   pe.Error(),
		},
	}

	gotFilenames := []string{}
	for f := range filenames {
		gotFilenames = append(gotFilenames, f)
	}
	sort.Strings(gotFilenames)
	sort.Strings(wantFilenames)
	if !reflect.DeepEqual(gotFilenames, wantFilenames) {
		t.Errorf("Watch: gotFilenames: %v, wantFilenames: %v",
			gotFilenames, wantFilenames)
	}
	if !reflect.DeepEqual(wantLogEntries, logger.GetEntries()) {
		t.Errorf("Watch: gotLogEntries: %v, want: %v",
			logger.GetEntries(), wantLogEntries)
	}
}

func TestGetExperimentFilenames(t *testing.T) {
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.yaml"), tempDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tempDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "flow.yaml"),
		tempDir,
		"flow.txt",
	)

	wantFilenames := []string{"debt.json", "debt.yaml", "flow.yaml"}
	gotFilenames, err := GetExperimentFilenames(tempDir)
	if err != nil {
		t.Fatalf("GetExperimentFilenames: %v", err)
	}

	sort.Strings(gotFilenames)
	sort.Strings(wantFilenames)
	if !reflect.DeepEqual(gotFilenames, wantFilenames) {
		t.Errorf("GetExperimentFilenames: gotFilenames: %v, wantFilenames: %v",
			gotFilenames, wantFilenames)
	}
}

func TestGetExperimentFilenames_errors(t *testing.T) {
	tempDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tempDir)
	dir := filepath.Join(tempDir, "non")
	wantFilenames := []string{}
	wantErr := &os.PathError{
		Op:   "open",
		Path: dir,
		Err:  errors.New("no such file or directory"),
	}
	gotFilenames, err := GetExperimentFilenames(dir)
	if err == nil || err.Error() != wantErr.Error() {
		t.Fatalf("GetExperimentFilenames: gotErr: %v, wantErr: %v", err, wantErr)
	}

	if !reflect.DeepEqual(gotFilenames, wantFilenames) {
		t.Errorf("GetExperimentFilenames: gotFilenames: %v, wantFilenames: %v",
			gotFilenames, wantFilenames)
	}
}
