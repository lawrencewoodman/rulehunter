/*
	rulehunter - A server to find rules in data based on user specified goals
	Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program; see the file COPYING.  If not, see
	<http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
)

var RootCmd = &cobra.Command{
	Use:   "rulehunter",
	Short: "Rulehunter finds rules in data based on user defined goals",
	Long: `Rulehunter finds rules in data based on user defined goals.
                Complete documentation is available at http://rulehunter.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.NewSvcLogger()
		return runRoot(l, flagConfigDir)
	},
}

var (
	flagUser      string
	flagConfigDir string
)

func init() {
	RootCmd.PersistentFlags().StringVar(
		&flagConfigDir,
		"configdir",
		".",
		"config file directory (default: .)",
	)
	RootCmd.AddCommand(ServeCmd)
	RootCmd.AddCommand(ServiceCmd)
}

func runRoot(l logger.Logger, configDir string) error {
	q := quitter.New()
	defer q.Quit()
	s, err := InitSetup(l, q, configDir)
	if err != nil {
		return err
	}
	if err := s.prg.ProcessDir(s.cfg.ExperimentsDir); err != nil {
		return fmt.Errorf("Errors while processing dir")
	}
	return nil
}
