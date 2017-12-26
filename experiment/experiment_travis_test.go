// Test database access under Travis
// +build travis

package experiment

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/ddataset/dcsv"
	"path/filepath"
	"testing"
)

func TestMakeDataset_travis(t *testing.T) {
	cases := []struct {
		driverName     string
		dataSourceName string
		query          string
		fields         []string
		want           ddataset.Dataset
	}{
		{driverName: "mysql",
			dataSourceName: "travis@/master",
			query:          "select * from flow",
			fields:         []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{driverName: "mysql",
			dataSourceName: "travis@/master",
			query:          "select grp,district,flow from flow",
			fields:         []string{"grp", "district", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
		{driverName: "postgres",
			dataSourceName: "user=postgres dbname=master",
			query:          "select * from flow",
			fields:         []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{driverName: "postgres",
			dataSourceName: "user=postgres dbname=master",
			query:          "select grp,district,flow from flow",
			fields:         []string{"grp", "district", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow_three_columns.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "flow"},
			),
		},
	}
	for _, c := range cases {
		e := &experimentFile{
			Dataset: "sql",
			Fields:  c.fields,
			Sql: &sqlDesc{
				DriverName:     c.driverName,
				DataSourceName: c.dataSourceName,
				Query:          c.query,
			},
		}
		got, err := makeDataset(e)
		if err != nil {
			t.Errorf("makeDataset(%v) query: %s, err: %v",
				c.query, e, err)
			continue
		}
		if err := checkDatasetsEqual(got, c.want); err != nil {
			t.Errorf("checkDatasetsEqual: err: %v", err)
		}
	}
}
