/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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

// ManifestCmd return package's manifest
type ManifestCmd struct {
	Cmd    *cobra.Command
	home   string
	filter string
	format string
}

func NewManifestCmd(artHome string) *ManifestCmd {
	c := &ManifestCmd{
		Cmd: &cobra.Command{
			Use:   "manifest [flags] name:tag",
			Short: "returns the package manifest",
			Long:  ``,
		},
		home: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.filter, "filter", "f", "", "--filter=JSONPath or -f=JSONPath")
	c.Cmd.Flags().StringVarP(&c.format, "format", "o", "json", "--format=mdf or -o=mdf\n"+
		"available formats are 'json' (in std output) or 'mdf' (creates a markdown file)\n")
	c.Cmd.Run = c.Run
	return c
}

func (c *ManifestCmd) Run(cmd *cobra.Command, args []string) {
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
