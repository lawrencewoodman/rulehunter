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
)

// AddLEF represents a rule determining if fieldA + fieldB <= floatValue
type AddLEF struct {
	fieldA string
	fieldB string
	value  *dlit.Literal
}

func init() {
	registerGenerator("AddLEF", generateAddLEF)
}

func NewAddLEF(fieldA string, fieldB string, value *dlit.Literal) *AddLEF {
	return &AddLEF{fieldA: fieldA, fieldB: fieldB, value: value}
}

func (r *AddLEF) String() string {
	return r.fieldA + " + " + r.fieldB + " <= " + r.value.String()
}

func (r *AddLEF) Value() *dlit.Literal {
	return r.value
}

func (r *AddLEF) Fields() []string {
	return []string{r.fieldA, r.fieldB}
}

// IsTrue returns whether the rule is true for this record.
// This rule relies on making sure that the two fields when
// added will not overflow, so this must have been checked
// before hand by looking at their max/min in the input description.
func (r *AddLEF) IsTrue(record ddataset.Record) (bool, error) {
	vA, ok := record[r.fieldA]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	vB, ok := record[r.fieldB]
	if !ok {
		return false, InvalidRuleError{Rule: r}
	}

	if vAInt, vAIsInt := vA.Int(); vAIsInt {
		if vBInt, vBIsInt := vB.Int(); vBIsInt {
			if i, iIsInt := r.value.Int(); iIsInt {
				return vAInt+vBInt <= i, nil
			}
		}
	}

	if vAFloat, vAIsFloat := vA.Float(); vAIsFloat {
		if vBFloat, vBIsFloat := vB.Float(); vBIsFloat {
			if f, fIsFloat := r.value.Float(); fIsFloat {
				return vAFloat+vBFloat <= f, nil
			}
		}
	}
	return false, IncompatibleTypesRuleError{Rule: r}
}

func (r *AddLEF) Tweak(
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
	min := dexpr.Eval("aMin + bMin", dexprfuncs.CallFuncs, vars)
	max := dexpr.Eval("aMax + bMax", dexprfuncs.CallFuncs, vars)
	points := generateTweakPoints(r.value, min, max, maxDP, stage)
	for _, p := range points {
		r := NewAddLEF(r.fieldA, r.fieldB, p)
		rules = append(rules, r)
	}
	return rules
}

func (r *AddLEF) Overlaps(o Rule) bool {
	switch x := o.(type) {
	case *AddLEF:
		oFields := x.Fields()
		if r.fieldA == oFields[0] && r.fieldB == oFields[1] {
			return true
		}
	}
	return false
}

func (r *AddLEF) DPReduce() []Rule {
	return roundRules(r.value, func(p *dlit.Literal) Rule {
		return NewAddLEF(r.fieldA, r.fieldB, p)
	})
}

func generateAddLEF(
	inputDescription *description.Description,
	generationDesc GenerationDescriber,
	field string,
) []Rule {
	fd := inputDescription.Fields[field]
	if !generationDesc.Arithmetic() || fd.Kind != description.Number {
		return []Rule{}
	}
	fieldNum := description.CalcFieldNum(inputDescription.Fields, field)
	rules := make([]Rule, 0)

	for oField, oFd := range inputDescription.Fields {
		oFieldNum := description.CalcFieldNum(inputDescription.Fields, oField)
		if fieldNum < oFieldNum &&
			oFd.Kind == description.Number &&
			internal.IsStringInSlice(oField, generationDesc.Fields()) {
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
				r := NewAddLEF(field, oField, p)
				rules = append(rules, r)
			}
		}
	}
	return rules
}
