// Test database access under AppVeyor
// +build appveyor

package experiment

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"path/filepath"
	"testing"
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
	for i, c := range cases {
		got, err := makeDataset("trainDataset", c.fields, c.desc)
		if err != nil {
			t.Errorf("(%d) makeDataset: %s", i, err)
		} else if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("(%d) checkDatasetsEqual: %s", i, err)
		}
	}
}
