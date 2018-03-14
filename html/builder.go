// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package html

import (
	"fmt"
	"sync"
	"time"

	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

// Builder represents an html website builder
type Builder struct {
	cfg       *config.Config
	pm        *progress.Monitor
	log       logger.Logger
	isRunning bool
	sync.Mutex
}

func New(
	cfg *config.Config,
	pm *progress.Monitor,
	l logger.Logger,
) *Builder {
	return &Builder{
		cfg:       cfg,
		pm:        pm,
		log:       l,
		isRunning: false,
	}
}

// This should be run as a goroutine and it will periodically generate html
func (b *Builder) Run(q *quitter.Quitter) {
	q.Add()
	defer q.Done()
	b.isRunning = true
	defer func() {
		b.Lock()
		defer b.Unlock()
		b.isRunning = false
	}()
	if err := b.generateAll(); err != nil {
		b.log.Error(fmt.Errorf("Couldn't generate html: %s", err))
	}

	oldExperiments := b.pm.GetExperiments()

	tickChan := time.NewTicker(time.Second * 5).C
	for {
		select {
		case <-tickChan:
			currentExperiments := b.pm.GetExperiments()
			if !areReportsUptodate(oldExperiments, currentExperiments) {
				if err := b.generateReports(); err != nil {
					b.log.Error(fmt.Errorf("Couldn't generate html: %s", err))
				}
			}
			if !isProgressUptodate(oldExperiments, currentExperiments) {
				if err := b.generateProgress(); err != nil {
					b.log.Error(fmt.Errorf("Couldn't generate html: %s", err))
				}
			}
			oldExperiments = currentExperiments
		case <-q.C:
			if err := b.generateAll(); err != nil {
				b.log.Error(fmt.Errorf("Couldn't generate html: %s", err))
			}
			return
		}
	}
}

// Running returns whether the html builder is running
func (b *Builder) Running() bool {
	b.Lock()
	defer b.Unlock()
	return b.isRunning
}

type generator func(*config.Config, *progress.Monitor) error

func (b *Builder) generateAll() error {
	generators := []generator{
		generateActivityPage,
		generateFront,
		generateReports,
		generateTagPages,
		generateCategoryPages,
	}
	return b.generate(generators)
}

func (b *Builder) generateProgress() error {
	generators := []generator{
		generateActivityPage,
		generateFront,
	}
	return b.generate(generators)
}

func (b *Builder) generateReports() error {
	generators := []generator{
		generateActivityPage,
		generateFront,
		generateReports,
		generateTagPages,
		generateCategoryPages,
	}
	return b.generate(generators)
}

func (b *Builder) generate(generators []generator) error {
	for _, g := range generators {
		if err := g(b.cfg, b.pm); err != nil {
			return err
		}
	}
	return nil
}

func areReportsUptodate(
	oldExperiments, currentExperiments []*progress.Experiment,
) bool {
	for _, ce := range currentExperiments {
		if ce.Status.State == progress.Success &&
			time.Since(ce.Status.Stamp).Seconds() < 10 {
			return false
		}
	}
	return true
}

func isProgressUptodate(
	oldExperiments, currentExperiments []*progress.Experiment,
) bool {
	if len(currentExperiments) != len(oldExperiments) {
		return false
	}
	for _, ce := range currentExperiments {
		for _, oe := range oldExperiments {
			if !ce.IsEqual(oe) {
				return false
			}
		}
	}
	return true
}
