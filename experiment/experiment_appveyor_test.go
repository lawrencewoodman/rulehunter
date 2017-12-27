// Test database access under AppVeyor
// +build appveyor

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

func TestMakeDataset_appveyor(t *testing.T) {
	cases := []struct {
		desc           *datasetDesc
		dataSourceName string
		query          string
		fields         []string
		want           ddataset.Dataset
	}{
		{desc: &datasetDesc{
			SQL: &sqlDesc{
				DriverName:     "mssql",
				DataSourceName: "Server=127.0.0.1;Port=1433;Database=master;UID=sa,PWD=Password12!",
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
	}
	tmpDir := testhelpers.BuildConfigDirs(t, true)
	defer os.RemoveAll(tmpDir)
	cfg := &config.Config{
		MaxNumRecords: -1,
		BuildDir:      filepath.Join(tmpDir, "build"),
	}
	for i, c := range cases {
		got, err := makeDataset("trainDataset", cfg, c.fields, c.desc)
		if err != nil {
			t.Errorf("(%d) makeDataset: %s", i, err)
		} else if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("(%d) checkDatasetsEqual: %s", i, err)
		}
	}
}
