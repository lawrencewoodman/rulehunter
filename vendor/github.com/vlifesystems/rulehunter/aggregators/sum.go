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
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rulehunter/goal"
	"github.com/vlifesystems/rulehunter/internal/dexprfuncs"
)

type sum struct {
	name string
	sum  *dlit.Literal
	expr *dexpr.Expr
}

var sumExpr = dexpr.MustNew("sum+value")

func newSum(name string, exprStr string) (*sum, error) {
	expr, err := dexpr.New(exprStr)
	if err != nil {
		return nil, err
	}
	ca := &sum{
		name: name,
		sum:  dlit.MustNew(0),
		expr: expr,
	}
	return ca, nil
}

func (a *sum) CloneNew() Aggregator {
	return &sum{
		name: a.name,
		sum:  dlit.MustNew(0),
		expr: a.expr,
	}
}

func (a *sum) GetName() string {
	return a.name
}

func (a *sum) GetArg() string {
	return a.expr.String()
}

func (a *sum) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	if isRuleTrue {
		exprValue := a.expr.Eval(record, dexprfuncs.CallFuncs)
		_, valueIsFloat := exprValue.Float()
		if !valueIsFloat {
			return fmt.Errorf("Value isn't a float: %s", exprValue)
		}

		vars := map[string]*dlit.Literal{
			"sum":   a.sum,
			"value": exprValue,
		}
		a.sum = sumExpr.Eval(vars, dexprfuncs.CallFuncs)
	}
	return nil
}

func (a *sum) GetResult(
	aggregators []Aggregator,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	return a.sum
}

func (a *sum) IsEqual(o Aggregator) bool {
	if _, ok := o.(*sum); !ok {
		return false
	}
	return a.name == o.GetName() && a.GetArg() == o.GetArg()
}
