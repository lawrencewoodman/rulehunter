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

type accuracyAggregator struct{}

type accuracySpec struct {
	name string
	expr *dexpr.Expr
}

type accuracyInstance struct {
	spec       *accuracySpec
	numMatches int64
}

var accuracyExpr = dexpr.MustNew("roundto(100*numMatches/numRecords,2)")

func init() {
	Register("accuracy", &accuracyAggregator{})
}

func (a *accuracyAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr)
	if err != nil {
		return nil, err
	}
	d := &accuracySpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *accuracySpec) New() AggregatorInstance {
	return &accuracyInstance{
		spec:       ad,
		numMatches: 0,
	}
}

func (ad *accuracySpec) GetName() string {
	return ad.name
}

func (ad *accuracySpec) GetArg() string {
	return ad.expr.String()
}

func (ai *accuracyInstance) GetName() string {
	return ai.spec.name
}

func (ai *accuracyInstance) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	matchExprIsTrue, err := ai.spec.expr.EvalBool(record, dexprfuncs.CallFuncs)
	if err != nil {
		return err
	}
	if isRuleTrue == matchExprIsTrue {
		ai.numMatches++
	}
	return nil
}

func (ai *accuracyInstance) GetResult(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	if numRecords == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"numRecords": dlit.MustNew(numRecords),
		"numMatches": dlit.MustNew(ai.numMatches),
	}
	return accuracyExpr.Eval(vars, dexprfuncs.CallFuncs)
}
