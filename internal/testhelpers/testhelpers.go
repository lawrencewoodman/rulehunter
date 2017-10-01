// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

// Package testhelpers contains routines to help test rulehunter
package testhelpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/config"
	"gopkg.in/yaml.v2"
)

type errorReporter interface {
	Fatalf(format string, args ...interface{})
}

func MustWriteConfig(e errorReporter, baseDir string, maxNumRecords int) {
	const mode = 0600
	cfg := &config.Config{
		ExperimentsDir: filepath.Join(baseDir, "experiments"),
		WWWDir:         filepath.Join(baseDir, "www"),
		BuildDir:       filepath.Join(baseDir, "build"),
		MaxNumRecords:  maxNumRecords,
	}
	cfgFilename := filepath.Join(baseDir, "config.yaml")
	y, err := yaml.Marshal(cfg)
	if err != nil {
		e.Fatalf("Marshal: %s", err)
	}
	if err := ioutil.WriteFile(cfgFilename, y, mode); err != nil {
		e.Fatalf("WriteFile(%s, ...) err: %s", cfgFilename, err)
	}
}

func BuildConfigDirs(e errorReporter, buildAllDirs bool) string {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700
	var subDirs []string

	tmpDir := TempDir(e)

	if buildAllDirs {
		subDirs = []string{
			"experiments",
			"datasets",
			filepath.Join("www", "reports"),
			filepath.Join("build", "progress"),
			filepath.Join("build", "reports"),
			filepath.Join("build", "descriptions"),
		}
	} else {
		subDirs = []string{
			"experiments",
			"datasets",
		}
	}
	for _, subDir := range subDirs {
		fullSubDir := filepath.Join(tmpDir, subDir)
		if err := os.MkdirAll(fullSubDir, modePerm); err != nil {
			e.Fatalf("MkDirAll(%s, ...) err: %v", fullSubDir, err)
		}
	}

	return tmpDir
}

func CopyFile(e errorReporter, srcFilename, dstDir string, args ...string) {
	contents, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		e.Fatalf("ReadFile(%s) err: %v", srcFilename, err)
	}
	info, err := os.Stat(srcFilename)
	if err != nil {
		e.Fatalf("Stat(%s) err: %v", srcFilename, err)
	}
	mode := info.Mode()
	dstFilename := filepath.Join(dstDir, filepath.Base(srcFilename))
	if len(args) == 1 {
		dstFilename = filepath.Join(dstDir, args[0])
	}
	if err := ioutil.WriteFile(dstFilename, contents, mode); err != nil {
		e.Fatalf("WriteFile(%s, ...) err: %v", dstFilename, err)
	}
}

func TempDir(e errorReporter) string {
	tmpDir, err := ioutil.TempDir("", "rulehunter_test")
	if err != nil {
		e.Fatalf("TempDir() err: %s", err)
	}
	return tmpDir
}

func MustParse(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func GetFilesInDir(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatalf("ioutil.ReadDir(%s) err: %s", dir, err)
	}

	r := []string{}
	for _, file := range files {
		if !file.IsDir() {
			r = append(r, file.Name())
		}
	}
	return r
}
