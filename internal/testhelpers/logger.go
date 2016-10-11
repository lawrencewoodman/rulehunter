package testhelpers

import (
	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
	"time"
)

type Logger struct {
	entries []logger.Entry
	entryCh chan logger.Entry
}

func NewLogger() *Logger {
	return &Logger{
		entries: make([]logger.Entry, 0),
		entryCh: make(chan logger.Entry),
	}
}

func (l *Logger) Run(q *quitter.Quitter) {
	quitCheckInSec := time.Duration(2)
	q.Add()
	defer q.Done()

	go func() {
		for {
			if q.ShouldQuit() {
				close(l.entryCh)
				return
			}
			time.Sleep(quitCheckInSec * time.Second)
		}
	}()

	for e := range l.entryCh {
		l.entries = append(l.entries, e)
	}
}

func (l *Logger) SetSvcLogger(logger service.Logger) {
}

func (l *Logger) Error(msg string) {
	l.entryCh <- logger.Entry{
		Level: logger.Error,
		Msg:   msg,
	}
}

func (l *Logger) Info(msg string) {
	l.entryCh <- logger.Entry{
		Level: logger.Info,
		Msg:   msg,
	}
}

func (l *Logger) GetEntries() []logger.Entry {
	return l.entries
}
