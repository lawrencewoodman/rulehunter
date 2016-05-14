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
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
)

type SumAggregator struct {
	name string
	sum  *dlit.Literal
	expr *dexpr.Expr
}

var sumExpr = dexpr.MustNew("sum+value")

func NewSumAggregator(name string, exprStr string) (*SumAggregator, error) {
	expr, err := dexpr.New(exprStr)
	if err != nil {
		return nil, err
	}
	ca := &SumAggregator{
		name: name,
		sum:  dlit.MustNew(0),
		expr: expr,
	}
	return ca, nil
}

func (a *SumAggregator) CloneNew() Aggregator {
	return &SumAggregator{
		name: a.name,
		sum:  dlit.MustNew(0),
		expr: a.expr,
	}
}

func (a *SumAggregator) GetName() string {
	return a.name
}

func (a *SumAggregator) GetArg() string {
	return a.expr.String()
}

func (a *SumAggregator) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	if isRuleTrue {
		exprValue := a.expr.Eval(record, CallFuncs)
		_, valueIsFloat := exprValue.Float()
		if !valueIsFloat {
			return fmt.Errorf("Value isn't a float: %s", exprValue)
		}

		vars := map[string]*dlit.Literal{
			"sum":   a.sum,
			"value": exprValue,
		}
		a.sum = sumExpr.Eval(vars, CallFuncs)
	}
	return nil
}

func (a *SumAggregator) GetResult(
	aggregators []Aggregator,
	goals []*Goal,
	numRecords int64,
) *dlit.Literal {
	return a.sum
}

func (a *SumAggregator) IsEqual(o Aggregator) bool {
	if _, ok := o.(*SumAggregator); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
