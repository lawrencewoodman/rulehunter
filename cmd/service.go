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
	"github.com/spf13/cobra"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
)

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Control Rulehunter service",
	Long:  `Control Rulehunter operating system service.`,
}

var ServiceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Rulehunter as a service",
	Long:  `Install the Rulehunter server as an operating system service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.NewSvcLogger()
		return runInstallService(l, flagConfigFilename)
	},
}

var ServiceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Rulehunter service",
	Long:  `Uninstall the Rulehunter operating system service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.NewSvcLogger()
		return runUninstallService(l, flagConfigFilename)
	},
}

func init() {
	ServiceCmd.PersistentFlags().StringVar(
		&flagUser,
		"user",
		"",
		"The name of a user to run the service under",
	)
	ServiceCmd.AddCommand(ServiceInstallCmd)
	ServiceCmd.AddCommand(ServiceUninstallCmd)
}

func runInstallService(l logger.Logger, configFilename string) error {
	q := quitter.New()
	defer q.Quit()
	s, err := InitSetup(l, q, configFilename)
	if err != nil {
		return err
	}
	s.svc.Uninstall()
	if err := s.svc.Install(); err != nil {
		return err
	}
	return nil
}

func runUninstallService(l logger.Logger, configFilename string) error {
	q := quitter.New()
	defer q.Quit()
	s, err := InitSetup(l, q, configFilename)
	if err != nil {
		return err
	}
	return s.svc.Uninstall()
}
