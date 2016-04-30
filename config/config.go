/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ExperimentsDir string
	WWWDir         string
	ProgressDir    string // Records the progress processing each experiment
	BuildDir       string
}

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
	// TODO: verify fields present and correct
	return &c, nil
}
