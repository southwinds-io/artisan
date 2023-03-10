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
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// UtilStampCmd generates a timestamp
type UtilStampCmd struct {
	Cmd  *cobra.Command
	path string
}

func NewUtilStampCmd() *UtilStampCmd {
	c := &UtilStampCmd{
		Cmd: &cobra.Command{
			Use:   "stamp [flags]",
			Short: "prints the current timestamp in UTC Unix Nano format",
			Long:  `prints the current timestamp in UTC Unix Nano format`,
			Args:  cobra.ExactArgs(0),
		},
	}
	c.Cmd.Run = c.Run
	c.Cmd.Flags().StringVarP(&c.path, "file-path", "p", "", "if set, writes the timestamp to the file system path")
	return c
}

func (c *UtilStampCmd) Run(_ *cobra.Command, args []string) {
	if len(c.path) > 0 {
		path, _ := filepath.Abs(c.path)
		os.WriteFile(path, []byte(strconv.FormatInt(time.Now().UTC().UnixNano(), 10)), 0755)
	} else {
		fmt.Printf("%d", time.Now().UTC().UnixNano())
	}
}
