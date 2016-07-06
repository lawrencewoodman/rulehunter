package main

import (
	"fmt"
	"github.com/vlifesystems/rulehuntersrv/internal/testhelpers"
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
		l := logger.NewTestLogger()
		exitCode, err := subMain(c.flags, l)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%q) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%q) %s", c.flags, err)
		}
		if len(l.GetEntries()) != 0 {
			t.Errorf("GetEntries() got: %s, want: {}", l.GetEntries())
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
		configDir, err := testhelpers.BuildConfigDirs()
		if err != nil {
			t.Fatalf("buildConfigDirs() err: %s", err)
		}
		defer os.RemoveAll(configDir)
		c.flags.configDir = configDir

		l := logger.NewTestLogger()
		go func() {
			tryInSeconds := 4
			for i := 0; i < tryInSeconds*5; i++ {
				if reflect.DeepEqual(l.GetEntries(), c.wantEntries) {
					interruptProcess()
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
			interruptProcess()
		}()

		go func() {
			<-time.After(6 * time.Second)
			panic("Run() hasn't been stopped")
		}()
		if err := os.Chdir(configDir); err != nil {
			t.Fatalf("Chdir() err: %s", err)
		}
		exitCode, err := subMain(c.flags, l)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%q) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%q) %s", c.flags, err)
		}
		if !reflect.DeepEqual(l.GetEntries(), c.wantEntries) {
			t.Errorf("GetEntries() got: %s, want: %s", l.GetEntries(), c.wantEntries)
		}
	}
}

/*************************************
 *  Helper functions
 *************************************/

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
