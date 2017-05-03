package main

import (
	"bytes"
	"fmt"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestSubMain(t *testing.T) {
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
	cfgDir := testhelpers.BuildConfigDirs(t, false)
	flags := &cmdFlags{configDir: cfgDir}
	defer os.RemoveAll(cfgDir)
	if testing.Short() {
		mustWriteConfig(t, cfgDir, 100)
	} else {
		mustWriteConfig(t, cfgDir, 2000)
	}
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "0debt_broken.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.yaml"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt.jso"),
		filepath.Join(cfgDir, "experiments"),
	)
	testhelpers.CopyFile(
		t,
		filepath.Join("fixtures", "debt2.json"),
		filepath.Join(cfgDir, "experiments"),
	)
	l := testhelpers.NewLogger()
	exitCode, err := subMain(flags, l)
	if exitCode != 0 {
		t.Errorf("subMain(%v, l) exitCode: %d, want: %d", flags, exitCode, 0)
	}
	if err != nil {
		t.Errorf("subMain(%v, l) err: %s", flags, err)
	}
	if !reflect.DeepEqual(l.GetEntries(), wantEntries) {
		t.Errorf("GetEntries() got: %v, want: %v", l.GetEntries(), wantEntries)
	}
	// TODO: Test files generated
}

func TestSubMain_errors(t *testing.T) {
	tmpDir := testhelpers.BuildConfigDirs(t, false)
	defer os.RemoveAll(tmpDir)

	cases := []struct {
		flags        *cmdFlags
		wantErr      error
		wantExitCode int
	}{
		{
			flags: &cmdFlags{
				configDir: "",
				install:   true,
			},
			wantErr:      errNoConfigDirArg,
			wantExitCode: 1,
		},
		{
			flags: &cmdFlags{
				configDir: tmpDir,
				install:   true,
			},
			wantErr: errConfigLoad{
				filename: filepath.Join(tmpDir, "config.yaml"),
				err: &os.PathError{
					"open",
					filepath.Join(tmpDir, "config.yaml"),
					syscall.ENOENT,
				},
			},
			wantExitCode: 1,
		},
	}

	for _, c := range cases {
		l := testhelpers.NewLogger()
		exitCode, err := subMain(c.flags, l)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%v) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%v) %s", c.flags, err)
		}
		if len(l.GetEntries()) != 0 {
			t.Errorf("GetEntries() got: %s, want: {}", l.GetEntries())
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

func mustWriteConfig(t *testing.T, baseDir string, maxNumRecords int) {
	const mode = 0600
	cfg := &config.Config{
		ExperimentsDir: filepath.Join(baseDir, "experiments"),
		WWWDir:         filepath.Join(baseDir, "www"),
		BuildDir:       filepath.Join(baseDir, "build"),
		MaxNumRecords:  maxNumRecords,
	}
	cfgFilename := filepath.Join(baseDir, "config.yaml")
	y, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() err: %v", err)
	}
	if err := ioutil.WriteFile(cfgFilename, y, mode); err != nil {
		t.Fatalf("WriteFile(%s, ...) err: %v", cfgFilename, err)
	}
}

func runCmd(t *testing.T, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("runCmd(%v...), err: %v, out: %v", name, err, out.String())
	}
}
