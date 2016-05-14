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
package internal

import (
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
)

type AccuracyAggregator struct {
	name       string
	numMatches int64
	expr       *dexpr.Expr
}

var accuracyExpr = dexpr.MustNew("roundto(100*numMatches/numRecords,2)")

func NewAccuracyAggregator(name string, expr string) (*AccuracyAggregator, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	ca :=
		&AccuracyAggregator{name: name, numMatches: 0, expr: dexpr}
	return ca, nil
}

func (a *AccuracyAggregator) CloneNew() Aggregator {
	return &AccuracyAggregator{
		name:       a.name,
		numMatches: 0,
		expr:       a.expr,
	}
}

func (a *AccuracyAggregator) GetName() string {
	return a.name
}

func (a *AccuracyAggregator) GetArg() string {
	return a.expr.String()
}

func (a *AccuracyAggregator) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	matchExprIsTrue, err := a.expr.EvalBool(record, CallFuncs)
	if err != nil {
		return err
	}
	if isRuleTrue == matchExprIsTrue {
		a.numMatches++
	}
	return nil
}

func (a *AccuracyAggregator) GetResult(
	aggregators []Aggregator,
	goals []*Goal,
	numRecords int64,
) *dlit.Literal {
	if numRecords == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"numRecords": dlit.MustNew(numRecords),
		"numMatches": dlit.MustNew(a.numMatches),
	}
	return accuracyExpr.Eval(vars, CallFuncs)
}

func (a *AccuracyAggregator) IsEqual(o Aggregator) bool {
	if _, ok := o.(*AccuracyAggregator); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
