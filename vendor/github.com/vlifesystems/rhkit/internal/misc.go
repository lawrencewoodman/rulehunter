/*
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of rhkit.

	rhkit is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	rhkit is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with rhkit; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

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

func GeneratePoints(min, max *dlit.Literal, maxDP int) []*dlit.Literal {
	// Nice DP to work with, in this order to prevent earlier stopping from
	// rounding a number beyond max at end
	dps := []int{maxDP, 3, 2, 1, 0}
	points := make(map[string]*dlit.Literal)
	vars := map[string]*dlit.Literal{
		"min":   min,
		"max":   max,
		"maxDP": dlit.MustNew(maxDP),
		"n":     min,
	}
	vars["diff"] = dexpr.Eval("max - min", dexprfuncs.CallFuncs, vars)
	vars["step"] = dexpr.Eval(
		"roundto(diff / 20, maxDP)",
		dexprfuncs.CallFuncs,
		vars,
	)
	if vars["step"].String() == "0" {
		vars["step"] = dlit.MustNew(1)
	}

	nextNExpr := dexpr.MustNew("n + step", dexprfuncs.CallFuncs)
	stopExpr := dexpr.MustNew("v >= max", dexprfuncs.CallFuncs)
	roundExpr := dexpr.MustNew("roundto(n, dp)", dexprfuncs.CallFuncs)
	stop := false
	for !stop {
		vars["n"] = nextNExpr.Eval(vars)
		for _, dp := range dps {
			vars["dp"] = dlit.MustNew(dp)
			v := roundExpr.Eval(vars)
			vars["v"] = v
			if shouldStop, err := stopExpr.EvalBool(vars); shouldStop || err != nil {
				stop = true
				break
			} else {
				points[v.String()] = v
			}
		}
	}

	r := make([]*dlit.Literal, len(points))
	i := 0
	for _, p := range points {
		r[i] = p
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
