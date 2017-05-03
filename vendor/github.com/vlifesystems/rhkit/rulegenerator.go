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

package rhkit

import (
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"github.com/vlifesystems/rhkit/internal/fieldtype"
	"github.com/vlifesystems/rhkit/rule"
	"sort"
	"strings"
)

type ruleGeneratorFunc func(*description.Description, []string, string) []rule.Rule

func GenerateRules(
	inputDescription *description.Description,
	ruleFields []string,
) []rule.Rule {
	rules := make([]rule.Rule, 1)
	ruleGenerators := []ruleGeneratorFunc{
		generateIntRules, generateFloatRules, generateValueRules,
		generateCompareNumericRules, generateCompareStringRules,
		generateAddRules, generateInRules,
	}
	rules[0] = rule.NewTrue()
	for field := range inputDescription.Fields {
		if stringInSlice(field, ruleFields) {
			for _, ruleGenerator := range ruleGenerators {
				newRules := ruleGenerator(inputDescription, ruleFields, field)
				rules = append(rules, newRules...)
			}
		}
	}

	if len(ruleFields) == 2 {
		cRules := CombineRules(rules)
		rules = append(rules, cRules...)
	}
	rule.Sort(rules)
	return rules
}

func CombineRules(rules []rule.Rule) []rule.Rule {
	rule.Sort(rules)
	combinedRules := make([]rule.Rule, 0)
	numRules := len(rules)
	for i := 0; i < numRules-1; i++ {
		for j := i + 1; j < numRules; j++ {
			if andRule, err := rule.NewAnd(rules[i], rules[j]); err == nil {
				combinedRules = append(combinedRules, andRule)
			}
			if orRule, err := rule.NewOr(rules[i], rules[j]); err == nil {
				combinedRules = append(combinedRules, orRule)
			}
		}
	}
	return rule.Uniq(combinedRules)
}

func stringInSlice(s string, strings []string) bool {
	for _, x := range strings {
		if x == s {
			return true
		}
	}
	return false
}

func generateValueRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	rulesMap := make(map[string]rule.Rule)
	values := fd.Values
	if len(values) < 2 {
		return []rule.Rule{}
	}
	switch fd.Kind {
	case fieldtype.Int:
		for _, vd := range values {
			if vd.Num < 2 {
				continue
			}
			n, isInt := vd.Value.Int()
			if !isInt {
				panic(fmt.Sprintf("value isn't int: %s", vd.Value))
			}
			eqRule := rule.NewEQFVI(field, n)
			neRule := rule.NewNEFVI(field, n)
			rulesMap[eqRule.String()] = eqRule
			rulesMap[neRule.String()] = neRule
		}
	case fieldtype.Float:
		for _, vd := range values {
			if vd.Num < 2 {
				continue
			}
			n, isFloat := vd.Value.Float()
			if !isFloat {
				panic(fmt.Sprintf("value isn't float: %s", vd.Value))
			}
			eqRule := rule.NewEQFVF(field, n)
			neRule := rule.NewNEFVF(field, n)
			rulesMap[eqRule.String()] = eqRule
			rulesMap[neRule.String()] = neRule
		}
	case fieldtype.String:
		for _, vd := range values {
			if vd.Num < 2 {
				continue
			}
			s := vd.Value.String()
			eqRule := rule.NewEQFVS(field, s)
			rulesMap[eqRule.String()] = eqRule
			if len(values) > 2 {
				neRule := rule.NewNEFVS(field, s)
				rulesMap[neRule.String()] = neRule
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules
}

func generateIntRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.Int {
		return []rule.Rule{}
	}
	rulesMap := make(map[string]rule.Rule)
	points := internal.GeneratePoints(fd.Min, fd.Max, fd.MaxDP)

	for _, p := range points {
		pInt, pIsInt := p.Int()
		if !pIsInt {
			continue
		}
		rL := rule.NewLEFVI(field, pInt)
		rG := rule.NewGEFVI(field, pInt)
		rulesMap[rL.String()] = rL
		rulesMap[rG.String()] = rG
	}

	for iL, pL := range points {
		for iH, pH := range points {
			pLInt, pLIsInt := pL.Int()
			pHInt, pHIsInt := pH.Int()
			if !pLIsInt || !pHIsInt {
				continue
			}
			if iH > iL {
				rB, err := rule.NewBetweenFVI(field, pLInt, pHInt)
				if err == nil {
					rulesMap[rB.String()] = rB
				}
				rO, err := rule.NewOutsideFVI(field, pLInt, pHInt)
				if err == nil {
					rulesMap[rO.String()] = rO
				}
			}
		}
	}

	return rulesMapToArray(rulesMap)
}

func generateFloatRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.Float {
		return []rule.Rule{}
	}
	rulesMap := make(map[string]rule.Rule)
	points := internal.GeneratePoints(fd.Min, fd.Max, fd.MaxDP)

	for _, p := range points {
		pFloat, pIsFloat := p.Float()
		if !pIsFloat {
			continue
		}
		rL := rule.NewLEFVF(field, pFloat)
		rG := rule.NewGEFVF(field, pFloat)
		rulesMap[rL.String()] = rL
		rulesMap[rG.String()] = rG
	}

	for iL, pL := range points {
		for iH, pH := range points {
			pLFloat, pLIsFloat := pL.Float()
			pHFloat, pHIsFloat := pH.Float()
			if pLIsFloat && pHIsFloat && iH > iL {
				rB, err := rule.NewBetweenFVF(field, pLFloat, pHFloat)
				if err == nil {
					rulesMap[rB.String()] = rB
				}
				rO, err := rule.NewOutsideFVF(field, pLFloat, pHFloat)
				if err == nil {
					rulesMap[rO.String()] = rO
				}
			}
		}
	}

	return rulesMapToArray(rulesMap)
}

func generateCompareNumericRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.Int && fd.Kind != fieldtype.Float {
		return []rule.Rule{}
	}
	fieldNum := calcFieldNum(inputDescription.Fields, field)
	rulesMap := make(map[string]rule.Rule)
	ruleNewFuncs := []func(string, string) rule.Rule{
		rule.NewLTFF,
		rule.NewLEFF,
		rule.NewEQFF,
		rule.NewNEFF,
		rule.NewGEFF,
		rule.NewGTFF,
	}

	for oField, oFd := range inputDescription.Fields {
		oFieldNum := calcFieldNum(inputDescription.Fields, oField)
		isComparable := hasComparableNumberRange(fd, oFd)
		if fieldNum < oFieldNum && isComparable &&
			stringInSlice(oField, ruleFields) {
			for _, ruleNewFunc := range ruleNewFuncs {
				r := ruleNewFunc(field, oField)
				rulesMap[r.String()] = r
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules
}

func generateCompareStringRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	if fd.Kind != fieldtype.String {
		return []rule.Rule{}
	}
	fieldNum := calcFieldNum(inputDescription.Fields, field)
	rulesMap := make(map[string]rule.Rule)
	ruleNewFuncs := []func(string, string) rule.Rule{
		rule.NewEQFF,
		rule.NewNEFF,
	}
	for oField, oFd := range inputDescription.Fields {
		if oFd.Kind == fieldtype.String {
			oFieldNum := calcFieldNum(inputDescription.Fields, oField)
			numSharedValues := calcNumSharedValues(fd, oFd)
			if fieldNum < oFieldNum && numSharedValues >= 2 &&
				stringInSlice(oField, ruleFields) {
				for _, ruleNewFunc := range ruleNewFuncs {
					r := ruleNewFunc(field, oField)
					rulesMap[r.String()] = r
				}
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules
}

func generateAddRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	if !isNumberField(fd) {
		return []rule.Rule{}
	}
	fieldNum := calcFieldNum(inputDescription.Fields, field)
	rules := make([]rule.Rule, 0)

	for oField, oFd := range inputDescription.Fields {
		oFieldNum := calcFieldNum(inputDescription.Fields, oField)
		if fieldNum < oFieldNum && isNumberField(oFd) &&
			stringInSlice(oField, ruleFields) {
			vars := map[string]*dlit.Literal{
				"min":  fd.Min,
				"max":  fd.Max,
				"oMin": oFd.Min,
				"oMax": oFd.Max,
			}
			min := dexpr.Eval("min + oMin", dexprfuncs.CallFuncs, vars)
			max := dexpr.Eval("max + oMax", dexprfuncs.CallFuncs, vars)
			maxDP := fd.MaxDP
			if oFd.MaxDP > maxDP {
				maxDP = oFd.MaxDP
			}
			points := internal.GeneratePoints(min, max, maxDP)
			for _, p := range points {
				rL := rule.NewAddLEF(field, oField, p)
				rG := rule.NewAddGEF(field, oField, p)
				rules = append(rules, rL)
				rules = append(rules, rG)
			}
		}
	}
	return rules
}

func calcNumSharedValues(
	fd1 *description.Field,
	fd2 *description.Field,
) int {
	numShared := 0
	for _, vd1 := range fd1.Values {
		if _, ok := fd2.Values[vd1.Value.String()]; ok {
			numShared++
		}
	}
	return numShared
}

func isNumberField(fd *description.Field) bool {
	return fd.Kind == fieldtype.Int || fd.Kind == fieldtype.Float
}

var compareExpr *dexpr.Expr = dexpr.MustNew(
	"min1 < max2 && max1 > min2",
	dexprfuncs.CallFuncs,
)

func hasComparableNumberRange(
	fd1 *description.Field,
	fd2 *description.Field,
) bool {
	if !isNumberField(fd1) || !isNumberField(fd2) {
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

func rulesMapToArray(rulesMap map[string]rule.Rule) []rule.Rule {
	rules := make([]rule.Rule, len(rulesMap))
	i := 0
	for _, expr := range rulesMap {
		rules[i] = expr
		i++
	}
	return rules
}

// TODO: Allow more numValues if only two ruleFields
func generateInRules(
	inputDescription *description.Description,
	ruleFields []string,
	field string,
) []rule.Rule {
	fd := inputDescription.Fields[field]
	numValues := len(fd.Values)
	if fd.Kind != fieldtype.String &&
		fd.Kind != fieldtype.Float &&
		fd.Kind != fieldtype.Int ||
		numValues <= 3 || numValues > 12 {
		return []rule.Rule{}
	}
	rulesMap := make(map[string]rule.Rule)
	for i := 3; ; i++ {
		numOnBits := calcNumOnBits(i)
		if numOnBits >= numValues {
			break
		}
		if numOnBits >= 2 && numOnBits <= 5 && numOnBits < (numValues-1) {
			compareValues := makeCompareValues(fd.Values, i)
			if len(compareValues) >= 2 {
				r := rule.NewInFV(field, compareValues)
				rulesMap[r.String()] = r
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules
}

func makeCompareValues(
	values map[string]description.Value,
	i int,
) []*dlit.Literal {
	bStr := fmt.Sprintf("%b", i)
	numValues := len(values)
	lits := valuesToLiterals(values)
	j := numValues - 1
	compareValues := []*dlit.Literal{}
	for _, b := range reverseString(bStr) {
		if b == '1' {
			lit := lits[numValues-1-j]
			if values[lit.String()].Num < 2 {
				return []*dlit.Literal{}
			}
			compareValues = append(compareValues, lit)
		}
		j -= 1
	}
	return compareValues
}

func valuesToLiterals(values map[string]description.Value) []*dlit.Literal {
	lits := make([]*dlit.Literal, len(values))
	keys := make([]string, len(values))
	i := 0
	for k := range values {
		keys[i] = k
		i++
	}
	// The keys are sorted to make it easier to test because maps aren't ordered
	sort.Strings(keys)
	j := 0
	for _, k := range keys {
		lits[j] = values[k].Value
		j++
	}
	return lits
}

func reverseString(s string) (r string) {
	for _, v := range s {
		r = string(v) + r
	}
	return
}

func calcNumOnBits(i int) int {
	bStr := fmt.Sprintf("%b", i)
	return strings.Count(bStr, "1")
}

func calcFieldNum(
	fieldDescriptions map[string]*description.Field,
	fieldN string,
) int {
	fields := make([]string, len(fieldDescriptions))
	i := 0
	for field := range fieldDescriptions {
		fields[i] = field
		i++
	}
	sort.Strings(fields)
	j := 0
	for _, field := range fields {
		if field == fieldN {
			return j
		}
		j++
	}
	panic("can't find field in fieldDescriptions")
}
