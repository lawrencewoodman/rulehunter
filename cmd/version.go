// Copyright (C) 2016-2017 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vlifesystems/rulehunter/version"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Report version of Rulehunter",
	Long:  "Report version of Rulehunter",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Rulehunter v%s\n", version.Version())
	},
}
