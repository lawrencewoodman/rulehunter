// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregator

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
) (Spec, error) {
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

func (ad *calcSpec) New() Instance {
	return &calcInstance{spec: ad}
}

func (ad *calcSpec) Name() string {
	return ad.name
}

func (ad *calcSpec) Kind() string {
	return "calc"
}

func (ad *calcSpec) Arg() string {
	return ad.expr.String()
}

func (ai *calcInstance) Name() string {
	return ai.spec.name
}

func (ai *calcInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	return nil
}

func (ai *calcInstance) Result(
	aggregatorInstances []Instance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	instancesMap, err :=
		InstancesToMap(aggregatorInstances, goals, numRecords, ai.Name())
	if err != nil {
		return dlit.MustNew(err)
	}
	return ai.spec.expr.Eval(instancesMap)
}
