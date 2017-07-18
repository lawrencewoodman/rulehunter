// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

// And represents a rule determining if ruleA AND ruleB
type And struct {
	ruleA Rule
	ruleB Rule
}

func NewAnd(ruleA Rule, ruleB Rule) (Rule, error) {
	_, ruleAIsTrue := ruleA.(True)
	_, ruleBIsTrue := ruleB.(True)
	if ruleAIsTrue || ruleBIsTrue {
		return nil, fmt.Errorf("can't And rule: %s, with: %s", ruleA, ruleB)
	}

	if skip, r := tryJoinRulesWithBetween(ruleA, ruleB); !skip {
		if r != nil {
			return r, nil
		}
		return nil, fmt.Errorf("can't And rule: %s, with: %s", ruleA, ruleB)
	}
	if skip, r := tryInRule(ruleA, ruleB); !skip {
		if r != nil {
			return r, nil
		}
		return nil, fmt.Errorf("can't And rule: %s, with: %s", ruleA, ruleB)
	}
	if skip, r := tryEqNeRule(ruleA, ruleB); !skip {
		if r != nil {
			return r, nil
		}
		return nil, fmt.Errorf("can't And rule: %s, with: %s", ruleA, ruleB)
	}
	return &And{ruleA: ruleA, ruleB: ruleB}, nil
}

func tryInRule(ruleA, ruleB Rule) (skip bool, newRule Rule) {
	_, ruleAIsIn := ruleA.(*InFV)
	_, ruleBIsIn := ruleB.(*InFV)
	if !ruleAIsIn && !ruleBIsIn {
		return true, nil
	}
	fieldsA := ruleA.Fields()
	fieldsB := ruleB.Fields()
	if len(fieldsA) == 1 && len(fieldsB) == 1 && fieldsA[0] == fieldsB[0] {
		return false, nil
	}
	return false, &And{ruleA: ruleA, ruleB: ruleB}
}

func tryEqNeRule(ruleA, ruleB Rule) (skip bool, newRule Rule) {
	ruleAFields := ruleA.Fields()
	ruleBFields := ruleB.Fields()
	if len(ruleAFields) != 1 && len(ruleBFields) != 1 {
		return true, nil
	}
	fieldA := ruleA.Fields()[0]
	fieldB := ruleB.Fields()[0]

	if fieldA != fieldB {
		return true, nil
	}
	switch ruleA.(type) {
	case *EQFV:
	case *NEFV:
	default:
		return true, nil
	}
	switch ruleB.(type) {
	case *EQFV:
	case *NEFV:
	default:
		return false, &And{ruleA: ruleA, ruleB: ruleB}
	}
	return false, nil
}

func tryJoinRulesWithBetween(
	ruleA Rule,
	ruleB Rule,
) (skip bool, newRule Rule) {
	var r Rule
	var err error

	if len(ruleA.Fields()) != 1 || len(ruleB.Fields()) != 1 {
		return true, nil
	}
	fieldA := ruleA.Fields()[0]
	fieldB := ruleB.Fields()[0]

	if fieldA != fieldB {
		return true, nil
	}
	_, ruleAIsBetweenFV := ruleA.(*BetweenFV)
	_, ruleBIsBetweenFV := ruleB.(*BetweenFV)
	OutsideFVRuleA, ruleAIsOutsideFV := ruleA.(*OutsideFV)
	OutsideFVRuleB, ruleBIsOutsideFV := ruleB.(*OutsideFV)

	if (ruleAIsBetweenFV && !ruleBIsOutsideFV) ||
		(!ruleAIsOutsideFV && ruleBIsBetweenFV) ||
		(ruleAIsOutsideFV && ruleBIsOutsideFV) {
		return false, nil
	}

	GEFVRuleA, ruleAIsGEFV := ruleA.(*GEFV)
	LEFVRuleA, ruleAIsLEFV := ruleA.(*LEFV)
	GEFVRuleB, ruleBIsGEFV := ruleB.(*GEFV)
	LEFVRuleB, ruleBIsLEFV := ruleB.(*LEFV)

	if (ruleAIsGEFV && ruleBIsGEFV) ||
		(ruleAIsLEFV && ruleBIsLEFV) {
		return false, nil
	}

	vars := map[string]*dlit.Literal{
		"ruleAIsLEFV":        dlit.MustNew(ruleAIsLEFV),
		"ruleBIsLEFV":        dlit.MustNew(ruleBIsLEFV),
		"ruleAIsGEFV":        dlit.MustNew(ruleAIsGEFV),
		"ruleBIsGEFV":        dlit.MustNew(ruleBIsGEFV),
		"ruleAIsOutsideFV":   dlit.MustNew(ruleAIsOutsideFV),
		"ruleBIsOutsideFV":   dlit.MustNew(ruleBIsOutsideFV),
		"LEFVRuleAValue":     dlit.MustNew(0),
		"LEFVRuleBValue":     dlit.MustNew(0),
		"GEFVRuleAValue":     dlit.MustNew(0),
		"GEFVRuleBValue":     dlit.MustNew(0),
		"OutsideFVRuleALow":  dlit.MustNew(0),
		"OutsideFVRuleAHigh": dlit.MustNew(0),
		"OutsideFVRuleBLow":  dlit.MustNew(0),
		"OutsideFVRuleBHigh": dlit.MustNew(0),
	}

	if ruleAIsLEFV {
		vars["LEFVRuleAValue"] = LEFVRuleA.Value()
	}
	if ruleBIsLEFV {
		vars["LEFVRuleBValue"] = LEFVRuleB.Value()
	}
	if ruleAIsGEFV {
		vars["GEFVRuleAValue"] = GEFVRuleA.Value()
	}
	if ruleBIsGEFV {
		vars["GEFVRuleBValue"] = GEFVRuleB.Value()
	}
	if ruleAIsOutsideFV {
		vars["OutsideFVRuleALow"] = OutsideFVRuleA.Low()
		vars["OutsideFVRuleAHigh"] = OutsideFVRuleA.High()
	}
	if ruleBIsOutsideFV {
		vars["OutsideFVRuleBLow"] = OutsideFVRuleB.Low()
		vars["OutsideFVRuleBHigh"] = OutsideFVRuleB.High()
	}
	invalidExpr, err := dexpr.EvalBool(
		"(ruleAIsOutsideFV && ruleBIsLEFV && "+
			" OutsideFVRuleAHigh >= LEFVRuleBValue) || "+
			"(ruleAIsOutsideFV && ruleBIsGEFV && "+
			" OutsideFVRuleALow <= GEFVRuleBValue) || "+
			"(ruleBIsOutsideFV && ruleAIsLEFV && "+
			" OutsideFVRuleBHigh >= LEFVRuleAValue) || "+
			"(ruleBIsOutsideFV && ruleAIsGEFV && "+
			"	OutsideFVRuleBLow <= GEFVRuleAValue)",
		dexprfuncs.CallFuncs,
		vars,
	)
	if invalidExpr || err != nil {
		return false, nil
	}

	if ruleAIsGEFV && ruleBIsLEFV {
		r, err = NewBetweenFV(fieldA, GEFVRuleA.Value(), LEFVRuleB.Value())
	} else if ruleAIsLEFV && ruleBIsGEFV {
		r, err = NewBetweenFV(fieldA, GEFVRuleB.Value(), LEFVRuleA.Value())
	} else {
		return true, nil
	}
	if err != nil {
		return false, nil
	}
	return false, r
}

func MustNewAnd(ruleA Rule, ruleB Rule) Rule {
	r, err := NewAnd(ruleA, ruleB)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *And) String() string {
	// TODO: Consider making this AND rather than &&
	aStr := r.ruleA.String()
	bStr := r.ruleB.String()
	switch r.ruleA.(type) {
	case *And:
		aStr = "(" + aStr + ")"
	case *Or:
		aStr = "(" + aStr + ")"
	case *BetweenFV:
		aStr = "(" + aStr + ")"
	case *OutsideFV:
		aStr = "(" + aStr + ")"
	}
	switch r.ruleB.(type) {
	case *And:
		bStr = "(" + bStr + ")"
	case *Or:
		bStr = "(" + bStr + ")"
	case *BetweenFV:
		bStr = "(" + bStr + ")"
	case *OutsideFV:
		bStr = "(" + bStr + ")"
	}
	return fmt.Sprintf("%s && %s", aStr, bStr)
}

func (r *And) IsTrue(record ddataset.Record) (bool, error) {
	lh, err := r.ruleA.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	rh, err := r.ruleB.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	return lh && rh, nil
}

func (r *And) Fields() []string {
	results := []string{}
	mResults := map[string]interface{}{}
	for _, f := range r.ruleA.Fields() {
		if _, ok := mResults[f]; !ok {
			mResults[f] = nil
			results = append(results, f)
		}
	}
	for _, f := range r.ruleB.Fields() {
		if _, ok := mResults[f]; !ok {
			mResults[f] = nil
			results = append(results, f)
		}
	}
	return results
}
