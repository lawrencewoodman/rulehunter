// Copyright (C) 2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import (
	"github.com/lawrencewoodman/ddataset"
	"github.com/lawrencewoodman/dexpr"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

// Dynamic represents a rule determining if supplied dynamic expression is
// true for a record
type Dynamic struct {
	dexpr *dexpr.Expr
}

func NewDynamic(expr string) (Rule, error) {
	dexpr, err := dexpr.New(expr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, InvalidExprError{Expr: expr}
	}
	return &Dynamic{dexpr: dexpr}, nil
}

func MakeDynamicRules(exprs []string) ([]Rule, error) {
	var err error
	r := make([]Rule, len(exprs))
	for i, expr := range exprs {
		r[i], err = NewDynamic(expr)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func (r *Dynamic) String() string {
	return r.dexpr.String()
}

func (r *Dynamic) IsTrue(record ddataset.Record) (bool, error) {
	isTrue, err := r.dexpr.EvalBool(record)
	if err == nil {
		return isTrue, nil
	}
	if x, ok := err.(dexpr.InvalidExprError); ok {
		if x.Err == dexpr.ErrIncompatibleTypes {
			return false, IncompatibleTypesRuleError{Rule: r}
		}
	}
	return false, InvalidRuleError{Rule: r}
}

func (r *Dynamic) Fields() []string {
	return []string{}
}
