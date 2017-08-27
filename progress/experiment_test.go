package progress

import (
	"fmt"
	"testing"
	"time"
)

func TestExperimentString(t *testing.T) {
	timeNow := time.Now()
	cases := []struct {
		experiment *Experiment
		want       string
	}{
		{experiment: &Experiment{
			Filename: "anicefile.yaml",
			Title:    "how to get ahead?",
			Tags:     []string{"good", "bad", "ugly"},
			Category: "responsibility",
			Status: &Status{
				Stamp:   timeNow,
				Msg:     "Assessing rules (1/5)",
				Percent: 23.9,
				State:   Processing,
			},
		},
			want: fmt.Sprintf(
				"{filename: anicefile.yaml, title: how to get ahead?, tags: [good bad ugly], category: responsibility, status: {stamp: %s, msg: Assessing rules (1/5), percent: 23.90, state: processing}}",
				timeNow,
			),
		},
		{experiment: &Experiment{
			Filename: "anicefile.yaml",
			Title:    "how to get ahead?",
			Tags:     []string{},
			Category: "responsibility",
			Status: &Status{
				Stamp:   timeNow,
				Msg:     "Assessing rules (1/5)",
				Percent: 23.9,
				State:   Processing,
			},
		},
			want: fmt.Sprintf(
				"{filename: anicefile.yaml, title: how to get ahead?, tags: [], category: responsibility, status: {stamp: %s, msg: Assessing rules (1/5), percent: 23.90, state: processing}}",
				timeNow,
			),
		},
	}
	for _, c := range cases {
		got := c.experiment.String()
		if got != c.want {
			t.Errorf("String - got: %s, want: %s", got, c.want)
		}
	}
}
