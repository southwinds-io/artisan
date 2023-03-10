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
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/runner"
)

type ExeCCmd struct {
	Cmd         *cobra.Command
	home        string
	interactive *bool
	credentials string
	path        string
	envFilename string
	network     string
}

func NewExeCCmd(artHome string) *ExeCCmd {
	c := &ExeCCmd{
		Cmd: &cobra.Command{
			Use:   "exec [flags] [package-name] [function-name]",
			Short: "runs a function within a package using an artisan runtime",
			Long: `runs a function within a package using an artisan runtime
* package-name: 
   mandatory - the fully qualified name of the package containing the function to execute
* function-name: 
   mandatory - the name of the function exported by the package that should be executed

NOTE: exec always pulls the package from its registry as it is done within the runtime and that is its only behaviour
   if the package is in a secure registry, then credentials must be specified via -u / --credentials flag
   if running in a linux host, ensure the user executing the exec command has UID/GID = 100000000 
   to avoid read / write issues from / to the host - e.g. public PGP key in the host artisan registry is required to open
     the package within the runtime - keys are accessible within the runtime using bind mounts
`,
		},
		home: artHome,
	}
	c.interactive = c.Cmd.Flags().BoolP("interactive", "i", false, "switches on interactive mode which prompts the user for information if not provided")
	c.Cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "the artisan registry user and password; e.g. -u USER:PASSWORD or -u USER")
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "the environment file to load; e.g. --env=.env or -e=.env")
	c.Cmd.Flags().StringVarP(&c.network, "network", "n", "", "attaches the container to the specified docker network; by default it is not specified so the container is not attached to any docker network; usage: --network my-net")
	c.Cmd.Run = c.Run
	return c
}

func (c *ExeCCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		core.RaiseErr("insufficient arguments")
	} else if len(args) > 2 {
		core.RaiseErr("too many arguments")
	}
	var (
		packageName = args[0]
		fxName      = args[1]
	)
	// create an instance of the runner
	run, err := runner.New()
	core.CheckErr(err, "cannot initialise runner")
	// load environment variables from file
	// NOTE: do not load from host environment to prevent clashes in the container
	env, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "failed to load environment file '%s'", c.envFilename)
	if len(c.credentials) == 0 {
		core.InfoLogger.Printf("no credentials have been provided, if you are connecting to a authenticated registry, you need to pass the -u flag\n")
	}
	// launch a runtime to execute the function
	err = run.ExeC(packageName, fxName, c.credentials, c.network, *c.interactive, env)
	i18n.Err(c.home, err, i18n.ERR_CANT_EXEC_FUNC_IN_PACKAGE, fxName, packageName)
}
