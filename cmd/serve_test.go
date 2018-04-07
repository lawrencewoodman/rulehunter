package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/vlifesystems/rulehunter/internal"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/quitter"
)

func TestRunServe(t *testing.T) {
	wantEntries := []testhelpers.Entry{
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 4: did not find expected key"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.json, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.json, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.yaml, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.yaml, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt2.json, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt2.json, mode: train"},
	}
	wantReportFiles := []string{
		// "debt2.json"
		internal.MakeBuildFilename(
			"train",
			"",
			"What is most likely to indicate success (2)",
		),
		// "debt.yaml"
		internal.MakeBuildFilename(
			"train",
			"testing",
			"What is most likely to indicate success",
		),
		// "debt.json"
		internal.MakeBuildFilename(
			"train",
			"",
			"What is most likely to indicate success",
		),
	}

	cfgDir := testhelpers.BuildConfigDirs(t, false)
	cfgFilename := filepath.Join(cfgDir, "config.yaml")
	defer os.RemoveAll(cfgDir)
	testhelpers.MustWriteConfig(t, cfgDir, 10)
	l := testhelpers.NewLogger()
	q := quitter.New()

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

	hasQuit := false
	tickerC := time.NewTicker(100 * time.Millisecond).C
	timeoutC := time.NewTimer(30 * time.Second).C
	for !hasQuit {
		select {
		case <-tickerC:
			gotReportFiles := testhelpers.GetFilesInDir(
				t,
				filepath.Join(cfgDir, "build", "reports"),
			)
			if len(gotReportFiles) == 3 {
				q.Quit()
				hasQuit = true
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
	if err := doLogEntriesMatch(gotEntries, wantEntries); err != nil {
		t.Errorf("GetEntries: %s", err)
	}
	// TODO: Test all files generated
}

func TestRunServe_http(t *testing.T) {
	httpPort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort: %d", err)
	}
	wantEntries := []testhelpers.Entry{
		{Level: testhelpers.Info,
			Msg: fmt.Sprintf("Starting http server on port: %d", httpPort)},
		{Level: testhelpers.Info,
			Msg: fmt.Sprintf("Shutdown http server on port: %d", httpPort)},
		{Level: testhelpers.Error,
			Msg: "Can't load experiment: 0debt_broken.yaml, yaml: line 4: did not find expected key"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.json, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.json, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt.yaml, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt.yaml, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Processing experiment: debt2.json, mode: train"},
		{Level: testhelpers.Info,
			Msg: "Successfully processed experiment: debt2.json, mode: train"},
	}

	cfgDir := testhelpers.BuildConfigDirs(t, false)
	cfgFilename := filepath.Join(cfgDir, "config.yaml")
	defer os.RemoveAll(cfgDir)
	testhelpers.MustWriteConfig(t, cfgDir, 10, httpPort)
	l := testhelpers.NewLogger()
	q := quitter.New()

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

	wantText := "Rulehunter"
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", httpPort))
	if err != nil {
		q.Quit()
		t.Fatalf("http.Get: %s", err)
	}
	if resp.StatusCode != 200 {
		q.Quit()
		t.Fatalf("received non-200 response: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		q.Quit()
		t.Fatal(err)
	}
	if !strings.Contains(string(body[:]), wantText) {
		t.Errorf("web page doesn't contain: %s", wantText)
	}

	hasQuit := false
	tickerC := time.NewTicker(100 * time.Millisecond).C
	timeoutC := time.NewTimer(30 * time.Second).C
	for !hasQuit {
		select {
		case <-tickerC:
			gotReportFiles := testhelpers.GetFilesInDir(
				t,
				filepath.Join(cfgDir, "build", "reports"),
			)
			if len(gotReportFiles) == 3 {
				q.Quit()
				hasQuit = true
				break
			}
		case <-timeoutC:
			q.Quit()
			t.Fatal("runServe hasn't stopped")
		}
	}

	time.Sleep(4 * time.Second)

	gotEntries := l.GetEntries()
	if err := doLogEntriesMatch(gotEntries, wantEntries); err != nil {
		t.Errorf("GetEntries: %s", err)
	}
	// TODO: Test all files generated
}
