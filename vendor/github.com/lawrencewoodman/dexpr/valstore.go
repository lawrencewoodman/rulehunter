/*
 * Copyright (C) 2017 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

import (
	"github.com/lawrencewoodman/dlit"
	"sync"
)

// The valStore is where Literal values are held to avoid continually
// creating new ones
type valStore struct {
	values map[string]*dlit.Literal
	mutex  *sync.Mutex
}

func newValStore() *valStore {
	return &valStore{
		values: make(map[string]*dlit.Literal),
		mutex:  &sync.Mutex{},
	}
}

// Use returns the string s, as a Literal.  It tries to recover it from the
// store where possible rather than recreating it.  If it doesn't exist
// then it will create it and add it to the store
func (vs *valStore) Use(s string) *dlit.Literal {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()
	if l, ok := vs.values[s]; ok {
		return l
	}
	l := dlit.NewString(s)
	vs.values[s] = l
	return l
}
