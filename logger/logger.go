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

package logger

import (
	"fmt"
	"github.com/kardianos/service"
	"sync"
)

type Entry struct {
	Level Level
	Msg   string
}

type Level int

const (
	Info Level = iota
	Error
)

type Runner func(svcLogger service.Logger, entries chan Entry)

func Run(svcLogger service.Logger, entries chan Entry) {
	for e := range entries {
		switch e.Level {
		case Info:
			svcLogger.Info(e.Msg)
		case Error:
			svcLogger.Error(e.Msg)
		default:
			panic(fmt.Sprintf("Unknown log level: %d", e.Level))
		}
	}
}

type TestLogger struct {
	entries []Entry
	sync.Mutex
}

func NewTestLogger() *TestLogger {
	return &TestLogger{
		entries: make([]Entry, 0),
	}
}

func (t *TestLogger) MakeRun() Runner {
	return func(svcLogger service.Logger, entries chan Entry) {
		for e := range entries {
			t.Lock()
			t.entries = append(t.entries, e)
			t.Unlock()
		}
	}
}

func (t *TestLogger) GetEntries() []Entry {
	t.Lock()
	defer t.Unlock()
	return t.entries
}
