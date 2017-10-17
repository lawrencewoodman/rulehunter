// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

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
		return runRoot(l, flagConfigFilename, flagFile)
	},
}

// The contents of the flags specified on the command line
var (
	flagFile           string
	flagUser           string
	flagConfigFilename string
)

func init() {
	RootCmd.PersistentFlags().StringVar(
		&flagConfigFilename,
		"config",
		"config.yaml",
		"config file",
	)
	RootCmd.Flags().StringVar(
		&flagFile,
		"file",
		"",
		"an experiment file to process",
	)
	RootCmd.AddCommand(ServeCmd)
	RootCmd.AddCommand(ServiceCmd)
	RootCmd.AddCommand(VersionCmd)
}

func runRoot(
	l logger.Logger,
	configFilename string,
	experimentFilename string,
) error {
	q := quitter.New()
	defer q.Quit()
	s, err := InitSetup(l, q, configFilename)
	if err != nil {
		return err
	}
	if experimentFilename != "" {
		if err := s.prg.ProcessFilename(experimentFilename); err != nil {
			return fmt.Errorf("Errors while processing file: %s", experimentFilename)
		}
	} else {
		if err := s.prg.ProcessDir(s.cfg.ExperimentsDir); err != nil {
			return fmt.Errorf("Errors while processing dir")
		}
	}
	return nil
}
