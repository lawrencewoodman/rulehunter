/*
 * A package for evaluating dynamic expressions
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"fmt"
	"github.com/lawrencewoodman/dlit"
	"go/ast"
	"go/token"
	"strconv"
)

type Expr struct {
	Expr string
	Node enode
}

type CallFun func([]*dlit.Literal) (*dlit.Literal, error)

func New(expr string, callFuncs map[string]CallFun) (*Expr, error) {
	node, err := parseExpr(expr)
	if err != nil {
		return &Expr{}, InvalidExprError{expr, ErrSyntax}
	}

	en := compile(node, callFuncs)
	if ee, ok := en.(enErr); ok {
		return &Expr{}, InvalidExprError{expr, ee.Err()}
	}
	return &Expr{Expr: expr, Node: en}, nil
}

func MustNew(expr string, callFuncs map[string]CallFun) *Expr {
	e, err := New(expr, callFuncs)
	if err != nil {
		panic(err.Error())
	}
	return e
}

func (expr *Expr) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	l := expr.Node.Eval(vars)
	if err := l.Err(); err != nil {
		return dlit.MustNew(InvalidExprError{expr.Expr, err})
	}
	return l
}

func (expr *Expr) EvalBool(vars map[string]*dlit.Literal) (bool, error) {
	l := expr.Eval(vars)
	if b, isBool := l.Bool(); isBool {
		return b, nil
	} else if err := l.Err(); err != nil {
		return false, err
	}
	return false, InvalidExprError{expr.Expr, ErrIncompatibleTypes}
}

func (expr *Expr) String() string {
	return expr.Expr
}

// kinds are the kinds of composite type
var kinds = map[string]*dlit.Literal{
	"lit": dlit.NewString("lit"),
}

func compile(node ast.Node, callFuncs map[string]CallFun) enode {
	var en enode
	inspector := func(n ast.Node) bool {
		eltStore := newEltStore()
		en = nodeToenode(callFuncs, eltStore, n)
		return false
	}
	ast.Inspect(node, inspector)
	if _, ok := en.(enErr); ok {
		return en
	}
	return en
}

func nodeToenode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	n ast.Node,
) enode {
	switch x := n.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.INT:
			fallthrough
		case token.FLOAT:
			return enLit{val: dlit.NewString(x.Value)}
		case token.CHAR:
			fallthrough
		case token.STRING:
			uc, err := strconv.Unquote(x.Value)
			if err != nil {
				return enErr{err: ErrSyntax}
			}
			return enLit{val: dlit.NewString(uc)}
		}
	case *ast.Ident:
		return enVar(x.Name)
	case *ast.ParenExpr:
		return nodeToenode(callFuncs, eltStore, x.X)
	case *ast.BinaryExpr:
		return binaryExprToenode(callFuncs, eltStore, x)
	case *ast.UnaryExpr:
		return unaryExprToenode(callFuncs, eltStore, x)
	case *ast.CallExpr:
		args := exprSliceToenodes(callFuncs, eltStore, x.Args)
		return enFunc{
			fn: func(vars map[string]*dlit.Literal) *dlit.Literal {
				lits := eNodesToDLiterals(vars, args)
				return callFun(callFuncs, x.Fun, lits)
			},
		}
	case *ast.CompositeLit:
		kindNode := nodeToenode(callFuncs, eltStore, x.Type)
		kind := kindNode.Eval(kinds)
		if kind.String() != "lit" {
			return enErr{err: ErrInvalidCompositeType}
		}
		elts := exprSliceToenodes(callFuncs, eltStore, x.Elts)
		rNum := eltStore.Add(elts)
		return enLit{val: dlit.MustNew(rNum)}
	case *ast.IndexExpr:
		return indexExprToenode(callFuncs, eltStore, x)
	case *ast.ArrayType:
		return nodeToenode(callFuncs, eltStore, x.Elt)
	}
	return enErr{err: ErrSyntax}
}

func indexExprToenode(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	ie *ast.IndexExpr,
) enode {
	var ii, ix int64
	var isInt bool

	indexX := nodeToenode(callFuncs, eltStore, ie.X)
	indexIndex := nodeToenode(callFuncs, eltStore, ie.Index)

	switch xx := indexX.(type) {
	case enErr:
		return indexX
	case enLit:
		switch xii := indexIndex.(type) {
		case enErr:
			return indexIndex
		case enLit:
			ii, isInt = xii.Int()
			if !isInt {
				return enErr{err: ErrSyntax}
			}
			if bl, ok := ie.X.(*ast.BasicLit); ok {
				if bl.Kind != token.STRING {
					return enErr{err: ErrTypeNotIndexable}
				}
				return enLit{val: dlit.MustNew(string(xx.String()[ii]))}
			}
			ix, isInt = xx.Int()
			if !isInt {
				return enErr{err: ErrSyntax}
			}
			elts := eltStore.Get(ix)
			if ii >= int64(len(elts)) {
				return enErr{err: ErrInvalidIndex}
			}
			return elts[ii]

		default:
			return enErr{err: ErrSyntax}
		}
	default:
		return enErr{err: ErrSyntax}
	}
}

func exprSliceToenodes(
	callFuncs map[string]CallFun,
	eltStore *eltStore,
	callArgs []ast.Expr,
) []enode {
	r := make([]enode, len(callArgs))
	for i, arg := range callArgs {
		r[i] = nodeToenode(callFuncs, eltStore, arg)
	}
	return r
}

func eNodesToDLiterals(
	vars map[string]*dlit.Literal,
	ens []enode,
) []*dlit.Literal {
	r := make([]*dlit.Literal, len(ens))
	for i, en := range ens {
		r[i] = en.Eval(vars)
	}
	return r
}

func callFun(
	callFuncs map[string]CallFun,
	name ast.Expr,
	args []*dlit.Literal,
) *dlit.Literal {
	id, ok := name.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("can't get name as *ast.Ident: %s", name))
	}
	f, exists := callFuncs[id.Name]
	if !exists {
		return dlit.MustNew(FunctionNotExistError(id.Name))
	}
	l, err := f(args)
	if err != nil {
		return dlit.MustNew(FunctionError{id.Name, err})
	}
	return l
}
