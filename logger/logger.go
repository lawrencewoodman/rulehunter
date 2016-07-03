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
	"github.com/vlifesystems/rulehuntersrv/quitter"
	"sync"
	"time"
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

type Logger interface {
	Run()
	Log(Level, string)
	SetSvcLogger(service.Logger)
}

type SvcLogger struct {
	svcLogger service.Logger
	entries   chan Entry
	quitter   *quitter.Quitter
}

func NewSvcLogger(quitter *quitter.Quitter) *SvcLogger {
	return &SvcLogger{
		svcLogger: nil,
		entries:   make(chan Entry),
		quitter:   quitter,
	}
}

func (l *SvcLogger) Run() {
	quitCheckInSec := time.Duration(2)
	if l.svcLogger == nil {
		panic("service logger not set")
	}
	l.quitter.Add()
	defer l.quitter.Done()

	go func() {
		for {
			if l.quitter.ShouldQuit() {
				close(l.entries)
			}
			time.Sleep(quitCheckInSec * time.Second)
		}
	}()

	for e := range l.entries {
		switch e.Level {
		case Info:
			l.svcLogger.Info(e.Msg)
		case Error:
			l.svcLogger.Error(e.Msg)
		default:
			panic(fmt.Sprintf("Unknown log level: %d", e.Level))
		}
	}
}

func (l *SvcLogger) SetSvcLogger(logger service.Logger) {
	l.svcLogger = logger
}

func (l *SvcLogger) Log(level Level, msg string) {
	l.entries <- Entry{
		Level: level,
		Msg:   msg,
	}
}

type TestLogger struct {
	entries []Entry
	quitter *quitter.Quitter
	sync.Mutex
}

func NewTestLogger(quitter *quitter.Quitter) *TestLogger {
	return &TestLogger{
		entries: make([]Entry, 0),
		quitter: quitter,
	}
}

func (l *TestLogger) Run() {
	l.quitter.Add()
	for !l.quitter.ShouldQuit() {
	}
	l.quitter.Done()
}

func (l *TestLogger) SetSvcLogger(logger service.Logger) {
}

func (l *TestLogger) Log(level Level, msg string) {
	e := Entry{
		Level: level,
		Msg:   msg,
	}
	l.Lock()
	defer l.Unlock()
	l.entries = append(l.entries, e)
}

func (t *TestLogger) GetEntries() []Entry {
	t.Lock()
	defer t.Unlock()
	return t.entries
}
