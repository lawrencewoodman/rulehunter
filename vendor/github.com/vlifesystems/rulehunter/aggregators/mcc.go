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

package aggregators

import (
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/goal"
	"github.com/vlifesystems/rulehunter/internal/dexprfuncs"
)

// mcc represents a Matthews correlation coefficient aggregator
// see: https://en.wikipedia.org/wiki/Matthews_correlation_coefficient
type mcc struct {
	name              string
	numTruePositives  int64
	numTrueNegatives  int64
	numFalsePositives int64
	numFalseNegatives int64
	expr              *dexpr.Expr
}

var mccExpr = dexpr.MustNew(
	"((tp*tn)-(fp*fn))/sqrt((tp+fp)*(tp+fn)*(tn+fp)*(tn+fn))",
)

func newMCC(name string, expr string) (*mcc, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	a := &mcc{
		name:              name,
		numTruePositives:  0,
		numTrueNegatives:  0,
		numFalsePositives: 0,
		numFalseNegatives: 0,
		expr:              dexpr,
	}
	return a, nil
}

func (a *mcc) CloneNew() Aggregator {
	return &mcc{
		name:              a.name,
		numTruePositives:  0,
		numTrueNegatives:  0,
		numFalsePositives: 0,
		numFalseNegatives: 0,
		expr:              a.expr,
	}
}

func (a *mcc) GetName() string {
	return a.name
}

func (a *mcc) GetArg() string {
	return a.expr.String()
}

func (a *mcc) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	matchExprIsTrue, err := a.expr.EvalBool(record, dexprfuncs.CallFuncs)
	if err != nil {
		return err
	}
	if matchExprIsTrue {
		if isRuleTrue {
			a.numTruePositives++
		} else {
			a.numFalseNegatives++
		}
	} else {
		if isRuleTrue {
			a.numFalsePositives++
		} else {
			a.numTrueNegatives++
		}
	}

	return nil
}

func (a *mcc) GetResult(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	if numRecords == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"tp": dlit.MustNew(a.numTruePositives),
		"tn": dlit.MustNew(a.numTrueNegatives),
		"fp": dlit.MustNew(a.numFalsePositives),
		"fn": dlit.MustNew(a.numFalseNegatives),
	}
	sums := (a.numTruePositives + a.numFalsePositives) *
		(a.numTruePositives + a.numFalseNegatives) *
		(a.numTrueNegatives + a.numFalsePositives) *
		(a.numTrueNegatives + a.numFalseNegatives)
	if sums == 0 {
		return dlit.MustNew(0)
	}
	return mccExpr.Eval(vars, dexprfuncs.CallFuncs)
}

func (a *mcc) IsEqual(o Aggregator) bool {
	if _, ok := o.(*mcc); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
