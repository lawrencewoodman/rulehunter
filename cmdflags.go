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
	"flag"
)

type cmdFlags struct {
	configDir string
	install   bool
	serve     bool
}

func parseFlags(args []string) *cmdFlags {
	flags := &cmdFlags{}
	fs := flag.NewFlagSet("cmd", flag.ExitOnError)

	fs.StringVar(
		&flags.configDir,
		"configdir",
		"",
		"The configuration directory",
	)
	fs.BoolVar(
		&flags.install,
		"install",
		false,
		"Install the server as a service",
	)
	fs.BoolVar(
		&flags.serve,
		"serve",
		false,
		"Run the program as a local server",
	)
	fs.Parse(args)
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
