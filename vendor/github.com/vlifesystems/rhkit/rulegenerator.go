/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package rhkit

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/rule"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ruleGeneratorFunc func(
	*Description,
	[]string,
	string,
) ([]rule.Rule, error)

func GenerateRules(
	inputDescription *Description,
	excludeFields []string,
) ([]rule.Rule, error) {
	rules := make([]rule.Rule, 1)
	ruleGenerators := []ruleGeneratorFunc{
		generateIntRules, generateFloatRules, generateStringRules,
		generateCompareNumericRules, generateCompareStringRules,
		generateInNiRules,
	}
	rules[0] = rule.NewTrue()
	for field, _ := range inputDescription.fields {
		if !stringInSlice(field, excludeFields) {
			for _, ruleGenerator := range ruleGenerators {
				newRules, err := ruleGenerator(inputDescription, excludeFields, field)
				if err != nil {
					return nil, err
				}
				rules = append(rules, newRules...)
			}
		}
	}
	return rules, nil
}

func CombineRules(rules []rule.Rule) []rule.Rule {
	combinedRules := make([]rule.Rule, 0)
	numRules := len(rules)
	for i := 0; i < numRules-1; i++ {
		for j := i + 1; j < numRules; j++ {
			_, ruleIIsTrue := rules[i].(rule.True)
			_, ruleJIsTrue := rules[j].(rule.True)
			if !ruleIIsTrue && !ruleJIsTrue {
				andRule := rule.NewAnd(rules[i], rules[j])
				orRule := rule.NewOr(rules[i], rules[j])
				combinedRules = append(combinedRules, andRule)
				combinedRules = append(combinedRules, orRule)
			}
		}
	}
	return combinedRules
}

func stringInSlice(s string, strings []string) bool {
	for _, x := range strings {
		if x == s {
			return true
		}
	}
	return false
}

func generateIntRules(
	inputDescription *Description,
	excludeFields []string,
	field string,
) ([]rule.Rule, error) {
	fd := inputDescription.fields[field]
	if fd.kind != ftInt {
		return []rule.Rule{}, nil
	}
	rulesMap := make(map[string]rule.Rule)
	min, _ := fd.min.Int()
	max, _ := fd.max.Int()
	values := fd.values
	diff := max - min
	step := diff / 10
	if step == 0 {
		step = 1
	}
	// i set to 0 to make more tweakable
	for i := int64(0); i < diff; i += step {
		n := min + i
		r := rule.NewGEFVI(field, n)
		rulesMap[r.String()] = r
	}

	for i := step; i <= diff; i += step {
		n := min + i
		r := rule.NewLEFVI(field, n)
		rulesMap[r.String()] = r
	}

	if len(values) >= 2 {
		for _, v := range values {
			n, isInt := v.Int()
			if !isInt {
				return nil, errors.New(fmt.Sprintf("value isn't int: %s", v))
			}
			eqRule := rule.NewEQFVI(field, n)
			rulesMap[eqRule.String()] = eqRule
			neRule := rule.NewNEFVI(field, n)
			rulesMap[neRule.String()] = neRule
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules, nil
}

var matchTrailingZerosRegexp = regexp.MustCompile("^(\\d+.[^ 0])(0+)$")
var matchAllZerosRegexp = regexp.MustCompile("^(0)(.0+)$")

func floatToString(f float64, maxDP int) string {
	fStr := fmt.Sprintf("%.*f", maxDP, f)
	fStr = matchTrailingZerosRegexp.ReplaceAllString(fStr, "$1")
	return matchAllZerosRegexp.ReplaceAllString(fStr, "0")
}

func truncateFloat(f float64, maxDP int) float64 {
	v := fmt.Sprintf("%.*f", maxDP, f)
	nf, _ := strconv.ParseFloat(v, 64)
	return nf
}

// TODO: For each rule give all dp numbers 0..maxDP
func generateFloatRules(
	inputDescription *Description,
	excludeFields []string, field string) ([]rule.Rule, error) {
	fd := inputDescription.fields[field]
	if fd.kind != ftFloat {
		return []rule.Rule{}, nil
	}
	rulesMap := make(map[string]rule.Rule)
	min, _ := fd.min.Float()
	max, _ := fd.max.Float()
	maxDP := fd.maxDP
	values := fd.values
	diff := max - min
	step := diff / 10.0

	// i set to 0 to make more tweakable
	for i := float64(0); i < diff; i += step {
		n := truncateFloat(min+i, maxDP)
		r := rule.NewGEFVF(field, n)
		rulesMap[r.String()] = r
	}

	for i := step; i <= diff; i += step {
		n := truncateFloat(min+i, maxDP)
		r := rule.NewLEFVF(field, n)
		rulesMap[r.String()] = r
	}

	if len(values) >= 2 {
		for _, v := range values {
			n, isFloat := v.Float()
			if !isFloat {
				return nil, errors.New(fmt.Sprintf("value isn't float: %s", v))
			}
			tn := truncateFloat(n, maxDP)
			eqRule := rule.NewEQFVF(field, tn)
			neRule := rule.NewNEFVF(field, tn)
			rulesMap[eqRule.String()] = eqRule
			rulesMap[neRule.String()] = neRule
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules, nil
}

func generateCompareNumericRules(
	inputDescription *Description,
	excludeFields []string,
	field string,
) ([]rule.Rule, error) {
	fd := inputDescription.fields[field]
	if fd.kind != ftInt && fd.kind != ftFloat {
		return []rule.Rule{}, nil
	}
	fieldNum := calcFieldNum(inputDescription.fields, field)
	rulesMap := make(map[string]rule.Rule)
	ruleNewFuncs := []func(string, string) rule.Rule{
		rule.NewLTFF,
		rule.NewLEFF,
		rule.NewEQFF,
		rule.NewNEFF,
		rule.NewGEFF,
		rule.NewGTFF,
	}

	for oField, oFd := range inputDescription.fields {
		oFieldNum := calcFieldNum(inputDescription.fields, oField)
		isComparable := hasComparableNumberRange(fd, oFd)
		if fieldNum < oFieldNum &&
			isComparable &&
			!stringInSlice(oField, excludeFields) {
			for _, ruleNewFunc := range ruleNewFuncs {
				r := ruleNewFunc(field, oField)
				rulesMap[r.String()] = r
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules, nil
}

func generateCompareStringRules(
	inputDescription *Description,
	excludeFields []string,
	field string,
) ([]rule.Rule, error) {
	fd := inputDescription.fields[field]
	if fd.kind != ftString {
		return []rule.Rule{}, nil
	}
	fieldNum := calcFieldNum(inputDescription.fields, field)
	rulesMap := make(map[string]rule.Rule)
	ruleNewFuncs := []func(string, string) rule.Rule{
		rule.NewEQFF,
		rule.NewNEFF,
	}
	for oField, oFd := range inputDescription.fields {
		if oFd.kind == ftString {
			oFieldNum := calcFieldNum(inputDescription.fields, oField)
			numSharedValues := calcNumSharedValues(fd, oFd)
			if fieldNum < oFieldNum &&
				numSharedValues >= 2 &&
				!stringInSlice(oField, excludeFields) {
				for _, ruleNewFunc := range ruleNewFuncs {
					r := ruleNewFunc(field, oField)
					rulesMap[r.String()] = r
				}
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules, nil
}

func calcNumSharedValues(
	fd1 *fieldDescription,
	fd2 *fieldDescription,
) int {
	numShared := 0
	for _, v1 := range fd1.values {
		for _, v2 := range fd2.values {
			if v1.String() == v2.String() {
				numShared++
			}
		}
	}
	return numShared
}

func generateStringRules(
	inputDescription *Description,
	excludeFields []string,
	field string,
) ([]rule.Rule, error) {
	fd := inputDescription.fields[field]
	if fd.kind != ftString {
		return []rule.Rule{}, nil
	}
	rulesMap := make(map[string]rule.Rule)

	for _, v := range fd.values {
		s := v.String()
		eqRule := rule.NewEQFVS(field, s)
		rulesMap[eqRule.String()] = eqRule
		if len(fd.values) > 2 {
			neRule := rule.NewNEFVS(field, s)
			rulesMap[neRule.String()] = neRule
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules, nil
}

func isNumberField(fd *fieldDescription) bool {
	return fd.kind == ftInt || fd.kind == ftFloat
}

func hasComparableNumberRange(
	fd1 *fieldDescription,
	fd2 *fieldDescription,
) bool {
	if !isNumberField(fd1) || !isNumberField(fd2) {
		return false
	}
	var isComparable bool
	vars := map[string]*dlit.Literal{
		"min1": fd1.min,
		"max1": fd1.max,
		"min2": fd2.min,
		"max2": fd2.max,
	}
	funcs := map[string]dexpr.CallFun{}
	expr, err := dexpr.New("min1 < max2 && max1 > min2")
	if err != nil {
		return false
	}
	isComparable, err = expr.EvalBool(vars, funcs)
	if err != nil {
		return false
	}
	return isComparable
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

func generateInNiRules(
	inputDescription *Description,
	excludeFields []string,
	field string,
) ([]rule.Rule, error) {
	fd := inputDescription.fields[field]
	numValues := len(fd.values)
	if fd.kind != ftString &&
		fd.kind != ftFloat &&
		fd.kind != ftInt ||
		numValues <= 3 || numValues > 12 {
		return []rule.Rule{}, nil
	}
	rulesMap := make(map[string]rule.Rule)
	ruleNewFuncs := []func(string, []*dlit.Literal) rule.Rule{
		rule.NewInFV,
		rule.NewNiFV,
	}
	for i := 3; ; i++ {
		numOnBits := calcNumOnBits(i)
		if numOnBits >= numValues {
			break
		}
		if numOnBits >= 2 && numOnBits <= 5 && numOnBits < (numValues-1) {
			compareValues := makeCompareValues(fd.values, i)
			for _, ruleNewFunc := range ruleNewFuncs {
				r := ruleNewFunc(field, compareValues)
				rulesMap[r.String()] = r
			}
		}
	}
	rules := rulesMapToArray(rulesMap)
	return rules, nil
}

func makeCompareValues(values []*dlit.Literal, i int) []*dlit.Literal {
	bStr := fmt.Sprintf("%b", i)
	numOnBits := calcNumOnBits(i)
	numValues := len(values)
	j := numValues - 1
	compareValues := make([]*dlit.Literal, numOnBits)
	k := 0
	for _, b := range reverseString(bStr) {
		if b == '1' {
			compareValues[k] = values[numValues-1-j]
			k++
		}
		j -= 1
	}
	return compareValues
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
	fieldDescriptions map[string]*fieldDescription,
	fieldN string,
) int {
	fields := make([]string, len(fieldDescriptions))
	i := 0
	for field, _ := range fieldDescriptions {
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
	panic("Can't field field in fieldDescriptions")
}
