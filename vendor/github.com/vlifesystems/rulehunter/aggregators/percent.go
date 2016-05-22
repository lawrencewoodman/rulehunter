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

type percent struct {
	name       string
	numRecords int64
	numMatches int64
	expr       *dexpr.Expr
}

var percentExpr = dexpr.MustNew("roundto(100*numMatches/numRecords,2)")

func newPercent(name string, expr string) (*percent, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	ca :=
		&percent{name: name, numMatches: 0, numRecords: 0, expr: dexpr}
	return ca, nil
}

func (a *percent) CloneNew() Aggregator {
	return &percent{
		name:       a.name,
		numMatches: 0,
		numRecords: 0,
		expr:       a.expr,
	}
}

func (a *percent) GetName() string {
	return a.name
}

func (a *percent) GetArg() string {
	return a.expr.String()
}

func (a *percent) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	countExprIsTrue, err := a.expr.EvalBool(record, dexprfuncs.CallFuncs)
	if err != nil {
		return err
	}
	if isRuleTrue {
		a.numRecords++
		if countExprIsTrue {
			a.numMatches++
		}
	}
	return nil
}

func (a *percent) GetResult(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	if a.numRecords == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"numRecords": dlit.MustNew(a.numRecords),
		"numMatches": dlit.MustNew(a.numMatches),
	}
	return percentExpr.Eval(vars, dexprfuncs.CallFuncs)
}

func (a *percent) IsEqual(o Aggregator) bool {
	if _, ok := o.(*percent); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
