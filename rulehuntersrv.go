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
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"github.com/lawrencewoodman/rulehuntersrv/config"
	"github.com/lawrencewoodman/rulehuntersrv/experiment"
	"github.com/lawrencewoodman/rulehuntersrv/html"
	"github.com/lawrencewoodman/rulehuntersrv/progress"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger service.Logger

type program struct {
	configDir       string
	config          *config.Config
	progressMonitor *progress.ProgressMonitor
}

func (p *program) Start(s service.Service) error {
	go p.startReportGenerator()
	go p.run()
	return nil
}

func (p *program) run() {
	sleepInSeconds := time.Duration(2)
	logWaitingForExperiments := true

	for {
		if logWaitingForExperiments {
			logWaitingForExperiments = false
			logger.Infof("Waiting for experiments to process")
		}
		experimentFilenames, err := p.getExperimentFilenames()
		if err != nil {
			logger.Error(err)
		}
		for _, experimentFilename := range experimentFilenames {
			err := p.progressMonitor.AddExperiment(experimentFilename)
			if err != nil {
				logger.Error(err)
			}
		}

		for _, experimentFilename := range experimentFilenames {
			logWaitingForExperiments = true
			logger.Infof("Processing experiment: %s", experimentFilename)

			err := experiment.Process(
				experimentFilename,
				p.config,
				p.progressMonitor,
			)
			if err != nil {
				logger.Errorf("Failed processing experiment: %s - %s",
					experimentFilename, err)
				err := p.moveExperimentToFail(experimentFilename)
				if err != nil {
					fullErr := fmt.Errorf("%s (Couldn't move experiment file: %s)", err)
					logger.Error(fullErr)
				}
			} else {
				err := p.moveExperimentToSuccess(experimentFilename)
				if err != nil {
					fullErr := fmt.Errorf("Couldn't move experiment file: %s", err)
					logger.Error(fullErr)
				} else {
					logger.Infof("Successfully processed experiment: %s",
						experimentFilename)
				}
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

func (p *program) moveExperimentToSuccess(experimentFilename string) error {
	experimentFullFilename :=
		filepath.Join(p.config.ExperimentsDir, experimentFilename)
	experimentSuccessFullFilename :=
		filepath.Join(p.config.ExperimentsDir, "success", experimentFilename)
	return os.Rename(experimentFullFilename, experimentSuccessFullFilename)
}

func (p *program) moveExperimentToFail(experimentFilename string) error {
	experimentFullFilename :=
		filepath.Join(p.config.ExperimentsDir, experimentFilename)
	experimentFailFullFilename :=
		filepath.Join(p.config.ExperimentsDir, "fail", experimentFilename)
	return os.Rename(experimentFullFilename, experimentFailFullFilename)
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) startReportGenerator() {
	sleepInSeconds := time.Duration(4)
	for {
		if err := html.Generate(p.config, p.progressMonitor); err != nil {
			fullErr := fmt.Errorf("Couldn't generate report: %s", err)
			// TODO: Work out if this is thread safe
			logger.Error(fullErr)
		}
		// Sleeping prevents 'excessive' cpu use and disk access
		time.Sleep(sleepInSeconds * time.Second)
	}
}

func main() {
	svcConfig := &service.Config{
		Name:        "GoTestService",
		DisplayName: "Go Test Service",
		Description: "A test Go service.",
	}
	prg := &program{}

	userPtr := flag.String("user", "", "The user to run the server as")
	configDirPtr := flag.String("configdir", "", "The configuration directory")
	installPtr := flag.Bool("install", false, "Install the server as a service")
	flag.Parse()

	if *userPtr != "" {
		svcConfig.UserName = *userPtr
	}

	if *configDirPtr != "" {
		svcConfig.Arguments = []string{fmt.Sprintf("-configdir=%s", *configDirPtr)}
		prg.configDir = *configDirPtr
	}

	configFilename := filepath.Join(prg.configDir, "config.json")
	config, err := config.Load(configFilename)
	if err != nil {
		log.Fatal(fmt.Sprintf("Couldn't load configuration %s: %s",
			configFilename, err))
	}
	prg.config = config

	prg.progressMonitor =
		progress.NewMonitor(filepath.Join(config.BuildDir, "progress"))

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if *installPtr {
		if *configDirPtr == "" {
			log.Fatal("No -configdir argument")
		}
		err = s.Install()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
	}
}
