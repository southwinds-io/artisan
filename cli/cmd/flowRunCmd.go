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
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/flow"
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/merge"
)

type FlowRunCmd struct {
	Cmd           *cobra.Command
	home          string
	envFilename   string
	credentials   string
	interactive   *bool
	flowPath      string
	runnerName    string
	buildFilePath string
	labels        []string
}

func NewFlowRunCmd(artHome string) *FlowRunCmd {
	c := &FlowRunCmd{
		Cmd: &cobra.Command{
			Use:   "run [flags] [/path/to/flow.yaml] [runner name]",
			Short: "merge and send a flow to a runner for execution",
			Long:  `merge and send a flow to a runner for execution`,
		},
		home: artHome,
	}
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "--env=.env or -e=.env; the path to a file containing environment variables to use")
	c.Cmd.Flags().StringVarP(&c.credentials, "user", "u", "", "USER:PASSWORD server user and password")
	c.interactive = c.Cmd.Flags().BoolP("interactive", "i", false, "switches on interactive mode which prompts the user for information if not provided")
	c.Cmd.Flags().StringVarP(&c.buildFilePath, "build-file-path", "b", ".", "--build-file-path=. or -b=.; the path to an artisan build.yaml file from which to pick required inputs")
	c.Cmd.Flags().StringSliceVarP(&c.labels, "label", "l", []string{}, "add one or more labels to the flow; -l label1=value1 -l label2=value2")
	c.Cmd.Run = c.Run
	return c
}

func (c *FlowRunCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) == 2 {
		c.flowPath = core.ToAbsPath(args[0])
		c.runnerName = args[1]
	} else if len(args) < 1 {
		i18n.Raise(c.home, i18n.ERR_INSUFFICIENT_ARGS)
	} else if len(args) > 1 {
		i18n.Raise(c.home, i18n.ERR_TOO_MANY_ARGS)
	}
	// add the build file level environment variables
	env := merge.NewEnVarFromSlice(os.Environ())
	// load vars from file
	env2, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "failed to load environment file '%s'", c.envFilename)
	// merge with existing environment
	env.Merge(env2)
	// loads a flow from the path
	f, err := flow.NewWithEnv(c.flowPath, c.buildFilePath, env, c.home)
	core.CheckErr(err, "cannot load flow")
	// add labels to the flow
	f.AddLabels(c.labels)
	err = f.Merge(*c.interactive)
	core.CheckErr(err, "cannot merge flow")
	err = f.Run(c.runnerName, c.credentials, *c.interactive)
	core.CheckErr(err, "cannot run flow")
}
