/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
