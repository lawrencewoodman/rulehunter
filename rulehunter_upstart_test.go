// +build upstart

package main

import "testing"

/*************************************
 *  Helper functions
 *************************************/

func startService(t *testing.T, serviceName string) {
	runOSCmd(t, true, "service", serviceName, "start")
}

func stopService(t *testing.T, serviceName string) {
	runOSCmd(t, false, "service", serviceName, "stop")
}
