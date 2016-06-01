package experiment

import (
	"errors"
	"fmt"
	"github.com/vlifesystems/rulehunter/aggregators"
	"github.com/vlifesystems/rulehunter/csvdataset"
	"github.com/vlifesystems/rulehunter/dataset"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/goal"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadExperiment(t *testing.T) {
	fieldNames := []string{"group", "district", "height", "flow"}
	cases := []struct {
		filename       string
		wantExperiment *experiment.Experiment
		wantTags       []string
	}{
		{filepath.Join("fixtures", "flow.json"),
			&experiment.Experiment{
				Title: "What would indicate good flow?",
				Dataset: mustNewCsvDataset(
					fieldNames,
					filepath.Join("fixtures", "flow.csv"),
					rune(','),
					true,
				),
				ExcludeFieldNames: []string{"flow"},
				Aggregators: []aggregators.Aggregator{
					aggregators.MustNew("goodFlowAccuracy", "accuracy", "flow > 60"),
				},
				Goals: []*goal.Goal{goal.MustNew("goodFlowAccuracy > 10")},
				SortOrder: []experiment.SortField{
					experiment.SortField{"goodFlowAccuracy", experiment.DESCENDING},
					experiment.SortField{"numMatches", experiment.DESCENDING},
				},
			},
			[]string{"test", "fred / ned"},
		},
	}
	for _, c := range cases {
		gotExperiment, gotTags, err := loadExperiment(c.filename)
		if err != nil {
			t.Errorf("loadExperiment(%s) err: %s", c.filename, err)
			return
		}
		doExperimentsMatch, msg := experimentMatch(gotExperiment, c.wantExperiment)
		if !doExperimentsMatch {
			t.Errorf("loadExperiment(%s) experiments don't match: %s\ngotExperiment: %s, wantExperiment: %s",
				c.filename, msg, gotExperiment, c.wantExperiment)
			return
		}
		if !reflect.DeepEqual(gotTags, c.wantTags) {
			t.Errorf("loadExperiment(%s) gotTags: %s, wantTags",
				c.filename, gotTags, c.wantTags)
			return
		}
	}
}

func TestLoadExperiment_error(t *testing.T) {
	cases := []struct {
		filename string
		wantErr  error
	}{
		{filepath.Join("fixtures", "flow_no_title.json"),
			errors.New("Experiment field missing: title")},
		{filepath.Join("fixtures", "flow_no_dataset.json"),
			errors.New("Experiment field missing: datasetFilename")},
	}
	for _, c := range cases {
		_, _, err := loadExperiment(c.filename)
		if err == nil {
			t.Errorf("loadExperiment(%s) no error, wantErr:%s",
				c.filename, c.wantErr)
			return
		}
		if err.Error() != c.wantErr.Error() {
			t.Errorf("loadExperiment(%s) gotErr: %s, wantErr:%s",
				c.filename, err, c.wantErr)
			return
		}
	}
}

/***********************
   Helper functions
************************/

func experimentMatch(
	e1 *experiment.Experiment,
	e2 *experiment.Experiment,
) (bool, string) {
	if e1.Title != e2.Title {
		return false, "Titles don't match"
	}
	if !areStringArraysEqual(e1.ExcludeFieldNames, e2.ExcludeFieldNames) {
		return false, "ExcludeFieldNames don't match"
	}
	if !areGoalExpressionsEqual(e1.Goals, e2.Goals) {
		return false, "Goals don't match"
	}
	if !areAggregatorsEqual(e1.Aggregators, e2.Aggregators) {
		return false, "Aggregators don't match"
	}
	if !areSortOrdersEqual(e1.SortOrder, e2.SortOrder) {
		return false, "Sort Orders don't match"
	}
	datasetsEqual, msg := areDatasetsEqual(e1.Dataset, e2.Dataset)
	return datasetsEqual, msg
}

func areDatasetsEqual(i1, i2 dataset.Dataset) (bool, string) {
	for {
		i1Next := i1.Next()
		i2Next := i2.Next()
		if i1Next != i2Next {
			return false, "Datasets don't finish at same point"
		}
		if !i1Next {
			break
		}

		i1Record, i1Err := i1.Read()
		i2Record, i2Err := i2.Read()
		if i1Err != i2Err {
			return false, "Datasets don't error at same point"
		} else if i1Err == nil && i2Err == nil {
			if !reflect.DeepEqual(i1Record, i2Record) {
				return false, "Datasets don't match"
			}
		}
	}
	if i1.Err() != i2.Err() {
		return false, "Datasets final error doesn't match"
	}
	return true, ""
}

func areStringArraysEqual(a1 []string, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, e := range a1 {
		if e != a2[i] {
			return false
		}
	}
	return true
}

func areGoalExpressionsEqual(g1 []*goal.Goal, g2 []*goal.Goal) bool {
	if len(g1) != len(g2) {
		return false
	}
	for i, g := range g1 {
		if g.String() != g2[i].String() {
			return false
		}
	}
	return true

}

func areAggregatorsEqual(
	a1 []aggregators.Aggregator,
	a2 []aggregators.Aggregator,
) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, e := range a1 {
		if !e.IsEqual(a2[i]) {
			return false
		}
	}
	return true
}

func areSortOrdersEqual(
	so1 []experiment.SortField,
	so2 []experiment.SortField,
) bool {
	if len(so1) != len(so2) {
		return false
	}
	for i, sf1 := range so1 {
		sf2 := so2[i]
		if sf1.Field != sf2.Field || sf1.Direction != sf2.Direction {
			return false
		}
	}
	return true
}

func mustNewCsvDataset(
	fieldNames []string,
	filename string,
	separator rune,
	skipFirstLine bool,
) dataset.Dataset {
	dataset, err :=
		csvdataset.New(fieldNames, filename, separator, skipFirstLine)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create Csv Dataset: %s", err))
	}
	return dataset
}
