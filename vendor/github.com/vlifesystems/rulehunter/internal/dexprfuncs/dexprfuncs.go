/*
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>
	This file is part of Rulehunter.

	Rulehunter is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Rulehunter is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Rulehunter; see the file COPYING.  If not, see
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
	"roundto": roundTo,
	"in":      in,
	"ni":      ni,
	"true":    alwaysTrue,
}

// This uses round half-up to tie-break
func roundTo(args []*dlit.Literal) (*dlit.Literal, error) {
	if len(args) > 2 {
		err := errors.New("Too many arguments")
		r := dlit.MustNew(err)
		return r, err
	}
	x, isFloat := args[0].Float()
	if !isFloat {
		if err, isErr := args[0].Err(); isErr {
			return args[0], err
		}
		err := errors.New(fmt.Sprintf("Can't convert to float: %s", args[0]))
		r := dlit.MustNew(err)
		return r, err
	}
	p, isInt := args[1].Int()
	if !isInt {
		if err, isErr := args[1].Err(); isErr {
			return args[1], err
		}
		err := errors.New(fmt.Sprintf("Can't convert to int: %s", args[0]))
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
		err := errors.New("Too few arguments")
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
		err := errors.New("Too few arguments")
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
