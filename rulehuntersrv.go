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
	"github.com/vlifesystems/rulehuntersrv/html"
	"github.com/vlifesystems/rulehuntersrv/html/cmd"
	"github.com/vlifesystems/rulehuntersrv/logger"
	"github.com/vlifesystems/rulehuntersrv/progress"
	"github.com/vlifesystems/rulehuntersrv/quitter"
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

	q := quitter.New()
	defer q.Quit()

	configFilename := filepath.Join(flags.configDir, "config.yaml")
	config, err := config.Load(configFilename)
	if err != nil {
		return 1, errConfigLoad{filename: configFilename, err: err}
	}

	htmlCmds := make(chan cmd.Cmd)
	pm, err := progress.NewMonitor(
		filepath.Join(config.BuildDir, "progress"),
		htmlCmds,
	)
	if err != nil {
		return 1, err
	}
	prg := newProgram(config, pm, l, q)

	s, err := newService(prg, config, flags)
	if err != nil {
		return 1, err
	}
	svcLogger, err := s.Logger(nil)
	if err != nil {
		return 1, err
	}

	l.SetSvcLogger(svcLogger)
	go l.Run(q)
	go html.Run(config, prg.progressMonitor, l, q, htmlCmds)

	if flags.install {
		if err = s.Install(); err != nil {
			return 1, err
		}
	} else if flags.serve {
		if err = s.Run(); err != nil {
			return 1, err
		}
	} else {
		if _, ifLoggedError := prg.ProcessDir(); ifLoggedError {
			return 1, fmt.Errorf("Errors while processing dir")
		}
	}
	return 0, nil
}

func newService(
	prg *program,
	cfg *config.Config,
	flags *cmdFlags,
) (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "RulehunterSrv",
		DisplayName: "Rulehunter server",
		Description: "Finds rules in data based on user specified goals.",
	}

	if cfg.User != "" {
		svcConfig.UserName = cfg.User
	}

	if flags.configDir != "" {
		svcConfig.Arguments =
			[]string{fmt.Sprintf("-configdir=%s", flags.configDir)}
	}
	s, err := service.New(prg, svcConfig)
	return s, err
}
