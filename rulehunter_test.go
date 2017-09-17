package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func runOSCmd(t *testing.T, fatalError bool, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	var cmdOut bytes.Buffer
	var cmdErr bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	if err := cmd.Run(); err != nil {
		if fatalError {
			t.Fatalf("runOSCmd(%s, %v), err: %v, stdout: %s, stderr: %s",
				name, arg, err, cmdOut.String(), cmdErr.String())
		} else {
			t.Errorf("runOSCmd(%s, %v), err: %v, stdout: %v, stderr: %s",
				name, arg, err, cmdOut.String(), cmdErr.String())
		}
	}
}
