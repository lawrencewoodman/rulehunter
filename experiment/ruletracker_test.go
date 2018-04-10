package experiment

import (
	"testing"

	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/rule"
)

func TestRuleTracker_track(t *testing.T) {
	cases := []struct {
		rules []rule.Rule
		want  []rule.Rule
	}{
		{rules: []rule.Rule{
			rule.NewEQFV("month", dlit.NewString("may")),
			rule.NewEQFV("month", dlit.NewString("june")),
			rule.NewGEFV("rate", dlit.MustNew(789.2)),
		},
			want: []rule.Rule{
				rule.NewEQFV("month", dlit.NewString("may")),
				rule.NewEQFV("month", dlit.NewString("june")),
				rule.NewGEFV("rate", dlit.MustNew(789.2)),
			},
		},
		{rules: []rule.Rule{
			rule.NewEQFV("month", dlit.NewString("may")),
			rule.NewEQFV("month", dlit.NewString("june")),
			rule.NewGEFV("rate", dlit.MustNew(789.2)),
		},
			want: []rule.Rule{},
		},
		{rules: []rule.Rule{
			rule.NewEQFV("month", dlit.NewString("april")),
			rule.NewEQFV("month", dlit.NewString("june")),
			rule.NewGEFV("rate", dlit.MustNew(789.2)),
		},
			want: []rule.Rule{
				rule.NewEQFV("month", dlit.NewString("april")),
			},
		},
	}

	rt := newRuleTracker()
	for _, c := range cases {
		got := rt.track(c.rules)
		if len(got) != len(c.want) {
			t.Errorf("track, got: %v, want: %v", got, c.want)
			continue
		}
		for i, r := range c.want {
			if r.String() != got[i].String() {
				t.Errorf("track, got: %v, want: %v", got, c.want)
				break
			}
		}
	}
}
