package report

import (
	"errors"
	"github.com/lawrencewoodman/dlit"
	"math"
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
