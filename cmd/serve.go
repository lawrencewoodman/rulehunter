// Copyright (C) 2016-2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run Rulehunter as a server",
	Long: `Rulehunter will run as a server continually monitoring an 'experiments'
         directory and processing its contents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.NewSvcLogger()
		q := quitter.New()
		defer q.Quit()
		return runServe(l, q, flagConfigFilename)
	},
}

func runServe(
	l logger.Logger,
	q *quitter.Quitter,
	configFilename string,
) error {
	s, err := InitSetup(l, q, configFilename)
	if err != nil {
		return err
	}
	go httpServer(s.cfg.HTTPPort, s.cfg.WWWDir, q, l)
	if err := s.svc.Run(); err != nil {
		return err
	}
	return nil
}
