/*
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
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

type countAggregator struct{}

type countSpec struct {
	name string
	expr *dexpr.Expr
}

type countInstance struct {
	spec       *countSpec
	numMatches int64
}

func init() {
	Register("count", &countAggregator{})
}

func (a *countAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, err
	}
	d := &countSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *countSpec) New() AggregatorInstance {
	return &countInstance{
		spec:       ad,
		numMatches: 0,
	}
}

func (ad *countSpec) Name() string {
	return ad.name
}

func (ad *countSpec) Kind() string {
	return "count"
}

func (ad *countSpec) Arg() string {
	return ad.expr.String()
}

func (ai *countInstance) Name() string {
	return ai.spec.name
}

func (ai *countInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	countExprIsTrue, err := ai.spec.expr.EvalBool(record)
	if err != nil {
		return err
	}
	if isRuleTrue && countExprIsTrue {
		ai.numMatches++
	}
	return nil
}

func (ai *countInstance) Result(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	return dlit.MustNew(ai.numMatches)
}
