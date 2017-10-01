// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package quitter

import "sync"

type Quitter struct {
	// C is closed to signal to processes to quit
	C  chan struct{}
	wg *sync.WaitGroup
}

func New() *Quitter {
	return &Quitter{C: make(chan struct{}), wg: &sync.WaitGroup{}}
}

// Add a process to wait to quit
func (q *Quitter) Add() {
	q.wg.Add(1)
}

// Done indicates that a process has finished
func (q *Quitter) Done() {
	q.wg.Done()
}

// Quit tell all the processes to quit and waits for them to finish
func (q *Quitter) Quit() {
	close(q.C)
	q.wg.Wait()
}
