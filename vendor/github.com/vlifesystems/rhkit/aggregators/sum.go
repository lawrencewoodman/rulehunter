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

package aggregators

import (
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

type sumAggregator struct{}

type sumSpec struct {
	name string
	expr *dexpr.Expr
}

type sumInstance struct {
	spec *sumSpec
	sum  *dlit.Literal
}

var sumExpr = dexpr.MustNew("sum+value")

func init() {
	Register("sum", &sumAggregator{})
}

func (a *sumAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	d := &sumSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *sumSpec) New() AggregatorInstance {
	return &sumInstance{
		spec: ad,
		sum:  dlit.MustNew(0),
	}
}

func (ad *sumSpec) GetName() string {
	return ad.name
}

func (ad *sumSpec) GetKind() string {
	return "sum"
}

func (ad *sumSpec) GetArg() string {
	return ad.expr.String()
}

func (ai *sumInstance) GetName() string {
	return ai.spec.name
}

func (ai *sumInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	if isRuleTrue {
		exprValue := ai.spec.expr.Eval(record, dexprfuncs.CallFuncs)
		_, valueIsFloat := exprValue.Float()
		if !valueIsFloat {
			return fmt.Errorf("sum aggregator: value isn't a float: %s", exprValue)
		}

		vars := map[string]*dlit.Literal{
			"sum":   ai.sum,
			"value": exprValue,
		}
		ai.sum = sumExpr.Eval(vars, dexprfuncs.CallFuncs)
	}
	return nil
}

func (ai *sumInstance) GetResult(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	return ai.sum
}
