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

type CountAggregator struct {
	name       string
	numMatches int64
	expr       *dexpr.Expr
}

func NewCountAggregator(name string, expr string) (*CountAggregator, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	ca := &CountAggregator{name: name, numMatches: 0, expr: dexpr}
	return ca, nil
}

func (a *CountAggregator) CloneNew() Aggregator {
	return &CountAggregator{name: a.name, numMatches: 0, expr: a.expr}
}

func (a *CountAggregator) GetName() string {
	return a.name
}

func (a *CountAggregator) GetArg() string {
	return a.expr.String()
}

func (a *CountAggregator) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	countExprIsTrue, err := a.expr.EvalBool(record, CallFuncs)
	if err != nil {
		return err
	}
	if isRuleTrue && countExprIsTrue {
		a.numMatches++
	}
	return nil
}

func (a *CountAggregator) GetResult(
	aggregators []Aggregator,
	goals []*Goal,
	numRecords int64,
) *dlit.Literal {
	l := dlit.MustNew(a.numMatches)
	return l
}

func (a *CountAggregator) IsEqual(o Aggregator) bool {
	if _, ok := o.(*CountAggregator); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
