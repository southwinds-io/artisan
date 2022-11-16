/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"southwinds.dev/artisan/core"
)

type RootCmd struct {
	Cmd *cobra.Command
}

// NewRootCmd
// https://textkool.com/en/ascii-art-generator?hl=default&vl=default&font=Broadway%20KB&text=artisan%0A
func NewRootCmd() *RootCmd {
	c := &RootCmd{
		Cmd: &cobra.Command{
			Use:   "art",
			Short: "Artisan: the Onix DevOps CLI",
			Long: fmt.Sprintf(`
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++
|         __    ___  _____  _   __    __    _           |
|        / /\  | |_)  | |  | | ( ('  / /\  | |\ |       |
|       /_/--\ |_| \  |_|  |_| _)_) /_/--\ |_| \|       |
|       build, package, publish and run everywhere      |
+++++++++++++++++| automation manager |++++++++++++++++++

version: %s`, core.Version),
			Version: core.Version,
		},
	}
	c.Cmd.SetVersionTemplate("version: {{.Version}}\n")
	cobra.OnInitialize(c.initConfig)
	return c
}

func (c *RootCmd) initConfig() {
}
