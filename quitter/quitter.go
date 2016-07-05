/*
	rulehuntersrv - A server to find rules in data based on user specified goals
	Copyright (C) 2016 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

// Package quitter handles stopping go routines cleanly
package quitter

import (
	"sync"
)

type Quitter struct {
	shouldQuit bool
	waitGroup  *sync.WaitGroup
}

func New() *Quitter {
	return &Quitter{
		shouldQuit: false,
		waitGroup:  &sync.WaitGroup{},
	}
}

// Add adds a go routine to wait for
func (q *Quitter) Add() {
	q.waitGroup.Add(1)
}

// Done indicates that a go routine has finished
func (q *Quitter) Done() {
	q.waitGroup.Done()
}

// Quit indicates to all the go routines that they should quit, it then waits
// for them to finish.
func (q *Quitter) Quit() {
	q.shouldQuit = true
	q.waitGroup.Wait()
}

// ShouldQuit returns if a go routine should quit
func (q *Quitter) ShouldQuit() bool {
	return q.shouldQuit
}
