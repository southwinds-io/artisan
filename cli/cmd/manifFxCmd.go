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
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/registry"
)

// ManifestFxCmd return package's manifest
type ManifestFxCmd struct {
	Cmd  *cobra.Command
	home string
}

func NewManifestFxCmd(artHome string) *ManifestFxCmd {
	c := &ManifestFxCmd{
		Cmd: &cobra.Command{
			Use:   "fx [flags] name:tag",
			Short: "returns the functions available in the package",
			Long:  ``,
			Example: `
art man fx mypackage:mytag
`,
		},
		home: artHome,
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *ManifestFxCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		core.RaiseErr("the package name:tag is required")
	} else if len(args) > 1 {
		core.RaiseErr("too many arguments")
	}
	// create a local registry
	local := registry.NewLocalRegistry(c.home)
	name, err := core.ParseName(args[0])
	i18n.Err("", err, i18n.ERR_INVALID_PACKAGE_NAME)
	// get the package manifest
	m := local.GetManifest(name)
	m.ListFxs(c.home)
}
