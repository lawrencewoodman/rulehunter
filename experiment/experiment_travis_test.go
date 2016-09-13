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
		query      string
		fieldNames []string
		want       ddataset.Dataset
	}{
		{query: "select * from flow",
			fieldNames: []string{"grp", "district", "height", "flow"},
			want: dcsv.New(
				filepath.Join("fixtures", "flow.csv"),
				true,
				rune(','),
				[]string{"grp", "district", "height", "flow"},
			),
		},
		{query: "select grp,district,flow from flow",
			fieldNames: []string{"grp", "district", "flow"},
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
			Dataset:    "sql",
			FieldNames: c.fieldNames,
			Sql: &sqlDesc{
				DriverName:     "mysql",
				DataSourceName: "travis@/master",
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
