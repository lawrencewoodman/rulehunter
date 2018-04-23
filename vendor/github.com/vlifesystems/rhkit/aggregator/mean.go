// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package aggregator

import (
	"strings"

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
	maxDP      int
}

var meanExpr = dexpr.MustNew("sum/n", dexprfuncs.CallFuncs)
var meanSumExpr = dexpr.MustNew("sum+value", dexprfuncs.CallFuncs)

func init() {
	Register("mean", &meanAggregator{})
}

func (a *meanAggregator) MakeSpec(
	name string,
	expr string,
) (Spec, error) {
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

func (ad *meanSpec) New() Instance {
	return &meanInstance{
		spec:       ad,
		sum:        dlit.MustNew(0),
		numRecords: 0,
		maxDP:      0,
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

		if dp := numDecPlaces(exprValue); dp > ai.maxDP {
			ai.maxDP = dp
		}

		vars := map[string]*dlit.Literal{
			"sum":   ai.sum,
			"value": exprValue,
		}
		ai.sum = meanSumExpr.Eval(vars)
		if err := ai.sum.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (ai *meanInstance) Result(
	aggregatorInstances []Instance,
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
	return roundTo(meanExpr.Eval(vars), ai.maxDP+2)
}

func numDecPlaces(l *dlit.Literal) int {
	if _, isInt := l.Int(); isInt {
		return 0
	}
	s := l.String()
	i := strings.IndexByte(s, '.')
	if i > -1 {
		s = strings.TrimRight(s, "0")
		return len(s) - i - 1
	}
	return 0
}
