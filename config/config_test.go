package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
			},
		},
	}

	for _, c := range cases {
		gotConfig, err := Load(c.filename)
		if err != nil {
			t.Errorf("Load(%s) err: %s", c.filename, err)
			return
		}
		if !configsMatch(gotConfig, c.wantConfig) {
			t.Errorf("Load(%s) got: %s, want: %s", c.filename, gotConfig, c.wantConfig)
			return
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
				errors.New("no such file or directory"),
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
		if gotErr == nil || gotErr.Error() != c.wantErr.Error() {
			t.Errorf("Load(%s) gotErr: %s, wantErr: %s",
				c.filename, gotErr, c.wantErr)
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
		c1.MaxNumProcesses == c2.MaxNumProcesses
}
