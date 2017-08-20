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
	"fmt"
	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/watcher"
	"time"
)

type program struct {
	config          *config.Config
	cmdFlags        cmdFlags
	progressMonitor *progress.Monitor
	logger          logger.Logger
	quit            *quitter.Quitter
	files           chan fileinfo.FileInfo
	shouldStop      chan struct{}
}

func newProgram(
	c *config.Config,
	p *progress.Monitor,
	l logger.Logger,
	q *quitter.Quitter,
) *program {
	return &program{
		config:          c,
		progressMonitor: p,
		logger:          l,
		quit:            q,
		files:           make(chan fileinfo.FileInfo, 100),
		shouldStop:      make(chan struct{}),
	}
}

func (p *program) Start(s service.Service) error {
	watchPeriod := 2.0 * time.Second
	go watcher.Watch(
		p.config.ExperimentsDir,
		watchPeriod,
		p.logger,
		p.quit,
		p.files,
	)
	go p.run()
	return nil
}

// ProcessFile tries to process an Experiment file.  It only returns an
// error if it is out of the ordinary for example if an error occurs when
// reporting to the progress monitor, not if it can't load an experiment
// nor if there is a problem processing the experiment.
func (p *program) ProcessFile(file fileinfo.FileInfo) error {
	var err error
	pm := p.progressMonitor
	stamp := time.Now()

	e, err := experiment.Load(p.config, file)
	if err != nil {
		logErr := fmt.Sprintf("Can't load experiment: %s, %s", file.Name(), err)
		p.logger.Error(logErr)
		if pmErr := pm.ReportLoadFailure(file.Name(), err); pmErr != nil {
			p.logger.Error(pmErr.Error())
			return pmErr
		}
		return nil
	}

	if pmErr := pm.AddExperiment(file.Name(), e.Title, e.Tags); pmErr != nil {
		p.logger.Error(pmErr.Error())
		return pmErr
	}

	isFinished, stamp := pm.GetFinishStamp(file.Name())

	ok, err := e.ShouldProcess(isFinished, stamp)
	if err != nil {
		logErr :=
			fmt.Sprintf("Failed processing experiment: %s, %s", file.Name(), err)
		p.logger.Error(logErr)
		if pmErr := pm.ReportFailure(file.Name(), err); pmErr != nil {
			p.logger.Error(pmErr.Error())
			return pmErr
		}
		return nil
	}
	if !ok {
		return nil
	}

	p.logger.Info("Processing experiment: " + file.Name())
	if err := e.Process(p.config, p.progressMonitor); err != nil {
		logErr :=
			fmt.Sprintf("Failed processing experiment: %s, %s", file.Name(), err)
		p.logger.Error(logErr)
		if pmErr := pm.ReportFailure(file.Name(), err); pmErr != nil {
			p.logger.Error(pmErr.Error())
			return pmErr
		}
		return nil
	}

	logInfo := "Successfully processed experiment: " + file.Name()
	p.logger.Info(logInfo)
	if pmErr := pm.ReportSuccess(file.Name()); pmErr != nil {
		p.logger.Error(pmErr.Error())
		return pmErr
	}
	return nil
}

func (p *program) run() {
	for {
		select {
		case <-p.quit.C:
			return
		case <-p.shouldStop:
			return
		case file := <-p.files:
			if file == nil {
				break
			}
			p.ProcessFile(file)
		}
	}
}

func (p *program) Stop(s service.Service) error {
	close(p.shouldStop)
	return nil
}
