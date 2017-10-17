// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package program

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/experiment"
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/watcher"
)

type Program struct {
	config          *config.Config
	progressMonitor *progress.Monitor
	logger          logger.Logger
	quit            *quitter.Quitter
	files           chan fileinfo.FileInfo
	shouldStop      chan struct{}
}

func New(
	c *config.Config,
	p *progress.Monitor,
	l logger.Logger,
	q *quitter.Quitter,
) *Program {
	return &Program{
		config:          c,
		progressMonitor: p,
		logger:          l,
		quit:            q,
		files:           make(chan fileinfo.FileInfo, 100),
		shouldStop:      make(chan struct{}),
	}
}

func (p *Program) Start(s service.Service) error {
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
func (p *Program) ProcessFile(file fileinfo.FileInfo) error {
	var err error
	pm := p.progressMonitor
	stamp := time.Now()

	e, err := experiment.Load(p.config, file)
	if err != nil {
		logErr := fmt.Errorf("Can't load experiment: %s, %s", file.Name(), err)
		p.logger.Error(logErr)
		if pmErr := pm.ReportLoadError(file.Name(), err); pmErr != nil {
			return p.logger.Error(pmErr)
		}
		return nil
	}

	isFinished, stamp := pm.GetFinishStamp(file.Name())

	if !isFinished {
		pmErr := pm.AddExperiment(file.Name(), e.Title, e.Tags, e.Category)
		if pmErr != nil {
			return p.logger.Error(pmErr)
		}
	}

	ok, err := e.ShouldProcess(isFinished, stamp)
	if err != nil {
		logErr :=
			fmt.Errorf("Error processing experiment: %s, %s", file.Name(), err)
		p.logger.Error(logErr)
		if pmErr := pm.ReportError(file.Name(), err); pmErr != nil {
			return p.logger.Error(pmErr)
		}
		return nil
	}
	if !ok {
		return nil
	}

	p.logger.Info("Processing experiment: " + file.Name())
	if err := e.Process(p.config, p.progressMonitor); err != nil {
		logErr :=
			fmt.Errorf("Error processing experiment: %s, %s", file.Name(), err)
		p.logger.Error(logErr)
		if pmErr := pm.ReportError(file.Name(), err); pmErr != nil {
			return p.logger.Error(pmErr)
		}
		return nil
	}

	logInfo := "Successfully processed experiment: " + file.Name()
	p.logger.Info(logInfo)
	if pmErr := pm.ReportSuccess(file.Name()); pmErr != nil {
		return p.logger.Error(pmErr)
	}
	return nil
}

func (p *Program) ProcessDir(dir string) error {
	files, err := watcher.GetExperimentFiles(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := p.ProcessFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (p *Program) ProcessFilename(filename string) error {
	file, err := os.Stat(filename)
	if err != nil {
		if pErr, ok := err.(*os.PathError); ok {
			err = pErr.Err
		}
		filename = path.Base(filename)
		logErr := fmt.Errorf("Can't load experiment: %s, %s", filename, err)
		p.logger.Error(logErr)
		pmErr := p.progressMonitor.ReportLoadError(filename, err)
		if pmErr != nil {
			return p.logger.Error(pmErr)
		}
		return nil
	}
	if err := p.ProcessFile(file); err != nil {
		return err
	}
	return nil
}

func (p *Program) run() {
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

func (p *Program) Stop(s service.Service) error {
	close(p.shouldStop)
	return nil
}
