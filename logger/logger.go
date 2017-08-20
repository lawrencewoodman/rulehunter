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

package logger

import (
	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/quitter"
)

type Logger interface {
	Run(*quitter.Quitter)
	Info(string)
	Error(error) error
	SetSvcLogger(service.Logger)
}

type SvcLogger struct {
	svcLogger service.Logger
}

func NewSvcLogger() *SvcLogger {
	return &SvcLogger{
		svcLogger: nil,
	}
}

func (l *SvcLogger) Run(quit *quitter.Quitter) {
	if l.svcLogger == nil {
		panic("service logger not set")
	}
	quit.Add()
	defer quit.Done()

	for {
		select {
		case <-quit.C:
			return
		}
	}
}

func (l *SvcLogger) SetSvcLogger(logger service.Logger) {
	l.svcLogger = logger
}

// Error logs an error and returns the same error
func (l *SvcLogger) Error(err error) error {
	l.svcLogger.Error(err.Error())
	return err
}

func (l *SvcLogger) Info(msg string) {
	l.svcLogger.Info(msg)
}
