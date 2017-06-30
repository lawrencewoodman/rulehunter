package html

import (
	"fmt"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// This checks if Run will quit properly when told to
func TestRun_quit(t *testing.T) {
	quit := quitter.New()
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
		t.Fatalf("NewMonitor() err: %v", err)
	}
	config := &config.Config{
		ExperimentsDir:    filepath.Join(cfgDir, "experiments"),
		WWWDir:            filepath.Join(cfgDir, "www"),
		BuildDir:          filepath.Join(cfgDir, "build"),
		MaxNumReportRules: 100,
	}
	hasQuitC := make(chan bool)
	go func() {
		Run(config, pm, l, quit, htmlCmds)
		hasQuitC <- true
	}()

	flushC := time.NewTimer(time.Second).C
	quitC := time.NewTimer(2 * time.Second).C
	timeoutC := time.NewTimer(5 * time.Second).C
	for {
		select {
		case <-flushC:
			htmlCmds <- cmd.Flush
		case <-quitC:
			quit.Quit()
		case <-timeoutC:
			t.Fatalf("Run() didn't quit")
		case <-hasQuitC:
			return
		}
	}
}

// Tests Run for cmd.Reports
func TestRun_cmd_reports(t *testing.T) {
	quit := quitter.New()
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
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "descriptions", "bank-loss.json"),
		filepath.Join(cfgDir, "build", "descriptions"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "descriptions", "bank-profit.json"),
		filepath.Join(cfgDir, "build", "descriptions"),
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
		filepath.Join(cfgDir, "www", "index.html"),
		filepath.Join(cfgDir, "www", "activity", "index.html"),
		filepath.Join(cfgDir, "www", "licence", "index.html"),
		filepath.Join(cfgDir, "www", "reports", "how-to-make-a-loss",
			"index.html"),
		filepath.Join(cfgDir, "www", "reports", "how-to-make-a-profit",
			"index.html"),
		filepath.Join(cfgDir, "www", "tag", "test",
			"index.html"),
		filepath.Join(cfgDir, "www", "tag", "bank",
			"index.html"),
		filepath.Join(cfgDir, "www", "tag", "fahrenheit-451",
			"index.html"),
		filepath.Join(cfgDir, "www", "tag", "fred-ned",
			"index.html"),
		filepath.Join(cfgDir, "www", "tag", "hot-in-the-city",
			"index.html"),
	}

	hasQuitC := make(chan bool)
	go func() {
		Run(config, pm, l, quit, htmlCmds)
		hasQuitC <- true
	}()

	reportsC := time.NewTimer(time.Second).C
	timeoutC := time.NewTimer(5 * time.Second).C
	ticker := time.NewTicker(time.Millisecond * 100).C
	for {
		select {
		case <-reportsC:
			htmlCmds <- cmd.Reports
		case <-ticker:
			if err := checkFilesExist(wantFiles); err == nil {
				quit.Quit()
			}
		case <-timeoutC:
			quit.Quit()
			if err := checkFilesExist(wantFiles); err != nil {
				t.Fatalf("Run: %s", err)
			}
		case <-hasQuitC:
			return
		}
	}
}

// Tests Run for cmd.Progress
func TestRun_cmd_progress(t *testing.T) {
	quit := quitter.New()
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
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "descriptions", "bank-loss.json"),
		filepath.Join(cfgDir, "build", "descriptions"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "descriptions", "bank-profit.json"),
		filepath.Join(cfgDir, "build", "descriptions"),
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

	hasQuitC := make(chan bool)
	go func() {
		Run(config, pm, l, quit, htmlCmds)
		hasQuitC <- true
	}()

	progressC := time.NewTimer(time.Second).C
	timeoutC := time.NewTimer(5 * time.Second).C
	ticker := time.NewTicker(time.Millisecond * 100).C
	for {
		select {
		case <-progressC:
			htmlCmds <- cmd.Progress
		case <-ticker:
			if err := checkFilesExist(wantFiles); err == nil {
				quit.Quit()
			}
		case <-timeoutC:
			quit.Quit()
			if err := checkFilesExist(wantFiles); err != nil {
				t.Fatalf("Run: %s", err)
			}
		case <-hasQuitC:
			return
		}
	}
}

func TestGenReportURLDir(t *testing.T) {
	cases := []struct {
		title   string
		wantDir string
	}{
		{"This could be very interesting",
			"reports/this-could-be-very-interesting/",
		},
	}
	for _, c := range cases {
		got := genReportURLDir(c.title)
		if got != c.wantDir {
			t.Errorf("genReportFilename(%s) got: %s, want: %s",
				c.title, got, c.wantDir)
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

/***********************************************
 *   Helper Functions
 ***********************************************/

func checkFilesExist(wantFiles []string) error {
	for _, wantFile := range wantFiles {
		if _, err := os.Stat(wantFile); os.IsNotExist(err) {
			return fmt.Errorf("file doesn't exist: %s", wantFile)
		}
	}
	return nil
}
