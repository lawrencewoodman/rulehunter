// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregator

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

var recallExpr = dexpr.MustNew(
	"roundto(numTP/(numTP+numFN),4)",
	dexprfuncs.CallFuncs,
)

func init() {
	Register("recall", &recallAggregator{})
}

func (a *recallAggregator) MakeSpec(
	name string,
	expr string,
) (Spec, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, err
	}
	d := &recallSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *recallSpec) New() Instance {
	return &recallInstance{
		spec:  ad,
		numTP: 0,
		numFN: 0,
	}
}

func (ad *recallSpec) Name() string {
	return ad.name
}

func (ad *recallSpec) Kind() string {
	return "recall"
}

func (ad *recallSpec) Arg() string {
	return ad.expr.String()
}

func (ai *recallInstance) Name() string {
	return ai.spec.name
}

func (ai *recallInstance) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	matchExprIsTrue, err := ai.spec.expr.EvalBool(record)
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

func (ai *recallInstance) Result(
	aggregatorInstances []Instance,
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
	return roundTo(recallExpr.Eval(vars), 4)
}
