/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"errors"
	"fmt"
	"go/token"
)

var ErrUnderflowOverflow = errors.New("underflow/overflow")
var ErrDivByZero = errors.New("divide by zero")
var ErrIncompatibleTypes = errors.New("incompatible types")
var ErrInvalidCompositeType = errors.New("invalid composite type")
var ErrInvalidIndex = errors.New("index out of range")
var ErrTypeNotIndexable = errors.New("type does not support indexing")
var ErrSyntax = errors.New("syntax error")

type ErrInvalidExpr struct {
	Expr string
	Err  error
}

func (e ErrInvalidExpr) Error() string {
	return fmt.Sprintf("invalid expression: %s (%s)", e.Expr, e.Err)
}

type ErrInvalidOp token.Token

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("invalid operator: %s", token.Token(e))
}

type ErrFunctionNotExist string

func (e ErrFunctionNotExist) Error() string {
	return fmt.Sprintf("function doesn't exist: %s", string(e))
}

type ErrVarNotExist string

func (e ErrVarNotExist) Error() string {
	return fmt.Sprintf("variable doesn't exist: %s", string(e))
}

type ErrFunctionError struct {
	FnName string
	Err    error
}

func (e ErrFunctionError) Error() string {
	return fmt.Sprintf("function: %s, returned error: %s", e.FnName, e.Err)
}
