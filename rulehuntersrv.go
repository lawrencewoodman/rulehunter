/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */
package main

import (
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"github.com/lawrencewoodman/rulehuntersrv/config"
	"github.com/lawrencewoodman/rulehuntersrv/experiment"
	"github.com/lawrencewoodman/rulehuntersrv/html"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger service.Logger

type program struct {
	configDir string
	config    *config.Config
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
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
				if err := experiment.Process(file.Name(), p.config); err != nil {
					logger.Errorf("Failed processing experiment: %s - %s", file.Name(), err)
				} else {
					logger.Infof("Successfully processed experiment: %s", file.Name())
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
	config, err := config.Load(configFilename)
	if err != nil {
		log.Fatal(fmt.Sprintf("Couldn't load configuration %s: %s",
			configFilename, err))
	}
	prg.config = config
	if err = html.GenerateReports(config); err != nil {
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
