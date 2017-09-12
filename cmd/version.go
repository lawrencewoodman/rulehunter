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