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
	"southwinds.dev/artisan/registry"
)

type TagCmd struct {
	Cmd  *cobra.Command
	home string
}

func NewTagCmd(artHome string) *TagCmd {
	c := &TagCmd{
		Cmd: &cobra.Command{
			Use:     "tag",
			Short:   "add a tag to an existing package",
			Long:    `create a tag TARGET_PACKAGE that refers to SOURCE_PACKAGE`,
			Example: `art tag SOURCE_PACKAGE[:TAG] TARGET_PACKAGE[:TAG]`,
		},
		home: artHome,
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *TagCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		core.RaiseErr("source and target package tags are required")
	}
	l := registry.NewLocalRegistry(c.home)
	core.CheckErr(l.Tag(args[0], args[1]), "cannot tag package")
}
