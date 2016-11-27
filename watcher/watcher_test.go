package watcher

import (
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
	go logger.Run(quit)
	go Watch(dir, logger, quit, filenames)
	time.Sleep(1.0 * time.Second)
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
	go logger.Run(quit)
	go Watch(tempDir, logger, quit, filenames)
	time.Sleep(1.0 * time.Second)

	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.yaml"), tempDir)
	time.Sleep(3.0 * time.Second)
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
	go logger.Run(quit)
	go Watch(tempDir, logger, quit, filenames)
	time.Sleep(1.0 * time.Second)

	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tempDir)
	time.Sleep(3.0 * time.Second)
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
