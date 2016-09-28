// +build upstart

package main

import "testing"

/*************************************
 *  Helper functions
 *************************************/

func startService(t *testing.T, serviceName string) {
	runCmd(t, "service", "start", serviceName)
}

func stopService(t *testing.T, serviceName string) {
	runCmd(t, "service", "stop", serviceName)
}
