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

// Package watcher is used to find experiment files that need processing
package watcher

import (
	"github.com/vlifesystems/rulehunter/fileinfo"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type FileStats struct {
	size    int64
	mode    os.FileMode
	modTime time.Time
}

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
	quit.Add()
	defer quit.Done()
	lastLogMsg := ""
	ticker := time.NewTicker(period).C
	allFiles, err := getFilesToMap(dir)
	if err != nil {
		lastLogMsg = err.Error()
		l.Error(err.Error())
	}

	for _, file := range allFiles {
		files <- file
	}

	// Used to only send old files every other run
	flipFlop := false
	for {
		select {
		case <-quit.C:
			close(files)
			return
		case <-ticker:
			newFiles, err := getFilesToMap(dir)
			if err != nil {
				if lastLogMsg != err.Error() {
					lastLogMsg = err.Error()
					l.Error(err.Error())
				}
				break
			}
			lastLogMsg = ""

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
