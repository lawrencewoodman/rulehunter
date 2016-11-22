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

package main

import (
	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"io/ioutil"
	"path/filepath"
	"time"
)

type program struct {
	config          *config.Config
	cmdFlags        cmdFlags
	progressMonitor *progress.ProgressMonitor
	logger          logger.Logger
	quit            <-chan struct{}
	shouldStop      bool
}

func newProgram(
	c *config.Config,
	p *progress.ProgressMonitor,
	l logger.Logger,
	q <-chan struct{},
) *program {
	return &program{
		config:          c,
		progressMonitor: p,
		logger:          l,
		quit:            q,
		shouldStop:      false,
	}
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

// Returns number of files processed and ifLoggedError
func (p *program) ProcessDir() (int, bool) {
	experimentFilenames, err := p.getExperimentFilenames()
	if err != nil {
		p.logger.Error(err.Error())
		return 0, true
	}
	for _, experimentFilename := range experimentFilenames {
		if err := p.progressMonitor.AddExperiment(experimentFilename); err != nil {
			p.logger.Error(err.Error())
			return 0, true
		}
	}

	numProcessed := 0
	ifLoggedError := false
	for _, experimentFilename := range experimentFilenames {
		numProcessed++
		err := experiment.Process(
			experimentFilename,
			p.config,
			p.logger,
			p.progressMonitor,
		)
		if err != nil {
			ifLoggedError = true
		}
	}
	return numProcessed, ifLoggedError
}

func (p *program) run() {
	logWaitingForExperiments := true

	for {
		select {
		case <-p.quit:
			return
		default:
			if p.shouldStop {
				return
			}
			if logWaitingForExperiments {
				logWaitingForExperiments = false
				p.logger.Info("Waiting for experiments to process")
			}

			if n, _ := p.ProcessDir(); n >= 1 {
				logWaitingForExperiments = true
			}

			// Sleeping prevents 'excessive' cpu use and disk access
			time.Sleep(2.0 * time.Second)
		}
	}
}

func (p *program) getExperimentFilenames() ([]string, error) {
	experimentFilenames := make([]string, 0)
	files, err := ioutil.ReadDir(p.config.ExperimentsDir)
	if err != nil {
		return experimentFilenames, err
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (ext == ".json" || ext == ".yaml") {
			experimentFilenames = append(experimentFilenames, file.Name())
		}
	}
	return experimentFilenames, nil
}

func (p *program) Stop(s service.Service) error {
	p.shouldStop = true
	return nil
}
