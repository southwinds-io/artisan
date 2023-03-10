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
	"southwinds.dev/artisan/flow"
	"southwinds.dev/artisan/merge"

	"os"
	"path"
	"path/filepath"
	"strings"
)

// EnvFlowCmd collects variables required by a flow
type EnvFlowCmd struct {
	Cmd           *cobra.Command
	buildFilePath string
	stdout        *bool
	out           string
	flowPath      string
}

func NewEnvFlowCmd() *EnvFlowCmd {
	c := &EnvFlowCmd{
		Cmd: &cobra.Command{
			Use:   "flow [flags] [/path/to/flow_bare.yaml]",
			Short: "return the variables required by a given flow and can include a build.yaml",
			Long:  `return the variables required by a given flow and can include a build.yaml`,
		},
	}
	c.Cmd.Flags().StringVarP(&c.buildFilePath, "build-file-path", "b", "", "--build-file-path=. or -b=.; the path to an artisan build.yaml file from which to pick required inputs")
	c.Cmd.Flags().StringVarP(&c.out, "output", "o", "env", "--output yaml or -o yaml; the output format (e.g. env, json, yaml)")
	c.stdout = c.Cmd.Flags().Bool("stdout", false, "prints the output to the console")
	c.Cmd.Run = c.Run
	return c
}

func (c *EnvFlowCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		c.flowPath = core.ToAbsPath(args[0])
	} else if len(args) < 1 {
		core.RaiseErr("insufficient arguments: need the path to the bare flow file")
	} else if len(args) > 1 {
		core.RaiseErr("too many arguments: only need the path to the bare flow file")
	}
	// loads a bare flow from the path
	f, err := flow.LoadFlow(c.flowPath, "")
	core.CheckErr(err, "cannot load bare flow")

	// loads the build.yaml
	var b *data.BuildFile
	// if there is a build file, load it
	if len(c.buildFilePath) > 0 {
		b, err = data.LoadBuildFile(path.Join(c.buildFilePath, "build.yaml"))
	}
	// discover the input required by the flow / build file
	input, err := f.GetInputDefinition(b, merge.NewEnVarFromSlice([]string{}))
	core.CheckErr(err, "cannot get inputs")
	var output []byte
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
		os.Stdout.WriteString(fmt.Sprintf("%s\n", string(output)))
	} else {
		// save to disk
		dir := filepath.Dir(c.flowPath)
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
		core.CheckErr(os.WriteFile(path.Join(dir, filename), output, os.ModePerm), "cannot write '%s' file", filename)
	}
}
