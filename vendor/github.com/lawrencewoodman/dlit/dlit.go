/*
 * A package for easing the use of dynamic literals
 *
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dlit handles dynamic typed Literals
package dlit

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// Literal represents a dynamically typed value
type Literal struct {
	i          int64
	f          float64
	s          string
	b          bool
	e          error
	canBeInt   canBeKind
	canBeFloat canBeKind
	canBeBool  canBeKind
	canBeError canBeKind
}

type canBeKind int

const (
	unknown canBeKind = iota
	yes
	no
)

// New creates a Literal from any of the following types:
// int, int64, float32, float64, string, bool, error
func New(v interface{}) (*Literal, error) {
	switch e := v.(type) {
	case int:
		return &Literal{i: int64(e), canBeInt: yes}, nil
	case int64:
		return &Literal{i: e, canBeInt: yes}, nil
	case float32:
		return &Literal{f: float64(e), canBeFloat: yes}, nil
	case float64:
		return &Literal{f: e, canBeFloat: yes}, nil
	case string:
		return &Literal{s: e}, nil
	case bool:
		return &Literal{b: e, canBeBool: yes}, nil
	case error:
		return newErrorLiteral(e), nil
	}
	err := InvalidKindError(reflect.TypeOf(v).String())
	return newErrorLiteral(err), err
}

// NewString creates a Literal from a string
func NewString(s string) *Literal {
	return &Literal{s: s}
}

// MustNew creates a New Literal and panic if it fails
func MustNew(v interface{}) *Literal {
	l, err := New(v)
	if err != nil {
		panic(err.Error())
	}
	return l
}

// Int returns Literal as an int64 and whether it can be an int64
func (l *Literal) Int() (value int64, canBeInt bool) {
	switch l.canBeInt {
	case yes:
		return l.i, true
	case unknown:
		str := trailingZerosRegexp.ReplaceAllString(l.String(), "")
		i, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			l.canBeInt = yes
			l.i = i
			return i, true
		}
	}
	l.canBeInt = no
	return 0, false
}

// Float returns Literal as a float64 and whether it can be a float64
func (l *Literal) Float() (value float64, canBeFloat bool) {
	switch l.canBeFloat {
	case yes:
		return l.f, true
	case unknown:
		f, err := strconv.ParseFloat(l.String(), 64)
		if err == nil {
			l.canBeFloat = yes
			l.f = f
			return f, true
		}
	}
	l.canBeFloat = no
	return 0, false
}

// Bool returns Literal as a bool and whether it can be a bool
func (l *Literal) Bool() (value bool, canBeBool bool) {
	switch l.canBeBool {
	case yes:
		return l.b, true
	case unknown:
		if l.canBeInt == yes {
			if l.i == 0 {
				l.canBeBool = yes
				l.b = false
				return false, true
			} else if l.i == 1 {
				l.canBeBool = yes
				l.b = true
				return true, true
			}
		} else if l.canBeFloat == yes {
			if l.f == 0.0 {
				l.canBeBool = yes
				l.b = false
				return false, true
			} else if l.f == 1.0 {
				l.canBeBool = yes
				l.b = true
				return true, true
			}
		} else {
			b, err := strconv.ParseBool(l.s)
			if err == nil {
				l.canBeBool = yes
				l.b = b
				return b, true
			}
		}
	}
	l.canBeBool = no
	return false, false
}

// String returns Literal as a string
func (l *Literal) String() string {
	if len(l.s) > 0 {
		return l.s
	}
	switch true {
	case l.canBeInt == yes:
		l.s = strconv.FormatInt(l.i, 10)
	case l.canBeFloat == yes:
		l.s = strconv.FormatFloat(l.f, 'f', -1, 64)
	case l.canBeBool == yes:
		if l.b {
			l.s = "true"
		} else {
			l.s = "false"
		}
	case l.canBeError == yes:
		l.s = l.e.Error()
	}
	return l.s
}

// Err returns an error if can be an error or nil
func (l *Literal) Err() error {
	if l.canBeError == yes {
		return l.e
	}
	return nil
}

// InvalidKindError indicates that a Literal can't be created from this type
type InvalidKindError string

// Error returns the error as a string
func (e InvalidKindError) Error() string {
	return fmt.Sprintf("can't create Literal from type: %s", string(e))
}

func newErrorLiteral(e error) *Literal {
	return &Literal{e: e, canBeInt: no, canBeFloat: no, canBeBool: no,
		canBeError: yes}
}

var trailingZerosRegexp = regexp.MustCompile("\\.0*$")
