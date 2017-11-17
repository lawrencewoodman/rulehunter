// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package testhelpers

import (
	"sort"

	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/quitter"
)

type Logger struct {
	entries   []Entry
	isRunning bool
}

type Entry struct {
	Level Level
	Msg   string
}

type Level int

const (
	Info Level = iota
	Error
)

func NewLogger() *Logger {
	return &Logger{
		entries:   make([]Entry, 0),
		isRunning: false,
	}
}

func (l *Logger) Run(quit *quitter.Quitter) {
	quit.Add()
	defer quit.Done()
	l.isRunning = true
	defer func() { l.isRunning = false }()
	for {
		select {
		case <-quit.C:
			return
		}
	}
}

// Running returns whether Logger is running
func (l *Logger) Running() bool {
	return l.isRunning
}

func (l *Logger) SetSvcLogger(logger service.Logger) {
}

func (l *Logger) Error(err error) error {
	entry := Entry{
		Level: Error,
		Msg:   err.Error(),
	}
	l.entries = append(l.entries, entry)
	return err
}

func (l *Logger) Info(msg string) {
	entry := Entry{
		Level: Info,
		Msg:   msg,
	}
	l.entries = append(l.entries, entry)
}

// GetEntries returns the entries from the log.  If an argument is passed,
// it represents whether to make the errors unique.
func (l *Logger) GetEntries(args ...bool) []Entry {
	uniqueErrors := false
	if len(args) == 1 {
		uniqueErrors = args[0]
	}
	if !uniqueErrors {
		return l.entries
	}
	r := []Entry{}
	for _, e := range l.entries {
		found := false
		if e.Level == Error {
			for _, re := range r {
				if e.Level == re.Level && e.Msg == re.Msg {
					found = true
					break
				}
			}
		}
		if !found {
			r = append(r, e)
		}
	}
	return r
}

func SortLogEntries(entries []Entry) []Entry {
	r := make([]Entry, len(entries))
	copy(r, entries)
	sort.SliceStable(
		r,
		func(i, j int) bool {
			if r[i].Level < r[j].Level {
				return true
			} else if r[i].Level == r[j].Level && r[i].Msg < r[j].Msg {
				return true
			}
			return false
		},
	)
	return r
}
