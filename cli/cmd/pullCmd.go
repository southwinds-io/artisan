/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
	home        string
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
		home: artHome,
	}
	c.Cmd.Run = c.Run
	c.Cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD server user and password")
	return c
}

func (c *PullCmd) Run(cmd *cobra.Command, args []string) {
	// check an package name has been provided
	if len(args) == 0 {
		log.Fatal("name of the package to pull is required")
	}
	// get the name of the package to push
	nameTag := args[0]
	// validate the name
	packageName, err := core.ParseName(nameTag)
	i18n.Err("", err, i18n.ERR_INVALID_PACKAGE_NAME)
	// create a local registry
	local := registry.NewLocalRegistry(c.home)
	// attempt pull from remote registry
	local.Pull(packageName, c.credentials, true)
}
