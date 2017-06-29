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
		instanceName string
		port         int
		query        string
		fields       []string
		want         ddataset.Dataset
	}{
		{instanceName: "SQL2014",
			port:   1433,
			query:  "select * from flow",
			fields: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
	}
	for _, c := range cases {
		e := &descFile{
			Dataset: "sql",
			Fields:  c.fields,
			Sql: &sqlDesc{
				DriverName: "mssql",
				DataSourceName: fmt.Sprintf(
					"Server=127.0.0.1;Port=%d;Database=master;UID=sa,PWD=Password12!",
					c.port,
				),
				Query: c.query,
			},
		}
		got, err := makeDataset(e)
		if err != nil {
			t.Errorf("makeDataset(%v) instanceName: %s, err: %v",
				c.instanceName, e, err)
			continue
		}
		if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("checkDatasetsEqual: instanceName: %s, err: %v",
				c.instanceName, err)
		}
	}
}
