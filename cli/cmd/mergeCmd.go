package cmd

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
import (
	"github.com/spf13/cobra"
	"os"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/merge"
)

// MergeCmd merges environment variables into one or more files
type MergeCmd struct {
	Cmd         *cobra.Command
	envFilename string
}

func NewMergeCmd() *MergeCmd {
	c := &MergeCmd{
		Cmd: &cobra.Command{
			Use:   "merge [flags] [template1 template2 template3 ...]",
			Short: "merges environment variables in the specified template files",
			Long: `
	merges environment variables in the specified template files
	merge merges variables stored in an .env file into one or more merge template files
	merge creates new merged files after the name of the templates without their extension`,
		},
	}
	c.Cmd.Run = c.Run
	c.Cmd.Flags().StringVarP(&c.envFilename, "env", "e", ".env", "--env=.env or -e=.env")
	return c
}

func (c *MergeCmd) Run(cmd *cobra.Command, args []string) {
	env := merge.NewEnVarFromSlice(os.Environ())
	env2, err := merge.NewEnVarFromFile(c.envFilename)
	core.CheckErr(err, "cannot load .env file")
	env.Merge(env2)
	m, _ := merge.NewTemplMerger()
	err = m.LoadTemplates(args)
	core.CheckErr(err, "cannot load templates")
	err = m.Merge(env)
	core.CheckErr(err, "cannot merge templates")
	err = m.Save()
	core.CheckErr(err, "cannot save templates")
}
