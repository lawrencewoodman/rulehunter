package html

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

// This checks if Run will quit properly when told to
func TestRun_quit(t *testing.T) {
	q := quitter.New()
	l := testhelpers.NewLogger()
	htmlCmds := make(chan cmd.Cmd)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Getwd() err: ", err)
	}
	defer os.Chdir(wd)
	cfgDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(cfgDir)

	pm, err := progress.NewMonitor(
		filepath.Join(cfgDir, "build", "progress"),
		htmlCmds,
	)
	if err != nil {
		t.Fatalf("NewMonitor: %s", err)
	}
	config := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumReportRules: 100,
	}
	h := New(config, pm, l, htmlCmds)
	go h.Run(q)
	for !h.Running() {
	}

	flushC := time.NewTimer(time.Second).C
	quitC := time.NewTimer(2 * time.Second).C
	isRunningC := time.NewTicker(100 * time.Millisecond).C
	timeoutC := time.NewTimer(5 * time.Second).C
	for {
		select {
		case <-flushC:
			htmlCmds <- cmd.Flush
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

// Tests Run for cmd.All
func TestRun_cmd_all(t *testing.T) {
	q := quitter.New()
	defer q.Quit()
	l := testhelpers.NewLogger()
	htmlCmds := make(chan cmd.Cmd)
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
		htmlCmds,
	)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumReportRules: 100,
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
			"index.html",
		),
		filepath.Join(
			cfgDir,
			"www",
			"reports",
			"category",
			"groupb",
			"how-to-make-a-profit",
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

	h := New(config, pm, l, htmlCmds)
	go h.Run(q)

	var allC <-chan time.Time
	timeoutC := time.NewTimer(5 * time.Second).C
	ticker := time.NewTicker(time.Millisecond * 100).C
	filesRemoved := false
	for {
		select {
		case <-allC:
			htmlCmds <- cmd.All
		case <-ticker:
			if err := checkFilesExist(wantFiles); err == nil {
				if !filesRemoved {
					// Files are first removed because cmd.All is passed at start
					if err := removeFiles(wantFiles); err != nil {
						t.Fatalf("removeFiles: %s", err)
					}
					filesRemoved = true
					allC = time.NewTimer(time.Second).C
				} else {
					return
				}
			}
		case <-timeoutC:
			if err := checkFilesExist(wantFiles); err != nil {
				t.Fatalf("Run: %s, log: %v", err, l)
			}
			return
		}
	}
}

// Tests Run for cmd.Reports
func TestRun_cmd_reports(t *testing.T) {
	q := quitter.New()
	defer q.Quit()
	l := testhelpers.NewLogger()
	htmlCmds := make(chan cmd.Cmd)
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
		htmlCmds,
	)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumReportRules: 100,
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
			"index.html",
		),
		filepath.Join(
			cfgDir,
			"www",
			"reports",
			"category",
			"groupb",
			"how-to-make-a-profit",
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

	h := New(config, pm, l, htmlCmds)
	go h.Run(q)

	var reportsC <-chan time.Time
	timeoutC := time.NewTimer(5 * time.Second).C
	ticker := time.NewTicker(time.Millisecond * 100).C
	filesRemoved := false
	for {
		select {
		case <-reportsC:
			htmlCmds <- cmd.Reports
		case <-ticker:
			if err := checkFilesExist(wantFiles); err == nil {
				if !filesRemoved {
					// Files are first removed because cmd.All is passed at start
					if err := removeFiles(wantFiles); err != nil {
						t.Fatalf("removeFiles: %s", err)
					}
					filesRemoved = true
					reportsC = time.NewTimer(time.Second).C
				} else {
					return
				}
			}
		case <-timeoutC:
			if err := checkFilesExist(wantFiles); err != nil {
				t.Fatalf("Run: %s", err)
			}
			return
		}
	}
}

// Tests Run for cmd.Progress
func TestRun_cmd_progress(t *testing.T) {
	q := quitter.New()
	defer q.Quit()
	l := testhelpers.NewLogger()
	htmlCmds := make(chan cmd.Cmd)
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
		htmlCmds,
	)
	if err != nil {
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumReportRules: 100,
	}
	wantFiles := []string{
		filepath.Join(cfgDir, "www", "activity", "index.html"),
	}

	h := New(config, pm, l, htmlCmds)
	go h.Run(q)

	var progressC <-chan time.Time
	timeoutC := time.NewTimer(5 * time.Second).C
	ticker := time.NewTicker(time.Millisecond * 100).C
	filesRemoved := false
	for {
		select {
		case <-progressC:
			htmlCmds <- cmd.Progress
		case <-ticker:
			if err := checkFilesExist(wantFiles); err == nil {
				if !filesRemoved {
					// Files are first removed because cmd.All is passed at start
					if err := removeFiles(wantFiles); err != nil {
						t.Fatalf("removeFiles: %s", err)
					}
					filesRemoved = true
					progressC = time.NewTimer(time.Second).C
				} else {
					return
				}
			}
		case <-timeoutC:
			if err := checkFilesExist(wantFiles); err != nil {
				t.Fatalf("Run: %s", err)
			}
			return
		}
	}
}

func TestGenReportURLDir(t *testing.T) {
	cases := []struct {
		category string
		title    string
		want     string
	}{
		{category: "",
			title: "This could be very interesting",
			want:  "reports/nocategory/this-could-be-very-interesting/",
		},
		{category: "acme or emca",
			title: "This could be very interesting",
			want:  "reports/category/acme-or-emca/this-could-be-very-interesting/",
		},
	}
	for _, c := range cases {
		got := genReportURLDir(c.category, c.title)
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
