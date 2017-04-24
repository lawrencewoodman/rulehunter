/*
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"github.com/lawrencewoodman/dlit"
)

type enode interface {
	Eval(map[string]*dlit.Literal) *dlit.Literal
}

type enErr struct {
	err error
}

type enFunc struct {
	fn func(map[string]*dlit.Literal) *dlit.Literal
}

type enLit struct {
	val *dlit.Literal
}

type enVar string

func (ee enErr) Err() error {
	return ee.err
}

func (ee enErr) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	return dlit.MustNew(ee)
}

func (ef enFunc) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	return ef.fn(vars)
}

func (el enLit) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	return el.val
}

func (el enLit) Int() (int64, bool) {
	i, isInt := el.val.Int()
	return i, isInt
}

func (el enLit) String() string {
	return el.val.String()
}

func (ev enVar) Eval(vars map[string]*dlit.Literal) *dlit.Literal {
	if l, ok := vars[string(ev)]; ok {
		return l
	}
	return dlit.MustNew(VarNotExistError(ev))
}
