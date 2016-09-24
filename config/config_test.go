package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		filename   string
		wantConfig *Config
	}{
		{filepath.Join("fixtures", "config_onemaxnumprocesses.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				User:              "",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   1,
				MaxNumRecords:     -1,
			},
		},
		{filepath.Join("fixtures", "config_somemaxnumrecords.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				User:              "",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   4,
				MaxNumRecords:     150,
			},
		},
		{filepath.Join("fixtures", "config_zeromaxnumrecords.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   4,
				MaxNumRecords:     -1,
			},
		},
		{filepath.Join("fixtures", "config.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				User:              "rhuser",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   4,
				MaxNumRecords:     -1,
			},
		},
		{filepath.Join("fixtures", "config_nomaxnumprocesses.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   runtime.NumCPU(),
				MaxNumRecords:     -1,
			},
		},
		{filepath.Join("fixtures", "config_nomaxnumreportrules.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				MaxNumReportRules: 100,
				MaxNumProcesses:   4,
				MaxNumRecords:     -1,
			},
		},
		{filepath.Join("fixtures", "config_nosourceurl.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://github.com/vlifesystems/rulehuntersrv",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   4,
				MaxNumRecords:     -1,
			},
		},
		{filepath.Join("fixtures", "config.yaml"),
			&Config{
				ExperimentsDir:    "experiments",
				WWWDir:            "www",
				BuildDir:          "build",
				SourceURL:         "https://example.com/rulehuntersrv/src",
				User:              "rhuser",
				MaxNumReportRules: 2000,
				MaxNumProcesses:   4,
				MaxNumRecords:     -1,
			},
		},
	}

	for _, c := range cases {
		gotConfig, err := Load(c.filename)
		if err != nil {
			t.Errorf("Load(%s) err: %s", c.filename, err)
			continue
		}
		if !configsMatch(gotConfig, c.wantConfig) {
			t.Errorf("Load(%s) got: %v, want: %v", c.filename, gotConfig, c.wantConfig)
		}
	}
}

func TestLoad_errors(t *testing.T) {
	cases := []struct {
		filename string
		wantErr  error
	}{
		{filepath.Join("fixtures", "config_nonexistant.yaml"),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "config_nonexistant.yaml"),
				syscall.ENOENT,
			},
		},
		{filepath.Join("fixtures", "config_noexperimentsdir.yaml"),
			errors.New("missing field: experimentsDir")},
		{filepath.Join("fixtures", "config_nowwwdir.yaml"),
			errors.New("missing field: wwwDir")},
		{filepath.Join("fixtures", "config_nobuilddir.yaml"),
			errors.New("missing field: buildDir")},
		{filepath.Join("fixtures", "config.json"), InvalidExtError(".json")},
		{filepath.Join("fixtures", "config_nonexistant.yaml"),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "config_nonexistant.yaml"),
				syscall.ENOENT,
			},
		},
		{filepath.Join("fixtures", "config_invalidyaml.yaml"),
			errors.New("yaml: line 2: did not find expected key")},
	}

	for _, c := range cases {
		_, gotErr := Load(c.filename)
		if err := checkErrorMatch(gotErr, c.wantErr); err != nil {
			t.Errorf("Load(%s) %s", c.filename, err)
			return
		}
	}
}

func TestInvalidExtErrorError(t *testing.T) {
	ext := ".exe"
	err := InvalidExtError(ext)
	want := "invalid extension: " + ext
	if got := err.Error(); got != want {
		t.Errorf("Error() got: %v, want: %v", got, want)
	}
}

/*****************************
 *   Helper functions
 *****************************/
func configsMatch(c1, c2 *Config) bool {
	return c1.ExperimentsDir == c2.ExperimentsDir &&
		c1.WWWDir == c2.WWWDir &&
		c1.BuildDir == c2.BuildDir &&
		c1.SourceURL == c2.SourceURL &&
		c1.User == c2.User &&
		c1.MaxNumReportRules == c2.MaxNumReportRules &&
		c1.MaxNumProcesses == c2.MaxNumProcesses &&
		c1.MaxNumRecords == c2.MaxNumRecords
}

func checkErrorMatch(got, want error) error {
	if perr, ok := want.(*os.PathError); ok {
		return checkPathErrorMatch(got, perr)
	}
	if ieErr, ok := want.(InvalidExtError); ok {
		return checkInvalidExtErrorMatch(got, ieErr)
	}
	if got.Error() != want.Error() {
		return fmt.Errorf("got err: %v, want err: %v", got, want)
	}
	return nil
}

func checkInvalidExtErrorMatch(checkErr error, wantErr InvalidExtError) error {
	ieErr, ok := checkErr.(InvalidExtError)
	if !ok {
		return fmt.Errorf("got err type: %T, want error type: InvalidExtError",
			checkErr)
	}
	if string(ieErr) != string(wantErr) {
		return fmt.Errorf("InvalidExtError got ext: %s, want: %s",
			string(ieErr), string(wantErr))
	}
	return nil
}

func checkPathErrorMatch(checkErr error, wantErr *os.PathError) error {
	perr, ok := checkErr.(*os.PathError)
	if !ok {
		return fmt.Errorf("got err type: %T, want error type: os.PathError",
			checkErr)
	}
	if perr.Op != wantErr.Op {
		return fmt.Errorf("got perr.Op: %s, want: %s", perr.Op, wantErr.Op)
	}
	if filepath.Clean(perr.Path) != filepath.Clean(wantErr.Path) {
		return fmt.Errorf("got perr.Path: %s, want: %s", perr.Path, wantErr.Path)
	}
	if perr.Err != wantErr.Err {
		return fmt.Errorf("got perr.Err: %s, want: %s", perr.Err, wantErr.Err)
	}
	return nil
}
