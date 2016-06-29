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
	"github.com/vlifesystems/rulehuntersrv/progress"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flags := parseFlags()
	exitCode, err := subMain(flags)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(exitCode)
}

func subMain(flags *cmdFlags) (int, error) {
	prg := &program{}

	if err := handleFlags(flags); err != nil {
		return 1, err
	}

	svcConfig := makeServiceConfig(flags)
	configFilename := filepath.Join(flags.configDir, "config.json")
	config, err := config.Load(configFilename)
	if err != nil {
		return 1, errConfigLoad{
			filename: configFilename,
			err:      err,
		}
	}
	prg.config = config

	htmlCmds := make(chan cmd.Cmd)
	prg.progressMonitor =
		progress.NewMonitor(filepath.Join(config.BuildDir, "progress"), htmlCmds)

	s, err := service.New(prg, svcConfig)
	if err != nil {
		return 1, err
	}
	prg.logger, err = s.Logger(nil)
	if err != nil {
		return 1, err
	}

	go html.Run(config, prg.progressMonitor, prg.logger, htmlCmds)
	htmlCmds <- cmd.All

	if flags.install {
		if err = s.Install(); err != nil {
			return 1, err
		}
	} else {
		if err = s.Run(); err != nil {
			return 1, err
		}
	}
	return 0, nil
}

func makeServiceConfig(flags *cmdFlags) *service.Config {
	svcConfig := &service.Config{
		Name:        "GoTestService",
		DisplayName: "Go Test Service",
		Description: "A test Go service.",
	}

	if flags.user != "" {
		svcConfig.UserName = flags.user
	}

	if flags.configDir != "" {
		svcConfig.Arguments =
			[]string{fmt.Sprintf("-configdir=%s", flags.configDir)}
	}
	return svcConfig
}
