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
