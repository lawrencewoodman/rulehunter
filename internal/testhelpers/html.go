package testhelpers

import "github.com/vlifesystems/rulehunter/html/cmd"

type htmlCmdMonitor struct {
	htmlCmds     <-chan cmd.Cmd
	cmdsReceived []cmd.Cmd
}

func NewHtmlCmdMonitor(cmds <-chan cmd.Cmd) *htmlCmdMonitor {
	return &htmlCmdMonitor{cmds, []cmd.Cmd{}}
}

func (h *htmlCmdMonitor) Run() {
	for c := range h.htmlCmds {
		h.cmdsReceived = append(h.cmdsReceived, c)
	}
}

func (h *htmlCmdMonitor) GetCmdsReceived() []cmd.Cmd {
	return h.cmdsReceived
}
