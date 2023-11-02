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

// LogAddCmd adds a record to a log table
type LogAddCmd struct {
	Cmd     *cobra.Command
	values  string
	artHome string
}

func NewLogAddCmd(artHome string) *LogAddCmd {
	c := &LogAddCmd{
		artHome: artHome,
		Cmd: &cobra.Command{
			Use:   "add [table-name] [flags]",
			Short: "adds a record to a log table",
			Long:  `adds a record to a log table`,
			Example: `
# adds a new record to my-data table
art add my-data -h "value1|value2|value3"

NOTE: if the number of values is greater than the number of headers then the surplus values are discarded and a warning is issued
`,
		},
	}
	c.Cmd.Flags().StringVarP(&c.values, "values", "v", "", "-v v1|v2|v3...; a pipe separated list of table record values")
	c.Cmd.Run = c.Run
	return c
}

func (c *LogAddCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		core.RaiseErr("invalid arguments, expecting table name; got %d arguments instead", len(args))
	}
	l, err := core.NewLog(c.artHome)
	core.CheckErr(err, "cannot create logger")
	core.CheckErr(l.New(args[0], strings.Split(c.values, "|")), "cannot add row in log table %s", args[0])
}
