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

// PullCmd pull a package from a remote registry
type PullCmd struct {
	Cmd         *cobra.Command
	artHome     string
	credentials string
	path        string
}

func NewPullCmd(artHome string) *PullCmd {
	c := &PullCmd{
		Cmd: &cobra.Command{
			Use:   "pull [FLAGS] NAME[:TAG]",
			Short: "downloads an package from the package registry",
			Long:  ``,
		},
		artHome: artHome,
	}
	c.Cmd.Run = c.Run
	c.Cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD server user and password")
	return c
}

func (c *PullCmd) Run(cmd *cobra.Command, args []string) {
	// check a package name has been provided
	if len(args) == 0 {
		log.Fatal("name of the package to pull is required")
	}
	// get the name of the package to push
	nameTag := args[0]
	// validate the name
	packageName, err := core.ParseName(nameTag)
	i18n.Err(c.artHome, err, i18n.ERR_INVALID_PACKAGE_NAME)
	// create a local registry
	local := registry.NewLocalRegistry(c.artHome)
	// attempt pull from remote registry
	_, err = local.Pull(packageName, c.credentials, true)
	core.CheckErr(err, "cannot pull package '%s'", packageName)
}
