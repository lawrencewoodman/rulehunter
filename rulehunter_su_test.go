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

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"github.com/vlifesystems/rulehunter/logger"
)

func TestMain(m *testing.M) {
	if strings.HasPrefix(os.Args[1], "-configdir") {
		cfgDir := strings.Split(os.Args[1], "=")[1]
		flags := &cmdFlags{
			configDir: cfgDir,
			install:   len(os.Args) == 3 && os.Args[2] == "-install",
			serve:     len(os.Args) == 3 && os.Args[2] == "-serve",
		}
		l := logger.NewSvcLogger()
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("os.Getwd: %v", err)
		}
		if err := os.Chdir(cfgDir); err != nil {
			log.Fatalf("os.Chdir: %v", err)
		}
		defer os.Chdir(pwd)
		exitCode, err := subMain(flags, l)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(exitCode)
	}

	os.Exit(m.Run())
}

func TestRulehunter_service(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t)
	defer os.RemoveAll(cfgDir)
	mustWriteConfig(t, cfgDir, 10)

	runCmd(t,
		os.Args[0],
		fmt.Sprintf("-configdir=%s", cfgDir),
		"-install",
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

	testStart := time.Now()
	waitSeconds := 10.0
	gotCorrectFiles := false
	files := []string{}
	wantFiles := []string{
		"debt2_datasets.json",
		"debt_datasets.json",
		"debt_datasets.yaml",
	}
	for !gotCorrectFiles && time.Since(testStart).Seconds() < waitSeconds {
		files = getFilesInDir(t, filepath.Join(cfgDir, "build", "reports"))
		if reflect.DeepEqual(files, wantFiles) {
			gotCorrectFiles = true
		}
	}
	if !gotCorrectFiles {
		t.Errorf("didn't generate correct files with time period, got: %v, want: %v",
			files, wantFiles)
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
