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
)

// And represents a rule determening if ruleA AND ruleB
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
	fieldsA := ruleA.GetFields()
	fieldsB := ruleB.GetFields()
	if len(fieldsA) == 1 && len(fieldsB) == 1 && fieldsA[0] == fieldsB[0] {
		return false, nil
	}
	return false, &And{ruleA: ruleA, ruleB: ruleB}
}

func tryEqNeRule(ruleA, ruleB Rule) (skip bool, newRule Rule) {
	ruleAFields := ruleA.GetFields()
	ruleBFields := ruleB.GetFields()
	if len(ruleAFields) != 1 && len(ruleBFields) != 1 {
		return true, nil
	}
	fieldA := ruleA.GetFields()[0]
	fieldB := ruleB.GetFields()[0]

	if fieldA != fieldB {
		return true, nil
	}
	switch ruleA.(type) {
	case *EQFVI:
	case *EQFVF:
	case *EQFVS:
	case *NEFVI:
	case *NEFVF:
	case *NEFVS:
	default:
		return true, nil
	}
	switch ruleB.(type) {
	case *EQFVI:
	case *EQFVF:
	case *EQFVS:
	case *NEFVI:
	case *NEFVF:
	case *NEFVS:
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
		OutsideFVIRuleA, ruleAIsOutsideFVI := ruleA.(*OutsideFVI)
		OutsideFVFRuleA, ruleAIsOutsideFVF := ruleA.(*OutsideFVF)
		OutsideFVIRuleB, ruleBIsOutsideFVI := ruleB.(*OutsideFVI)
		OutsideFVFRuleB, ruleBIsOutsideFVF := ruleB.(*OutsideFVF)

		if (ruleAIsBetweenFVI && !ruleBIsOutsideFVI) ||
			(ruleAIsBetweenFVF && !ruleBIsOutsideFVF) ||
			(!ruleAIsOutsideFVI && ruleBIsBetweenFVI) ||
			(!ruleAIsOutsideFVF && ruleBIsBetweenFVF) ||
			(ruleAIsOutsideFVI && ruleBIsOutsideFVI) ||
			(ruleAIsOutsideFVF && ruleBIsOutsideFVF) {
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

		if (ruleAIsOutsideFVI && ruleBIsLEFVI &&
			OutsideFVIRuleA.GetHigh() >= LEFVIRuleB.GetValue()) ||
			(ruleAIsOutsideFVI && ruleBIsGEFVI &&
				OutsideFVIRuleA.GetLow() <= GEFVIRuleB.GetValue()) ||
			(ruleAIsOutsideFVF && ruleBIsLEFVF &&
				OutsideFVFRuleA.GetHigh() >= LEFVFRuleB.GetValue()) ||
			(ruleAIsOutsideFVF && ruleBIsGEFVF &&
				OutsideFVFRuleA.GetLow() <= GEFVFRuleB.GetValue()) ||
			(ruleBIsOutsideFVI && ruleAIsLEFVI &&
				OutsideFVIRuleB.GetHigh() >= LEFVIRuleA.GetValue()) ||
			(ruleBIsOutsideFVI && ruleAIsGEFVI &&
				OutsideFVIRuleB.GetLow() <= GEFVIRuleA.GetValue()) ||
			(ruleBIsOutsideFVF && ruleAIsLEFVF &&
				OutsideFVFRuleB.GetHigh() >= LEFVFRuleA.GetValue()) ||
			(ruleBIsOutsideFVF && ruleAIsGEFVF &&
				OutsideFVFRuleB.GetLow() <= GEFVFRuleA.GetValue()) {
			return false, nil
		}

		if ruleAIsGEFVI && ruleBIsLEFVI {
			r, err = NewBetweenFVI(
				fieldA,
				GEFVIRuleA.GetValue(),
				LEFVIRuleB.GetValue(),
			)
		} else if ruleAIsLEFVI && ruleBIsGEFVI {
			r, err = NewBetweenFVI(
				fieldA,
				GEFVIRuleB.GetValue(),
				LEFVIRuleA.GetValue(),
			)
		} else if ruleAIsGEFVF && ruleBIsLEFVF {
			r, err = NewBetweenFVF(
				fieldA,
				GEFVFRuleA.GetValue(),
				LEFVFRuleB.GetValue(),
			)
		} else if ruleAIsLEFVF && ruleBIsGEFVF {
			r, err = NewBetweenFVF(
				fieldA,
				GEFVFRuleB.GetValue(),
				LEFVFRuleA.GetValue(),
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

func (r *And) GetFields() []string {
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
