/*
	rulehuntersrv - A server to find rules in data based on user specified goals
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
	"fmt"
	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehuntersrv/config"
	"github.com/vlifesystems/rulehuntersrv/experiment"
	"github.com/vlifesystems/rulehuntersrv/logger"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/quitter"
	"io/ioutil"
	"path/filepath"
	"time"
)

type program struct {
	config          *config.Config
	cmdFlags        cmdFlags
	progressMonitor *progress.ProgressMonitor
	logger          logger.Logger
	quitter         *quitter.Quitter
}

func newProgram(
	c *config.Config,
	p *progress.ProgressMonitor,
	l logger.Logger,
	q *quitter.Quitter,
) *program {
	return &program{
		config:          c,
		progressMonitor: p,
		quitter:         q,
		logger:          l,
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
		err := experiment.Process(
			experimentFilename,
			p.config,
			p.logger,
			p.progressMonitor,
		)
		if err != nil {
			msg := fmt.Sprintf("Failed processing experiment: %s - %s",
				experimentFilename, err)
			p.logger.Error(msg)
			ifLoggedError = true
		}
	}
	return numProcessed, ifLoggedError
}

func (p *program) run() {
	sleepInSeconds := time.Duration(2)
	logWaitingForExperiments := true
	p.quitter.Add()
	defer p.quitter.Done()

	for !p.quitter.ShouldQuit() {
		if logWaitingForExperiments {
			logWaitingForExperiments = false
			p.logger.Info("Waiting for experiments to process")
		}

		if n, _ := p.ProcessDir(); n >= 1 {
			logWaitingForExperiments = true
		}

		// Sleeping prevents 'excessive' cpu use and disk access
		time.Sleep(sleepInSeconds * time.Second)
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
	p.quitter.Quit()
	return nil
}
