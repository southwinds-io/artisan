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
	"strings"
)

// LogNewCmd creates a new log file
type LogNewCmd struct {
	Cmd     *cobra.Command
	headers string
	artHome string
}

func NewLogNewCmd(artHome string) *LogNewCmd {
	c := &LogNewCmd{
		artHome: artHome,
		Cmd: &cobra.Command{
			Use:   "new [table-name] [flags]",
			Short: "creates a new log table",
			Long:  `creates a new log table`,
			Example: `
# creates a new table within the log called my-data with headers: header1, header2 and header3
art new my-data -h "header1|header2|header3"

NOTE: if a table is already defined it will be deleted and a warning issued
`,
		},
	}
	c.Cmd.Flags().StringVarP(&c.headers, "headers", "h", "action", "-h h1|h2|h3...; a pipe separated list of table headers")
	c.Cmd.Run = c.Run
	return c
}

func (c *LogNewCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		core.RaiseErr("invalid arguments, expecting table name; got %d arguments instead", len(args))
	}
	l, err := core.NewLog(c.artHome)
	core.CheckErr(err, "cannot create logger")
	core.CheckErr(l.New(args[0], strings.Split(c.headers, "|")), "cannot create new table in log")
}
