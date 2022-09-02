/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"github.com/spf13/cobra"
	"southwinds.dev/artisan/i18n"
)

// LangUpdateCmd add missing entries in language dictionary
type LangUpdateCmd struct {
	Cmd *cobra.Command
}

func NewLangUpdateCmd() *LangUpdateCmd {
	c := &LangUpdateCmd{
		Cmd: &cobra.Command{
			Use:   "update [path/to/lang/file]",
			Short: "add missing entries in language dictionary, added values in english",
			Long:  `add missing entries in language dictionary, added values in english`,
		},
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *LangUpdateCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		i18n.Raise("", i18n.ERR_INSUFFICIENT_ARGS)
	}
	if len(args) > 1 {
		i18n.Raise("", i18n.ERR_TOO_MANY_ARGS)
	}
	err := i18n.Update(args[0])
	i18n.Err("", err, i18n.ERR_CANT_UPDATE_LANG_FILE)
}
