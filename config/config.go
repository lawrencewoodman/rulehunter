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

// Package config handles the loading of a config file
package config

import (
	"encoding/json"
	"errors"
	"os"
	"runtime"
)

type Config struct {
	ExperimentsDir   string
	WWWDir           string
	BuildDir         string
	SourceURL        string
	NumRulesInReport int
	MaxNumProcesses  int
}

// Load the configuration file from filename
func Load(filename string) (*Config, error) {
	var c Config
	var f *os.File
	var err error

	f, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err = dec.Decode(&c); err != nil {
		return nil, err
	}

	if c.MaxNumProcesses < 1 {
		c.MaxNumProcesses = runtime.NumCPU()
	}

	if c.NumRulesInReport < 1 {
		c.NumRulesInReport = 100
	}

	if c.SourceURL == "" {
		c.SourceURL = "https://github.com/vlifesystems/rulehuntersrv"
	}

	if err := checkConfigValid(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func checkConfigValid(c *Config) error {
	if len(c.ExperimentsDir) == 0 {
		return errors.New("missing field: experimentsDir")
	}
	if len(c.WWWDir) == 0 {
		return errors.New("missing field: wwwDir")
	}
	if len(c.BuildDir) == 0 {
		return errors.New("missing field: buildDir")
	}
	return nil
}
