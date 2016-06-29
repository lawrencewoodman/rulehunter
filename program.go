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
	"github.com/vlifesystems/rulehuntersrv/progress"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type program struct {
	config          *config.Config
	cmdFlags        cmdFlags
	progressMonitor *progress.ProgressMonitor
	logger          service.Logger
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	sleepInSeconds := time.Duration(2)
	logWaitingForExperiments := true

	for {
		if logWaitingForExperiments {
			logWaitingForExperiments = false
			p.logger.Infof("Waiting for experiments to process")
		}
		experimentFilenames, err := p.getExperimentFilenames()
		if err != nil {
			p.logger.Error(err)
		}
		for _, experimentFilename := range experimentFilenames {
			err := p.progressMonitor.AddExperiment(experimentFilename)
			if err != nil {
				p.logger.Error(err)
			}
		}

		for _, experimentFilename := range experimentFilenames {
			logWaitingForExperiments = true
			p.logger.Infof("Processing experiment: %s", experimentFilename)

			err := experiment.Process(
				experimentFilename,
				p.config,
				p.progressMonitor,
			)
			if err != nil {
				p.logger.Errorf("Failed processing experiment: %s - %s",
					experimentFilename, err)
			} else {
				p.logger.Infof("Successfully processed experiment: %s",
					experimentFilename)
			}
			if err := p.moveExperimentToProcessed(experimentFilename); err != nil {
				fullErr := fmt.Errorf("Couldn't move experiment file: %s", err)
				p.logger.Error(fullErr)
			}
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
		if !file.IsDir() {
			experimentFilenames = append(experimentFilenames, file.Name())
		}
	}
	return experimentFilenames, nil
}

func (p *program) moveExperimentToProcessed(experimentFilename string) error {
	experimentFullFilename :=
		filepath.Join(p.config.ExperimentsDir, experimentFilename)
	experimentProcessedFullFilename :=
		filepath.Join(p.config.ExperimentsDir, "processed", experimentFilename)
	return os.Rename(experimentFullFilename, experimentProcessedFullFilename)
}

func (p *program) Stop(s service.Service) error {
	// TODO: Put code here to stop processing cleanly
	return nil
}
