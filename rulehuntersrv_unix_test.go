// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/vlifesystems/rulehuntersrv/internal/testhelpers"
	"github.com/vlifesystems/rulehuntersrv/logger"
)

func TestSubMain_interrupt(t *testing.T) {
	cases := []struct {
		flags        *cmdFlags
		wantErr      error
		wantExitCode int
		wantEntries  []logger.Entry
	}{
		{
			flags: &cmdFlags{
				user:    "fred",
				install: false,
				serve:   true,
			},
			wantErr:      nil,
			wantExitCode: 0,
			wantEntries: []logger.Entry{
				{logger.Info, "Waiting for experiments to process"},
			},
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() err: ", err)
	}
	defer os.Chdir(wd)

	for _, c := range cases {
		configDir := testhelpers.BuildConfigDirs(t)
		defer os.RemoveAll(configDir)
		testhelpers.CopyFile(t, filepath.Join("fixtures", "config.json"), configDir)
		c.flags.configDir = configDir

		l := testhelpers.NewLogger()
		go func() {
			tryInSeconds := 4
			for i := 0; i < tryInSeconds*5; i++ {
				if reflect.DeepEqual(l.GetEntries(), c.wantEntries) {
					interruptProcess(t)
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
			interruptProcess(t)
		}()

		go func() {
			<-time.After(6 * time.Second)
			t.Fatal("Run() hasn't been stopped")
		}()
		if err := os.Chdir(configDir); err != nil {
			t.Fatalf("Chdir() err: %s", err)
		}
		exitCode, err := subMain(c.flags, l)
		if exitCode != c.wantExitCode {
			t.Errorf("subMain(%q) exitCode: %d, want: %d",
				c.flags, exitCode, c.wantExitCode)
		}
		if err := checkErrorMatch(err, c.wantErr); err != nil {
			t.Errorf("subMain(%q) %s", c.flags, err)
		}
		if !reflect.DeepEqual(l.GetEntries(), c.wantEntries) {
			t.Errorf("GetEntries() got: %s, want: %s", l.GetEntries(), c.wantEntries)
		}
	}
}
