/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
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

func binaryExprToLiteral(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	be *ast.BinaryExpr,
) *dlit.Literal {
	lh := nodeToLiteral(vars, callFuncs, eltStore, be.X)
	rh := nodeToLiteral(vars, callFuncs, eltStore, be.Y)
	if lh.Err() != nil {
		return lh
	} else if rh.Err() != nil {
		return rh
	}

	switch be.Op {
	case token.LSS:
		return opLss(lh, rh)
	case token.LEQ:
		return opLeq(lh, rh)
	case token.EQL:
		return opEql(lh, rh)
	case token.NEQ:
		return opNeq(lh, rh)
	case token.GTR:
		return opGtr(lh, rh)
	case token.GEQ:
		return opGeq(lh, rh)
	case token.LAND:
		return opLand(lh, rh)
	case token.LOR:
		return opLor(lh, rh)
	case token.ADD:
		return opAdd(lh, rh)
	case token.SUB:
		return opSub(lh, rh)
	case token.MUL:
		return opMul(lh, rh)
	case token.QUO:
		return opQuo(lh, rh)
	}
	return dlit.MustNew(ErrInvalidOp(be.Op))
}

func opLss(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt < rhInt)
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat < rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opLeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt <= rhInt)
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat <= rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opGtr(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt > rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat > rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opGeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt >= rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat >= rhFloat)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opEql(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt == rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat == rhFloat)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	lhErr := lh.Err()
	rhErr := rh.Err()
	if lhErr != nil || rhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	return dlit.MustNew(lh.String() == rh.String())
}

func opNeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		return dlit.MustNew(lhInt != rhInt)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		dlit.MustNew(lhFloat != rhFloat)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	lhErr := lh.Err()
	rhErr := rh.Err()
	if lhErr != nil || rhErr != nil {
		return dlit.MustNew(ErrIncompatibleTypes)
	}

	return dlit.MustNew(lh.String() != rh.String())
}

func opLand(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		return dlit.MustNew(lhBool && rhBool)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opLor(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		return dlit.MustNew(lhBool || rhBool)
	}

	return dlit.MustNew(ErrIncompatibleTypes)
}

func opAdd(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt + rhInt
		if (r < lhInt) != (rhInt < 0) {
			return dlit.MustNew(ErrOverflow)
		}
		return dlit.MustNew(r)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat + rhFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}

func opSub(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt - rhInt
		if (r > lhInt) != (rhInt < 0) {
			return dlit.MustNew(ErrOverflow)
		}
		return dlit.MustNew(r)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat - rhFloat)
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
		if lhInt == math.MinInt64 || rhInt == math.MinInt64 {
			return dlit.MustNew(ErrOverflow)
		}
		r := lhInt * rhInt
		if r/rhInt != lhInt {
			return dlit.MustNew(ErrOverflow)
		}
		return dlit.MustNew(r)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		return dlit.MustNew(lhFloat * rhFloat)
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
		return dlit.MustNew(lhFloat / rhFloat)
	}
	return dlit.MustNew(ErrIncompatibleTypes)
}
