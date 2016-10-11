/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

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
	de, err := dexpr.New(expr)
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
	ok, err := whenExpr.EvalBool(vars, map[string]dexpr.CallFun{})
	if err != nil {
		return false, err
	}
	return ok, nil
}
