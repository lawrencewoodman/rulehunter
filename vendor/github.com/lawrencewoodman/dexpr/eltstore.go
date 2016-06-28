/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import "github.com/lawrencewoodman/dlit"

// The eltStore is where the composite types store there elements
type eltStore struct {
	elts map[int64][]*dlit.Literal
	num  int64
}

func newEltStore() *eltStore {
	return &eltStore{elts: map[int64][]*dlit.Literal{}, num: 0}
}

// Get returns the elements for n from eltStore
func (e *eltStore) Get(n int64) []*dlit.Literal {
	return e.elts[n]
}

// Add adds a slice of literals to eltStore and returns the number
// that these are stored under for use by Get
func (e *eltStore) Add(ls []*dlit.Literal) int64 {
	rNum := e.num
	e.elts[e.num] = ls
	e.num++
	return rNum
}
