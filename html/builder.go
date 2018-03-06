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

	tickChan := time.NewTicker(time.Second * 5).C
	for {
		select {
		case <-tickChan:
			// TODO: Alter this so that it checks whether progress has changed and
			// generates accordingly
			if err := b.generateReports(); err != nil {
				b.log.Error(fmt.Errorf("Couldn't generate html: %s", err))
			}
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
