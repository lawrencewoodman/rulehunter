// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package experiment

import (
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"time"
)

func makeWhenExpr(expr string) (*dexpr.Expr, error) {
	if expr == "" {
		expr = "!hasRun"
	}
	funcs := map[string]dexpr.CallFun{}
	de, err := dexpr.New(expr, funcs)
	return de, err
}

// TODO: Add dayName := {MonTueWedThurFriSatSun, etc} may want to allow
//       day as 60 * 60 * 24
// TODO: Add monthName := {JanFebMarAprMayJunJulAugSepOctNovDec}
// TODO: Add sinceLastRunDays try working out using:
//         https://play.golang.org/p/nTcjGZQKAa
//         https://groups.google.com/forum/#!topic/golang-nuts/O2NaRAH94GI
func evalWhenExpr(
	now time.Time,
	isFinished bool,
	stamp time.Time,
	whenExpr *dexpr.Expr,
) (bool, error) {
	nISOWeekYear, nISOWeekWeek := now.ISOWeek()
	stampISOWeekYear, stampISOWeekWeek := stamp.ISOWeek()

	vars := map[string]*dlit.Literal{
		"hasRun": dlit.MustNew(isFinished),
		"hasRunToday": dlit.MustNew(
			isFinished &&
				stamp.Year() == now.Year() &&
				stamp.YearDay() == now.YearDay(),
		),
		"hasRunThisWeek": dlit.MustNew(
			isFinished &&
				stampISOWeekYear == nISOWeekYear &&
				stampISOWeekWeek == nISOWeekWeek,
		),
		"hasRunThisMonth": dlit.MustNew(
			isFinished && stamp.Year() == now.Year() && stamp.Month() == now.Month(),
		),
		"hasRunThisYear":      dlit.MustNew(isFinished && stamp.Year() == now.Year()),
		"sinceLastRunMinutes": dlit.MustNew(int64(now.Sub(stamp).Minutes())),
		"sinceLastRunHours":   dlit.MustNew(int64(now.Sub(stamp).Hours())),
		"isWeekday": dlit.MustNew(
			now.Weekday() != time.Saturday && now.Weekday() != time.Sunday,
		),
	}
	ok, err := whenExpr.EvalBool(vars)
	if err != nil {
		return false, err
	}
	return ok, nil
}
