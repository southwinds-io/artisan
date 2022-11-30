/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
	"regexp"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/registry"
	"southwinds.dev/os"
)

// UtilReplaceCmd client url issues http requests within a retry framework
type UtilReplaceCmd struct {
	Cmd         *cobra.Command
	regex       string
	replacement string
}

func NewUtilReplaceCmd() *UtilReplaceCmd {
	c := &UtilReplaceCmd{
		Cmd: &cobra.Command{
			Use:   "replace [flags] URI",
			Short: "replaces a set of characters found by regex with replacement string",
			Long:  `replaces a set of characters found by regex with replacement string`,
			Args:  cobra.ExactArgs(1),
			Example: `$ art u replace file.txt -r "image:(?:.*)$" "replaced-string"
`,
		},
	}
	c.Cmd.Flags().StringVarP(&c.regex, "regex", "r", "", `-r "image:(?:.*)$"`)
	c.Cmd.Flags().StringVarP(&c.replacement, "replace", "e", "", `-e "hello"`)
	c.Cmd.Args = cobra.ExactArgs(1)
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilReplaceCmd) Run(cmd *cobra.Command, args []string) {
	// load the file
	abs, err := filepath.Abs(args[0])
	core.CheckErr(err, "cannot make path absolute")
	content, err := os.ReadFile(abs, "")
	core.CheckErr(err, "cannot read file")
	r, err := regexp.Compile(c.regex)
	core.CheckErr(err, "cannot compile regex")
	replaced := r.ReplaceAll(content, []byte(c.replacement))
	core.CheckErr(registry.MoveFile(abs, fmt.Sprintf("%s.old", abs)), "cannot move file")
	core.CheckErr(os.WriteFile(replaced, abs, ""), "cannot write file")
}
