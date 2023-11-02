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
)

// LogCmd output information to the file system
type LogCmd struct {
	Cmd   *cobra.Command
	Table string
}

func NewLogCmd() *LogCmd {
	c := &LogCmd{
		Cmd: &cobra.Command{
			Use:   "log",
			Short: "writes structured data to the file system",
			Long:  `writes structured data to the file system`,
			Example: `
# restart a new log
art log clear

# creates a new table called my-data with headers header1, header2 and header3
art log new my-data -h "header1|header2|header3"

# writes a record in my-data table
art log add my-data -v "value1|value2|value3"writes structured data to the file system

# add data from a yaml dictionary into the log
art log add my-data -f my-data.yaml

# writes the content of the log to stdout
art log print
`,
		},
	}
	return c
}
