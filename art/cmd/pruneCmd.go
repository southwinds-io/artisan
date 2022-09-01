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

// PruneCmd remove all dangling packages
type PruneCmd struct {
	Cmd *cobra.Command
}

func NewPruneCmd() *PruneCmd {
	c := &PruneCmd{
		Cmd: &cobra.Command{
			Use:   "prune",
			Short: "remove all dangling packages",
			Long:  `remove all dangling packages`,
		},
	}
	c.Cmd.Run = c.Run
	return c
}

func (b *PruneCmd) Run(cmd *cobra.Command, args []string) {
	local := registry.NewLocalRegistry("")
	core.CheckErr(local.Prune(), "")
}
