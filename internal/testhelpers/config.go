// Package testhelpers contains routines to help test rulehuntersrv
package testhelpers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func BuildConfigDirs() (string, error) {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700

	tmpDir, err := ioutil.TempDir("", "rulehuntersrv")
	if err != nil {
		return "", errors.New("TempDir() couldn't create dir")
	}

	// TODO: Create the www/* and build/* subdirectories from rulehuntersrv code
	subDirs := []string{
		"experiments",
		filepath.Join("www", "reports"),
		filepath.Join("www", "progress"),
		filepath.Join("build", "reports")}
	for _, subDir := range subDirs {
		fullSubDir := filepath.Join(tmpDir, subDir)
		if err := os.MkdirAll(fullSubDir, modePerm); err != nil {
			return "", fmt.Errorf("can't make directory: %s", subDir)
		}
	}

	err = copyFile(filepath.Join("fixtures", "config.json"), tmpDir)
	return tmpDir, err
}

func copyFile(srcFilename, dstDir string) error {
	contents, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcFilename)
	if err != nil {
		return err
	}
	mode := info.Mode()
	dstFilename := filepath.Join(dstDir, filepath.Base(srcFilename))
	return ioutil.WriteFile(dstFilename, contents, mode)
}
