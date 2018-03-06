package html

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

// This checks if Run will quit properly when told to
func TestRun_quit(t *testing.T) {
	q := quitter.New()
	l := testhelpers.NewLogger()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Getwd() err: ", err)
	}
	defer os.Chdir(wd)
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)

	pm, err := progress.NewMonitor(
		filepath.Join(cfgDir, "build", "progress"),
	)
	if err != nil {
		t.Fatalf("NewMonitor: %s", err)
	}
	config := &config.Config{
		ExperimentsDir: filepath.Join(cfgDir, "experiments"),
		WWWDir:         filepath.Join(cfgDir, "www"),
		BuildDir:       filepath.Join(cfgDir, "build"),
	}
	h := New(config, pm, l)
	go h.Run(q)
	for !h.Running() {
	}

	quitC := time.NewTimer(2 * time.Second).C
	isRunningC := time.NewTicker(100 * time.Millisecond).C
	timeoutC := time.NewTimer(5 * time.Second).C
	for {
		select {
		case <-isRunningC:
			if !h.Running() {
				return
			}
		case <-quitC:
			q.Quit()
		case <-timeoutC:
			t.Fatalf("Run() didn't quit")
		}
	}
}

func TestGenerateAll(t *testing.T) {
	q := quitter.New()
	defer q.Quit()
	l := testhelpers.NewLogger()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Getwd() err: ", err)
	}
	defer os.Chdir(wd)
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)

	testFiles := []string{
		"bank-loss.json",
		"bank-profit.json",
		"bank-notagsnocats.json",
	}
	for _, f := range testFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", "reports", f),
			filepath.Join(cfgDir, "build", "reports"),
		)
	}

	pm, err := progress.NewMonitor(
		filepath.Join(cfgDir, "build", "progress"),
	)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir: filepath.Join(cfgDir, "experiments"),
		WWWDir:         filepath.Join(cfgDir, "www"),
		BuildDir:       filepath.Join(cfgDir, "build"),
	}
	wantFiles := []string{
		filepath.Join(cfgDir, "www", "index.html"),
		filepath.Join(cfgDir, "www", "activity", "index.html"),
		filepath.Join(
			cfgDir,
			"www",
			"reports",
			"category",
			"groupa",
			"how-to-make-a-loss",
			"train",
			"index.html",
		),
		filepath.Join(
			cfgDir,
			"www",
			"reports",
			"category",
			"groupb",
			"how-to-make-a-profit",
			"train",
			"index.html",
		),
		filepath.Join(cfgDir, "www", "reports", "notag", "index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "test",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "bank",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "fahrenheit-451",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "fred-ned",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "hot-in-the-city",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "nocategory", "index.html"),
		filepath.Join(cfgDir, "www", "reports", "category", "groupa",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "category", "groupb",
			"index.html"),
	}

	h := New(config, pm, l)

	if err := h.generateAll(); err != nil {
		t.Fatalf("generateAll: %s", err)
	}

	if err := checkFilesExist(wantFiles); err != nil {
		t.Errorf("checkFilesExist: %s", err)
	}
}

func TestGenerateReports(t *testing.T) {
	q := quitter.New()
	defer q.Quit()
	l := testhelpers.NewLogger()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Getwd() err: ", err)
	}
	defer os.Chdir(wd)
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)

	testFiles := []string{
		"bank-loss.json",
		"bank-profit.json",
		"bank-notagsnocats.json",
	}
	for _, f := range testFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", "reports", f),
			filepath.Join(cfgDir, "build", "reports"),
		)
	}

	pm, err := progress.NewMonitor(
		filepath.Join(cfgDir, "build", "progress"),
	)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir: filepath.Join(cfgDir, "experiments"),
		WWWDir:         filepath.Join(cfgDir, "www"),
		BuildDir:       filepath.Join(cfgDir, "build"),
	}
	wantFiles := []string{
		filepath.Join(cfgDir, "www", "index.html"),
		filepath.Join(cfgDir, "www", "activity", "index.html"),
		filepath.Join(
			cfgDir,
			"www",
			"reports",
			"category",
			"groupa",
			"how-to-make-a-loss",
			"train",
			"index.html",
		),
		filepath.Join(
			cfgDir,
			"www",
			"reports",
			"category",
			"groupb",
			"how-to-make-a-profit",
			"train",
			"index.html",
		),
		filepath.Join(cfgDir, "www", "reports", "notag", "index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "test",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "bank",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "fahrenheit-451",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "fred-ned",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "tag", "hot-in-the-city",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "nocategory", "index.html"),
		filepath.Join(cfgDir, "www", "reports", "category", "groupa",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "category", "groupb",
			"index.html"),
	}

	h := New(config, pm, l)

	if err := h.generateAll(); err != nil {
		t.Fatalf("generateReports: %s", err)
	}

	if err := checkFilesExist(wantFiles); err != nil {
		t.Errorf("checkFilesExist: %s", err)
	}
}

func TestGenerateProgress(t *testing.T) {
	q := quitter.New()
	defer q.Quit()
	l := testhelpers.NewLogger()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Getwd() err: ", err)
	}
	defer os.Chdir(wd)
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "reports", "bank-loss.json"),
		filepath.Join(cfgDir, "build", "reports"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "reports", "bank-profit.json"),
		filepath.Join(cfgDir, "build", "reports"),
	)

	pm, err := progress.NewMonitor(
		filepath.Join(cfgDir, "build", "progress"),
	)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir: filepath.Join(cfgDir, "experiments"),
		WWWDir:         filepath.Join(cfgDir, "www"),
		BuildDir:       filepath.Join(cfgDir, "build"),
	}
	wantFiles := []string{
		filepath.Join(cfgDir, "www", "activity", "index.html"),
	}

	h := New(config, pm, l)

	if err := h.generateProgress(); err != nil {
		t.Fatalf("generateProgress: %s", err)
	}

	if err := checkFilesExist(wantFiles); err != nil {
		t.Errorf("checkFilesExist: %s", err)
	}
}
