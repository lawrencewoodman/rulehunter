// +build su

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/cmd"
	"github.com/vlifesystems/rulehunter/internal"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestMain(m *testing.M) {
	if len(os.Args) >= 2 && (os.Args[1] == "serve" || os.Args[1] == "service") {
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--config") {
				cfgFilename := strings.Split(arg, "=")[1]
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
		}
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestRulehunter_service_install(t *testing.T) {
	for _, user := range knownUsers {
		t.Logf("user: %s", user)
		cfgDir := testhelpers.BuildConfigDirs(t, false)
		defer os.RemoveAll(cfgDir)
		testhelpers.MustWriteConfig(t, cfgDir, 10)

		if user != "" {
			runOSCmd(t,
				true,
				os.Args[0],
				"service",
				"install",
				fmt.Sprintf("--config=%s", filepath.Join(cfgDir, "config.yaml")),
				fmt.Sprintf("--user=%s", user),
			)
		} else {
			runOSCmd(t,
				true,
				os.Args[0],
				"service",
				"install",
				fmt.Sprintf("--config=%s", filepath.Join(cfgDir, "config.yaml")),
			)
		}

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

		wantReportFiles := []string{
			// "debt2_datasets.json"
			internal.MakeBuildFilename(
				"",
				"What is most likely to indicate success (2)",
			),
			// "debt_datasets.yaml"
			internal.MakeBuildFilename(
				"testing",
				"What is most likely to indicate success",
			),
			// "debt_datasets.json"
			internal.MakeBuildFilename("", "What is most likely to indicate success"),
		}
		isFinished := false
		files := []string{}
		timeoutC := time.NewTimer(20 * time.Second).C
		tickerC := time.NewTicker(400 * time.Millisecond).C
		for !isFinished {
			select {
			case <-tickerC:
				files = testhelpers.GetFilesInDir(
					t,
					filepath.Join(cfgDir, "build", "reports"),
				)
				if reflect.DeepEqual(files, wantReportFiles) {
					isFinished = true
					break
				}
			case <-timeoutC:
				t.Errorf("(user: %s) didn't generate correct files within time period, got: %v, want: %v",
					user, files, wantReportFiles)
				isFinished = true
				break
			}
		}
		stopService(t, "rulehunter")
	}
}

func TestRulehunter_service_uninstall(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, false)
	defer os.RemoveAll(cfgDir)
	testhelpers.MustWriteConfig(t, cfgDir, 10)
	runOSCmd(t,
		true,
		os.Args[0],
		"service",
		"uninstall",
		fmt.Sprintf("--config=%s", filepath.Join(cfgDir, "config.yaml")),
	)
	runOSCmd(t,
		true,
		os.Args[0],
		"service",
		"install",
		fmt.Sprintf("--config=%s", filepath.Join(cfgDir, "config.yaml")),
	)

	startService(t, "rulehunter")
	defer stopService(t, "rulehunter")
	runOSCmd(t,
		true,
		os.Args[0],
		"service",
		"uninstall",
		fmt.Sprintf("--config=%s", filepath.Join(cfgDir, "config.yaml")),
	)
}
