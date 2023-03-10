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
	"os"
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/merge"
)

// RunCmd runs the function specified in the project's build.yaml file
type RunCmd struct {
	Cmd         *cobra.Command
	home        string
	envFilename string
	interactive *bool
}

func NewRunCmd(artHome string) *RunCmd {
	c := &RunCmd{
		Cmd: &cobra.Command{
			Use:   "run [function name] [project path]",
			Short: "runs the function commands specified in the project's build.yaml file",
			Long:  ``,
		},
		home: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "--env=.env or -e=.env; the path to a file containing environment variables to use")
	c.interactive = c.Cmd.Flags().BoolP("interactive", "i", false, "switches on interactive mode which prompts the user for information if not provided")
	c.Cmd.Run = c.Run
	return c
}

func (c *RunCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		core.RaiseErr("At least function name is required")
	}
	var function = args[0]
	var path = "."
	if len(args) > 1 {
		path = args[1]
	}
	builder := build.NewBuilder(c.home)
	// add the build file level environment variables
	env := merge.NewEnVarFromSlice(os.Environ())
	// load vars from file
	env2, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "failed to load environment file '%s'", c.envFilename)
	// merge with existing environment
	env.Merge(env2)
	// execute the function
	core.CheckErr(builder.Run(function, path, *c.interactive, env), "")
}
