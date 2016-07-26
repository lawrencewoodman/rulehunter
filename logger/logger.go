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
	Run(*quitter.Quitter)
	Log(Level, string)
	SetSvcLogger(service.Logger)
}

type SvcLogger struct {
	svcLogger service.Logger
	entries   chan Entry
	quitter   *quitter.Quitter
}

func NewSvcLogger() *SvcLogger {
	return &SvcLogger{
		svcLogger: nil,
		entries:   make(chan Entry),
	}
}

func (l *SvcLogger) Run(q *quitter.Quitter) {
	quitCheckInSec := time.Duration(2)
	if l.svcLogger == nil {
		panic("service logger not set")
	}
	q.Add()
	defer q.Done()

	go func() {
		for {
			if q.ShouldQuit() {
				close(l.entries)
				return
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
