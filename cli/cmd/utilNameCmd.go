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
	"southwinds.dev/artisan/core"
)

// UtilNameCmd generates passwords
type UtilNameCmd struct {
	Cmd          *cobra.Command
	number       *int
	specialChars *bool
}

func NewUtilNameCmd() *UtilNameCmd {
	c := &UtilNameCmd{
		Cmd: &cobra.Command{
			Use:   "name [flags]",
			Short: "generates a random name",
			Long:  `generates a random name`,
		},
	}
	c.number = c.Cmd.Flags().IntP("max-number", "n", 0, "adds a random number at the end of the name ranging from 0 to max-number")
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilNameCmd) Run(_ *cobra.Command, _ []string) {
	fmt.Printf("%s", core.RandomName(*c.number))
}
