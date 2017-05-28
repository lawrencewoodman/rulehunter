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

// Package rule implements rules to be tested against a dataset
package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"sort"
	"strings"
)

type Rule interface {
	fmt.Stringer
	IsTrue(record ddataset.Record) (bool, error)
	Fields() []string
}

type Tweaker interface {
	Tweak(desc *description.Description, stage int) []Rule
}

type Overlapper interface {
	Overlaps(o Rule) bool
}

type DPReducer interface {
	DPReduce() []Rule
}

type Valuer interface {
	Value() *dlit.Literal
}

type InvalidRuleError struct {
	Rule Rule
}

type IncompatibleTypesRuleError struct {
	Rule Rule
}

func (e InvalidRuleError) Error() string {
	return "invalid rule: " + e.Rule.String()
}

func (e IncompatibleTypesRuleError) Error() string {
	return "incompatible types in rule: " + e.Rule.String()
}

// Sort sorts the rules in place using their .String() method
func Sort(rules []Rule) {
	sort.Sort(byString(rules))
}

// Uniq returns the slices of Rules with duplicates removed
func Uniq(rules []Rule) []Rule {
	results := []Rule{}
	mResults := map[string]interface{}{}
	for _, r := range rules {
		if _, ok := mResults[r.String()]; !ok {
			mResults[r.String()] = nil
			results = append(results, r)
		}
	}
	return results
}

// ReduceDP returns decimal number reduced rules if they can be reduced.
// Adds True rule at end.
func ReduceDP(rules []Rule) []Rule {
	newRules := make([]Rule, 0)
	for _, r := range rules {
		switch x := r.(type) {
		case DPReducer:
			rules := x.DPReduce()
			newRules = append(newRules, rules...)
		}
	}
	newRules = append(newRules, NewTrue())
	return Uniq(newRules)
}

func commaJoinValues(values []*dlit.Literal) string {
	str := fmt.Sprintf("\"%s\"", values[0].String())
	for _, v := range values[1:] {
		str += fmt.Sprintf(",\"%s\"", v)
	}
	return str
}

// byString implements sort.Interface for []Rule
type byString []Rule

func (rs byString) Len() int { return len(rs) }
func (rs byString) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}
func (rs byString) Less(i, j int) bool {
	return strings.Compare(rs[i].String(), rs[j].String()) == -1
}

func generateTweakPoints(
	value, min, max *dlit.Literal,
	maxDP int,
	stage int,
) []*dlit.Literal {
	vars := map[string]*dlit.Literal{
		"min":   min,
		"max":   max,
		"maxDP": dlit.MustNew(maxDP),
		"stage": dlit.MustNew(stage),
		"value": value,
	}
	vars["step"] =
		dexpr.Eval(
			"roundto((max - min) / (10 * stage), maxDP)",
			dexprfuncs.CallFuncs,
			vars,
		)
	validValueExpr := dexpr.MustNew(
		"rp > min && rp != value && rp < max",
		dexprfuncs.CallFuncs,
	)
	low := dexpr.Eval("max(min, value - step)", dexprfuncs.CallFuncs, vars)
	high := dexpr.Eval("min(max, value + step)", dexprfuncs.CallFuncs, vars)
	points := internal.GeneratePoints(low, high, maxDP)
	tweakPoints := map[string]*dlit.Literal{}

	for _, p := range points {
		vars["p"] = p
		vars["rp"] = internal.RoundLit(p, maxDP)
		if ok, err := validValueExpr.EvalBool(vars); ok && err == nil {
			tweakPoints[vars["rp"].String()] = vars["rp"]
		}
	}

	return internal.MapLitNumsToSlice(tweakPoints)
}

type makeRoundRule func(*dlit.Literal) Rule

func roundRules(v *dlit.Literal, makeRule makeRoundRule) []Rule {
	rulesMap := map[string]Rule{}
	rules := []Rule{}
	for dp := 200; dp >= 0; dp-- {
		p := internal.RoundLit(v, dp)
		r := makeRule(p)
		if _, ok := rulesMap[r.String()]; !ok {
			rulesMap[r.String()] = r
			rules = append(rules, r)
		}
	}
	return rules
}
