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
	"gopkg.in/yaml.v2"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/registry"

	"os"
	"strings"
)

// EnvPackageCmd work out the variables required by a given package to run
type EnvPackageCmd struct {
	Cmd           *cobra.Command
	buildFilePath string
	stdout        *bool
	out           string
	artHome       string
}

func NewEnvPackageCmd(artHome string) *EnvPackageCmd {
	c := &EnvPackageCmd{
		Cmd: &cobra.Command{
			Use: "package [flags] [package name] [function-name (optional)]",
			Short: "return the variables required by a given package to run\n " +
				"if a function name is not specified then variables for all functions are retrieved",
			Long: "return the variables required by a given package to run\n " +
				"if a function name is not specified then variables for all functions are retrieved",
		},
		artHome: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.buildFilePath, "build-file-path", "b", "", "--build-file-path=. or -b=.; the path to an artisan build.yaml file from which to pick required inputs")
	c.Cmd.Flags().StringVarP(&c.out, "output", "o", "env", "--output yaml or -o yaml; the output format (e.g. env, json, yaml)")
	c.stdout = c.Cmd.Flags().Bool("stdout", false, "prints the output to the console")
	c.Cmd.Run = c.Run
	return c
}

func (c *EnvPackageCmd) Run(cmd *cobra.Command, args []string) {
	var input *data.Input
	if len(args) > 0 && len(args) < 3 {
		name, err := core.ParseName(args[0])
		core.CheckErr(err, "invalid package name: %s", name)
		local := registry.NewLocalRegistry(c.artHome)
		manifest := local.GetManifest(name)
		if len(args) == 2 {
			fxName := args[1]
			fx := manifest.Fx(fxName)
			if fx == nil {
				core.RaiseErr(fmt.Sprintf("function %s either does not exist or it has not been exported", fxName))
			}
			input = fx.Input
		} else {
			if len(manifest.Functions) == 0 {
				core.RaiseErr(fmt.Sprintf("no functions found in package manifest"))
			}
			for i, function := range manifest.Functions {
				if i == 0 {
					input = function.Input
				} else {
					input.Merge(function.Input)
				}
			}
		}
		// add the credentials to download the package
		input.SurveyRegistryCreds(name.Group, name.Name, "", name.Domain, false, true, merge.NewEnVarFromSlice([]string{}))
	} else if len(args) < 2 {
		i18n.Raise("", i18n.ERR_INSUFFICIENT_ARGS)
	} else if len(args) > 2 {
		i18n.Raise("", i18n.ERR_TOO_MANY_ARGS)
	}

	var (
		output []byte
		err    error
	)
	switch strings.ToLower(c.out) {
	// if the requested format is env
	case "env":
		output = input.ToEnvFile()
	case "yaml":
		output, err = yaml.Marshal(input)
		core.CheckErr(err, "cannot marshal input")
	case "json":
		output, err = json.MarshalIndent(input, "", " ")
		core.CheckErr(err, "cannot marshal input")
	}
	if *c.stdout {
		// print to console
		fmt.Println(string(output))
	} else {
		// save to disk
		var filename string
		switch strings.ToLower(c.out) {
		case "yaml":
			fallthrough
		case "yml":
			filename = "env.yaml"
		case "json":
			filename = "env.json"
		default:
			filename = ".env"
		}
		err = os.WriteFile(filename, output, os.ModePerm)
		core.CheckErr(err, "cannot write '%s' file", filename)
	}
}
