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

import (
	"encoding/json"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Monitor represents a progress monitor.
type Monitor struct {
	filename    string
	htmlCmds    chan<- cmd.Cmd
	experiments []*Experiment
}

type progressFile struct {
	Experiments []*Experiment
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
	htmlCmds chan<- cmd.Cmd,
) (*Monitor, error) {
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

	m := &Monitor{
		filename:    filename,
		htmlCmds:    htmlCmds,
		experiments: experiments,
	}
	for _, e := range experiments {
		e.monitor = m
	}
	sort.Sort(m)
	return m, nil
}

func (m *Monitor) AddExperiment(filename string) (*Experiment, error) {
	e := m.findExperiment(filename)
	if e == nil {
		newExperiment := &Experiment{
			Title:    "",
			Tags:     []string{},
			Stamp:    time.Now(),
			Filename: filename,
			Msg:      "Waiting to be processed",
			Percent:  0,
			Status:   Waiting,
			monitor:  m,
		}
		m.experiments = append(m.experiments, newExperiment)
		e = newExperiment
	} else {
		if isFinished(e) {
			return e, nil
		}
		e.Title = ""
		e.Tags = []string{}
		e.Stamp = time.Now()
		e.Msg = "Waiting to be processed"
		e.Percent = 0
		e.Status = Waiting
	}
	if err := m.writeJSON(); err != nil {
		return e, err
	}
	m.htmlCmds <- cmd.Progress
	return e, nil
}

func (m *Monitor) GetExperiments() []*Experiment {
	return m.experiments
}

// Returns experiment if found experiment or nil if not found
func (m *Monitor) findExperiment(filename string) *Experiment {
	for _, experiment := range m.experiments {
		if experiment.Filename == filename {
			return experiment
		}
	}
	return nil
}

func (m *Monitor) writeJSON() error {
	// File mode permission:
	// No special permission bits
	// User: Read, Write
	// Group: Read
	// Other: None
	const modePerm = 0640

	sort.Sort(m)
	successfulExperiments := []*Experiment{}
	for _, e := range m.experiments {
		if e.Status == Success {
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

// Implements sort.Interface
func (m *Monitor) Len() int {
	return len(m.experiments)
}
func (m *Monitor) Swap(i, j int) {
	m.experiments[i], m.experiments[j] =
		m.experiments[j], m.experiments[i]
}

func (m *Monitor) Less(i, j int) bool {
	return m.experiments[j].Stamp.Before(m.experiments[i].Stamp)
}
