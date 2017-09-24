package main

import "testing"

// Users that are known on a Windows system to try using
var knownUsers = []string{"", "administrator"}

func TestSubMain_interrupt(t *testing.T) {
	t.Skip("This test isn't implemented for Windows yet")
}

/*************************************
 *  Helper functions
 *************************************/

func startService(t *testing.T, serviceName string) {
	runOSCmd(t, true, "net", "start", serviceName)
}

func stopService(t *testing.T, serviceName string) {
	runOSCmd(t, false, "net", "stop", serviceName)
}
