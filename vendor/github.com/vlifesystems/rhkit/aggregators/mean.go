// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregators

import (
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

type meanAggregator struct{}

type meanSpec struct {
	name string
	expr *dexpr.Expr
}

type meanInstance struct {
	spec       *meanSpec
	sum        *dlit.Literal
	numRecords int64
}

var meanExpr = dexpr.MustNew("sum/n", dexprfuncs.CallFuncs)
var meanSumExpr = dexpr.MustNew("sum+value", dexprfuncs.CallFuncs)

func init() {
	Register("mean", &meanAggregator{})
}

func (a *meanAggregator) MakeSpec(
	name string,
	expr string,
) (AggregatorSpec, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, err
	}
	d := &meanSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *meanSpec) New() AggregatorInstance {
	return &meanInstance{
		spec:       ad,
		sum:        dlit.MustNew(0),
		numRecords: 0,
	}
}

func (ad *meanSpec) Name() string {
	return ad.name
}

func (ad *meanSpec) Kind() string {
	return "mean"
}

func (ad *meanSpec) Arg() string {
	return ad.expr.String()
}

func (ai *meanInstance) Name() string {
	return ai.spec.name
}

func (ai *meanInstance) NextRecord(
	record map[string]*dlit.Literal,
	isRuleTrue bool,
) error {
	if isRuleTrue {
		ai.numRecords++
		exprValue := ai.spec.expr.Eval(record)
		if err := exprValue.Err(); err != nil {
			return err
		}

		vars := map[string]*dlit.Literal{
			"sum":   ai.sum,
			"value": exprValue,
		}
		ai.sum = meanSumExpr.Eval(vars)
	}
	return nil
}

func (ai *meanInstance) Result(
	aggregatorInstances []AggregatorInstance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	if ai.numRecords == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"sum": ai.sum,
		"n":   dlit.MustNew(ai.numRecords),
	}
	return meanExpr.Eval(vars)
}
