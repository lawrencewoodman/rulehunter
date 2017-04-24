/*
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"math"
	"strconv"
)

func unaryExprToenode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	ue *ast.UnaryExpr,
) enode {
	rh := nodeToenode(callFuncs, eltStore, ue.X)
	if _, ok := rh.(enErr); ok {
		return rh
	}
	switch ue.Op {
	case token.NOT:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callUnaryFn(opNot, rh, vars)
			},
		}
	case token.SUB:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callUnaryFn(opNeg, rh, vars)
			},
		}
	}
	return enErr{err: InvalidOpError(ue.Op)}
}

func callUnaryFn(
	fn unaryFn,
	l enode,
	vars map[string]*dlit.Literal,
) *dlit.Literal {
	lV := l.Eval(vars)
	if lV.Err() != nil {
		return lV
	}
	return fn(lV)
}

type unaryFn func(*dlit.Literal) *dlit.Literal

func opNot(l *dlit.Literal) *dlit.Literal {
	lBool, lIsBool := l.Bool()
	if !lIsBool {
		return dlit.MustNew(ErrIncompatibleTypes)
	}
	if lBool {
		return falseLiteral
	}
	return trueLiteral
}

func opNeg(l *dlit.Literal) *dlit.Literal {
	lInt, lIsInt := l.Int()
	if lIsInt {
		return dlit.MustNew(0 - lInt)
	}

	strMinInt64 := strconv.FormatInt(int64(math.MinInt64), 10)
	posMinInt64 := strMinInt64[1:]
	if l.String() == posMinInt64 {
		return dlit.MustNew(int64(math.MinInt64))
	}

	lFloat, lIsFloat := l.Float()
	if lIsFloat {
		return dlit.MustNew(0 - lFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}
