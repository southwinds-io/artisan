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
	"fmt"
	"github.com/spf13/cobra"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/flow"
	"southwinds.dev/artisan/merge"

	"os"
	"path/filepath"
)

// FlowMergeCmd merge a flow with env variables
type FlowMergeCmd struct {
	Cmd           *cobra.Command
	home          string
	envFilename   string
	buildFilePath string
	stdout        *bool
	tkn           *bool
	out           string
	interactive   *bool
	labels        []string
}

func NewFlowMergeCmd(artHome string) *FlowMergeCmd {
	c := &FlowMergeCmd{
		Cmd: &cobra.Command{
			Use:   "merge [flags] [/path/to/flow_bare.yaml]",
			Short: "fills in a bare flow by adding the required variables, secrets and keys",
			Long:  `fills in a bare flow by adding the required variables, secrets and keys`,
		},
		home: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "--env=.env or -e=.env; the path to a file containing environment variables to use")
	c.Cmd.Flags().StringVarP(&c.buildFilePath, "build-file-path", "b", "", "--build-file-path=. or -b=.; the path to an artisan build.yaml file from which to pick required inputs")
	c.stdout = c.Cmd.Flags().Bool("stdout", false, "prints the output to the console")
	c.Cmd.Flags().StringVarP(&c.out, "output", "o", "yaml", "--output json or -o json; the output format for the written flow; available formats are:\n"+
		"yaml: output in YAML format\n"+
		"json: output in JSON format\n"+
		"ojson: output as an Onix configuration item format\n")
	c.interactive = c.Cmd.Flags().BoolP("interactive", "i", false, "switches on interactive mode which prompts the user for information if not provided")
	c.Cmd.Flags().StringSliceVarP(&c.labels, "label", "l", []string{}, "add one or more labels to the flow; -l label1=value1 -l label2=value2")
	c.Cmd.Run = c.Run
	return c
}

func (c *FlowMergeCmd) Run(_ *cobra.Command, args []string) {
	var flowPath string
	if len(args) == 1 {
		flowPath = core.ToAbsPath(args[0])
	} else if len(args) < 1 {
		core.RaiseErr("insufficient arguments: need the path to the bare flow file")
	} else if len(args) > 1 {
		core.RaiseErr("too many arguments: only need the path to the bare flow file")
	}
	// add the build file level environment variables
	env := merge.NewEnVarFromSlice(os.Environ())
	// load vars from file
	env2, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "failed to load environment file '%s'", c.envFilename)
	// merge with existing environment
	env.Merge(env2)
	// loads a bare flow from the path
	f, err := flow.NewWithEnv(flowPath, c.buildFilePath, env, c.home)
	core.CheckErr(err, "cannot load bare flow")
	// add labels to the flow
	f.AddLabels(c.labels)
	// merges input, surveying for required data if in interactive mode
	err = f.Merge(*c.interactive)
	core.CheckErr(err, "cannot merge bare flow")
	// if stdout required
	if *c.stdout {
		if c.out == "yaml" {
			// marshals the flow to YAML
			yaml, err := f.YamlString()
			core.CheckErr(err, "cannot marshal bare flow")
			// print to stdout
			fmt.Println(yaml)
		} else if c.out == "json" {
			// marshals the flow to YAML
			json, err := f.JsonString()
			core.CheckErr(err, "cannot marshal bare flow")
			// print to stdout
			fmt.Println(json)
		} else {
			core.RaiseErr("invalid format '%s'", c.out)
		}
	} else {
		// save the flow to file
		if c.out == "yaml" {
			err = f.SaveYAML()
		} else if c.out == "json" {
			err = f.SaveJSON()
		} else if c.out == "ojson" {
			err = f.SaveOnixJSON()
		} else {
			core.RaiseErr("invalid format '%s'", c.out)
		}
		core.CheckErr(err, "cannot save bare flow")
	}
}

func tknPath(path string) string {
	dir, file := filepath.Split(path)
	filename := core.FilenameWithoutExtension(file)
	return filepath.Join(dir, fmt.Sprintf("%s_tkn.yaml", filename[0:len(filename)-len("_bare")]))
}
