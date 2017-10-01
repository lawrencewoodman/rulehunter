// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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
