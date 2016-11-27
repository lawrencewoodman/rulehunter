// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestSubMain_interrupt(t *testing.T) {
	wantExitCode := 0

	configDir := testhelpers.BuildConfigDirs(t)
	flags := &cmdFlags{install: false, serve: true, configDir: configDir}
	defer os.RemoveAll(configDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "config.yaml"), configDir)

	l := testhelpers.NewLogger()
	go func() {
		time.Sleep(1 * time.Second)
		interruptProcess(t)
	}()

	go func() {
		<-time.After(6 * time.Second)
		t.Fatal("Run() hasn't been stopped")
	}()

	exitCode, err := subMain(flags, l)
	if exitCode != wantExitCode {
		t.Errorf("subMain(%v) exitCode: %d, want: %d",
			flags, exitCode, wantExitCode)
	}
	if err != nil {
		t.Errorf("subMain(%v): %s", flags, err)
	}
}
