// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/description"
	"github.com/vlifesystems/rhkit/internal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
	"github.com/vlifesystems/rhkit/internal/fieldtype"
)

// MulLEF represents a rule determining if fieldA * fieldB <= value
type MulLEF struct {
	fieldA string
	fieldB string
	value  *dlit.Literal
}

func init() {
	registerGenerator("MulLEF", generateMulLEF)
}

func NewMulLEF(fieldA string, fieldB string, value *dlit.Literal) *MulLEF {
	return &MulLEF{fieldA: fieldA, fieldB: fieldB, value: value}
}

func (r *MulLEF) String() string {
	return r.fieldA + " * " + r.fieldB + " <= " + r.value.String()
}

func (r *MulLEF) Fields() []string {
	return []string{r.fieldA, r.fieldB}
}

func (r *MulLEF) Value() *dlit.Literal {
	return r.value
}

// IsTrue returns whether the rule is true for this record.
// This rule relies on making sure that the two fields when
// added will not overflow, so this must have been checked
// beforehand by looking at their max/min in the input description.
func (r *MulLEF) IsTrue(record ddataset.Record) (bool, error) {
	vA, ok := record[r.fieldA]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	vB, ok := record[r.fieldB]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	vAInt, vAIsInt := vA.Int()
	if vAIsInt {
		vBInt, vBIsInt := vB.Int()
		if vBIsInt {
			if i, ok := r.value.Int(); ok {
				return vAInt*vBInt <= i, nil
			}
		}
	}

	vAFloat, vAIsFloat := vA.Float()
	vBFloat, vBIsFloat := vB.Float()
	valueFloat, valueIsFloat := r.value.Float()
	if !vAIsFloat || !vBIsFloat || !valueIsFloat {
		return false, IncompatibleTypesRuleError{Rule: r}
	}

	return vAFloat*vBFloat <= valueFloat, nil
}

func (r *MulLEF) Tweak(
	inputDescription *description.Description,
	stage int,
) []Rule {
	vars := map[string]*dlit.Literal{
		"aMin": inputDescription.Fields[r.fieldA].Min,
		"bMin": inputDescription.Fields[r.fieldB].Min,
		"aMax": inputDescription.Fields[r.fieldA].Max,
		"bMax": inputDescription.Fields[r.fieldB].Max,
	}
	maxDP := inputDescription.Fields[r.fieldA].MaxDP
	bMaxDP := inputDescription.Fields[r.fieldB].MaxDP
	if bMaxDP > maxDP {
		maxDP = bMaxDP
	}
	rules := make([]Rule, 0)
	min := dexpr.Eval("aMin * bMin", dexprfuncs.CallFuncs, vars)
	max := dexpr.Eval("aMax * bMax", dexprfuncs.CallFuncs, vars)
	points := generateTweakPoints(r.value, min, max, maxDP, stage)
	for _, p := range points {
		r := NewMulLEF(r.fieldA, r.fieldB, p)
		rules = append(rules, r)
	}
	return rules
}

func (r *MulLEF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *MulLEF:
		oFields := x.Fields()
		if r.fieldA == oFields[0] && r.fieldB == oFields[1] {
			return true
		}
	}
	return false
}

func (r *MulLEF) DPReduce() []Rule {
	return roundRules(r.value, func(p *dlit.Literal) Rule {
		return NewMulLEF(r.fieldA, r.fieldB, p)
	})
}

func generateMulLEF(
	inputDescription *description.Description,
	ruleFields []string,
	complexity Complexity,
	field string,
) []Rule {
	fd := inputDescription.Fields[field]
	if !complexity.Arithmetic || fd.Kind != fieldtype.Number {
		return []Rule{}
	}
	fieldNum := description.CalcFieldNum(inputDescription.Fields, field)
	rules := make([]Rule, 0)

	for oField, oFd := range inputDescription.Fields {
		oFieldNum := description.CalcFieldNum(inputDescription.Fields, oField)
		if fieldNum < oFieldNum &&
			oFd.Kind == fieldtype.Number &&
			internal.StringInSlice(oField, ruleFields) {
			vars := map[string]*dlit.Literal{
				"min":  fd.Min,
				"max":  fd.Max,
				"oMin": oFd.Min,
				"oMax": oFd.Max,
			}
			min := dexpr.Eval("min * oMin", dexprfuncs.CallFuncs, vars)
			max := dexpr.Eval("max * oMax", dexprfuncs.CallFuncs, vars)
			maxDP := fd.MaxDP
			if oFd.MaxDP > maxDP {
				maxDP = oFd.MaxDP
			}
			points := internal.GeneratePoints(min, max, maxDP)
			for _, p := range points {
				r := NewMulLEF(field, oField, p)
				rules = append(rules, r)
			}
		}
	}
	return rules
}
