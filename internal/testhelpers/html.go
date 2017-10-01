// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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
	for {
		select {
		case c, ok := <-h.htmlCmds:
			if ok {
				h.cmdsReceived = append(h.cmdsReceived, c)
			} else {
				return h.cmdsReceived
			}
		default:
			return h.cmdsReceived
		}
	}
}
