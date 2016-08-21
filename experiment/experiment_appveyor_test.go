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
		fieldNames   []string
		want         ddataset.Dataset
	}{
		{instanceName: "SQL2008R2SP2",
			port:       12008,
			query:      "select * from flow",
			fieldNames: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{instanceName: "SQL2008R2SP2",
			port:       12008,
			query:      "select grp,district,flow from flow",
			fieldNames: []string{"grp", "district", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
		{instanceName: "SQL2012SP1",
			port:       12012,
			query:      "select * from flow",
			fieldNames: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{instanceName: "SQL2014",
			port:       12014,
			query:      "select * from flow",
			fieldNames: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
	}
	for _, c := range cases {
		e := &experimentFile{
			Dataset:    "sql",
			FieldNames: c.fieldNames,
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
