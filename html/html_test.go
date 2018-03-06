package html

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/report"
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

func TestGenReportURLDir(t *testing.T) {
	cases := []struct {
		mode     report.ModeKind
		category string
		title    string
		want     string
	}{
		{mode: report.Train,
			category: "",
			title:    "This could be very interesting",
			want:     "reports/nocategory/this-could-be-very-interesting/train/",
		},
		{mode: report.Train,
			category: "acme or emca",
			title:    "This could be very interesting",
			want:     "reports/category/acme-or-emca/this-could-be-very-interesting/train/",
		},
		{mode: report.Test,
			category: "",
			title:    "This could be very interesting",
			want:     "reports/nocategory/this-could-be-very-interesting/test/",
		},
		{mode: report.Test,
			category: "acme or emca",
			title:    "This could be very interesting",
			want:     "reports/category/acme-or-emca/this-could-be-very-interesting/test/",
		},
	}
	for _, c := range cases {
		got := genReportURLDir(c.mode, c.category, c.title)
		if got != c.want {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.category, c.title, got, c.want)
		}
	}
}

func TestEscapeString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"This is a TITLE",
			"this-is-a-title"},
		{"  hello how are % you423 33  today __ --",
			"hello-how-are-you423-33-today"},
		{"--  hello how are %^& you423 33  today __ --",
			"hello-how-are-you423-33-today"},
		{"hello((_ how are % you423 33  today",
			"hello-how-are-you423-33-today"},
		{"This is it's TITLE",
			"this-is-its-title"},
		{"", ""},
	}
	for _, c := range cases {
		got := escapeString(c.in)
		if got != c.want {
			t.Errorf("escapeString(%s) got: %s, want: %s", c.in, got, c.want)
		}
	}
}

func TestCreatePageErrorError(t *testing.T) {
	err := CreatePageError{
		Filename: "/tmp/somefilename.html",
		Op:       "execute",
		Err:      errors.New("can't write to file"),
	}
	want := "can't create html page for filename: /tmp/somefilename.html, can't write to file (execute)"
	got := err.Error()
	if got != want {
		t.Errorf("Error - got: %s, want: %s", got, want)
	}
}

/***********************************************
 *   Helper Functions
 ***********************************************/

func checkFilesExist(files []string) error {
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("file doesn't exist: %s", f)
		}
	}
	return nil
}

func removeFiles(files []string) error {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
