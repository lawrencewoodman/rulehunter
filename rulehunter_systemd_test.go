// +build systemd

package main

import "testing"

/*************************************
 *  Helper functions
 *************************************/

func startService(t *testing.T, serviceName string) {
	runOSCmd(t, true, "systemctl", "start", serviceName)
}

func stopService(t *testing.T, serviceName string) {
	runOSCmd(t, false, "systemctl", "stop", serviceName)
}
