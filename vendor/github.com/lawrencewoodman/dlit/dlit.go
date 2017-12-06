/*
 * A package for easing the use of dynamic literals
 *
 * Copyright (C) 2016-2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

// Package dlit handles dynamic typed Literals
package dlit

import (
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
)

// Literal represents a dynamically typed value
type Literal struct {
	i          int64
	f          atomic.Value
	s          atomic.Value
	b          int32
	e          atomic.Value
	canBeInt   int32
	canBeFloat int32
	canBeBool  int32
	canBeError int32
}

type canBeKind int32

const (
	unknown canBeKind = iota
	yes
	no
)

// New creates a Literal from any of the following types:
// int, int64, float32, float64, string, bool, error
func New(v interface{}) (*Literal, error) {
	var err error
	l := &Literal{canBeInt: int32(unknown), canBeFloat: int32(unknown),
		canBeBool: int32(unknown), canBeError: int32(no)}
	s := ""
	switch e := v.(type) {
	case int:
		atomic.StoreInt64(&l.i, int64(e))
		l.canBeInt = int32(yes)
	case int64:
		atomic.StoreInt64(&l.i, e)
		l.canBeInt = int32(yes)
	case float32:
		l.f.Store(float64(e))
		l.canBeFloat = int32(yes)
	case float64:
		l.f.Store(e)
		l.canBeFloat = int32(yes)
	case string:
		s = e
	case bool:
		l.setBool(e)
		l.canBeBool = int32(yes)
	case error:
		l.e.Store(e)
		l.canBeInt = int32(no)
		l.canBeFloat = int32(no)
		l.canBeBool = int32(no)
		l.canBeError = int32(yes)
	default:
		err = InvalidKindError(reflect.TypeOf(v).String())
		l.e.Store(err)
		l.canBeInt = int32(no)
		l.canBeFloat = int32(no)
		l.canBeBool = int32(no)
		l.canBeError = int32(yes)
	}
	l.s.Store(s)
	return l, err
}

// NewString creates a Literal from a string
func NewString(s string) *Literal {
	l := &Literal{}
	l.canBeInt = int32(unknown)
	l.canBeFloat = int32(unknown)
	l.canBeBool = int32(unknown)
	l.canBeError = int32(no)
	l.s.Store(s)
	return l
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
	switch loadCanBeKind(&l.canBeInt) {
	case yes:
		return atomic.LoadInt64(&l.i), true
	case no:
		return 0, false
	case unknown:
		v, ok := parseInt(l.String())
		if ok {
			atomic.StoreInt64(&l.i, v)
			storeCanBeKind(&l.canBeInt, yes)
			return v, true
		}
	}
	storeCanBeKind(&l.canBeInt, no)
	return 0, false
}

// parseInt returns a string as an int64 value if it can and says whether
// this was successful.  If the string has a decimal point in it but is
// still an integer then this number will be converted successfully;
// e.g. -6, 6, 6., 6.0, 6.000 will all be fine.
func parseInt(s string) (value int64, ok bool) {
	dpPos := -1
	for i, r := range s {
		if r == '.' {
			dpPos = i
		} else if dpPos > -1 && r != '0' {
			return 0, false
		}
	}
	if dpPos >= 0 {
		s = s[:dpPos]
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, true
	}
	return 0, false
}

// Float returns Literal as a float64 and whether it can be a float64
func (l *Literal) Float() (value float64, canBeFloat bool) {
	switch loadCanBeKind(&l.canBeFloat) {
	case yes:
		return l.f.Load().(float64), true
	case no:
		return 0, false
	case unknown:
		f, err := strconv.ParseFloat(l.String(), 64)
		if err == nil {
			l.f.Store(f)
			storeCanBeKind(&l.canBeFloat, yes)
			return f, true
		}
	}
	storeCanBeKind(&l.canBeFloat, no)
	return 0, false
}

// Bool returns Literal as a bool and whether it can be a bool
func (l *Literal) Bool() (value bool, canBeBool bool) {
	switch loadCanBeKind(&l.canBeBool) {
	case yes:
		return l.getBool(), true
	case no:
		return false, false
	case unknown:
		if l.isInt() {
			v := atomic.LoadInt64(&l.i)
			if v == 0 {
				l.setBool(false)
				storeCanBeKind(&l.canBeBool, yes)
				return false, true
			} else if v == 1 {
				l.setBool(true)
				storeCanBeKind(&l.canBeBool, yes)
				return true, true
			}
		} else if l.isFloat() {
			v := l.f.Load().(float64)
			if v == 0.0 {
				l.setBool(false)
				storeCanBeKind(&l.canBeBool, yes)
				return false, true
			} else if v == 1.0 {
				l.setBool(true)
				storeCanBeKind(&l.canBeBool, yes)
				return true, true
			}
		} else {
			b, err := strconv.ParseBool(l.String())
			if err == nil {
				l.setBool(b)
				storeCanBeKind(&l.canBeBool, yes)
				return b, true
			}
		}
	}
	storeCanBeKind(&l.canBeBool, no)
	return false, false
}

// String returns Literal as a string
func (l *Literal) String() string {
	s := l.s.Load().(string)
	if len(s) > 0 {
		return s
	}
	switch true {
	case l.isInt():
		s = strconv.FormatInt(atomic.LoadInt64(&l.i), 10)
	case l.isFloat():
		s = strconv.FormatFloat(l.f.Load().(float64), 'f', -1, 64)
	case l.isBool():
		if l.getBool() {
			s = "true"
		} else {
			s = "false"
		}
	case l.isError():
		s = l.Err().Error()
	}
	l.s.Store(s)
	return s
}

// Err returns an error if can be an error or nil
func (l *Literal) Err() error {
	if l.isError() {
		return l.e.Load().(error)
	}
	return nil
}

func (l *Literal) isInt() bool {
	return loadCanBeKind(&l.canBeInt) == yes
}

func (l *Literal) isFloat() bool {
	return loadCanBeKind(&l.canBeFloat) == yes
}

func (l *Literal) isBool() bool {
	return loadCanBeKind(&l.canBeBool) == yes
}

func (l *Literal) isError() bool {
	return loadCanBeKind(&l.canBeError) == yes
}

// setBool sets l.b to b using an atomic operation
func (l *Literal) setBool(b bool) {
	if b {
		atomic.StoreInt32(&l.b, 1)
	} else {
		atomic.StoreInt32(&l.b, 0)
	}
}

// getBool returns l.b using an atomic operation
func (l *Literal) getBool() bool {
	if atomic.LoadInt32(&l.b) == 1 {
		return true
	}
	return false
}

// loadCanBeKind gets the value of x using an atomic operation
func loadCanBeKind(x *int32) canBeKind {
	return canBeKind(atomic.LoadInt32(x))
}

// storeCanBeKind sets x to v using an atomic operation
func storeCanBeKind(x *int32, v canBeKind) {
	atomic.StoreInt32(x, int32(v))
}

// InvalidKindError indicates that a Literal can't be created from this type
type InvalidKindError string

// Error returns the error as a string
func (e InvalidKindError) Error() string {
	return fmt.Sprintf("can't create Literal from type: %s", string(e))
}
