// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/quitter"
)

func TestRunServe_interrupt(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, false)
	cfgFilename := filepath.Join(cfgDir, "config.yaml")
	defer os.RemoveAll(cfgDir)
	if testing.Short() {
		testhelpers.MustWriteConfig(t, cfgDir, 100)
	} else {
		testhelpers.MustWriteConfig(t, cfgDir, 2000)
	}
	l := testhelpers.NewLogger()
	q := quitter.New()
	defer q.Quit()

	go func() {
		if err := runServe(l, q, cfgFilename); err != nil {
			t.Errorf("runServe: %s", err)
		}
	}()

	if !testing.Short() {
		time.Sleep(2 * time.Second)
	}

	experimentFiles := []string{
		"debt.json",
		"debt.yaml",
		"0debt_broken.yaml",
		"debt2.json",
		"debt.jso",
	}
	for _, f := range experimentFiles {
		testhelpers.CopyFile(
			t,
			filepath.Join("fixtures", f),
			filepath.Join(cfgDir, "experiments"),
		)
	}

	hasInterrupted := false
	tickerC := time.NewTicker(400 * time.Millisecond).C
	timeoutC := time.NewTimer(20 * time.Second).C
	for !hasInterrupted {
		select {
		case <-tickerC:
			gotReportFiles := testhelpers.GetFilesInDir(
				t,
				filepath.Join(cfgDir, "build", "reports"),
			)
			if len(gotReportFiles) >= 1 {
				interruptProcess(t)
				hasInterrupted = true
				break
			}
		case <-timeoutC:
			t.Fatal("runServe hasn't stopped")
		}
	}

	time.Sleep(4 * time.Second)

	gotReportFiles := testhelpers.GetFilesInDir(
		t,
		filepath.Join(cfgDir, "build", "reports"),
	)
	if len(gotReportFiles) != 1 {
		t.Errorf("runServe - gotReportFiles: %v, len(want): 1", gotReportFiles)
	}
}
