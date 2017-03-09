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
	"github.com/vlifesystems/rulehunter/html"
	"github.com/vlifesystems/rulehunter/html/cmd"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/progress"
	"github.com/vlifesystems/rulehunter/quitter"
	"github.com/vlifesystems/rulehunter/watcher"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flags := parseFlags()
	l := logger.NewSvcLogger()
	exitCode, err := subMain(flags, l)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(exitCode)
}

// subMain is split from main to improve testing. quitter is used to stop
// the routine.
func subMain(
	flags *cmdFlags,
	l logger.Logger,
) (int, error) {
	if err := handleFlags(flags); err != nil {
		return 1, err
	}

	quit := quitter.New()
	defer quit.Quit()

	configFilename := filepath.Join(flags.configDir, "config.yaml")
	config, err := config.Load(configFilename)
	if err != nil {
		return 1, errConfigLoad{filename: configFilename, err: err}
	}

	if err := buildConfigDirs(config); err != nil {
		return 1, err
	}

	htmlCmds := make(chan cmd.Cmd, 100)
	pm, err := progress.NewMonitor(
		filepath.Join(config.BuildDir, "progress"),
		htmlCmds,
	)
	if err != nil {
		return 1, err
	}
	prg := newProgram(config, pm, l, quit)

	s, err := newService(prg, config, flags)
	if err != nil {
		return 1, err
	}
	svcLogger, err := s.Logger(nil)
	if err != nil {
		return 1, err
	}

	l.SetSvcLogger(svcLogger)
	go l.Run(quit)
	go html.Run(config, prg.progressMonitor, l, quit, htmlCmds)

	if flags.install {
		s.Uninstall()
		if err = s.Install(); err != nil {
			return 1, err
		}
	} else if flags.serve {
		if err = s.Run(); err != nil {
			return 1, err
		}
	} else {
		if err := processDir(prg, config); err != nil {
			return 1, fmt.Errorf("Errors while processing dir")
		}
	}
	return 0, nil
}

func processDir(p *program, config *config.Config) error {
	files, err := watcher.GetExperimentFiles(config.ExperimentsDir)
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

func newService(
	prg *program,
	cfg *config.Config,
	flags *cmdFlags,
) (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "rulehunter",
		DisplayName: "Rulehunter server",
		Description: "Rulehunter finds rules in data based on user specified goals.",
	}

	if cfg.User != "" {
		svcConfig.UserName = cfg.User
	}

	if flags.configDir != "" {
		svcConfig.Arguments =
			[]string{fmt.Sprintf("-configdir=%s", flags.configDir), "-serve"}
	}
	s, err := service.New(prg, svcConfig)
	return s, err
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
