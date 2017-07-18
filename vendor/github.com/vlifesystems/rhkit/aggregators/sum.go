// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregators

import (
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

var sumExpr = dexpr.MustNew("sum+value", dexprfuncs.CallFuncs)

func init() {
	Register("sum", &sumAggregator{})
}

func (a *sumAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
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

func (ad *sumSpec) Name() string {
	return ad.name
}

func (ad *sumSpec) Kind() string {
	return "sum"
}

func (ad *sumSpec) Arg() string {
	return ad.expr.String()
}

func (ai *sumInstance) Name() string {
	return ai.spec.name
}

func (ai *sumInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	if isRuleTrue {
		exprValue := ai.spec.expr.Eval(record)
		if err := exprValue.Err(); err != nil {
			return err
		}

		vars := map[string]*dlit.Literal{
			"sum":   ai.sum,
			"value": exprValue,
		}
		ai.sum = sumExpr.Eval(vars)
	}
	return nil
}

func (ai *sumInstance) Result(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	return ai.sum
}
