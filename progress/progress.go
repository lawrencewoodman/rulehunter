/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package progress

import (
	"encoding/json"
	"fmt"
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
	epr := &ExperimentProgressReporter{pm, experimentFilename}
	return epr, pm.AddExperiment(experimentFilename)
}

func (epr *ExperimentProgressReporter) UpdateDetails(
	title string,
	tags []string,
) error {
	return epr.pm.updateExperimentDetails(
		epr.experimentFilename,
		title,
		tags,
	)
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
	return epr.pm.updateExperiment(
		epr.experimentFilename,
		Success,
		"Finished processing successfully",
	)
}

type ProgressMonitor struct {
	progressFilename string
}

type StatusKind int

const (
	Waiting StatusKind = iota
	Processing
	Success
	Failure
)

type Progress struct {
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

func NewMonitor(progressDir string) *ProgressMonitor {
	return &ProgressMonitor{
		filepath.Join(progressDir, "progress.json"),
	}
}

func (pm *ProgressMonitor) AddExperiment(
	experimentFilename string,
) error {
	experiments, err := pm.GetExperiments()
	if err != nil {
		return err
	}
	newExperiment := &Experiment{
		"",
		[]string{},
		time.Now(),
		experimentFilename,
		"Waiting to be processed",
		Waiting,
	}
	newExperiments := make([]*Experiment, len(experiments))
	i := 0
	for _, experiment := range experiments {
		if experiment.ExperimentFilename != experimentFilename {
			newExperiments[i] = experiment
			i++
		}
	}
	newExperiments = newExperiments[:i]
	newExperiments = append(newExperiments, newExperiment)
	return pm.writeJson(newExperiments)
}

func (pm *ProgressMonitor) GetExperiments() ([]*Experiment, error) {
	var progress Progress
	f, err := os.Open(pm.progressFilename)
	if os.IsNotExist(err) {
		return []*Experiment{}, nil
	}
	if err != nil {
		return []*Experiment{}, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&progress); err != nil {
		return []*Experiment{}, err
	}
	sort.Sort(sort.Reverse(progress))
	return progress.Experiments, nil
}

func (pm *ProgressMonitor) updateExperimentDetails(
	experimentFilename string,
	title string,
	tags []string,
) error {
	experiments, err := pm.GetExperiments()
	if err != nil {
		return err
	}
	newExperiment := &Experiment{
		title,
		tags,
		time.Now(),
		experimentFilename,
		"Waiting to be processed",
		Waiting,
	}
	if i := pm.findExperiment(experiments, experimentFilename); i >= 0 {
		experiments[i] = newExperiment
	} else {
		return fmt.Errorf("Can't update experiment details for: %s",
			experimentFilename)
	}
	return pm.writeJson(experiments)
}

func (pm *ProgressMonitor) updateExperiment(
	experimentFilename string,
	status StatusKind,
	msg string,
) error {
	experiments, err := pm.GetExperiments()
	if err != nil {
		return err
	}
	if i := pm.findExperiment(experiments, experimentFilename); i >= 0 {
		experiments[i].Stamp = time.Now()
		experiments[i].Status = status
		experiments[i].Msg = msg
	} else {
		return fmt.Errorf("Can't update experiment with filename: %s",
			experimentFilename)
	}
	return pm.writeJson(experiments)
}

// Returns index of found experiment or -1 if not found
func (pm *ProgressMonitor) findExperiment(
	experiments []*Experiment,
	experimentFilename string,
) int {
	for i, experiment := range experiments {
		if experiment.ExperimentFilename == experimentFilename {
			return i
		}
	}
	return -1
}

func (pm *ProgressMonitor) writeJson(experiments []*Experiment) error {
	progress := &Progress{experiments}
	json, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pm.progressFilename, json, 0640)
}

// Implements sort.Interface for Progress
func (p Progress) Len() int { return len(p.Experiments) }
func (p Progress) Swap(i, j int) {
	p.Experiments[i], p.Experiments[j] =
		p.Experiments[j], p.Experiments[i]
}

func (p Progress) Less(i, j int) bool {
	return p.Experiments[i].Stamp.Before(p.Experiments[j].Stamp)
}
