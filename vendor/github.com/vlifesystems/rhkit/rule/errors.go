// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package rule

import "errors"

var ErrNoRuleFieldsSpecified = errors.New("no rule fields specified")

type InvalidRuleError struct {
	Rule Rule
}

type IncompatibleTypesRuleError struct {
	Rule Rule
}

type InvalidExprError struct {
	Expr string
}

func (e InvalidRuleError) Error() string {
	return "invalid rule: " + e.Rule.String()
}

func (e IncompatibleTypesRuleError) Error() string {
	return "incompatible types in rule: " + e.Rule.String()
}

func (e InvalidExprError) Error() string {
	return "invalid expression in rule: " + e.Expr
}

type InvalidRuleFieldError string

func (e InvalidRuleFieldError) Error() string {
	return "invalid rule field: " + string(e)
}
