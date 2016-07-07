// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import (
	"os"
	"testing"
)

func interruptProcess(t *testing.T) {
	pid := os.Getpid()
	p, err := os.FindProcess(pid)
	if err != nil {
		t.Fatal("Can't find process to Quit")
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal("Can't send os.Interrupt signal: %s", err)
	}
}
