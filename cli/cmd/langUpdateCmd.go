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
	"southwinds.dev/artisan/i18n"
)

// LangUpdateCmd add missing entries in language dictionary
type LangUpdateCmd struct {
	Cmd  *cobra.Command
	home string
}

func NewLangUpdateCmd(artHome string) *LangUpdateCmd {
	c := &LangUpdateCmd{
		Cmd: &cobra.Command{
			Use:   "update [path/to/lang/file]",
			Short: "add missing entries in language dictionary, added values in english",
			Long:  `add missing entries in language dictionary, added values in english`,
		},
		home: artHome,
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *LangUpdateCmd) Run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		i18n.Raise(c.home, i18n.ERR_INSUFFICIENT_ARGS)
	}
	if len(args) > 1 {
		i18n.Raise(c.home, i18n.ERR_TOO_MANY_ARGS)
	}
	err := i18n.Update(args[0])
	i18n.Err(c.home, err, i18n.ERR_CANT_UPDATE_LANG_FILE)
}
