// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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
	"sync"
)

var (
	generatorsMu sync.RWMutex
	generators   = make(map[string]generatorFunc)
)

// GenerationDescriber describes what sort of rules should be generated
type GenerationDescriber interface {
	// Fields indicates which fields should be used to generate rules
	Fields() []string
	// Arithmetic indicates whether to generate arithmetic rules
	Arithmetic() bool
}

type generatorFunc func(
	desc *description.Description,
	generationDesc GenerationDescriber,
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

// Generate generates rules for rules that have registered a generator.
func Generate(
	inputDescription *description.Description,
	generationDesc GenerationDescriber,
) ([]Rule, error) {
	err := checkRuleFieldsValid(inputDescription, generationDesc.Fields())
	if err != nil {
		return []Rule{}, err
	}
	rules := make([]Rule, 1)
	rules[0] = NewTrue()

	for _, generator := range generators {
		newRules := generator(inputDescription, generationDesc)
		rules = append(rules, newRules...)
	}

	Sort(rules)
	return Uniq(rules), nil
}

// Combine combines rules together using And and Or
func Combine(rules []Rule, maxNumRules int) []Rule {
	Sort(rules)
	combinedRules := make([]Rule, 0)
	numRules := len(rules)
	for i := 0; i < numRules-1; i++ {
		for j := i + 1; j < numRules; j++ {
			if andRule, err := NewAnd(rules[i], rules[j]); err == nil {
				combinedRules = append(combinedRules, andRule)
			}
			if len(combinedRules) >= maxNumRules {
				break
			}
			if orRule, err := NewOr(rules[i], rules[j]); err == nil {
				combinedRules = append(combinedRules, orRule)
			}
			if len(combinedRules) >= maxNumRules {
				break
			}
		}
		if len(combinedRules) >= maxNumRules {
			break
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
	if fd1.Kind != description.Number || fd2.Kind != description.Number {
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

func checkRuleFieldsValid(
	inputDescription *description.Description,
	ruleFields []string,
) error {
	fields := inputDescription.FieldNames()
	for _, ruleField := range ruleFields {
		if !internal.IsStringInSlice(ruleField, fields) {
			return InvalidRuleFieldError(ruleField)
		}
	}
	return nil
}

func countNumOnBits(mask string) int {
	return strings.Count(mask, "1")
}

func makeMask(numPlaces, i int) string {
	return fmt.Sprintf("%0*b", numPlaces, i)
}

func getMaskStrings(mask string, values []string) []string {
	r := []string{}
	for j, b := range mask {
		if j >= len(values) {
			break
		}
		if b == '1' {
			v := values[j]
			r = append(r, v)
		}
	}
	return r
}

func stringCombinations(values []string, min, max int) [][]string {
	r := [][]string{}
	for i := 3; ; i++ {
		mask := makeMask(len(values), i)
		numOnBits := countNumOnBits(mask)
		if len(mask) > len(values) {
			break
		}
		if numOnBits >= min && numOnBits <= max && numOnBits <= len(values) {
			r = append(r, getMaskStrings(mask, values))
		}
	}
	return r
}
