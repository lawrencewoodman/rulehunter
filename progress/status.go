/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

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

package progress

import "time"

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

// IsFinished returns whether the status is either Success or Error
func (s *Status) IsFinished() bool {
	return s.State == Success || s.State == Error
}

// SetProgress sets the progress with a message
// and percentage progress (0.0-1.0).
func (s *Status) SetProgress(msg string, percent float64) {
	s.Stamp = time.Now()
	s.Msg = msg
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
