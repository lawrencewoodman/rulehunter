/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of rhkit.

	rhkit is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	rhkit is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with rhkit; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

// Package to handle functions to be used by dexpr
package dexprfuncs

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"math"
)

var CallFuncs = map[string]dexpr.CallFun{
	"in":      in,
	"ni":      ni,
	"pow":     pow,
	"roundto": roundTo,
	"sqrt":    sqrt,
	"true":    alwaysTrue,
}

type WrongNumOfArgsError struct {
	Got  int
	Want int
}

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
	r, err := dlit.New(math.Sqrt(x))
	return r, err
}

// Pow returns the base raised to the power of the exponent
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
	r, err := dlit.New(math.Pow(x, y))
	return r, err
}

// This uses round half-up to tie-break
func roundTo(args []*dlit.Literal) (*dlit.Literal, error) {
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
		r := dlit.MustNew(err)
		return r, err
	}
	p, isInt := args[1].Int()
	if !isInt {
		if err := args[1].Err(); err != nil {
			return args[1], err
		}
		err := CantConvertToTypeError{Kind: "int", Value: args[0]}
		r := dlit.MustNew(err)
		return r, err
	}
	shift := math.Pow(10, float64(p))
	r, err := dlit.New(math.Floor(.5+x*shift) / shift)
	return r, err
}

// Is a string IN a list of strings
func in(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) < 2 {
		err := errors.New("too few arguments")
		r := dlit.MustNew(err)
		return r, err
	}
	needle := args[0]
	haystack := args[1:]
	for _, v := range haystack {
		if needle.String() == v.String() {
			r, err := dlit.New(true)
			return r, err
		}
	}
	r, err := dlit.New(false)
	return r, err
}

// Is a string NI a list of strings
func ni(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) < 2 {
		err := errors.New("too few arguments")
		r := dlit.MustNew(err)
		return r, err
	}
	needle := args[0]
	haystack := args[1:]
	for _, v := range haystack {
		if needle.String() == v.String() {
			r, err := dlit.New(false)
			return r, err
		}
	}
	r, err := dlit.New(true)
	return r, err
}

// Returns true
func alwaysTrue(args []*dlit.Literal) (*dlit.Literal, error) {
	return dlit.MustNew(true), nil
}
