/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
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
	case *BetweenFVI:
		aStr = "(" + aStr + ")"
	case *BetweenFVF:
		aStr = "(" + aStr + ")"
	case *OutsideFVI:
		aStr = "(" + aStr + ")"
	case *OutsideFVF:
		aStr = "(" + aStr + ")"
	}
	switch r.ruleB.(type) {
	case *And:
		bStr = "(" + bStr + ")"
	case *Or:
		bStr = "(" + bStr + ")"
	case *BetweenFVI:
		bStr = "(" + bStr + ")"
	case *BetweenFVF:
		bStr = "(" + bStr + ")"
	case *OutsideFVI:
		bStr = "(" + bStr + ")"
	case *OutsideFVF:
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

func (r *Or) GetFields() []string {
	results := []string{}
	mResults := map[string]interface{}{}
	for _, f := range r.ruleA.GetFields() {
		if _, ok := mResults[f]; !ok {
			mResults[f] = nil
			results = append(results, f)
		}
	}
	for _, f := range r.ruleB.GetFields() {
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

	if len(ruleA.GetFields()) != 1 || len(ruleB.GetFields()) != 1 {
		return true, nil
	}
	fieldA := ruleA.GetFields()[0]
	fieldB := ruleB.GetFields()[0]

	if fieldA == fieldB {
		_, ruleAIsBetweenFVI := ruleA.(*BetweenFVI)
		_, ruleAIsBetweenFVF := ruleA.(*BetweenFVF)
		_, ruleBIsBetweenFVI := ruleB.(*BetweenFVI)
		_, ruleBIsBetweenFVF := ruleB.(*BetweenFVF)
		_, ruleAIsOutsideFVI := ruleA.(*OutsideFVI)
		_, ruleAIsOutsideFVF := ruleA.(*OutsideFVF)
		_, ruleBIsOutsideFVI := ruleB.(*OutsideFVI)
		_, ruleBIsOutsideFVF := ruleB.(*OutsideFVF)
		if ruleAIsBetweenFVI || ruleBIsBetweenFVI ||
			ruleAIsBetweenFVF || ruleBIsBetweenFVF {
			return true, nil
		}

		if ruleAIsOutsideFVI || ruleBIsOutsideFVI ||
			ruleAIsOutsideFVF || ruleBIsOutsideFVF {
			return false, nil
		}

		GEFVIRuleA, ruleAIsGEFVI := ruleA.(*GEFVI)
		LEFVIRuleA, ruleAIsLEFVI := ruleA.(*LEFVI)
		GEFVIRuleB, ruleBIsGEFVI := ruleB.(*GEFVI)
		LEFVIRuleB, ruleBIsLEFVI := ruleB.(*LEFVI)
		GEFVFRuleA, ruleAIsGEFVF := ruleA.(*GEFVF)
		LEFVFRuleA, ruleAIsLEFVF := ruleA.(*LEFVF)
		GEFVFRuleB, ruleBIsGEFVF := ruleB.(*GEFVF)
		LEFVFRuleB, ruleBIsLEFVF := ruleB.(*LEFVF)

		if (ruleAIsGEFVI && ruleBIsGEFVI) ||
			(ruleAIsLEFVI && ruleBIsLEFVI) ||
			(ruleAIsGEFVF && ruleBIsGEFVF) ||
			(ruleAIsLEFVF && ruleBIsLEFVF) {
			return false, nil
		}

		if ruleAIsGEFVI && ruleBIsLEFVI {
			r, err = NewOutsideFVI(
				fieldA,
				LEFVIRuleB.GetValue(),
				GEFVIRuleA.GetValue(),
			)
		} else if ruleAIsLEFVI && ruleBIsGEFVI {
			r, err = NewOutsideFVI(
				fieldA,
				LEFVIRuleA.GetValue(),
				GEFVIRuleB.GetValue(),
			)
		} else if ruleAIsGEFVF && ruleBIsLEFVF {
			r, err = NewOutsideFVF(
				fieldA,
				LEFVFRuleB.GetValue(),
				GEFVFRuleA.GetValue(),
			)
		} else if ruleAIsLEFVF && ruleBIsGEFVF {
			r, err = NewOutsideFVF(
				fieldA,
				LEFVFRuleA.GetValue(),
				GEFVFRuleB.GetValue(),
			)
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
	if ruleA.GetFields()[0] != ruleB.GetFields()[0] {
		return &Or{ruleA: ruleA, ruleB: ruleB}
	}
	newValues := []*dlit.Literal{}
	mNewValues := map[string]interface{}{}
	for _, v := range ruleA.GetValues() {
		if _, ok := mNewValues[v.String()]; !ok {
			mNewValues[v.String()] = nil
			newValues = append(newValues, v)
		}
	}
	for _, v := range ruleB.GetValues() {
		if _, ok := mNewValues[v.String()]; !ok {
			mNewValues[v.String()] = nil
			newValues = append(newValues, v)
		}
	}
	return NewInFV(ruleA.GetFields()[0], newValues)
}
