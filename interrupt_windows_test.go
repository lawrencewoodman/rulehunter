// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.
//
// This is based on code from the os/signal tests

package main

import (
	"fmt"
	"os"
	"syscall"
)

func interruptProcess() {
	pid := os.Getpid()
	fmt.Printf("interruptProcess() pid: %d\n", os.Getpid())
	d, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		panic(fmt.Sprintf("LoadDLL: %v", err))
	}
	p, err := d.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		panic(fmt.Sprintf("FindProc: %v", err))
	}
	r, _, err := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		panic(fmt.Sprintf("GenerateConsoleCtrlEvent: %v", err))
	}
}
