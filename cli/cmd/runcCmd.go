/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
