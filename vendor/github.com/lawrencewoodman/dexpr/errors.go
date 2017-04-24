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

var ErrDivByZero = errors.New("divide by zero")
var ErrUnderflowOverflow = errors.New("underflow/overflow")
var ErrIncompatibleTypes = errors.New("incompatible types")
var ErrInvalidCompositeType = errors.New("invalid composite type")
var ErrInvalidIndex = errors.New("index out of range")
var ErrTypeNotIndexable = errors.New("type does not support indexing")
var ErrSyntax = errors.New("syntax error")

type InvalidExprError struct {
	Expr string
	Err  error
}

func (e InvalidExprError) Error() string {
	return fmt.Sprintf("invalid expression: %s (%s)", e.Expr, e.Err)
}

type InvalidOpError token.Token

func (e InvalidOpError) Error() string {
	return fmt.Sprintf("invalid operator: %s", token.Token(e))
}

type FunctionNotExistError string

func (e FunctionNotExistError) Error() string {
	return fmt.Sprintf("function doesn't exist: %s", string(e))
}

type VarNotExistError string

func (e VarNotExistError) Error() string {
	return fmt.Sprintf("variable doesn't exist: %s", string(e))
}

type FunctionError struct {
	FnName string
	Err    error
}

func (e FunctionError) Error() string {
	return fmt.Sprintf("function: %s, returned error: %s", e.FnName, e.Err)
}
