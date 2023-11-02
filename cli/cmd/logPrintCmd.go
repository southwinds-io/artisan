/*
   Artisan Core - Automation Manager
   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cmd

import (
	"github.com/spf13/cobra"
	"southwinds.dev/artisan/core"
)

// LogPrintCmd collects variables required by a flow
type LogPrintCmd struct {
	Cmd     *cobra.Command
	artHome string
}

func NewLogPrintCmd(artHome string) *LogPrintCmd {
	c := &LogPrintCmd{
		artHome: artHome,
		Cmd: &cobra.Command{
			Use:   "print",
			Short: "prints the content of the current log to the std output in json format",
			Long:  `prints the content of the current log to the std output in json format`,
			Example: `
# writes the content of the log to stdout
art log print
`,
		},
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *LogPrintCmd) Run(cmd *cobra.Command, args []string) {
	l, err := core.NewLog(c.artHome)
	core.CheckErr(err, "cannot create logger")
	core.CheckErr(l.Print(), "cannot print log")
}
