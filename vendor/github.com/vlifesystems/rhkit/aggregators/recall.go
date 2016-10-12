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
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

type recallAggregator struct{}

type recallSpec struct {
	name string
	expr *dexpr.Expr
}

type recallInstance struct {
	spec  *recallSpec
	numTP int64
	numFN int64
}

var recallExpr = dexpr.MustNew("roundto(numTP/(numTP+numFN),4)")

func init() {
	Register("recall", &recallAggregator{})
}

func (a *recallAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	d := &recallSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *recallSpec) New() AggregatorInstance {
	return &recallInstance{
		spec:  ad,
		numTP: 0,
		numFN: 0,
	}
}

func (ad *recallSpec) GetName() string {
	return ad.name
}

func (ad *recallSpec) GetArg() string {
	return ad.expr.String()
}

func (ai *recallInstance) GetName() string {
	return ai.spec.name
}

func (ai *recallInstance) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	matchExprIsTrue, err := ai.spec.expr.EvalBool(record, dexprfuncs.CallFuncs)
	if err != nil {
		return err
	}
	if isRuleTrue {
		if matchExprIsTrue {
			ai.numTP++
		}
	} else {
		if matchExprIsTrue {
			ai.numFN++
		}
	}
	return nil
}

func (ai *recallInstance) GetResult(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	if ai.numTP == 0 && ai.numFN == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"numTP": dlit.MustNew(ai.numTP),
		"numFN": dlit.MustNew(ai.numFN),
	}
	return recallExpr.Eval(vars, dexprfuncs.CallFuncs)
}
