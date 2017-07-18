// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package internal

import (
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"sort"
	"strings"
)

func NumDecPlaces(s string) int {
	i := strings.IndexByte(s, '.')
	if i > -1 {
		s = strings.TrimRight(s, "0")
		return len(s) - i - 1
	}
	return 0
}

// GeneratePoints will generate numbers between two points (min..max).
// It will round each number to maxDP decimal places
func GeneratePoints(min, max *dlit.Literal, maxDP int) []*dlit.Literal {
	points := make(map[string]*dlit.Literal)
	vars := map[string]*dlit.Literal{
		"min": min,
		"max": max,
		"n":   min,
	}
	vars["diff"] = dexpr.Eval("max - min", dexprfuncs.CallFuncs, vars)
	vars["step"] = dexpr.Eval("diff / 20", dexprfuncs.CallFuncs, vars)
	if vars["step"].String() == "0" {
		vars["step"] = dlit.MustNew(1)
	}

	nextNExpr := dexpr.MustNew("n + step", dexprfuncs.CallFuncs)
	stopExpr := dexpr.MustNew("v >= max", dexprfuncs.CallFuncs)
	tooLowExpr := dexpr.MustNew("v <= min", dexprfuncs.CallFuncs)
	stop := false
	for !stop {
		vars["n"] = nextNExpr.Eval(vars)
		vars["v"] = RoundLit(vars["n"], maxDP)
		if shouldStop, err := stopExpr.EvalBool(vars); shouldStop || err != nil {
			stop = true
			break
		}
		if tooLow, err := tooLowExpr.EvalBool(vars); !tooLow && err == nil {
			points[vars["v"].String()] = vars["v"]
		}
	}

	return MapLitNumsToSlice(points)
}

var roundExpr = dexpr.MustNew("roundto(n, dp)", dexprfuncs.CallFuncs)

func RoundLit(l *dlit.Literal, dp int) *dlit.Literal {
	vars := map[string]*dlit.Literal{"n": l, "dp": dlit.MustNew(dp)}
	return roundExpr.Eval(vars)
}

func MapLitNumsToSlice(nums map[string]*dlit.Literal) []*dlit.Literal {
	r := make([]*dlit.Literal, len(nums))
	i := 0
	for _, n := range nums {
		r[i] = n
		i++
	}

	sort.Sort(byNumber(r))
	return r
}

// byNumber implements sort.Interface for []*dlit.Literal
type byNumber []*dlit.Literal

func (bn byNumber) Len() int { return len(bn) }
func (bn byNumber) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}

var compareLitExpr = dexpr.MustNew("a < b", dexprfuncs.CallFuncs)

func (bn byNumber) Less(i, j int) bool {
	vars := map[string]*dlit.Literal{
		"a": bn[i],
		"b": bn[j],
	}
	if r, err := compareLitExpr.EvalBool(vars); err != nil {
		panic(err)
	} else {
		return r
	}
}

func StringInSlice(s string, strings []string) bool {
	for _, x := range strings {
		if x == s {
			return true
		}
	}
	return false
}
