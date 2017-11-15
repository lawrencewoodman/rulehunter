// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package progress

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/report"
)

// Monitor represents an experiment progress monitor.
type Monitor struct {
	filename    string
	htmlCmds    chan<- cmd.Cmd
	experiments map[string]*Experiment
	sync.Mutex
}

type progressFile struct {
	Experiments []*Experiment `json:"experiments"`
}

type ExperimentNotFoundError struct {
	filename string
}

func (e ExperimentNotFoundError) Error() string {
	return fmt.Sprintf("progress for experiment file not found: %s", e.filename)
}

func NewMonitor(
	progressDir string,
	htmlCmds chan<- cmd.Cmd,
) (*Monitor, error) {
	var progress progressFile
	experiments := map[string]*Experiment{}
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
		for _, e := range progress.Experiments {
			experiments[e.Filename] = e
		}
	}

	return &Monitor{
		filename:    filename,
		htmlCmds:    htmlCmds,
		experiments: experiments,
	}, nil
}

func (m *Monitor) AddExperiment(
	filename string,
	title string,
	tags []string,
	category string,
) error {
	m.Lock()
	m.experiments[filename] = newExperiment(filename, title, tags, category)
	m.Unlock()
	m.htmlCmds <- cmd.Progress
	return nil
}

// ReportProgress reports a message and percent progress (0.0-1.0) for
// an experiment
func (m *Monitor) ReportProgress(
	file string,
	mode report.ModeKind,
	msg string,
	percent float64,
) error {
	e := m.getExperiment(file)
	if e == nil {
		return ExperimentNotFoundError{file}
	}
	e.Status.SetProgress(mode, msg, percent)
	if err := m.writeJSON(); err != nil {
		return err
	}
	m.htmlCmds <- cmd.Progress
	return nil
}

func (m *Monitor) getExperiment(file string) *Experiment {
	m.Lock()
	defer m.Unlock()
	e, ok := m.experiments[file]
	if !ok {
		return nil
	}
	return e
}

// ReportLoadError reports that an experiment failed to load
func (m *Monitor) ReportLoadError(file string, err error) error {
	m.AddExperiment(file, "", []string{}, "")
	fullErr := fmt.Errorf("Error loading experiment: %s", err)
	return m.ReportError(file, fullErr)
}

// ReportError sets experiment to having failed with an error
func (m *Monitor) ReportError(file string, err error) error {
	e := m.getExperiment(file)
	if e == nil {
		return ExperimentNotFoundError{file}
	}
	e.Status.SetError(err)
	if err := m.writeJSON(); err != nil {
		return err
	}
	m.htmlCmds <- cmd.Progress
	m.htmlCmds <- cmd.Reports
	return nil
}

// ReportSuccess reports experiment has been successful
func (m *Monitor) ReportSuccess(file string) error {
	e := m.getExperiment(file)
	if e == nil {
		return ExperimentNotFoundError{file}
	}
	e.Status.SetSuccess()
	if err := m.writeJSON(); err != nil {
		return err
	}
	m.htmlCmds <- cmd.Progress
	m.htmlCmds <- cmd.Reports
	return nil
}

// GetFinishStamp returns whether a file has finished and its last update
// time stamp.  If the file isn't known then it will return false and the
// current time.
func (m *Monitor) GetFinishStamp(file string) (bool, time.Time) {
	m.Lock()
	defer m.Unlock()
	e, ok := m.experiments[file]
	if !ok {
		return false, time.Now()
	}
	return e.Status.IsFinished(), e.Status.Stamp
}

func (m *Monitor) GetExperiments() []*Experiment {
	m.Lock()
	defer m.Unlock()
	experiments := make([]*Experiment, len(m.experiments))
	i := 0
	for f, e := range m.experiments {
		experiments[i] = &Experiment{
			Filename: f,
			Title:    e.Title,
			Tags:     e.Tags,
			Category: e.Category,
			Status:   e.Status,
		}
		i++
	}
	sort.Slice(experiments, func(i, j int) bool {
		return experiments[j].Status.Stamp.Before(experiments[i].Status.Stamp)
	})
	return experiments
}

func (m *Monitor) writeJSON() error {
	// File mode permission:
	// No special permission bits
	// User: Read, Write
	// Group: Read
	// Other: None
	const modePerm = 0640

	experiments := m.GetExperiments()
	successfulExperiments := []*Experiment{}
	for _, e := range experiments {
		if e.Status.State == Success {
			successfulExperiments = append(successfulExperiments, e)
		}
	}
	progress := &progressFile{successfulExperiments}
	json, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.filename, json, modePerm)
}
