package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"
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
		exitCode, err := subMain(c.flags)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%q) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%q) %s", c.flags, err)
		}
	}
}

/*************************************
 *  Helper functions
 *************************************/

func checkErrorMatch(got, want error) error {
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
