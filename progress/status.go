// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package progress

import (
	"fmt"
	"strings"
	"time"

	"github.com/vlifesystems/rulehunter/report"
)

// StatusKind represents the status of an experiment
type StatusKind int

const (
	Waiting StatusKind = iota
	Processing
	Success
	Error
)

func (s StatusKind) String() string {
	switch s {
	case Waiting:
		return "waiting"
	case Processing:
		return "processing"
	case Success:
		return "success"
	case Error:
		return "error"
	}
	panic("Unrecognized status")
}

type Status struct {
	Stamp   time.Time  `json:"stamp"` // Time of last update
	Msg     string     `json:"msg"`
	Percent float64    `json:"-"`
	State   StatusKind `json:"state"`
}

func NewStatus() *Status {
	return &Status{
		Stamp:   time.Now(),
		Msg:     "Waiting to be processed",
		Percent: 0,
		State:   Waiting,
	}
}

func (s *Status) String() string {
	return fmt.Sprintf(
		"{stamp: %s, msg: %s, percent: %.2f, state: %s}",
		s.Stamp, s.Msg, s.Percent, s.State,
	)
}

// IsFinished returns whether the status is either Success or Error
func (s *Status) IsFinished() bool {
	return s.State == Success || s.State == Error
}

// SetProgress sets the progress with a message
// and percentage progress (0.0-1.0).
func (s *Status) SetProgress(mode report.ModeKind, msg string, percent float64) {
	s.Stamp = time.Now()
	s.Msg = strings.Title(mode.String()) + " > " + msg
	s.Percent = percent
	s.State = Processing
}

// SetError sets Msg to the error and State to Error.
func (s *Status) SetError(err error) {
	s.Stamp = time.Now()
	s.Msg = err.Error()
	s.Percent = 0.0
	s.State = Error
}

// SetSuccess sets the Msg to report success and State to Success
func (s *Status) SetSuccess() {
	s.Stamp = time.Now()
	s.Msg = "Finished processing successfully"
	s.Percent = 0.0
	s.State = Success
}
