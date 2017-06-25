package report

import (
	"errors"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit"
	"github.com/vlifesystems/rhkit/rule"
	"math"
	"reflect"
	"testing"
)

func TestCalcTrueAggregatorDiff(t *testing.T) {
	trueAggregators := map[string]*dlit.Literal{
		"numMatches": dlit.MustNew(176),
		"profit":     dlit.MustNew(23),
		"bigNum":     dlit.MustNew(int64(math.MaxInt64)),
	}
	cases := []struct {
		name  string
		value *dlit.Literal
		want  string
	}{
		{name: "numMatches", value: dlit.MustNew(192), want: "16"},
		{name: "numMatches", value: dlit.MustNew(165), want: "-11"},
		{name: "bigNum",
			value: dlit.MustNew(int64(math.MinInt64)),
			want: dlit.MustNew(
				float64(math.MinInt64) - float64(math.MaxInt64),
			).String(),
		},
		{name: "bigNum",
			value: dlit.MustNew(errors.New("some error")),
			want:  "N/A",
		},
	}

	for _, c := range cases {
		got := calcTrueAggregatorDiff(trueAggregators, c.name, c.value)
		if got != c.want {
			t.Errorf("calcTrueAggregatorDifference(trueAggregators, %v, %v) got: %s, want: %s",
				c.name, c.value, got, c.want)
		}
	}
}

func TestGetSortedAggregatorNames(t *testing.T) {
	aggregators := map[string]*dlit.Literal{
		"numMatches": dlit.MustNew(176),
		"profit":     dlit.MustNew(23),
		"bigNum":     dlit.MustNew(int64(math.MaxInt64)),
	}
	want := []string{"bigNum", "numMatches", "profit"}
	got := getSortedAggregatorNames(aggregators)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getSortedAggregatorNames - got: %v, want: %v", got, want)
	}
}

func TestGetTrueAggregators(t *testing.T) {
	assessment := &rhkit.Assessment{
		NumRecords: 20,
		RuleAssessments: []*rhkit.RuleAssessment{
			&rhkit.RuleAssessment{
				Rule: rule.NewEQFV("month", dlit.NewString("may")),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("2142"),
					"percentMatches": dlit.MustNew("242"),
					"numIncomeGt2":   dlit.MustNew("22"),
					"goalsScore":     dlit.MustNew(20.1),
				},
				Goals: []*rhkit.GoalAssessment{
					&rhkit.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkit.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkit.RuleAssessment{
				Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("3142"),
					"percentMatches": dlit.MustNew("342"),
					"numIncomeGt2":   dlit.MustNew("32"),
					"goalsScore":     dlit.MustNew(30.1),
				},
				Goals: []*rhkit.GoalAssessment{
					&rhkit.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkit.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkit.RuleAssessment{
				Rule: rule.NewTrue(),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("142"),
					"percentMatches": dlit.MustNew("42"),
					"numIncomeGt2":   dlit.MustNew("2"),
					"goalsScore":     dlit.MustNew(0.1),
				},
				Goals: []*rhkit.GoalAssessment{
					&rhkit.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkit.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
		},
	}
	want := map[string]*dlit.Literal{
		"numMatches":     dlit.MustNew("142"),
		"percentMatches": dlit.MustNew("42"),
		"numIncomeGt2":   dlit.MustNew("2"),
		"goalsScore":     dlit.MustNew(0.1),
	}

	got, err := getTrueAggregators(assessment)
	if err != nil {
		t.Fatalf("getTrueAggregators: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getTrueAggregators - got: %v, want: %v", got, want)
	}
}

func TestGetTrueAggregators_error(t *testing.T) {
	assessment := &rhkit.Assessment{
		NumRecords: 20,
		RuleAssessments: []*rhkit.RuleAssessment{
			&rhkit.RuleAssessment{
				Rule: rule.NewEQFV("month", dlit.NewString("may")),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("2142"),
					"percentMatches": dlit.MustNew("242"),
					"numIncomeGt2":   dlit.MustNew("22"),
					"goalsScore":     dlit.MustNew(20.1),
				},
				Goals: []*rhkit.GoalAssessment{
					&rhkit.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkit.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkit.RuleAssessment{
				Rule: rule.NewTrue(),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("142"),
					"percentMatches": dlit.MustNew("42"),
					"numIncomeGt2":   dlit.MustNew("2"),
					"goalsScore":     dlit.MustNew(0.1),
				},
				Goals: []*rhkit.GoalAssessment{
					&rhkit.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkit.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
			&rhkit.RuleAssessment{
				Rule: rule.NewGEFV("rate", dlit.MustNew(789.2)),
				Aggregators: map[string]*dlit.Literal{
					"numMatches":     dlit.MustNew("3142"),
					"percentMatches": dlit.MustNew("342"),
					"numIncomeGt2":   dlit.MustNew("32"),
					"goalsScore":     dlit.MustNew(30.1),
				},
				Goals: []*rhkit.GoalAssessment{
					&rhkit.GoalAssessment{"numIncomeGt2 == 1", false},
					&rhkit.GoalAssessment{"numIncomeGt2 == 2", true},
				},
			},
		},
	}
	wantErr := errors.New("can't find true() rule")

	_, err := getTrueAggregators(assessment)
	if err == nil || err.Error() != wantErr.Error() {
		t.Errorf("getTrueAggregators: err: %s, wantErr: %s", err, wantErr)
	}
}
