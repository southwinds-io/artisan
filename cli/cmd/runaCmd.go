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
	"southwinds.dev/artisan/runner"
)

// RunACmd runs application in a runtime
type RunACmd struct {
	Cmd         *cobra.Command
	home        string
	envFilename string
	path        string
	credentials string
	detached    bool
	clean       bool
}

func NewRunACmd(artHome string) *RunACmd {
	c := &RunACmd{
		Cmd: &cobra.Command{
			Use:   "runa [package-name]",
			Short: "runs a packaged application",
			Long:  `runs a packaged application`,
			Example: `
# assuming that the package "localhost:8082/app/artr:latest" has the following labels""
 - app:entrypoint: artr
 - app:var@ARTR_ADMIN_USER: required,default=admin
 - app:var@ARTR_ADMIN_PWD: required,default=adm1n
 - app:var@ARTR_READ_USER: optional
 - app:var@ARTR_READ_PWD: optional
 - app:volume@DATA_PATH: 0

# launches the artr application
art runa localhost:8082/app/artr:latest

# App Labels

- app:entrypoint = defines the relative path of the command to call in order to launch the application
- app:var@VAR_NAME = defines an environment variable needed by the application to run
- app:volume@VAR_NAME = defines a generic data volume mapped to VAR_NAME (e.g. VAR_NAME=/volume_0)
`,
		},
		home: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD artisan registry user and password")
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "the environment file to load; e.g. --env=.env or -e=.env")
	c.Cmd.Flags().StringVarP(&c.path, "path", "p", ".", "the path where application files should be placed")
	c.Cmd.Flags().BoolVarP(&c.detached, "detached", "d", false, "runs the application in the background")
	c.Cmd.Flags().BoolVarP(&c.clean, "clean", "c", false, "removes the application package from the local registry after opening it")
	c.Cmd.Args = cobra.ExactArgs(1)
	c.Cmd.Run = c.Run
	return c
}

func (c *RunACmd) Run(_ *cobra.Command, args []string) {
	name, err := core.ParseName(args[0])
	core.CheckErr(err, "invalid package name")
	core.CheckErr(runner.RunApp(name, c.credentials, c.detached, c.clean, c.path, c.home, nil, nil), "cannot run application")
}
