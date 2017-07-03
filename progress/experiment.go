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
package progress

import (
	"fmt"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"time"
)

// StatusKind represents the status of an experiment
type StatusKind int

const (
	Waiting StatusKind = iota
	Processing
	Success
	Failure
)

type Experiment struct {
	Title    string
	Tags     []string
	Stamp    time.Time // Time of last update
	Filename string
	Msg      string
	Percent  float64
	Status   StatusKind
	monitor  *Monitor
}

// isFinished returns whether an experiment has finished being processed
func isFinished(e *Experiment) bool {
	return e.Status == Success || e.Status == Failure
}

func (e *Experiment) String() string {
	fmtStr := "{Title: %s, Tags: %s, Stamp: %s, Filename: %s, Msg: %s, Percent: %f, Status: %s}"
	return fmt.Sprintf(
		fmtStr, e.Title, e.Tags,
		e.Stamp.Format(time.RFC3339Nano),
		e.Filename, e.Msg, e.Percent, e.Status,
	)
}

// UpdateDetails updates the experiement's title and tags.  It also
// updates the time stamp to the current time.
func (e *Experiment) UpdateDetails(
	title string,
	tags []string,
) error {
	e.Title = title
	e.Tags = tags
	e.Stamp = time.Now()
	if err := e.monitor.writeJSON(); err != nil {
		return err
	}
	e.monitor.htmlCmds <- cmd.Progress
	return nil
}

// ReportProgress updates the progress of the experiment with a message
// and percentage progress (0.0-1.0).
func (e *Experiment) ReportProgress(
	msg string,
	progress float64,
) error {
	return e.updateExperiment(
		e.Filename,
		Processing,
		msg,
		progress,
	)
}

// ReportError updates the progress of the experiment with an error.
// It returns the error passed as a parameter unless there is a
// problem updating the progress file in which case this error will
// be returned.
func (e *Experiment) ReportError(err error) error {
	updateErr := e.updateExperiment(
		e.Filename,
		Failure,
		err.Error(),
		0,
	)
	if updateErr != nil {
		return updateErr
	}
	return err
}

// ReportSuccess updates the progress of the experiment to report success.
func (e *Experiment) ReportSuccess() error {
	err := e.updateExperiment(
		e.Filename,
		Success,
		"Finished processing successfully",
		0,
	)
	if err == nil {
		e.monitor.htmlCmds <- cmd.Reports
	}
	return err
}

// GetFinishStamp returns whether an experiment has finished
// (Success or Failure) and when it finished.  If it hasn't
// finished then it returns the current time.
func (e *Experiment) GetFinishStamp() (bool, time.Time) {
	if isFinished(e) {
		return true, e.Stamp
	}
	return false, time.Now()
}

func (e *Experiment) updateExperiment(
	experimentFilename string,
	status StatusKind,
	msg string,
	percent float64,
) error {
	e.Stamp = time.Now()
	e.Status = status
	e.Msg = msg
	e.Percent = percent
	if err := e.monitor.writeJSON(); err != nil {
		return err
	}
	e.monitor.htmlCmds <- cmd.Progress
	return nil
}
