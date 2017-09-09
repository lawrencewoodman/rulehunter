// +build su

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/cmd"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestMain(m *testing.M) {
	if len(os.Args) >= 2 && (os.Args[1] == "serve" || os.Args[1] == "service") {
		if len(os.Args) >= 3 && strings.HasPrefix(os.Args[2], "--config") {
			cfgFilename := strings.Split(os.Args[2], "=")[1]
			cfgDir := filepath.Dir(cfgFilename)
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("os.Getwd: %s", err)
			}
			if err := os.Chdir(cfgDir); err != nil {
				log.Fatalf("os.Chdir: %s", err)
			}
			defer os.Chdir(pwd)
		}
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestRulehunter_service(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, false)
	defer os.RemoveAll(cfgDir)
	testhelpers.MustWriteConfig(t, cfgDir, 10)

	runOSCmd(t,
		os.Args[0],
		"service",
		fmt.Sprintf("--config=%s", filepath.Join(cfgDir, "config.yaml")),
	)

	startService(t, "rulehunter")
	defer stopService(t, "rulehunter")

	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.csv"),
		filepath.Join(cfgDir, "datasets"),
	)

	if !testing.Short() {
		time.Sleep(4 * time.Second)
	}

	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_datasets.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_datasets.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt_datasets.jso"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt2_datasets.json"),
		filepath.Join(cfgDir, "experiments"),
	)

	files := []string{}
	wantFiles := []string{
		"debt2_datasets.json",
		"debt_datasets.json",
		"debt_datasets.yaml",
	}
	timeoutC := time.NewTimer(10 * time.Second).C
	tickerC := time.NewTicker(400 * time.Millisecond).C
	for {
		select {
		case <-tickerC:
			files = getFilesInDir(t, filepath.Join(cfgDir, "build", "reports"))
			if reflect.DeepEqual(files, wantFiles) {
				return
			}
		case <-timeoutC:
			t.Errorf("didn't generate correct files with time period, got: %v, want: %v",
				files, wantFiles)
			return
		}
	}
}

func getFilesInDir(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatalf("ioutil.ReadDir(%s) err: %s", dir, err)
	}

	r := []string{}
	for _, file := range files {
		if !file.IsDir() {
			r = append(r, file.Name())
		}
	}
	return r

}
