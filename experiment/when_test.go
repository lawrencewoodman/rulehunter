package experiment

import (
	"github.com/lawrencewoodman/dexpr"
	"testing"
	"time"
)

var evalWhenExprCases = []struct {
	now        time.Time
	isFinished bool
	stamp      time.Time
	when       string
	want       bool
}{
	{time.Now(), true, time.Now(), "!hasRun", false},
	{time.Now(), false, time.Now(), "!hasRun", true},

	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunToday",
		true,
	},
	{mustNewTime("2017-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunToday",
		false,
	},
	{mustNewTime("2016-06-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunToday",
		false,
	},
	{mustNewTime("2016-05-06T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunToday",
		false,
	},
	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		false,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"hasRunToday",
		false,
	},

	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisWeek",
		true,
	},
	{mustNewTime("2016-05-06T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisWeek",
		true,
	},
	{mustNewTime("2017-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisWeek",
		false,
	},
	{mustNewTime("2016-06-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisWeek",
		false,
	},
	{mustNewTime("2016-05-09T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisWeek",
		false,
	},
	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		false,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"hasRunThisWeek",
		false,
	},

	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisMonth",
		true,
	},
	{mustNewTime("2016-05-30T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisMonth",
		true,
	},
	{mustNewTime("2017-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisMonth",
		false,
	},
	{mustNewTime("2016-06-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisMonth",
		false,
	},
	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		false,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"hasRunThisMonth",
		false,
	},

	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisYear",
		true,
	},
	{mustNewTime("2016-11-30T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisYear",
		true,
	},
	{mustNewTime("2017-05-05T09:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T19:37:58.0+00:00"),
		"hasRunThisYear",
		false,
	},
	{mustNewTime("2016-05-05T09:37:58.0+00:00"),
		false,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"hasRunThisYear",
		false,
	},

	{mustNewTime("2016-05-05T09:40:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"sinceLastRunMinutes == 3",
		true,
	},
	{mustNewTime("2016-05-05T09:40:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"sinceLastRunMinutes == 4",
		false,
	},

	{mustNewTime("2016-05-05T13:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"sinceLastRunHours == 4",
		true,
	},
	{mustNewTime("2016-05-05T14:40:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"sinceLastRunHours == 4",
		false,
	},

	// Monday
	{mustNewTime("2016-04-04T9:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-07T09:37:58.0+00:00"),
		"isWeekday",
		true,
	},
	// Friday
	{mustNewTime("2016-04-08T9:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-07T09:37:58.0+00:00"),
		"isWeekday",
		true,
	},
	// Saturday
	{mustNewTime("2016-04-09T9:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"isWeekday",
		false,
	},
	// Sunday
	{mustNewTime("2016-04-10T9:37:58.0+00:00"),
		true,
		mustNewTime("2016-05-05T09:37:58.0+00:00"),
		"isWeekday",
		false,
	},
}

func TestEvalWhenExpr(t *testing.T) {
	for _, c := range evalWhenExprCases {
		whenExpr := dexpr.MustNew(c.when)
		got, err := evalWhenExpr(c.now, c.isFinished, c.stamp, whenExpr)
		if err != nil {
			t.Errorf("evalWhenExpr(%v, %t, %v, %v) err: %v",
				c.now, c.isFinished, c.stamp, c.when, err)
			continue
		}
		if got != c.want {
			t.Errorf("evalWhenExpr(%v, %t, %v, %v) got: %t, want: %t",
				c.now, c.isFinished, c.stamp, c.when, got, c.want)
		}
	}
}

func TestEvalWhenExpr_errors(t *testing.T) {
	now := time.Now()
	isFinished := true
	stamp := now
	when := "!hasTwoLegs"
	wantErr := dexpr.ErrInvalidExpr{
		Expr: "!hasTwoLegs",
		Err:  dexpr.ErrVarNotExist("hasTwoLegs"),
	}
	whenExpr := dexpr.MustNew(when)
	got, err := evalWhenExpr(now, isFinished, stamp, whenExpr)
	if got != false {
		t.Errorf("evalWhenExpr(%v, %t, %v, %v) got: %t, want: %t",
			now, isFinished, stamp, when, got, false)
	}
	if err != wantErr {
		t.Errorf("evalWhenExpr(%v, %t, %v, %v) err: %v, want: %v",
			now, isFinished, stamp, when, err, wantErr)
	}
}

/***********************
   Helper functions
************************/
func mustNewTime(stamp string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		panic(err)
	}
	return t
}
