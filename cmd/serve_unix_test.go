// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestRunServe(t *testing.T) {
	wantEntries := []testhelpers.Entry{
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 3: did not find expected key"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.json"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.yaml"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.yaml"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt2.json"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt2.json"},
	}
	wantReportFiles := []string{"debt.json", "debt.yaml", "debt2.json"}

	cfgDir := testhelpers.BuildConfigDirs(t, false)
	cfgFilename := filepath.Join(cfgDir, "config.yaml")
	defer os.RemoveAll(cfgDir)
	testhelpers.MustWriteConfig(t, cfgDir, 100)
	l := testhelpers.NewLogger()

	go func() {
		if err := runServe(l, cfgFilename); err != nil {
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
	tickerC := time.NewTicker(100 * time.Millisecond).C
	timeoutC := time.NewTimer(20 * time.Second).C
	for !hasInterrupted {
		select {
		case <-tickerC:
			gotReportFiles := testhelpers.GetFilesInDir(
				t,
				filepath.Join(cfgDir, "build", "reports"),
			)
			if len(gotReportFiles) == 3 {
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
	if !reflect.DeepEqual(gotReportFiles, wantReportFiles) {
		t.Errorf("GetFilesInDir - got: %v\n want: %v",
			gotReportFiles, wantReportFiles)
	}

	gotEntries := l.GetEntries()
	sort.Slice(gotEntries, func(i, j int) bool {
		return gotEntries[i].Msg < gotEntries[j].Msg
	})
	sort.Slice(wantEntries, func(i, j int) bool {
		return wantEntries[i].Msg < wantEntries[j].Msg
	})
	if !reflect.DeepEqual(gotEntries, wantEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v", gotEntries, wantEntries)
	}
	// TODO: Test all files generated
}

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

	go func() {
		if err := runServe(l, cfgFilename); err != nil {
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
