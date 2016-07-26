// Package testhelpers contains routines to help test rulehuntersrv
package testhelpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func BuildConfigDirs(t *testing.T) string {
	// File mode permission:
	// No special permission bits
	// User: Read, Write Execute
	// Group: None
	// Other: None
	const modePerm = 0700

	tmpDir := TempDir(t)

	// TODO: Create the www/* and build/* subdirectories from rulehuntersrv code
	subDirs := []string{
		"experiments",
		"datasets",
		filepath.Join("www", "reports"),
		filepath.Join("www", "progress"),
		filepath.Join("build", "progress"),
		filepath.Join("build", "reports"),
	}
	for _, subDir := range subDirs {
		fullSubDir := filepath.Join(tmpDir, subDir)
		if err := os.MkdirAll(fullSubDir, modePerm); err != nil {
			t.Fatalf("MkDirAll(%s, ...) err: %v", fullSubDir, err)
		}
	}

	return tmpDir
}

func CopyFile(t *testing.T, srcFilename, dstDir string) {
	contents, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		t.Fatalf("ReadFile(%s) err: %v", srcFilename, err)
	}
	info, err := os.Stat(srcFilename)
	if err != nil {
		t.Fatalf("Stat(%s) err: %v", srcFilename, err)
	}
	mode := info.Mode()
	dstFilename := filepath.Join(dstDir, filepath.Base(srcFilename))
	if err := ioutil.WriteFile(dstFilename, contents, mode); err != nil {
		t.Fatalf("WriteFile(%s, ...) err: %v", dstFilename, err)
	}
}

func TempDir(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "rulehuntersrv_test")
	if err != nil {
		t.Fatalf("TempDir() err: %s", err)
	}
	return tempDir
}
