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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yalp/jsonpath"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/registry"

	"os"
	"path"
)

// ManifestGetCmd return package's manifest
type ManifestGetCmd struct {
	Cmd    *cobra.Command
	home   string
	filter string
	format string
}

func NewManifestGetCmd(artHome string) *ManifestGetCmd {
	c := &ManifestGetCmd{
		Cmd: &cobra.Command{
			Use:   "get [flags] name:tag",
			Short: "returns the package manifest",
			Long:  ``,
			Example: `
art man get mypackage:mytag
`,
		},
		home: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.filter, "filter", "f", "", "--filter=JSONPath or -f=JSONPath")
	c.Cmd.Flags().StringVarP(&c.format, "format", "o", "json", "--format=mdf or -o=mdf\n"+
		"available formats are 'json' (in std output) or 'mdf' (creates a markdown file)\n")
	c.Cmd.Run = c.Run
	return c
}

func (c *ManifestGetCmd) Run(cmd *cobra.Command, args []string) {
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
	if c.format == "json" {
		// marshal the manifest
		bytes, err := json.MarshalIndent(m, "", "  ")
		core.CheckErr(err, "cannot marshal manifest")
		// if no filter is set then return the whole manifest
		if len(c.filter) == 0 {
			fmt.Printf("%v\n", string(bytes))
		} else {
			var jason interface{}
			err := json.Unmarshal(bytes, &jason)
			// otherwise apply the jsonpath to extract a value from the manifest
			result, err := jsonpath.Read(jason, c.filter)
			core.CheckErr(err, "cannot apply filter expression '%s'", c.filter)
			fmt.Printf("%v", result)
		}
	} else if c.format == "mdf" {
		bytes := m.ToMarkDownBytes(name.String())
		os.WriteFile(path.Join(core.WorkDir(), "manifest.md"), bytes, os.ModePerm)
	}
}
