package cmd

/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
