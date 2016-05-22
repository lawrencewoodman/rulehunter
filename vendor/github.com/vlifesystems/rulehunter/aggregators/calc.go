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

type calc struct {
	name string
	expr *dexpr.Expr
}

func newCalc(name string, expr string) (*calc, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	ca := &calc{name: name, expr: dexpr}
	return ca, nil
}

func (a *calc) CloneNew() Aggregator {
	return &calc{name: a.name, expr: a.expr}
}

func (a *calc) GetName() string {
	return a.name
}

func (a *calc) GetArg() string {
	return a.expr.String()
}

func (a *calc) NextRecord(
	record map[string]*dlit.Literal, isRuleTrue bool) error {
	return nil
}

func (a *calc) GetResult(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	aggregatorsMap, err :=
		AggregatorsToMap(aggregators, goals, numRecords, a.name)
	if err != nil {
		return dlit.MustNew(err)
	}
	return a.expr.Eval(aggregatorsMap, dexprfuncs.CallFuncs)
}

func (a *calc) IsEqual(o Aggregator) bool {
	if _, ok := o.(*calc); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
