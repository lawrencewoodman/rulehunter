/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger service.Logger

type program struct {
	configDir string
	config    *config
}

type config struct {
	ExperimentsDir string
	ReportsDir     string
	ProgressDir    string // Records the progress processing each experiment
	BuildDir       string
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func loadConfig(filename string) (*config, error) {
	var c config
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
	return &c, nil
}

func (p *program) run() {
	var experimentFiles []os.FileInfo
	var err error
	sleepInSeconds := time.Duration(2)
	logWaitingForExperiments := true

	for {
		if logWaitingForExperiments {
			logWaitingForExperiments = false
			logger.Infof("Waiting for experiments to process")
		}
		experimentFiles, err = ioutil.ReadDir(p.config.ExperimentsDir)
		if err != nil {
			logger.Error(err)
		}

		for _, file := range experimentFiles {
			if !file.IsDir() {
				logWaitingForExperiments = true
				logger.Infof("Processing experiment: %s", file.Name())
				err := processExperiment(file.Name(), p.config)
				if err == nil {
					logger.Infof("Successfully processed experiment: %s", file.Name())
					err := moveExperimentToSuccess(file.Name(), p.config)
					if err != nil {
						logger.Errorf("Couldn't move experiment file: %s", err)
					}
				} else {
					logger.Errorf("Failed processing experiment: %s - %s",
						file.Name(), err)
					err := moveExperimentToFail(file.Name(), p.config)
					if err != nil {
						logger.Errorf("Couldn't move experiment file: %s", err)
					}
				}
			}
		}

		// Sleeping prevents 'excessive' cpu use and disk access
		time.Sleep(sleepInSeconds * time.Second)
	}
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "GoTestService",
		DisplayName: "Go Test Service",
		Description: "A test Go service.",
	}
	prg := &program{}

	userPtr := flag.String("user", "", "The user to run the server as")
	configDirPtr := flag.String("configdir", "", "The configuration directory")
	installPtr := flag.Bool("install", false, "Install the server as a service")
	flag.Parse()

	if *userPtr != "" {
		svcConfig.UserName = *userPtr
	}

	if *configDirPtr != "" {
		svcConfig.Arguments = []string{fmt.Sprintf("-configdir=%s", *configDirPtr)}
		prg.configDir = *configDirPtr
	}

	configFilename := filepath.Join(prg.configDir, "config.json")
	config, err := loadConfig(configFilename)
	if err != nil {
		log.Fatal(fmt.Sprintf("Couldn't load configuration %s: %s",
			configFilename, err))
	}
	prg.config = config
	if err = writeIndexHTML(config.BuildDir, config.ReportsDir); err != nil {
		log.Fatal(err)
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if *installPtr {
		if *configDirPtr == "" {
			log.Fatal("No -configdir argument")
		}
		err = s.Install()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
	}
}
