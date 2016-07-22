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
		{filepath.Join("fixtures", "config_onemaxnumprocesses.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://example.com/rulehuntersrv/src",
				NumRulesInReport: 100,
				MaxNumProcesses:  1,
				MaxNumRecords:    -1,
			},
		},
		{filepath.Join("fixtures", "config_somemaxnumrecords.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://example.com/rulehuntersrv/src",
				NumRulesInReport: 100,
				MaxNumProcesses:  4,
				MaxNumRecords:    150,
			},
		},
		{filepath.Join("fixtures", "config_zeromaxnumrecords.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://example.com/rulehuntersrv/src",
				NumRulesInReport: 100,
				MaxNumProcesses:  4,
				MaxNumRecords:    -1,
			},
		},
		{filepath.Join("fixtures", "config.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://example.com/rulehuntersrv/src",
				NumRulesInReport: 100,
				MaxNumProcesses:  4,
				MaxNumRecords:    -1,
			},
		},
		{filepath.Join("fixtures", "config_nomaxnumprocesses.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://example.com/rulehuntersrv/src",
				NumRulesInReport: 100,
				MaxNumProcesses:  runtime.NumCPU(),
				MaxNumRecords:    -1,
			},
		},
		{filepath.Join("fixtures", "config_nonumrulesinreport.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://example.com/rulehuntersrv/src",
				NumRulesInReport: 100,
				MaxNumProcesses:  4,
				MaxNumRecords:    -1,
			},
		},
		{filepath.Join("fixtures", "config_nosourceurl.json"),
			&Config{
				ExperimentsDir:   "experiments",
				WWWDir:           "www",
				BuildDir:         "build",
				SourceURL:        "https://github.com/vlifesystems/rulehuntersrv",
				NumRulesInReport: 100,
				MaxNumProcesses:  4,
				MaxNumRecords:    -1,
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
			t.Errorf("Load(%s) got: %s, want: %s", c.filename, gotConfig, c.wantConfig)
			continue
		}
	}
}

func TestLoad_errors(t *testing.T) {
	cases := []struct {
		filename string
		wantErr  error
	}{
		{filepath.Join("fixtures", "config_nonexistant.json"),
			&os.PathError{
				"open",
				filepath.Join("fixtures", "config_nonexistant.json"),
				syscall.ENOENT,
			},
		},
		{filepath.Join("fixtures", "config_noexperimentsdir.json"),
			errors.New("missing field: experimentsDir")},
		{filepath.Join("fixtures", "config_nowwwdir.json"),
			errors.New("missing field: wwwDir")},
		{filepath.Join("fixtures", "config_nobuilddir.json"),
			errors.New("missing field: buildDir")},
		{filepath.Join("fixtures", "config_invalidjson.json"),
			errors.New("invalid character '\"' after object key:value pair")},
	}

	for _, c := range cases {
		_, gotErr := Load(c.filename)
		if err := checkErrorMatch(gotErr, c.wantErr); err != nil {
			t.Errorf("Load(%s) %s", c.filename, err)
			return
		}
	}
}

/*****************************
 *   Helper functions
 *****************************/
func configsMatch(c1, c2 *Config) bool {
	return c1.ExperimentsDir == c2.ExperimentsDir &&
		c1.WWWDir == c2.WWWDir &&
		c1.BuildDir == c2.BuildDir &&
		c1.NumRulesInReport == c2.NumRulesInReport &&
		c1.MaxNumProcesses == c2.MaxNumProcesses &&
		c1.MaxNumRecords == c2.MaxNumRecords
}

func checkErrorMatch(got, want error) error {
	if perr, ok := want.(*os.PathError); ok {
		return checkPathErrorMatch(got, perr)
	}
	if got.Error() != want.Error() {
		return fmt.Errorf("got err: %v, want err: %v", got, want)
	}
	return nil
}

func checkPathErrorMatch(
	checkErr error,
	wantErr *os.PathError,
) error {
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
