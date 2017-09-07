package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func runOSCmd(t *testing.T, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("runOSCmd(%s, %v), err: %v, out: %v", name, arg, err, out.String())
	}
}
