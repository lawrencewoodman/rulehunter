// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"fmt"
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dlit"
)

// Or represents a rule determining if ruleA OR ruleB
type Or struct {
	ruleA Rule
	ruleB Rule
}

func NewOr(ruleA Rule, ruleB Rule) (Rule, error) {
	_, ruleAIsTrue := ruleA.(True)
	_, ruleBIsTrue := ruleB.(True)
	if ruleAIsTrue || ruleBIsTrue {
		return nil, fmt.Errorf("can't Or rule: %s, with: %s", ruleA, ruleB)
	}
	inRuleA, ruleAIsIn := ruleA.(*InFV)
	inRuleB, ruleBIsIn := ruleB.(*InFV)
	if ruleAIsIn && ruleBIsIn {
		return handleInRules(inRuleA, inRuleB), nil
	}

	if skip, r := tryJoinRulesWithOutside(ruleA, ruleB); !skip {
		if r != nil {
			return r, nil
		}
		return nil, fmt.Errorf("can't Or rule: %s, with: %s", ruleA, ruleB)
	}

	return &Or{ruleA: ruleA, ruleB: ruleB}, nil
}

func MustNewOr(ruleA Rule, ruleB Rule) Rule {
	r, err := NewOr(ruleA, ruleB)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Or) String() string {
	// TODO: Consider making this OR rather than ||
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
	return fmt.Sprintf("%s || %s", aStr, bStr)
}

func (r *Or) IsTrue(record ddataset.Record) (bool, error) {
	lh, err := r.ruleA.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	rh, err := r.ruleB.IsTrue(record)
	if err != nil {
		return false, InvalidRuleError{Rule: r}
	}
	return lh || rh, nil
}

func (r *Or) Fields() []string {
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

func tryJoinRulesWithOutside(
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

	if fieldA == fieldB {
		_, ruleAIsBetweenFV := ruleA.(*BetweenFV)
		_, ruleBIsBetweenFV := ruleB.(*BetweenFV)
		_, ruleAIsOutsideFV := ruleA.(*OutsideFV)
		_, ruleBIsOutsideFV := ruleB.(*OutsideFV)
		if ruleAIsBetweenFV || ruleBIsBetweenFV {
			return true, nil
		}

		if ruleAIsOutsideFV || ruleBIsOutsideFV {
			return false, nil
		}

		GEFVRuleA, ruleAIsGEFV := ruleA.(*GEFV)
		LEFVRuleA, ruleAIsLEFV := ruleA.(*LEFV)
		GEFVRuleB, ruleBIsGEFV := ruleB.(*GEFV)
		LEFVRuleB, ruleBIsLEFV := ruleB.(*LEFV)

		if (ruleAIsGEFV && ruleBIsGEFV) || (ruleAIsLEFV && ruleBIsLEFV) {
			return false, nil
		}

		if ruleAIsGEFV && ruleBIsLEFV {
			r, err = NewOutsideFV(fieldA, LEFVRuleB.Value(), GEFVRuleA.Value())
		} else if ruleAIsLEFV && ruleBIsGEFV {
			r, err = NewOutsideFV(fieldA, LEFVRuleA.Value(), GEFVRuleB.Value())
		} else {
			return true, nil
		}
		if err != nil {
			return false, nil
		}
		return false, r
	}
	return true, nil
}

func handleInRules(ruleA, ruleB *InFV) Rule {
	if ruleA.Fields()[0] != ruleB.Fields()[0] {
		return &Or{ruleA: ruleA, ruleB: ruleB}
	}
	newValues := []*dlit.Literal{}
	mNewValues := map[string]interface{}{}
	for _, v := range ruleA.Values() {
		if _, ok := mNewValues[v.String()]; !ok {
			mNewValues[v.String()] = nil
			newValues = append(newValues, v)
		}
	}
	for _, v := range ruleB.Values() {
		if _, ok := mNewValues[v.String()]; !ok {
			mNewValues[v.String()] = nil
			newValues = append(newValues, v)
		}
	}
	return NewInFV(ruleA.Fields()[0], newValues)
}
