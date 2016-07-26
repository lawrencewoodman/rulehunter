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
	"flag"
)

type cmdFlags struct {
	user      string
	configDir string
	install   bool
	serve     bool
}

func parseFlags() *cmdFlags {
	flags := &cmdFlags{}

	flag.StringVar(
		&flags.user,
		"user",
		"",
		"The user to run the server as",
	)
	flag.StringVar(
		&flags.configDir,
		"configdir",
		"",
		"The configuration directory",
	)
	flag.BoolVar(
		&flags.install,
		"install",
		false,
		"Install the server as a service",
	)
	flag.BoolVar(
		&flags.serve,
		"serve",
		false,
		"Run the program as a local server",
	)
	flag.Parse()
	return flags
}

func handleFlags(flags *cmdFlags) error {
	if flags.install && flags.serve {
		return errInstallAndServeArg
	}
	if flags.install && flags.configDir == "" {
		return errNoConfigDirArg
	}
	return nil
}
