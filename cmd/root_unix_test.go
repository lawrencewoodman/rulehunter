// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/vlifesystems/rulehunter/internal/testhelpers"
)

func TestRunRoot_interrupt(t *testing.T) {
	cfgDir := testhelpers.BuildConfigDirs(t, false)
	defer os.RemoveAll(cfgDir)
	testhelpers.MustWriteConfig(t, cfgDir, 100)

	l := testhelpers.NewLogger()
	hasQuitC := make(chan bool)
	go func() {
		if err := runRoot(l, cfgDir); err != nil {
			t.Errorf("runRoot: %s", err)
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
			t.Fatal("runRoot hasn't stopped")
		case <-hasQuitC:
			return
		}
	}
}
