// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package watcher is used to find experiment files that need processing
package watcher

import (
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
	"io/ioutil"
	"path/filepath"
	"time"
)

// DirError indicates that there was an error reading the directory
type DirError string

func (e DirError) Error() string {
	return "can not watch directory: " + string(e)
}

// Watch sends experiment filenames that need processing
// to the filenames channel.  It checks every period of time.
func Watch(
	dir string,
	period time.Duration,
	l logger.Logger,
	quit *quitter.Quitter,
	files chan<- fileinfo.FileInfo,
) {
	var lastLogErr error
	quit.Add()
	defer quit.Done()
	ticker := time.NewTicker(period).C
	allFiles, err := getFilesToMap(dir)
	if err != nil {
		lastLogErr = l.Error(err)
	}

	for _, file := range allFiles {
		files <- file
	}

	// Used to only send old files every other run
	flipFlop := true
	for {
		select {
		case <-quit.C:
			close(files)
			return
		case <-ticker:
			newFiles, err := getFilesToMap(dir)
			if err != nil {
				if lastLogErr == nil || lastLogErr.Error() != err.Error() {
					lastLogErr = l.Error(err)
				}
				break
			}
			lastLogErr = nil

			for filename, file := range newFiles {
				if lastFile, ok := allFiles[filename]; ok {
					isNew := !fileinfo.IsEqual(lastFile, file)
					if isNew {
						files <- file
					} else if flipFlop {
						files <- file
					}
				} else {
					files <- file
				}
				allFiles[filename] = file
			}
			allFiles = newFiles
			flipFlop = !flipFlop
		}
	}
}

func GetExperimentFiles(dir string) ([]fileinfo.FileInfo, error) {
	experimentFiles := make([]fileinfo.FileInfo, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return experimentFiles, DirError(dir)
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (ext == ".json" || ext == ".yaml") {
			experimentFiles = append(experimentFiles, file)
		}
	}
	return experimentFiles, nil
}

func getFilesToMap(dir string) (map[string]fileinfo.FileInfo, error) {
	filesMap := make(map[string]fileinfo.FileInfo)
	files, err := GetExperimentFiles(dir)
	if err != nil {
		return filesMap, err
	}
	for _, file := range files {
		filesMap[file.Name()] = file
	}
	return filesMap, nil
}
