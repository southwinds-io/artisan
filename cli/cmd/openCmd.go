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
	"log"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/registry"
)

// OpenCmd opens a package in the specified path
type OpenCmd struct {
	cmd         *cobra.Command
	home        string
	credentials string
}

func NewOpenCmd(artHome string) *OpenCmd {
	c := &OpenCmd{
		cmd: &cobra.Command{
			Use:   "open NAME[:TAG] [path]",
			Short: "opens an package in the specified path",
			Long:  ``,
		},
		home: artHome,
	}
	c.cmd.Run = c.Run
	c.cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD server user and password")
	return c
}

func (c *OpenCmd) Run(cmd *cobra.Command, args []string) {
	// check a package name has been provided
	if len(args) < 1 {
		log.Fatal("name of the package to open is required")
	}
	// get the name of the package to push
	nameTag := args[0]
	path := ""
	if len(args) == 2 {
		path = args[1]
	}
	// validate the name
	name, err := core.ParseName(nameTag)
	i18n.Err("", err, i18n.ERR_INVALID_PACKAGE_NAME)
	// create a local registry
	local := registry.NewLocalRegistry(c.home)
	// attempt to open from local registry
	core.CheckErr(local.Open(name, c.credentials, path, nil, nil, []string{}), "")
}
