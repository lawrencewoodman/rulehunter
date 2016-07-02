package main

import (
	"errors"
	"fmt"
	"github.com/vlifesystems/rulehuntersrv/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"
)

func TestSubMain_errors(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "rulehuntersrv")
	if err != nil {
		t.Fatal("TempDir() couldn't create dir")
	}
	defer os.RemoveAll(tmpDir)

	cases := []struct {
		flags        *cmdFlags
		wantErr      error
		wantExitCode int
	}{
		{
			flags: &cmdFlags{
				user:      "fred",
				configDir: "",
				install:   true,
			},
			wantErr:      errNoConfigDirArg,
			wantExitCode: 1,
		},
		{
			flags: &cmdFlags{
				user:      "fred",
				configDir: tmpDir,
				install:   true,
			},
			wantErr: errConfigLoad{
				filename: filepath.Join(tmpDir, "config.json"),
				err: &os.PathError{
					"open",
					filepath.Join(tmpDir, "config.json"),
					syscall.ENOENT,
				},
			},
			wantExitCode: 1,
		},
	}

	for _, c := range cases {
		testLogger := logger.NewTestLogger()
		quitter := newQuitter()
		exitCode, err := subMain(c.flags, testLogger.MakeRun(), quitter)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%q) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%q) %s", c.flags, err)
		}
		if len(testLogger.GetEntries()) != 0 {
			t.Errorf("GetEntries() got: %s, want: {}", testLogger.GetEntries())
		}
	}
}

func TestSubMain(t *testing.T) {
	cases := []struct {
		flags        *cmdFlags
		wantErr      error
		wantExitCode int
		wantEntries  []logger.Entry
	}{
		{
			flags: &cmdFlags{
				user:    "fred",
				install: false,
			},
			wantErr:      nil,
			wantExitCode: 0,
			wantEntries: []logger.Entry{
				{logger.Info, "Waiting for experiments to process"},
			},
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() err: ", err)
	}
	defer os.Chdir(wd)

	for _, c := range cases {
		configDir, err := buildConfigDirs()
		if err != nil {
			t.Fatalf("buildConfigDirs() err: %s", err)
		}
		defer os.RemoveAll(c.flags.configDir)
		c.flags.configDir = configDir

		testLogger := logger.NewTestLogger()
		quitter := newQuitter()
		go func() {
			tryInSeconds := 5
			for i := 0; i < tryInSeconds*5; i++ {
				if reflect.DeepEqual(testLogger.GetEntries(), c.wantEntries) {
					quitter.Quit()
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
			quitter.Quit()
		}()
		if err := os.Chdir(configDir); err != nil {
			t.Fatalf("Chdir() err: %s", err)
		}
		exitCode, err := subMain(c.flags, testLogger.MakeRun(), quitter)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%q) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%q) %s", c.flags, err)
		}
		if !reflect.DeepEqual(testLogger.GetEntries(), c.wantEntries) {
			t.Errorf("GetEntries() got: %s, want: %s",
				testLogger.GetEntries(), c.wantEntries)
		}
	}
}

/*************************************
 *  Helper functions
 *************************************/

func buildConfigDirs() (string, error) {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700

	tmpDir, err := ioutil.TempDir("", "rulehuntersrv")
	if err != nil {
		return "", errors.New("TempDir() couldn't create dir")
	}

	// TODO: Create the www/* and build/* subdirectories from rulehuntersrv code
	subDirs := []string{
		"experiments",
		filepath.Join("www", "reports"),
		filepath.Join("www", "progress"),
		filepath.Join("build", "reports")}
	for _, subDir := range subDirs {
		fullSubDir := filepath.Join(tmpDir, subDir)
		if err := os.MkdirAll(fullSubDir, modePerm); err != nil {
			return "", fmt.Errorf("can't make directory: %s", subDir)
		}
	}

	err = copyFile(filepath.Join("fixtures", "config.json"), tmpDir)
	return tmpDir, err
}

func checkErrorMatch(got, want error) error {
	if got == nil && want == nil {
		return nil
	}
	if got == nil || want == nil {
		return fmt.Errorf("got err: %s, want err: %s", got, want)
	}
	switch x := want.(type) {
	case *os.PathError:
		return checkPathErrorMatch(got, x)
	case errConfigLoad:
		return checkErrConfigLoadMatch(got, x)
	}
	if got.Error() != want.Error() {
		return fmt.Errorf("got err: %s, want err: %s", got, want)
	}
	return nil
}

func checkPathErrorMatch(checkErr error, wantErr error) error {
	cerr, ok := checkErr.(*os.PathError)
	if !ok {
		return fmt.Errorf("got err type: %T, want error type: os.PathError",
			checkErr)
	}
	werr, ok := wantErr.(*os.PathError)
	if !ok {
		panic("wantErr isn't type *os.PathError")
	}
	if cerr.Op != werr.Op {
		return fmt.Errorf("got cerr.Op: %s, want: %s", cerr.Op, werr.Op)
	}
	if filepath.Clean(cerr.Path) != filepath.Clean(werr.Path) {
		return fmt.Errorf("got cerr.Path: %s, want: %s", cerr.Path, werr.Path)
	}
	if cerr.Err != werr.Err {
		return fmt.Errorf("got cerr.Err: %s, want: %s", cerr.Err, werr.Err)
	}
	return nil
}

func checkErrConfigLoadMatch(checkErr error, wantErr errConfigLoad) error {
	cerr, ok := checkErr.(errConfigLoad)
	if !ok {
		return fmt.Errorf("got err type: %T, want error type: errConfigLoad",
			checkErr)
	}
	if filepath.Clean(cerr.filename) != filepath.Clean(wantErr.filename) {
		return fmt.Errorf("got cerr.Path: %s, want: %s",
			cerr.filename, wantErr.filename)
	}
	return checkPathErrorMatch(cerr.err, wantErr.err)
}

func copyFile(srcFilename, dstDir string) error {
	contents, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcFilename)
	if err != nil {
		return err
	}
	mode := info.Mode()
	dstFilename := filepath.Join(dstDir, filepath.Base(srcFilename))
	return ioutil.WriteFile(dstFilename, contents, mode)
}
