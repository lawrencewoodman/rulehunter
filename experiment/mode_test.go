package experiment

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/lawrencewoodman/ddataset/dtruncate"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestShouldProcessMode(t *testing.T) {
	funcs := map[string]dexpr.CallFun{}
	cases := []struct {
		file       fileinfo.FileInfo
		when       *dexpr.Expr
		isFinished bool
		pmStamp    time.Time
		want       bool
	}{
		{file: testhelpers.NewFileInfo("bank-divorced.json", time.Now()),
			when:       dexpr.MustNew("!hasRun", funcs),
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00"),
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-divorced.json",
			testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00")),
			when:       dexpr.MustNew("!hasRun", funcs),
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-04T14:53:00.570347516+01:00"),
			want: false,
		},
		{file: testhelpers.NewFileInfo("bank-tiny.json", time.Now()),
			when:       dexpr.MustNew("!hasRun", funcs),
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00"),
			want: true,
		},
		{file: testhelpers.NewFileInfo("bank-tiny.json",
			testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00")),
			when:       dexpr.MustNew("!hasRun", funcs),
			isFinished: true,
			pmStamp: testhelpers.MustParse(time.RFC3339Nano,
				"2016-05-05T09:37:58.220312223+01:00"),
			want: false,
		},
		{file: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			when:       dexpr.MustNew("!hasRun", funcs),
			isFinished: false,
			pmStamp:    time.Now(),
			want:       true,
		},
		{file: testhelpers.NewFileInfo("bank-full-divorced.json", time.Now()),
			when:       dexpr.MustNew("!hasRun", funcs),
			isFinished: false,
			pmStamp:    time.Now(),
			want:       true,
		},
	}

	for i, c := range cases {
		got, err := shouldProcessMode(c.when, c.file, c.isFinished, c.pmStamp)
		if err != nil {
			t.Errorf("(%d) shouldProcessMode: %s", i, err)
			continue
		}
		if got != c.want {
			t.Errorf("(%d) shouldProcessMode, got: %t, want: %t", i, got, c.want)
		}
	}
}

func TestMakeDataset(t *testing.T) {
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cases := []struct {
		desc           *datasetDesc
		dataSourceName string
		query          string
		config         *config.Config
		want           ddataset.Dataset
	}{
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select * from flow",
			},
			Fields: []string{"grp", "district", "height", "flow"},
		},
			config: &config.Config{
				MaxNumRecords: -1,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select * from flow",
			},
			Fields: []string{"grp", "district", "height", "flow"},
		},
			config: &config.Config{
				MaxNumRecords: 1000,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select * from flow",
			},
			Fields: []string{"grp", "district", "height", "flow"},
		},
			config: &config.Config{
				MaxNumRecords: 4,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dtruncate.New(
				dcsv.New(
					filepath.Join("fixtures", "flow.csv"),
					true,
					rune(','),
					[]string{"grp", "district", "height", "flow"},
				),
				4,
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "sqlite3",
				DataSourceName: filepath.Join("fixtures", "flow.db"),
				Query:          "select grp,district,flow from flow",
			},
			Fields: []string{"grp", "district", "flow"},
		},
			config: &config.Config{
				MaxNumRecords: -1,
				BuildDir:      filepath.Join(tmpDir, "build"),
			},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
	}
	for i, c := range cases {
		got, err := makeDataset(c.config, c.desc)
		if err != nil {
			t.Errorf("(%d) makeDataset: %s", i, err)
		} else if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("(%d) checkDatasetsEqual: %s", i, err)
		}
	}
}

func TestMakeDataset_err(t *testing.T) {
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cases := []struct {
		desc              *datasetDesc
		wantOpenErrRegexp *regexp.Regexp
	}{
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "mysql",
				DataSourceName: "invalid:invalid@tcp(127.0.0.1:9999)/master",
				Query:          "select * from invalid",
			},
			Fields: []string{},
		},
			wantOpenErrRegexp: regexp.MustCompile("^dial tcp 127.0.0.1:9999.*?connection.*?refused.*$"),
		},
		{desc: &datasetDesc{
			CSV: &csvDesc{
				Filename:  filepath.Join("fixtures", "nonexistant.csv"),
				HasHeader: false,
				Separator: ",",
			},
			Fields: []string{},
		},
			wantOpenErrRegexp: regexp.MustCompile(
				// Replace used because in Windows the backslash in the path is
				// altering the meaning of the regexp
				strings.Replace(
					fmt.Sprintf(
						"^%s$",
						&os.PathError{
							Op:   "open",
							Path: filepath.Join("fixtures", "nonexistant.csv"),
							Err:  syscall.ENOENT,
						},
					),
					"\\",
					"\\\\",
					-1,
				),
			),
		},
	}
	cfg := &config.Config{
		MaxNumRecords: -1,
		BuildDir:      filepath.Join(tmpDir, "build"),
	}
	for i, c := range cases {
		_, err := makeDataset(cfg, c.desc)
		if !c.wantOpenErrRegexp.MatchString(err.Error()) {
			t.Fatalf("(%d) makeDataset: %s", i, err)
		}
	}
}
