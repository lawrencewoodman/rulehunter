package main

import (
	"fmt"
	"os"
	"syscall"
)

func interruptProcess() {
	pid := os.Getpid()
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		panic(fmt.Sprintf("LoadDLL: %v", e))
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		panic(fmt.Sprintf("FindProc: %v", e))
	}
	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		panic(fmt.Sprintf("GenerateConsoleCtrlEvent: %v", e))
	}
}
