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

type count struct {
	name       string
	numMatches int64
	expr       *dexpr.Expr
}

func newCount(name string, expr string) (*count, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	ca := &count{name: name, numMatches: 0, expr: dexpr}
	return ca, nil
}

func (a *count) CloneNew() Aggregator {
	return &count{name: a.name, numMatches: 0, expr: a.expr}
}

func (a *count) GetName() string {
	return a.name
}

func (a *count) GetArg() string {
	return a.expr.String()
}

func (a *count) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	countExprIsTrue, err := a.expr.EvalBool(record, dexprfuncs.CallFuncs)
	if err != nil {
		return err
	}
	if isRuleTrue && countExprIsTrue {
		a.numMatches++
	}
	return nil
}

func (a *count) GetResult(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	l := dlit.MustNew(a.numMatches)
	return l
}

func (a *count) IsEqual(o Aggregator) bool {
	if _, ok := o.(*count); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
