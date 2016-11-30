/*
	rulehunter - A server to find rules in data based on user specified goals
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
