package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestRunRoot(t *testing.T) {
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
	if testing.Short() {
		testhelpers.MustWriteConfig(t, cfgDir, 100)
	} else {
		testhelpers.MustWriteConfig(t, cfgDir, 2000)
	}

	experimentFiles := []string{
		"0debt_broken.yaml",
		"debt.json",
		"debt.yaml",
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

	l := testhelpers.NewLogger()
	if err := runRoot(l, cfgFilename); err != nil {
		t.Errorf("runRoot: %s", err)
	}
	gotReportFiles := testhelpers.GetFilesInDir(
		t,
		filepath.Join(cfgDir, "build", "reports"),
	)
	if !reflect.DeepEqual(gotReportFiles, wantReportFiles) {
		t.Errorf("GetFilesInDir - got: %v\n want: %v",
			gotReportFiles, wantReportFiles)
	}

	if !reflect.DeepEqual(l.GetEntries(), wantEntries) {
		t.Errorf("GetEntries() got: %v\n want: %v", l.GetEntries(), wantEntries)
	}
	// TODO: Test all files generated
}

func TestRunRoot_errors(t *testing.T) {
	tmpDir := testhelpers.BuildConfigDirs(t, false)
	defer os.RemoveAll(tmpDir)

	cases := []struct {
		configFilename string
		wantErr        error
	}{
		{
			configFilename: "config.yaml",
			wantErr: errConfigLoad{
				filename: "config.yaml",
				err:      &os.PathError{"open", "config.yaml", syscall.ENOENT},
			},
		},
		{
			configFilename: filepath.Join(tmpDir, "config.yaml"),
			wantErr: errConfigLoad{
				filename: filepath.Join(tmpDir, "config.yaml"),
				err: &os.PathError{
					"open",
					filepath.Join(tmpDir, "config.yaml"),
					syscall.ENOENT,
				},
			},
		},
	}

	for i, c := range cases {
		l := testhelpers.NewLogger()
		err := runRoot(l, c.configFilename)
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("(%d) runRoot: %s", i, err)
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
