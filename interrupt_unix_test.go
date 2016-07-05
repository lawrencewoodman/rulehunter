package main

import (
	"fmt"
	"os"
)

func interruptProcess() {
	pid := os.Getpid()
	p, err := os.FindProcess(pid)
	if err != nil {
		panic("Can't find process to Quit")
	}
	if err := p.Signal(os.Interrupt); err != nil {
		panic(fmt.Sprintf("Can't send os.Interrupt signal: %s", err))
	}
}
