/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

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

package cmd

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/vlifesystems/rulehunter/config"
	"github.com/vlifesystems/rulehunter/html"
	"github.com/vlifesystems/rulehunter/html/cmd"
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
	configDir string,
) (*Setup, error) {
	configFilename := filepath.Join(configDir, "config.yaml")
	config, err := config.Load(configFilename)
	if err != nil {
		return nil, errConfigLoad{filename: configFilename, err: err}
	}

	if err := buildConfigDirs(config); err != nil {
		return nil, err
	}

	htmlCmds := make(chan cmd.Cmd, 100)
	pm, err := progress.NewMonitor(
		filepath.Join(config.BuildDir, "progress"),
		htmlCmds,
	)
	if err != nil {
		return nil, err
	}
	prg := program.New(config, pm, l, q)

	s, err := newService(prg, flagUser, configDir)
	if err != nil {
		return nil, err
	}
	svcLogger, err := s.Logger(nil)
	if err != nil {
		return nil, err
	}

	l.SetSvcLogger(svcLogger)
	go l.Run(q)
	go html.Run(config, pm, l, q, htmlCmds)
	return &Setup{
		prg: prg,
		svc: s,
		cfg: config,
	}, nil
}

func newService(
	prg *program.Program,
	user string,
	configDir string,
) (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "rulehunter",
		DisplayName: "Rulehunter server",
		Description: "Rulehunter finds rules in data based on user defined goals.",
	}

	if user != "" {
		svcConfig.UserName = user
	}

	svcConfig.Arguments = os.Args[1:]
	if len(svcConfig.Arguments) >= 1 && svcConfig.Arguments[0] == "service" {
		svcConfig.Arguments[0] = "serve"
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
		filepath.Join(cfg.BuildDir, "descriptions"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, modePerm); err != nil {
			return err
		}
	}

	return nil
}
