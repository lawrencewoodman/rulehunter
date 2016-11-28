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
	"github.com/vlifesystems/rulehunter/logger"
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

// Watch sends experiment filenames that need processing
// to the filenames channel.  It checks every period of time.
func Watch(
	dir string,
	period time.Duration,
	l logger.Logger,
	quit <-chan struct{},
	filenames chan<- string,
) {
	lastLogMsg := ""
	ticker := time.NewTicker(period).C
	allFileStats, err := getFileStats(dir)
	if err != nil {
		lastLogMsg = err.Error()
		l.Error(err.Error())
	}

	for filename, _ := range allFileStats {
		filenames <- filename
	}

	for {
		select {
		case <-quit:
			close(filenames)
			return
		case <-ticker:
			newFileStats, err := getFileStats(dir)
			if err != nil {
				if lastLogMsg != err.Error() {
					lastLogMsg = err.Error()
					l.Error(err.Error())
				}
				break
			}
			lastLogMsg = ""

			for filename, fileStats := range newFileStats {
				if lastFileStats, ok := allFileStats[filename]; ok {
					if !lastFileStats.isEqual(fileStats) {
						allFileStats[filename] = fileStats
						filenames <- filename
					}
				} else {
					allFileStats[filename] = fileStats
					filenames <- filename
				}
			}
			allFileStats = newFileStats
		}
	}
}

func GetExperimentFilenames(dir string) ([]string, error) {
	experimentFilenames := make([]string, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (ext == ".json" || ext == ".yaml") {
			experimentFilenames = append(experimentFilenames, file.Name())
		}
	}
	return experimentFilenames, nil
}

func getFileStats(dir string) (map[string]*FileStats, error) {
	fileStats := make(map[string]*FileStats)
	filenames, err := GetExperimentFilenames(dir)
	if err != nil {
		return fileStats, err
	}
	for _, filename := range filenames {
		fs, err := newFileStats(filepath.Join(dir, filename))
		if err != nil {
			return fileStats, err
		}
		fileStats[filename] = fs
	}
	return fileStats, nil
}

func newFileStats(filename string) (*FileStats, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	return &FileStats{
		size:    fi.Size(),
		mode:    fi.Mode(),
		modTime: fi.ModTime(),
	}, nil
}

func (fs *FileStats) isEqual(otherFileStats *FileStats) bool {
	return fs.size == otherFileStats.size &&
		fs.mode == otherFileStats.mode &&
		fs.modTime == otherFileStats.modTime
}
