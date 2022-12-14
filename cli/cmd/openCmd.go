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
	core.CheckErr(local.Open(name, c.credentials, path, nil, []string{}), "")
}
