// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package to handle functions to be used by dexpr
package dexprfuncs

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"math"
	"strings"
)

var CallFuncs = map[string]dexpr.CallFun{}

func init() {
	CallFuncs["in"] = in
	CallFuncs["ni"] = ni
	CallFuncs["min"] = min
	CallFuncs["max"] = max
	CallFuncs["pow"] = pow
	CallFuncs["roundto"] = roundTo
	CallFuncs["sqrt"] = sqrt
	CallFuncs["true"] = alwaysTrue
}

var trueLiteral = dlit.MustNew(true)
var falseLiteral = dlit.MustNew(false)

type WrongNumOfArgsError struct {
	Got  int
	Want int
}

var ErrTooFewArguments = errors.New("too few arguments")
var ErrIncompatibleTypes = errors.New("incompatible types")

func (e WrongNumOfArgsError) Error() string {
	return fmt.Sprintf("wrong number of arguments got: %d, expected: %d",
		e.Got, e.Want)
}

type CantConvertToTypeError struct {
	Kind  string
	Value *dlit.Literal
}

func (e CantConvertToTypeError) Error() string {
	return fmt.Sprintf("can't convert to %s: %s", e.Kind, e.Value)
}

// sqrt returns the square root of a number
func sqrt(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) != 1 {
		err := WrongNumOfArgsError{Got: len(args), Want: 1}
		r := dlit.MustNew(err)
		return r, err
	}
	x, isFloat := args[0].Float()
	if !isFloat {
		if err := args[0].Err(); err != nil {
			return args[0], err
		}
		err := CantConvertToTypeError{Kind: "float", Value: args[0]}
		r := dlit.MustNew(err)
		return r, err
	}
	return dlit.New(math.Sqrt(x))
}

// pow returns the base raised to the power of the exponent
func pow(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) != 2 {
		err := WrongNumOfArgsError{Got: len(args), Want: 2}
		r := dlit.MustNew(err)
		return r, err
	}
	x, isFloat := args[0].Float()
	if !isFloat {
		if err := args[0].Err(); err != nil {
			return args[0], err
		}
		err := CantConvertToTypeError{Kind: "float", Value: args[0]}
		return dlit.MustNew(err), err
	}
	y, isFloat := args[1].Float()
	if !isFloat {
		if err := args[1].Err(); err != nil {
			return args[1], err
		}
		err := CantConvertToTypeError{Kind: "float", Value: args[1]}
		return dlit.MustNew(err), err
	}
	return dlit.New(math.Pow(x, y))
}

// roundTo returns a number rounded to a number of decimal places.
// This uses round half-up to tie-break
func roundTo(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) != 2 {
		err := WrongNumOfArgsError{Got: len(args), Want: 2}
		r := dlit.MustNew(err)
		return r, err
	}

	if _, isInt := args[0].Int(); isInt {
		return args[0], nil
	}

	x, isFloat := args[0].Float()
	if !isFloat {
		if err := args[0].Err(); err != nil {
			return args[0], err
		}
		err := CantConvertToTypeError{Kind: "float", Value: args[0]}
		r := dlit.MustNew(err)
		return r, err
	}
	dp, isInt := args[1].Int()
	if !isInt {
		if err := args[1].Err(); err != nil {
			return args[1], err
		}
		err := CantConvertToTypeError{Kind: "int", Value: args[1]}
		r := dlit.MustNew(err)
		return r, err
	}

	// Prevent rounding errors where too high dp is used
	xNumDP := numDecPlaces(args[0].String())
	if dp > int64(xNumDP) {
		dp = int64(xNumDP)
	}
	shift := math.Pow(10, float64(dp))
	return dlit.New(math.Floor(.5+x*shift) / shift)
}

// in returns whether a string is in a slice of strings
func in(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) < 2 {
		r := dlit.MustNew(ErrTooFewArguments)
		return r, ErrTooFewArguments
	}
	needle := args[0]
	haystack := args[1:]
	if err := needle.Err(); err != nil {
		return needle, err
	}
	for _, v := range haystack {
		if err := v.Err(); err != nil {
			return v, err
		}
		if needle.String() == v.String() {
			return trueLiteral, nil
		}
	}
	return falseLiteral, nil
}

// ni returns whether a string is not in a slice of strings
func ni(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) < 2 {
		r := dlit.MustNew(ErrTooFewArguments)
		return r, ErrTooFewArguments
	}
	needle := args[0]
	haystack := args[1:]
	if err := needle.Err(); err != nil {
		return needle, err
	}
	for _, v := range haystack {
		if err := v.Err(); err != nil {
			return v, err
		}
		if needle.String() == v.String() {
			return falseLiteral, nil
		}
	}
	return trueLiteral, nil
}

var isSmallerExpr = dexpr.MustNew("v < min", CallFuncs)

// min returns the smallest number of those supplied
func min(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) < 2 {
		r := dlit.MustNew(ErrTooFewArguments)
		return r, ErrTooFewArguments
	}

	min := args[0]
	for _, v := range args[1:] {
		vars := map[string]*dlit.Literal{"min": min, "v": v}
		isSmaller, err := isSmallerExpr.EvalBool(vars)
		if err != nil {
			if x, ok := err.(dexpr.InvalidExprError); ok {
				if x.Err == dexpr.ErrIncompatibleTypes {
					return dlit.MustNew(ErrIncompatibleTypes), ErrIncompatibleTypes
				}
				return dlit.MustNew(x.Err), x.Err
			}
			return dlit.MustNew(err), err
		}
		if isSmaller {
			min = v
		}
	}
	return min, nil
}

var isBiggerExpr = dexpr.MustNew("v > max", CallFuncs)

// max returns the biggest number of those supplied
func max(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) < 2 {
		r := dlit.MustNew(ErrTooFewArguments)
		return r, ErrTooFewArguments
	}

	max := args[0]
	for _, v := range args[1:] {
		vars := map[string]*dlit.Literal{"max": max, "v": v}
		isBigger, err := isBiggerExpr.EvalBool(vars)
		if err != nil {
			if x, ok := err.(dexpr.InvalidExprError); ok {
				if x.Err == dexpr.ErrIncompatibleTypes {
					return dlit.MustNew(ErrIncompatibleTypes), ErrIncompatibleTypes
				}
				return dlit.MustNew(x.Err), x.Err
			}
			return dlit.MustNew(err), err
		}
		if isBigger {
			max = v
		}
	}
	return max, nil
}

// alwaysTrue returns true
func alwaysTrue(args []*dlit.Literal) (*dlit.Literal, error) {
	return trueLiteral, nil
}

func numDecPlaces(s string) int {
	i := strings.IndexByte(s, '.')
	if i > -1 {
		s = strings.TrimRight(s, "0")
		return len(s) - i - 1
	}
	return 0
}
