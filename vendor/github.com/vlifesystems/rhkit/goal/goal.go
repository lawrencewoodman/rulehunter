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

// Package goal handles goal expressions and their results
package goal

import (
	"fmt"
	"github.com/lawrencewoodman/dexpr"
	"github.com/lawrencewoodman/dlit"
	"github.com/vlifesystems/rhkit/internal/dexprfuncs"
)

type Goal struct {
	expr *dexpr.Expr
}

type InvalidGoalError string

func (e InvalidGoalError) Error() string {
	return "invalid goal: " + string(e)
}

func New(exprStr string) (*Goal, error) {
	expr, err := dexpr.New(exprStr, dexprfuncs.CallFuncs)
	if err != nil {
		return nil, InvalidGoalError(exprStr)
	}
	return &Goal{expr}, nil
}

// This should only be used for testing
func MustNew(expr string) *Goal {
	g, err := New(expr)
	if err != nil {
		panic(fmt.Sprintf("Can't create goal: %s", err))
	}
	return g
}

func (g *Goal) String() string {
	return g.expr.String()
}

func (g *Goal) Assess(aggregators map[string]*dlit.Literal) (bool, error) {
	passed, err := g.expr.EvalBool(aggregators)
	return passed, err
}
