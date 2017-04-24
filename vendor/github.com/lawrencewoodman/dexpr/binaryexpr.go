/*
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"math"
)

var trueLiteral = dlit.MustNew(true)
var falseLiteral = dlit.MustNew(false)

func binaryExprToenode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	be *ast.BinaryExpr,
) enode {
	lh := nodeToenode(callFuncs, eltStore, be.X)
	rh := nodeToenode(callFuncs, eltStore, be.Y)
	if _, ok := lh.(enErr); ok {
		return lh
	} else if _, ok := rh.(enErr); ok {
		return rh
	}

	switch be.Op {
	case token.LSS:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opLss, lh, rh, vars)
			},
		}
	case token.LEQ:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opLeq, lh, rh, vars)
			},
		}
	case token.EQL:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opEql, lh, rh, vars)
			},
		}
	case token.NEQ:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opNeq, lh, rh, vars)
			},
		}
	case token.GTR:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opGtr, lh, rh, vars)
			},
		}
	case token.GEQ:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opGeq, lh, rh, vars)
			},
		}
	case token.LAND:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opLand, lh, rh, vars)
			},
		}
	case token.LOR:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opLor, lh, rh, vars)
			},
		}
	case token.ADD:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opAdd, lh, rh, vars)
			},
		}
	case token.SUB:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opSub, lh, rh, vars)
			},
		}
	case token.MUL:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opMul, lh, rh, vars)
			},
		}
	case token.QUO:
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				return callBinaryFn(opQuo, lh, rh, vars)
			},
		}
	}
	return enErr{err: InvalidOpError(be.Op)}
}

func callBinaryFn(
	fn binaryFn,
	lh enode,
	rh enode,
	vars map[string]*dlit.Literal,
) *dlit.Literal {
	lhV := lh.Eval(vars)
	rhV := rh.Eval(vars)
	if lhV.Err() != nil {
		return lhV
	}
	if rhV.Err() != nil {
		return rhV
	}
	return fn(lhV, rhV)
}

type binaryFn func(*dlit.Literal, *dlit.Literal) *dlit.Literal

func opLss(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		if rhInt, rhIsInt := rh.Int(); rhIsInt {
			if lhInt < rhInt {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		if rhFloat, rhIsFloat := rh.Float(); rhIsFloat {
			if lhFloat < rhFloat {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opLeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		if rhInt, rhIsInt := rh.Int(); rhIsInt {
			if lhInt <= rhInt {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		if rhFloat, rhIsFloat := rh.Float(); rhIsFloat {
			if lhFloat <= rhFloat {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opGtr(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		if rhInt, rhIsInt := rh.Int(); rhIsInt {
			if lhInt > rhInt {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		if rhFloat, rhIsFloat := rh.Float(); rhIsFloat {
			if lhFloat > rhFloat {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opGeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		if rhInt, rhIsInt := rh.Int(); rhIsInt {
			if lhInt >= rhInt {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		if rhFloat, rhIsFloat := rh.Float(); rhIsFloat {
			if lhFloat >= rhFloat {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opEql(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		if rhInt, rhIsInt := rh.Int(); rhIsInt {
			if lhInt == rhInt {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		if rhFloat, rhIsFloat := rh.Float(); rhIsFloat {
			if lhFloat == rhFloat {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	if lhErr := lh.Err(); lhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	if rhErr := rh.Err(); rhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	if lh.String() == rh.String() {
		return trueLiteral
	}
	return falseLiteral
}

func opNeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	if lhIsInt {
		if rhInt, rhIsInt := rh.Int(); rhIsInt {
			if lhInt != rhInt {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat {
		if rhFloat, rhIsFloat := rh.Float(); rhIsFloat {
			if lhFloat != rhFloat {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools

	if lhErr := lh.Err(); lhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	if rhErr := rh.Err(); rhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	if lh.String() != rh.String() {
		return trueLiteral
	}
	return falseLiteral
}

func opLand(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhBool, lhIsBool := lh.Bool()
	if lhIsBool {
		if rhBool, rhIsBool := rh.Bool(); rhIsBool {
			if lhBool && rhBool {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opLor(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhBool, lhIsBool := lh.Bool()
	if lhIsBool {
		if rhBool, rhIsBool := rh.Bool(); rhIsBool {
			if lhBool || rhBool {
				return trueLiteral
			} else {
				return falseLiteral
			}
		}
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opAdd(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt + rhInt
		if (r < lhInt) == (rhInt < 0) {
			return dlit.MustNew(r)
		}
		// If overflow then use Float
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		r := lhFloat + rhFloat
		if !math.IsInf(r, 0) {
			return dlit.MustNew(r)
		}
		return dlit.MustNew(ErrUnderflowOverflow)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opSub(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt - rhInt
		if (r > lhInt) == (rhInt < 0) {
			return dlit.MustNew(r)
		}
		// If overflow then use Float
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		r := lhFloat - rhFloat
		if !math.IsInf(r, 0) {
			return dlit.MustNew(r)
		}
		return dlit.MustNew(ErrUnderflowOverflow)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opMul(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		// Overflow detection inspired by suggestion from Rob Pike on Go-nuts group:
		//   https://groups.google.com/d/msg/Golang-nuts/h5oSN5t3Au4/KaNQREhZh0QJ
		if lhInt == 0 || rhInt == 0 || lhInt == 1 || rhInt == 1 {
			return dlit.MustNew(lhInt * rhInt)
		}
		if lhInt != math.MinInt64 && rhInt != math.MinInt64 {
			r := lhInt * rhInt
			if r/rhInt == lhInt {
				return dlit.MustNew(r)
			}
		}
		// If overflow then use Float
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		r := lhFloat * rhFloat
		if !math.IsInf(r, 0) {
			return dlit.MustNew(r)
		}
		return dlit.MustNew(ErrUnderflowOverflow)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opQuo(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()

	if rhIsInt && rhInt == 0 {
		return dlit.MustNew(ErrDivByZero)
	}
	if lhIsInt && rhIsInt && lhInt%rhInt == 0 {
		return dlit.MustNew(lhInt / rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		r := lhFloat / rhFloat
		if !math.IsInf(r, 0) {
			return dlit.MustNew(r)
		}
		return dlit.MustNew(ErrUnderflowOverflow)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}
