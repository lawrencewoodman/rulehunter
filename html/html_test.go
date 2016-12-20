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
	"strconv"
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
	cfgDir := testhelpers.BuildConfigDirs(t)
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

func TestGenReportURLDir(t *testing.T) {
	cases := []struct {
		stamp   time.Time
		title   string
		wantDir string
	}{
		{time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			"This could be very interesting",
			fmt.Sprintf("/reports/2009/11/10/%s_this-could-be-very-interesting/",
				genStampMagicString(
					time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
				),
			),
		},
	}
	for _, c := range cases {
		got := genReportURLDir(c.stamp, c.title)
		if got != c.wantDir {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.stamp, c.title, got, c.wantDir)
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

func TestGenStampMagicString(t *testing.T) {
	cases := []struct {
		in       time.Time
		wantDiff uint64
	}{
		{time.Date(2009, time.November, 10, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2009, time.November, 11, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2009, time.December, 11, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2010, time.December, 11, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2009, time.November, 10, 22, 19, 19, 17, time.UTC), 1},
		{time.Date(2009, time.November, 10, 22, 19, 29, 17, time.UTC), 11},
		{time.Date(2009, time.November, 10, 22, 20, 18, 17, time.UTC), 60},
		{time.Date(2009, time.November, 10, 23, 19, 18, 17, time.UTC), 3600},
	}

	initStamp := time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC)
	initMagicStr := genStampMagicString(initStamp)

	initMagicNum, err := strconv.ParseUint(initMagicStr, 36, 64)
	if err != nil {
		t.Errorf("ParseUint(%s, 36, 64) err: %s", initMagicStr, err)
		return
	}

	for _, c := range cases {
		magicStr := genStampMagicString(c.in)
		magicNum, err := strconv.ParseUint(magicStr, 36, 64)
		if err != nil {
			t.Errorf("ParseUint(%s, 36, 64) err: %s", magicStr, err)
			return
		}
		diff := magicNum - initMagicNum
		if diff != c.wantDiff {
			t.Errorf("diff != wantDiff for stamp: %s got: %d, want: %d",
				c.in, diff, c.wantDiff)
		}
	}
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
