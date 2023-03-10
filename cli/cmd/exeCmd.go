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
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/merge"
)

// ExeCmd executes an exported function
type ExeCmd struct {
	cmd           *cobra.Command
	home          string
	interactive   *bool
	credentials   string
	path          string
	envFilename   string
	preserveFiles *bool
}

func NewExeCmd(artHome string) *ExeCmd {
	c := &ExeCmd{
		cmd: &cobra.Command{
			Use:   "exe [package name] [function]",
			Short: "runs a function within a package on the current host",
			Long:  `runs a function within a package on the current host`,
		},
		home: artHome,
	}
	c.interactive = c.cmd.Flags().BoolP("interactive", "i", false, "switches on interactive mode which prompts the user for information if not provided")
	c.cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD server user and password")
	c.cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "--env=.env or -e=.env")
	c.cmd.Flags().StringVar(&c.path, "path", "", "--path=/path/to/package/files - specify the location where the Artisan package must be open. If not specified, Artisan opens the package in a temporary folder under a randomly generated name.")
	c.preserveFiles = c.cmd.Flags().BoolP("preserve-files", "f", false, "use -f to preserve the open package files")
	c.cmd.Run = c.Run
	return c
}

func (c *ExeCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		core.RaiseErr("package and function names are required")
	}
	var (
		pack     = args[0]
		function = args[1]
	)
	name, err := core.ParseName(pack)
	i18n.Err("", err, i18n.ERR_INVALID_PACKAGE_NAME)
	// add the build file level environment variables
	env := merge.NewEnVarFromSlice(os.Environ())
	// load vars from file
	env2, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "failed to load environment file '%s'", c.envFilename)
	// merge with existing environment
	env.Merge(env2)
	// get a builder handle
	builder := build.NewBuilder(c.home)
	// run the function on the open package
	err = builder.Execute(name, function, c.credentials, *c.interactive, c.path, *c.preserveFiles, env, []string{}, false)
	core.CheckErr(err, "failed to execute function")
}
