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

// ExperimentReporter represents an experiment file that is being
// reported on.
type ExperimentReporter struct {
	m        *Monitor
	filename string
}

// NewExperimentReporter creates a new ExperimentReporter
func NewExperimentReporter(
	m *Monitor,
	filename string,
) (*ExperimentReporter, error) {
	er := &ExperimentReporter{
		m:        m,
		filename: filename,
	}
	return er, m.AddExperiment(filename)
}

// UpdateDetails updates the experiement's title and tags.  It also
// updates the time stamp to the current time.
func (er *ExperimentReporter) UpdateDetails(
	title string,
	tags []string,
) error {
	e := er.m.findExperiment(er.filename)
	if e == nil {
		return fmt.Errorf("Can't update experiment details for: %s",
			er.filename)
	}
	e.Title = title
	e.Tags = tags
	e.Stamp = time.Now()
	if err := er.m.writeJSON(); err != nil {
		return err
	}
	er.m.htmlCmds <- cmd.Progress
	return nil
}

// ReportProgress updates the progress of the experiment with a message
// and percentage progress (0.0-1.0).
func (er *ExperimentReporter) ReportProgress(
	msg string,
	progress float64,
) error {
	return er.m.updateExperiment(
		er.filename,
		Processing,
		msg,
		progress,
	)
}

// ReportError updates the progress of the experiment with an error.
// It returns the error passed as a parameter unless there is a
// problem updating the progress file in which case this error will
// be returned.
func (er *ExperimentReporter) ReportError(err error) error {
	updateErr := er.m.updateExperiment(
		er.filename,
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
func (er *ExperimentReporter) ReportSuccess() error {
	err := er.m.updateExperiment(
		er.filename,
		Success,
		"Finished processing successfully",
		0,
	)
	if err == nil {
		er.m.htmlCmds <- cmd.Reports
	}
	return err
}
