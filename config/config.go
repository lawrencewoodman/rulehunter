// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package config handles the loading of a config file
package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type Config struct {
	ExperimentsDir    string `yaml:"experimentsDir"`
	WWWDir            string `yaml:"wwwDir"`
	BuildDir          string `yaml:"buildDir"`
	BaseURL           string `yaml:"baseUrl"`
	MaxNumReportRules int    `yaml:"maxNumReportRules"`
	MaxNumProcesses   int    `yaml:"maxNumProcesses"`
	MaxNumRecords     int    `yaml:"maxNumRecords"`
}

// InvalidExtError indicates that a config file has an invalid extension
type InvalidExtError string

func (e InvalidExtError) Error() string {
	return "invalid extension: " + string(e)
}

// Load the configuration file from filename
func Load(filename string) (*Config, error) {
	var c *Config
	var err error

	c, err = loadYAML(filename)
	if err != nil {
		return nil, err
	}

	if c.MaxNumReportRules < 1 {
		c.MaxNumReportRules = 20
	}

	if c.MaxNumProcesses < 1 {
		c.MaxNumProcesses = runtime.NumCPU()
	}

	if c.MaxNumRecords < 1 {
		c.MaxNumRecords = -1
	}

	if c.BaseURL == "" {
		c.BaseURL = "/"
	}

	if err := checkConfigValid(c); err != nil {
		return nil, err
	}
	return c, nil
}

func loadYAML(filename string) (*Config, error) {
	var c Config

	if ext := filepath.Ext(filename); ext != ".yaml" {
		return nil, InvalidExtError(ext)
	}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
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
