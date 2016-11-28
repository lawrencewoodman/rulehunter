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
	"github.com/vlifesystems/rulehunter/watcher"
	"time"
)

type program struct {
	config          *config.Config
	cmdFlags        cmdFlags
	progressMonitor *progress.ProgressMonitor
	logger          logger.Logger
	quit            <-chan struct{}
	filenames       chan string
	shouldStop      chan struct{}
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
		filenames:       make(chan string, 100),
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
		p.filenames,
	)
	go p.run()
	return nil
}

func (p *program) run() {
	for {
		select {
		case <-p.quit:
			return
		case <-p.shouldStop:
			return
		case filename := <-p.filenames:
			if err := p.progressMonitor.AddExperiment(filename); err != nil {
				p.logger.Error(err.Error())
			}
			err := experiment.Process(
				filename,
				p.config,
				p.logger,
				p.progressMonitor,
			)
			if err != nil {
				p.logger.Error(err.Error())
			}
		}
	}
}

func (p *program) Stop(s service.Service) error {
	close(p.shouldStop)
	return nil
}
