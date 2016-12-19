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
	configDir := testhelpers.BuildConfigDirs(t)
	flags := &cmdFlags{install: false, serve: true, configDir: configDir}
	defer os.RemoveAll(configDir)
	testhelpers.CopyFile(t, filepath.Join("fixtures", "config.yaml"), configDir)

	l := testhelpers.NewLogger()
	hasQuitC := make(chan bool)
	go func() {
		wantExitCode := 0
		exitCode, err := subMain(flags, l)
		if exitCode != wantExitCode {
			t.Errorf("subMain(%v) exitCode: %d, want: %d",
				flags, exitCode, wantExitCode)
		}
		if err != nil {
			t.Errorf("subMain(%v): %s", flags, err)
		}
		hasQuitC <- true
	}()
	interruptC := time.NewTimer(time.Second).C
	timeoutC := time.NewTimer(6 * time.Second).C
	for {
		select {
		case <-interruptC:
			interruptProcess(t)
		case <-timeoutC:
			t.Fatal("subMain() hasn't stopped")
		case <-hasQuitC:
			return
		}
	}

}
