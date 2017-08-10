// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregator

import (
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/goal"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

type precisionAggregator struct{}

type precisionSpec struct {
	name string
	expr *dexpr.Expr
}

type precisionInstance struct {
	spec  *precisionSpec
	numTP int64
	numFP int64
}

var precisionExpr = dexpr.MustNew(
	"roundto(numTP/(numTP+numFP),4)",
	dexprfuncs.CallFuncs,
)

func init() {
	Register("precision", &precisionAggregator{})
}

func (a *precisionAggregator) MakeSpec(
	name string,
	expr string,
) (Spec, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, err
	}
	d := &precisionSpec{
		name: name,
		expr: dexpr,
	}
	return d, nil
}

func (ad *precisionSpec) New() Instance {
	return &precisionInstance{
		spec:  ad,
		numTP: 0,
		numFP: 0,
	}
}

func (ad *precisionSpec) Name() string {
	return ad.name
}

func (ad *precisionSpec) Kind() string {
	return "precision"
}

func (ad *precisionSpec) Arg() string {
	return ad.expr.String()
}

func (ai *precisionInstance) Name() string {
	return ai.spec.name
}

func (ai *precisionInstance) NextRecord(record map[string]*dlit.Literal,
	isRuleTrue bool) error {
	matchExprIsTrue, err := ai.spec.expr.EvalBool(record)
	if err != nil {
		return err
	}
	if isRuleTrue {
		if matchExprIsTrue {
			ai.numTP++
		} else {
			ai.numFP++
		}
	}
	return nil
}

func (ai *precisionInstance) Result(
	aggregatorInstances []Instance,
	goals []*goal.Goal,
	numRecords int64,
) *dlit.Literal {
	if ai.numTP == 0 && ai.numFP == 0 {
		return dlit.MustNew(0)
	}

	vars := map[string]*dlit.Literal{
		"numTP": dlit.MustNew(ai.numTP),
		"numFP": dlit.MustNew(ai.numFP),
	}
	return precisionExpr.Eval(vars)
}
