package watcher

import (
	"fmt"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/quitter"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

// Test initial filenames
func TestWatch_1(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.yaml"), tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tmpDir)

	files := make(chan fileinfo.FileInfo, 100)
	logger := testhelpers.NewLogger()
	quit := quitter.New()
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(tmpDir, period, logger, quit, files)
	time.Sleep(200 * time.Millisecond)
	quit.Quit()

	wantNewFiles := map[string]int{
		"debt.json": 1,
		"debt.yaml": 1,
		"flow.yaml": 1,
	}
	wantNonNewFiles := map[string]int{
		"debt.json": 2,
		"debt.yaml": 2,
		"flow.yaml": 2,
	}
	err := checkCorrectFileChan(wantNewFiles, wantNonNewFiles, files)
	if err != nil {
		t.Error("Watch:", err)
	}

	if logEntries := logger.GetEntries(); len(logEntries) != 0 {
		t.Errorf("Watch: gotLogEntries: %v, wanted: []", logEntries)
	}
}

// Test adding a filename to directory
func TestWatch_2(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tmpDir)

	files := make(chan fileinfo.FileInfo, 100)
	logger := testhelpers.NewLogger()
	quit := quitter.New()
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(tmpDir, period, logger, quit, files)
	time.Sleep(100 * time.Millisecond)

	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.yaml"), tmpDir)
	time.Sleep(60 * time.Millisecond)
	quit.Quit()

	wantNewFiles := map[string]int{
		"debt.json": 1,
		"debt.yaml": 1,
		"flow.yaml": 1,
	}
	wantNonNewFiles := map[string]int{
		"debt.json": 2,
		"flow.yaml": 2,
		"debt.yaml": 0,
	}
	err := checkCorrectFileChan(wantNewFiles, wantNonNewFiles, files)
	if err != nil {
		t.Error("Watch:", err)
	}
	if logEntries := logger.GetEntries(); len(logEntries) != 0 {
		t.Errorf("Watch: gotLogEntries: %v, wanted: []", logEntries)
	}
}

// Test changing a file
func TestWatch_3(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tmpDir)

	files := make(chan fileinfo.FileInfo, 100)
	logger := testhelpers.NewLogger()
	quit := quitter.New()
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(tmpDir, period, logger, quit, files)
	time.Sleep(100 * time.Millisecond)

	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tmpDir)
	time.Sleep(100 * time.Millisecond)
	quit.Quit()

	wantNewFiles := map[string]int{
		"debt.json": 2,
		"flow.yaml": 1,
	}
	wantNonNewFiles := map[string]int{
		"debt.json": 1,
		"flow.yaml": 2,
	}
	err := checkCorrectFileChan(wantNewFiles, wantNonNewFiles, files)
	if err != nil {
		t.Error("Watch:", err)
	}
	if logEntries := logger.GetEntries(); len(logEntries) != 0 {
		t.Errorf("Watch: gotLogEntries: %v, wanted: []", logEntries)
	}
}

func TestWatch_errors(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	os.RemoveAll(tmpDir)
	dir := filepath.Join(tmpDir, "non")
	files := make(chan fileinfo.FileInfo, 100)
	logger := testhelpers.NewLogger()
	quit := quitter.New()
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(dir, period, logger, quit, files)
	time.Sleep(100 * time.Millisecond)
	quit.Quit()

	wantNewFiles := map[string]int{}
	wantNonNewFiles := map[string]int{}
	wantLogEntries := []testhelpers.Entry{
		testhelpers.Entry{
			Level: testhelpers.Error,
			Msg:   DirError(dir).Error(),
		},
	}

	err := checkCorrectFileChan(wantNewFiles, wantNonNewFiles, files)
	if err != nil {
		t.Error("Watch:", err)
	}
	if !reflect.DeepEqual(wantLogEntries, logger.GetEntries()) {
		t.Errorf("Watch: gotLogEntries: %v, want: %v",
			logger.GetEntries(), wantLogEntries)
	}
}

// Test a directory being removed part way through watching
func TestWatch_errors2(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	files := make(chan fileinfo.FileInfo, 100)
	logger := testhelpers.NewLogger()
	quit := quitter.New()
	period := 50 * time.Millisecond
	go logger.Run(quit)
	go Watch(tmpDir, period, logger, quit, files)
	time.Sleep(100 * time.Millisecond)

	os.RemoveAll(tmpDir)
	time.Sleep(100 * time.Millisecond)
	quit.Quit()

	wantNewFiles := map[string]int{}
	wantNonNewFiles := map[string]int{}
	wantLogEntries := []testhelpers.Entry{
		testhelpers.Entry{
			Level: testhelpers.Error,
			Msg:   DirError(tmpDir).Error(),
		},
	}

	err := checkCorrectFileChan(wantNewFiles, wantNonNewFiles, files)
	if err != nil {
		t.Error("Watch:", err)
	}
	if !reflect.DeepEqual(wantLogEntries, logger.GetEntries()) {
		t.Errorf("Watch: gotLogEntries: %v, want: %v",
			logger.GetEntries(), wantLogEntries)
	}
}

func TestGetExperimentFilenames(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.json"), tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "debt.yaml"), tmpDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "flow.yaml"), tmpDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "flow.yaml"),
		tmpDir,
		"flow.txt",
	)

	wantFiles := []string{"debt.json", "debt.yaml", "flow.yaml"}
	gotFiles, err := GetExperimentFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetExperimentFilenames: %v", err)
	}
	if err := checkCorrectFiles(gotFiles, wantFiles); err != nil {
		t.Error("GetExperimentFiles:", err)
	}
}

func TestGetExperimentFilenames_errors(t *testing.T) {
	tmpDir := testhelpers.TempDir(t)
	defer os.RemoveAll(tmpDir)
	dir := filepath.Join(tmpDir, "non")
	wantFiles := []string{}
	wantErr := DirError(dir)
	gotFiles, err := GetExperimentFiles(dir)
	if err == nil || err.Error() != wantErr.Error() {
		t.Fatalf("GetExperimentFilenames: gotErr: %v, wantErr: %v", err, wantErr)
	}

	if err := checkCorrectFiles(gotFiles, wantFiles); err != nil {
		t.Error("GetExperimentFiles:", err)
	}
}

func TestDirErrorError(t *testing.T) {
	dir := "/tmp/someplace"
	want := "can not watch directory: /tmp/someplace"
	got := DirError(dir).Error()
	if got != want {
		t.Errorf("Error: got: %s, want: %s", got, want)
	}
}

/*********************************
 *    Helper functions
 *********************************/
func checkCorrectFiles(
	gotFiles []fileinfo.FileInfo,
	wantFilenames []string,
) error {
	allFilenames := []string{}
	for _, file := range gotFiles {
		allFilenames = append(allFilenames, file.Name())
	}
	if len(allFilenames) != len(wantFilenames) {
		return fmt.Errorf("gotFiles: %v, wantFilenames: %v",
			allFilenames, wantFilenames)
	}
	sort.Strings(allFilenames)
	sort.Strings(wantFilenames)
	if !reflect.DeepEqual(allFilenames, wantFilenames) {
		return fmt.Errorf("gotFiles: %v, wantFilenames: %v",
			allFilenames, wantFilenames)
	}
	return nil
}

// The int in the maps below indicates that has been seen at least x times
func checkCorrectFileChan(
	wantNewFiles map[string]int,
	wantNonNewFiles map[string]int,
	files <-chan fileinfo.FileInfo,
) error {
	allFiles := []fileinfo.FileInfo{}
	gotNewFiles := map[string]int{}
	gotNonNewFiles := map[string]int{}
	for file := range files {
		isNew := true
		for _, f := range allFiles {
			if fileinfo.IsEqual(file, f) {
				isNew = false
				break
			}
		}
		if isNew {
			gotNewFiles[file.Name()]++
		} else {
			gotNonNewFiles[file.Name()]++
		}
		allFiles = append(allFiles, file)
	}
	if len(gotNewFiles) != len(wantNewFiles) {
		return fmt.Errorf("gotNewFiles: %v, wantNewFiles: %v",
			gotNewFiles, wantNewFiles)
	}
	if len(gotNewFiles) > len(wantNewFiles) {
		return fmt.Errorf("gotNewFiles: %v, wantNewFiles: %v",
			gotNewFiles, wantNewFiles)
	}
	for name, n := range wantNewFiles {
		if n > gotNewFiles[name] {
			return fmt.Errorf("gotNewFiles: %v, wantNewFiles: %v",
				gotNewFiles, wantNewFiles)
		}
	}
	if len(gotNonNewFiles) > len(wantNonNewFiles) {
		return fmt.Errorf("gotNonNewFiles: %v, wantNonNewFiles: %v",
			gotNonNewFiles, wantNonNewFiles)
	}
	for name, n := range wantNonNewFiles {
		if n > gotNonNewFiles[name] {
			return fmt.Errorf("gotNonNewFiles: %v, wantNonNewFiles: %v",
				gotNonNewFiles, wantNonNewFiles)
		}
	}
	return nil
}
