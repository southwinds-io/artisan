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

// PushCmd push a package to a remote registry
type PushCmd struct {
	Cmd         *cobra.Command
	home        string
	credentials string
}

func NewPushCmd(artHome string) *PushCmd {
	c := &PushCmd{
		Cmd: &cobra.Command{
			Use:   "push [FLAGS] NAME[:TAG]",
			Short: "uploads an package to a remote package store",
			Long:  ``,
		},
		home: artHome,
	}
	c.Cmd.Run = c.Run
	c.Cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD server user and password")
	return c
}

func (c *PushCmd) Run(cmd *cobra.Command, args []string) {
	// check an package name has been provided
	if len(args) == 0 {
		log.Fatal("name of the package to push is required")
	}
	// get the name of the package to push
	nameTag := args[0]
	// validate the name
	packageName, err := core.ParseName(nameTag)
	i18n.Err(c.home, err, i18n.ERR_INVALID_PACKAGE_NAME)
	// create a local registry
	local := registry.NewLocalRegistry(c.home)
	// attempt upload to remote repository
	core.CheckErr(local.Push(packageName, c.credentials, true), i18n.Sprintf(c.home, i18n.ERR_CANT_PUSH_PACKAGE))
}
