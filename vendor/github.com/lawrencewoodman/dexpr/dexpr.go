/*
 * A package for evaluating dynamic expressions
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */
package dexpr

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"math"
	"strconv"
)

type Expr struct {
	Expr string
	Node ast.Node
}

type ErrInvalidExpr string

func (e ErrInvalidExpr) Error() string {
	return string(e)
}

type ErrInvalidOp token.Token

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("Invalid operator: %q", token.Token(e))
}

type CallFun func([]*dlit.Literal) (*dlit.Literal, error)

func New(expr string) (*Expr, error) {
	node, err := parseExpr(expr)
	if err != nil {
		return &Expr{}, ErrInvalidExpr(fmt.Sprintf("Invalid expression: %s", expr))
	}
	return &Expr{Expr: expr, Node: node}, nil
}

func MustNew(expr string) *Expr {
	e, err := New(expr)
	if err != nil {
		panic(err.Error())
	}
	return e
}

func (expr *Expr) Eval(
	vars map[string]*dlit.Literal, callFuncs map[string]CallFun) *dlit.Literal {
	var l *dlit.Literal
	inspector := func(n ast.Node) bool {
		l = nodeToLiteral(vars, callFuncs, n)
		return false
	}
	ast.Inspect(expr.Node, inspector)
	return l
}

func (expr *Expr) EvalBool(
	vars map[string]*dlit.Literal, callFuncs map[string]CallFun) (bool, error) {
	l := expr.Eval(vars, callFuncs)
	if b, isBool := l.Bool(); isBool {
		return b, nil
	} else if err, isErr := l.Err(); isErr {
		return false, ErrInvalidExpr(err.Error())
	} else {
		return false, ErrInvalidExpr("Expression doesn't return a bool")
	}
}

func (expr *Expr) String() string {
	return expr.Expr
}

func nodeToLiteral(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	n ast.Node) *dlit.Literal {
	var l *dlit.Literal
	var exists bool

	switch x := n.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.INT:
			l, _ = dlit.New(x.Value)
		case token.FLOAT:
			l, _ = dlit.New(x.Value)
		case token.STRING:
			uc, err := strconv.Unquote(x.Value)
			if err != nil {
				l = makeErrInvalidExprLiteral(err.Error())
			} else {
				l, _ = dlit.New(uc)
			}
		}
	case *ast.Ident:
		if l, exists = vars[x.Name]; !exists {
			l = makeErrInvalidExprLiteral(
				fmt.Sprintf("Variable doesn't exist: %s", x.Name))
		}
	case *ast.ParenExpr:
		l = nodeToLiteral(vars, callFuncs, x.X)
	case *ast.BinaryExpr:
		lh := nodeToLiteral(vars, callFuncs, x.X)
		rh := nodeToLiteral(vars, callFuncs, x.Y)
		if _, isErr := lh.Err(); isErr {
			l = lh
		} else if _, isErr := rh.Err(); isErr {
			l = rh
		} else {
			l = evalBinaryExpr(lh, rh, x.Op)
		}
	case *ast.UnaryExpr:
		rh := nodeToLiteral(vars, callFuncs, x.X)
		if _, isErr := rh.Err(); isErr {
			l = rh
		} else {
			l = evalUnaryExpr(rh, x.Op)
		}
	case *ast.CallExpr:
		args := callArgsToDLiterals(vars, callFuncs, x.Args)
		l = callFun(callFuncs, x.Fun, args)
	default:
		fmt.Printf("UNRECOGNIZED TYPE - x: %q", x)
	}
	return l
}

func callArgsToDLiterals(
	vars map[string]*dlit.Literal,
	callFuncs map[string]CallFun,
	callArgs []ast.Expr) []*dlit.Literal {
	argLits := make([]*dlit.Literal, len(callArgs))
	for i, arg := range callArgs {
		argLits[i] = nodeToLiteral(vars, callFuncs, arg)
	}
	return argLits
}

func callFun(
	callFuncs map[string]CallFun,
	name ast.Expr,
	args []*dlit.Literal) *dlit.Literal {
	// TODO: Find more direct way of getting name as a string
	nameString := fmt.Sprintf("%s", name)
	f, exists := callFuncs[nameString]
	if !exists {
		return makeErrInvalidExprLiteral(
			fmt.Sprintf("Function doesn't exist: %s", name))
	}
	l, err := f(args)
	if err != nil {
		return makeErrInvalidExprLiteral(err.Error())
	}
	return l
}

func evalBinaryExpr(lh *dlit.Literal, rh *dlit.Literal,
	op token.Token) *dlit.Literal {
	var r *dlit.Literal

	switch op {
	case token.LSS:
		r = opLss(lh, rh)
	case token.LEQ:
		r = opLeq(lh, rh)
	case token.EQL:
		r = opEql(lh, rh)
	case token.NEQ:
		r = opNeq(lh, rh)
	case token.GTR:
		r = opGtr(lh, rh)
	case token.GEQ:
		r = opGeq(lh, rh)
	case token.LAND:
		r = opLand(lh, rh)
	case token.LOR:
		r = opLor(lh, rh)
	case token.ADD:
		r = opAdd(lh, rh)
	case token.SUB:
		r = opSub(lh, rh)
	case token.MUL:
		r = opMul(lh, rh)
	case token.QUO:
		r = opQuo(lh, rh)
	default:
		r, _ = dlit.New(ErrInvalidOp(op))
	}

	return r
}

func evalUnaryExpr(rh *dlit.Literal, op token.Token) *dlit.Literal {
	var r *dlit.Literal
	switch op {
	case token.SUB:
		r = opNeg(rh)
	default:
		r, _ = dlit.New(ErrInvalidOp(op))
	}
	return r
}

func literalToQuotedString(l *dlit.Literal) string {
	_, isInt := l.Int()
	if isInt {
		return l.String()
	}
	_, isFloat := l.Float()
	if isFloat {
		return l.String()
	}
	_, isBool := l.Bool()
	if isBool {
		return l.String()
	}
	return fmt.Sprintf("\"%s\"", l.String())
}

func makeErrInvalidExprLiteral(msg string) *dlit.Literal {
	var l *dlit.Literal
	var err error
	err = ErrInvalidExpr(msg)
	l, _ = dlit.New(err)
	return l
}

func makeErrInvalidExprLiteralFmt(errFormattedMsg string, l1 *dlit.Literal,
	l2 *dlit.Literal) *dlit.Literal {
	l1s := literalToQuotedString(l1)
	l2s := literalToQuotedString(l2)
	return makeErrInvalidExprLiteral(fmt.Sprintf(errFormattedMsg, l1s, l2s))
}

func checkNewLitError(l *dlit.Literal, err error, errFormattedMsg string,
	l1 *dlit.Literal, l2 *dlit.Literal) *dlit.Literal {
	if err != nil {
		return makeErrInvalidExprLiteralFmt(errFormattedMsg, l1, l2)
	}
	return l
}

func opLss(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s < %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt < rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()

	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat < rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opLeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s <= %s"
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt <= rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	rhFloat, rhIsFloat := rh.Float()
	lhFloat, lhIsFloat := lh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat <= rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opGtr(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s > %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt > rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()

	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat > rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opGeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s >= %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt >= rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat >= rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opEql(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s == %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt == rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat == rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	_, lhIsErr := lh.Err()
	_, rhIsErr := rh.Err()
	if lhIsErr || rhIsErr {
		return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
	}

	l, err := dlit.New(lh.String() == rh.String())
	return checkNewLitError(l, err, errMsg, lh, rh)
}

func opNeq(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid comparison: %s != %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		l, err := dlit.New(lhInt != rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat != rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	// Don't compare bools as otherwise with the way that floats or ints
	// are cast to bools you would find that "True" == 1.0 because they would
	// both convert to true bools
	_, lhIsErr := lh.Err()
	_, rhIsErr := rh.Err()
	if lhIsErr || rhIsErr {
		return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
	}

	l, err := dlit.New(lh.String() != rh.String())
	return checkNewLitError(l, err, errMsg, lh, rh)
}

func opLand(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s && %s"

	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		l, err := dlit.New(lhBool && rhBool)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opLor(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s || %s"

	lhBool, lhIsBool := lh.Bool()
	rhBool, rhIsBool := rh.Bool()
	if lhIsBool && rhIsBool {
		l, err := dlit.New(lhBool || rhBool)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opAdd(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s + %s"
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt + rhInt
		if (r < lhInt) != (rhInt < 0) {
			thisErrMsg := fmt.Sprintf("%s (Underflow/Overflow)", errMsg)
			return makeErrInvalidExprLiteralFmt(thisErrMsg, lh, rh)
		}
		l, err := dlit.New(r)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat + rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}
	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opSub(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s - %s"
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		r := lhInt - rhInt
		if (r > lhInt) != (rhInt < 0) {
			thisErrMsg := fmt.Sprintf("%s (Underflow/Overflow)", errMsg)
			return makeErrInvalidExprLiteralFmt(thisErrMsg, lh, rh)
		}
		l, err := dlit.New(r)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat - rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}
	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opMul(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s * %s"
	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()
	if lhIsInt && rhIsInt {
		// Overflow detection inspired by suggestion from Rob Pike on Go-nuts group:
		//   https://groups.google.com/d/msg/Golang-nuts/h5oSN5t3Au4/KaNQREhZh0QJ
		if lhInt == 0 || rhInt == 0 || lhInt == 1 || rhInt == 1 {
			l, err := dlit.New(lhInt * rhInt)
			return checkNewLitError(l, err, errMsg, lh, rh)
		}
		if lhInt == math.MinInt64 || rhInt == math.MinInt64 {
			thisErrMsg := fmt.Sprintf("%s (Underflow/Overflow)", errMsg)
			return makeErrInvalidExprLiteralFmt(thisErrMsg, lh, rh)
		}
		r := lhInt * rhInt
		if r/rhInt != lhInt {
			thisErrMsg := fmt.Sprintf("%s (Underflow/Overflow)", errMsg)
			return makeErrInvalidExprLiteralFmt(thisErrMsg, lh, rh)
		}
		l, err := dlit.New(r)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat * rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}
	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opQuo(lh *dlit.Literal, rh *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: %s / %s"

	lhInt, lhIsInt := lh.Int()
	rhInt, rhIsInt := rh.Int()

	if rhIsInt && rhInt == 0 {
		errMsg := "Invalid operation: %s / %s (Divide by zero)"
		return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
	}
	if lhIsInt && rhIsInt && lhInt%rhInt == 0 {
		l, err := dlit.New(lhInt / rhInt)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}

	lhFloat, lhIsFloat := lh.Float()
	rhFloat, rhIsFloat := rh.Float()
	if lhIsFloat && rhIsFloat {
		l, err := dlit.New(lhFloat / rhFloat)
		return checkNewLitError(l, err, errMsg, lh, rh)
	}
	return makeErrInvalidExprLiteralFmt(errMsg, lh, rh)
}

func opNeg(l *dlit.Literal) *dlit.Literal {
	errMsg := "Invalid operation: -%s"

	lInt, lIsInt := l.Int()
	if lIsInt {
		r, err := dlit.New(0 - lInt)
		return checkNewLitError(r, err, errMsg, l, l)
	}

	strMinInt64 := strconv.FormatInt(int64(math.MinInt64), 10)
	posMinInt64 := strMinInt64[1:]
	if l.String() == posMinInt64 {
		r, err := dlit.New(int64(math.MinInt64))
		return checkNewLitError(r, err, errMsg, l, l)
	}

	lFloat, lIsFloat := l.Float()
	if lIsFloat {
		r, err := dlit.New(0 - lFloat)
		return checkNewLitError(r, err, errMsg, l, l)
	}
	return makeErrInvalidExprLiteralFmt(errMsg, l, l)
}
