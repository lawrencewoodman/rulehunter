// Test database access under Travis
// +build travis

package experiment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestMakeDataset_travis(t *testing.T) {
	cases := []struct {
		desc   *datasetDesc
		fields []string
		want   ddataset.Dataset
	}{
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "mysql",
				DataSourceName: "travis@/master",
				Query:          "select * from flow",
			},
		},
			fields: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "mysql",
				DataSourceName: "travis@/master",
				Query:          "select grp,district,flow from flow",
			},
		},
			fields: []string{"grp", "district", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "postgres",
				DataSourceName: "user=postgres",
				Query:          "select * from \"flow\"",
			},
		},
			fields: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "postgres",
				DataSourceName: "user=postgres",
				Query:          "select grp,district,flow from \"flow\"",
			},
		},
			fields: []string{"grp", "district", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
	}
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cfg := &config.Config{
		MaxNumRecords: -1,
		BuildDir:      filepath.Join(tmpDir, "build"),
	}
	for i, c := range cases {
		got, err := makeDataset("train", cfg, c.fields, c.desc)
		if err != nil {
			t.Errorf("(%d) makeDataset: %s", i, err)
		} else if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("checkDatasetsEqual: err: %s", err)
		}
	}
}
