package testhelpers

import (
	"github.com/kardianos/service"
)

type Logger struct {
	entries []Entry
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
		entries: make([]Entry, 0),
	}
}

func (l *Logger) Run(quit <-chan struct{}) {
	for {
		select {
		case <-quit:
			return
		}
	}
}

func (l *Logger) SetSvcLogger(logger service.Logger) {
}

func (l *Logger) Error(msg string) {
	entry := Entry{
		Level: Error,
		Msg:   msg,
	}
	l.entries = append(l.entries, entry)
}

func (l *Logger) Info(msg string) {
	entry := Entry{
		Level: Info,
		Msg:   msg,
	}
	l.entries = append(l.entries, entry)
}

func (l *Logger) GetEntries() []Entry {
	return l.entries
}
