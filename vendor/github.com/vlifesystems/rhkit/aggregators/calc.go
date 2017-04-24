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

type calcAggregator struct{}

type calcSpec struct {
	name string
	expr *dexpr.Expr
}

type calcInstance struct {
	spec *calcSpec
}

func init() {
	Register("calc", &calcAggregator{})
}

func (a *calcAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, err
	}
	d := &calcSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *calcSpec) New() AggregatorInstance {
	return &calcInstance{spec: ad}
}

func (ad *calcSpec) GetName() string {
	return ad.name
}

func (ad *calcSpec) GetKind() string {
	return "calc"
}

func (ad *calcSpec) GetArg() string {
	return ad.expr.String()
}

func (ai *calcInstance) GetName() string {
	return ai.spec.name
}

func (ai *calcInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	return nil
}

func (ai *calcInstance) GetResult(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	instancesMap, err :=
		InstancesToMap(aggregatorInstances, goals, numRecords, ai.GetName())
	if err != nil {
		return dlit.MustNew(err)
	}
	return ai.spec.expr.Eval(instancesMap)
}
