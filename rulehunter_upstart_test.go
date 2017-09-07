// +build upstart

package main

import "testing"

/*************************************
 *  Helper functions
 *************************************/

func startService(t *testing.T, serviceName string) {
	runOSCmd(t, "service", serviceName, "start")
}

func stopService(t *testing.T, serviceName string) {
	runOSCmd(t, "service", serviceName, "stop")
}
