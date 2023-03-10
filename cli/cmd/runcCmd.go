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
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/runner"
)

// RunCCmd runs a function specified in the project's build.yaml file within an artisan runtime
type RunCCmd struct {
	Cmd         *cobra.Command
	home        string
	interactive *bool
	envFilename string
	network     string
}

func NewRunCCmd(artHome string) *RunCCmd {
	c := &RunCCmd{
		Cmd: &cobra.Command{
			Use:   "runc [function name] [project path]",
			Short: "runs the function commands specified in the project's build.yaml file within an artisan runtime container",
			Long:  ``,
		},
		home: artHome,
	}
	c.interactive = c.Cmd.Flags().BoolP("interactive", "i", false, "switches on interactive mode which prompts the user for information if not provided")
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "the environment file to load; e.g. --env=.env or -e=.env")
	c.Cmd.Flags().StringVarP(&c.network, "network", "n", "", "attaches the container to the specified docker network; by default it is not specified so the container is not attached to any docker network; usage: --network my-net")
	c.Cmd.Run = c.Run
	return c
}

func (c *RunCCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		core.RaiseErr("At least function name is required")
	}
	var function = args[0]
	var path = "."
	if len(args) > 1 {
		path = args[1]
	}
	// create an instance of the runner
	run, err := runner.NewFromPath(path, c.home)
	core.CheckErr(err, "cannot initialise runner")
	// load environment variables from file
	// NOTE: do not pass any vars from the host to avoid clashing issues
	// if any vars are required load them directly into the container from the env file
	env, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "failed to load environment file '%s'", c.envFilename)
	// launch a runtime to execute the function
	err = run.RunC(function, *c.interactive, env, c.network)
	core.CheckErr(err, "cannot execute function '%s'", function)
}
