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

func binaryExprToLiteral(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	valStore *valStore,
	eltStore *eltStore,
	be *ast.BinaryExpr,
) *dlit.Literal {
	lh := nodeToLiteral(vars, callFuncs, valStore, eltStore, be.X)
	rh := nodeToLiteral(vars, callFuncs, valStore, eltStore, be.Y)
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
	return dlit.MustNew(InvalidOpError(be.Op))
}

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
		return dlit.MustNew(lhFloat + rhFloat)
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
