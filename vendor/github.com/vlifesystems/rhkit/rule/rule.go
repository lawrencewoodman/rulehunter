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
	"github.com/vlifesystems/rhkit/internal/fieldtype"
	"sort"
	"strings"
	"sync"
)

var (
	generatorsMu sync.RWMutex
	generators   = make(map[string]generatorFunc)
)

type generatorFunc func(
	desc *description.Description,
	ruleFields []string,
	complexity Complexity,
	field string,
) []Rule

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

type Complexity struct {
	Arithmetic bool
}

func (e InvalidRuleError) Error() string {
	return "invalid rule: " + e.Rule.String()
}

func (e IncompatibleTypesRuleError) Error() string {
	return "incompatible types in rule: " + e.Rule.String()
}

// Generate generates rules for rules that have registered a generator.
// complexity is used to indicate how complex and in turn how many rules
// to produce it takes a number 1 to 10.
func Generate(
	inputDescription *description.Description,
	ruleFields []string,
	complexity Complexity,
) []Rule {
	rules := make([]Rule, 1)
	rules[0] = NewTrue()
	for field := range inputDescription.Fields {
		if internal.StringInSlice(field, ruleFields) {
			for _, generator := range generators {
				newRules := generator(inputDescription, ruleFields, complexity, field)
				rules = append(rules, newRules...)
			}
		}
	}
	if len(ruleFields) == 2 {
		cRules := Combine(rules)
		rules = append(rules, cRules...)
	}
	Sort(rules)
	return Uniq(rules)
}

func Combine(rules []Rule) []Rule {
	Sort(rules)
	combinedRules := make([]Rule, 0)
	numRules := len(rules)
	for i := 0; i < numRules-1; i++ {
		for j := i + 1; j < numRules; j++ {
			if andRule, err := NewAnd(rules[i], rules[j]); err == nil {
				combinedRules = append(combinedRules, andRule)
			}
			if orRule, err := NewOr(rules[i], rules[j]); err == nil {
				combinedRules = append(combinedRules, orRule)
			}
		}
	}
	return Uniq(combinedRules)
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

// registerGenerator makes a rule generator available.
// If called twice with the same ruleName it panics.
func registerGenerator(ruleType string, generator generatorFunc) {
	generatorsMu.Lock()
	defer generatorsMu.Unlock()
	if _, dup := generators[ruleType]; dup {
		panic("registerGenerator called twice for ruleType: " + ruleType)
	}
	generators[ruleType] = generator
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

func calcNumSharedValues(fd1 *description.Field, fd2 *description.Field) int {
	numShared := 0
	for _, vd1 := range fd1.Values {
		if _, ok := fd2.Values[vd1.Value.String()]; ok {
			numShared++
		}
	}
	return numShared
}

var compareExpr *dexpr.Expr = dexpr.MustNew(
	"min1 < max2 && max1 > min2",
	dexprfuncs.CallFuncs,
)

func hasComparableNumberRange(
	fd1 *description.Field,
	fd2 *description.Field,
) bool {
	if fd1.Kind != fieldtype.Number || fd2.Kind != fieldtype.Number {
		return false
	}
	var isComparable bool
	vars := map[string]*dlit.Literal{
		"min1": fd1.Min,
		"max1": fd1.Max,
		"min2": fd2.Min,
		"max2": fd2.Max,
	}
	isComparable, err := compareExpr.EvalBool(vars)
	return err == nil && isComparable
}
