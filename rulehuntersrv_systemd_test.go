// +build systemd

package main

import "testing"

/*************************************
 *  Helper functions
 *************************************/

func startService(t *testing.T, serviceName string) {
	runCmd(t, "systemctl", "start", serviceName)
}

func stopService(t *testing.T, serviceName string) {
	runCmd(t, "systemctl", "stop", serviceName)
}
