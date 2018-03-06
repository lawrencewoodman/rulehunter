// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/html"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/program"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
)

type Setup struct {
	prg *program.Program
	svc service.Service
	cfg *config.Config
}

func InitSetup(
	l logger.Logger,
	q *quitter.Quitter,
	configFilename string,
) (*Setup, error) {
	config, err := config.Load(configFilename)
	if err != nil {
		return nil, errConfigLoad{filename: configFilename, err: err}
	}

	if err := buildConfigDirs(config); err != nil {
		return nil, err
	}

	pm, err := progress.NewMonitor(
		filepath.Join(config.BuildDir, "progress"),
	)
	if err != nil {
		return nil, err
	}
	prg := program.New(config, pm, l, q)

	s, err := newService(prg, flagUser, configFilename)
	if err != nil {
		return nil, err
	}
	svcLogger, err := s.Logger(nil)
	if err != nil {
		return nil, err
	}

	l.SetSvcLogger(svcLogger)
	h := html.New(config, pm, l)
	go l.Run(q)
	go h.Run(q)

	for i := 0; i < 10; i++ {
		// This helps to ensure that quitter is properly established
		// in goroutine before returning and being able to call Quit on quitter
		if l.Running() && h.Running() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return &Setup{
		prg: prg,
		svc: s,
		cfg: config,
	}, nil
}

func newService(
	prg *program.Program,
	user string,
	configFilename string,
) (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "rulehunter",
		DisplayName: "Rulehunter server",
		Description: "Rulehunter finds rules in data based on user defined goals.",
	}

	if user != "" {
		svcConfig.UserName = user
	}

	if len(os.Args) >= 2 && os.Args[1] == "service" {
		svcConfig.Arguments = []string{
			"serve",
			fmt.Sprintf("--config=%s", configFilename),
		}
	} else {
		svcConfig.Arguments = os.Args[1:]
	}
	return service.New(prg, svcConfig)
}

func buildConfigDirs(cfg *config.Config) error {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700

	dirs := []string{
		filepath.Join(cfg.WWWDir, "reports"),
		filepath.Join(cfg.WWWDir, "progress"),
		filepath.Join(cfg.BuildDir, "progress"),
		filepath.Join(cfg.BuildDir, "reports"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, modePerm); err != nil {
			return err
		}
	}

	return nil
}
