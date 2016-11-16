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
	"encoding/json"
	"fmt"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ExperimentProgressReporter struct {
	pm                 *ProgressMonitor
	experimentFilename string
}

func NewExperimentProgressReporter(
	pm *ProgressMonitor,
	experimentFilename string,
) (*ExperimentProgressReporter, error) {
	epr := &ExperimentProgressReporter{
		pm:                 pm,
		experimentFilename: experimentFilename,
	}
	return epr, pm.AddExperiment(experimentFilename)
}

func (epr *ExperimentProgressReporter) UpdateDetails(
	title string,
	tags []string,
) error {
	e := epr.pm.findExperiment(epr.experimentFilename)
	if e == nil {
		return fmt.Errorf("Can't update experiment details for: %s",
			epr.experimentFilename)
	}
	e.Title = title
	e.Tags = tags
	e.Stamp = time.Now()
	if err := epr.pm.writeJson(); err != nil {
		return err
	}
	epr.pm.htmlCmds <- cmd.Progress
	return nil
}

func (epr *ExperimentProgressReporter) ReportInfo(msg string) error {
	return epr.pm.updateExperiment(
		epr.experimentFilename,
		Processing,
		msg,
	)
}

// This returns the error passed as a parameter unless there is a problem
// updating the progress file in which case this error will be returned
func (epr *ExperimentProgressReporter) ReportError(err error) error {
	updateErr := epr.pm.updateExperiment(
		epr.experimentFilename,
		Failure,
		err.Error(),
	)
	if updateErr != nil {
		return updateErr
	}
	return err
}

func (epr *ExperimentProgressReporter) ReportSuccess() error {
	err := epr.pm.updateExperiment(
		epr.experimentFilename,
		Success,
		"Finished processing successfully",
	)
	if err == nil {
		epr.pm.htmlCmds <- cmd.Reports
	}
	return err
}

type ProgressMonitor struct {
	filename    string
	htmlCmds    chan cmd.Cmd
	experiments []*Experiment
}

type StatusKind int

const (
	Waiting StatusKind = iota
	Processing
	Success
	Failure
)

type progressFile struct {
	Experiments []*Experiment
}

type Experiment struct {
	Title              string
	Tags               []string
	Stamp              time.Time // Time of last update
	ExperimentFilename string
	Msg                string
	Status             StatusKind
}

func (s StatusKind) String() string {
	switch s {
	case Waiting:
		return "waiting"
	case Processing:
		return "processing"
	case Success:
		return "success"
	case Failure:
		return "failure"
	}
	panic("Unrecognized status")
}

func NewMonitor(
	progressDir string,
	htmlCmds chan cmd.Cmd,
) (*ProgressMonitor, error) {
	var progress progressFile
	experiments := []*Experiment{}
	filename := filepath.Join(progressDir, "progress.json")

	f, err := os.Open(filename)
	if os.IsNotExist(err) {
	} else if err != nil {
		return nil, err
	} else {
		defer f.Close()
		dec := json.NewDecoder(f)
		if err = dec.Decode(&progress); err != nil {
			return nil, err
		}
		experiments = progress.Experiments
	}

	pm := &ProgressMonitor{
		filename:    filename,
		htmlCmds:    htmlCmds,
		experiments: experiments,
	}
	sort.Sort(pm)
	return pm, nil
}

func (pm *ProgressMonitor) AddExperiment(
	experimentFilename string,
) error {
	e := pm.findExperiment(experimentFilename)
	if e == nil {
		newExperiment := &Experiment{
			"",
			[]string{},
			time.Now(),
			experimentFilename,
			"Waiting to be processed",
			Waiting,
		}
		pm.experiments = append(pm.experiments, newExperiment)
	} else {
		if isFinished(e) {
			return nil
		}
		e.Title = ""
		e.Tags = []string{}
		e.Stamp = time.Now()
		e.Msg = "Waiting to be processed"
		e.Status = Waiting
	}
	if err := pm.writeJson(); err != nil {
		return err
	}
	pm.htmlCmds <- cmd.Progress
	return nil
}

func (pm *ProgressMonitor) GetExperiments() []*Experiment {
	return pm.experiments
}

// GetFinishStamp returns whether an experiment has finished
// (Success or Failure) and when it finished.  If it hasn't
// finished then it returns the current time.
func (pm *ProgressMonitor) GetFinishStamp(
	experimentFilename string,
) (bool, time.Time) {
	e := pm.findExperiment(experimentFilename)
	if e == nil {
		return false, time.Now()
	}
	if isFinished(e) {
		return true, e.Stamp
	}
	return false, time.Now()
}

func (pm *ProgressMonitor) updateExperiment(
	experimentFilename string,
	status StatusKind,
	msg string,
) error {
	e := pm.findExperiment(experimentFilename)
	if e == nil {
		return fmt.Errorf("can't update experiment with filename: %s",
			experimentFilename)
	}
	e.Stamp = time.Now()
	e.Status = status
	e.Msg = msg
	if err := pm.writeJson(); err != nil {
		return err
	}
	pm.htmlCmds <- cmd.Progress
	return nil
}

// Returns experiment if found experiment or nil if not found
func (pm *ProgressMonitor) findExperiment(
	experimentFilename string,
) *Experiment {
	for _, experiment := range pm.experiments {
		if experiment.ExperimentFilename == experimentFilename {
			return experiment
		}
	}
	return nil
}

func (pm *ProgressMonitor) writeJson() error {
	sort.Sort(pm)
	successfulExperiments := []*Experiment{}
	for _, e := range pm.experiments {
		if e.Status == Success {
			successfulExperiments = append(successfulExperiments, e)
		}
	}
	progress := &progressFile{successfulExperiments}
	json, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pm.filename, json, 0640)
}

// Implements sort.Interface
func (pm *ProgressMonitor) Len() int { return len(pm.experiments) }
func (pm *ProgressMonitor) Swap(i, j int) {
	pm.experiments[i], pm.experiments[j] =
		pm.experiments[j], pm.experiments[i]
}

func (pm *ProgressMonitor) Less(i, j int) bool {
	return pm.experiments[j].Stamp.Before(pm.experiments[i].Stamp)
}

// isFinished returns whether an experiment has finished being processed
func isFinished(e *Experiment) bool {
	return e.Status == Success || e.Status == Failure
}
