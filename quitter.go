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

package main

import (
	"github.com/kardianos/service"
	"os"
	"sync"
)

type quitter struct {
	shouldQuit bool
	svc        service.Service
	waitGroup  *sync.WaitGroup
}

func newQuitter() *quitter {
	return &quitter{
		shouldQuit: false,
		waitGroup:  &sync.WaitGroup{},
	}
}

func (q *quitter) SetService(s service.Service) {
	q.svc = s
}

// Add adds a go routine to wait for
func (q *quitter) Add() {
	q.waitGroup.Add(1)
}

// Done indicates that a go routine has finished
func (q *quitter) Done() {
	q.waitGroup.Done()
}

// Quit indicates to all the go routines that they should quit, it then waits
// for them to finish. One they have all finished the os.Interrupt signal
// is sent to stop the process.
func (q *quitter) Quit() {
	q.shouldQuit = true
	q.waitGroup.Wait()
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		panic("Can't find process to Quit")
	}
	p.Signal(os.Interrupt)
}

// ShouldQuit returns if a go routine should quit
func (q *quitter) ShouldQuit() bool {
	return q.shouldQuit
}
