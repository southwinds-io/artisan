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
	Cmd           *cobra.Command
	regex         string
	replaceString string
	replaceFile   string
	keepOld       bool
}

func NewUtilReplaceCmd() *UtilReplaceCmd {
	c := &UtilReplaceCmd{
		Cmd: &cobra.Command{
			Use:   "replace [flags] URI",
			Short: "replaces a set of characters found by regex with replacement string",
			Long:  `replaces a set of characters found by regex with replacement string`,
			Args:  cobra.ExactArgs(1),
			Example: `$ art u replace -r "image:(?:.*)$" -e "replacement-content-string" file.txt 
$ art u replace -r "image:(?:.*)$" -f "replacement-content-filename" file.txt
`,
		},
	}
	c.Cmd.Flags().StringVarP(&c.regex, "regex", "r", "", `-r "image:(?:.*)$"; a golang regular expression to match the  test  to replace`)
	c.Cmd.Flags().StringVarP(&c.replaceString, "rep-str", "e", "", `-e "hello"; the new text that will replace the regex matched text`)
	c.Cmd.Flags().StringVarP(&c.replaceFile, "rep-file", "f", "", `-f "replacement.txt"; the name of the file containing the replacement text`)
	c.Cmd.Flags().BoolVarP(&c.keepOld, "keep-old", "k", false, `-k; keeps the old version of the file before it was replaced`)
	c.Cmd.Args = cobra.ExactArgs(1)
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilReplaceCmd) Run(cmd *cobra.Command, args []string) {
	if len(c.replaceString) == 0 && len(c.replaceFile) == 0 {
		core.RaiseErr("one of rep-str or rep-file flags should be provided")
	}
	if len(c.replaceString) > 0 && len(c.replaceFile) > 0 {
		core.RaiseErr("cannot set both rep-str and rep-file flags, choose one")
	}
	// load the file
	abs, err := filepath.Abs(args[0])
	core.CheckErr(err, "cannot make path absolute")
	content, err := os.ReadFile(abs, "")
	core.CheckErr(err, "cannot read file")
	r, err := regexp.Compile(c.regex)
	core.CheckErr(err, "cannot compile regex")
	var replaced []byte
	if len(c.replaceString) > 0 {
		replaced = r.ReplaceAll(content, []byte(c.replaceString))
	} else {
		replacementBytes, err := os.ReadFile(c.replaceFile, "")
		core.CheckErr(err, "cannot read replacement file")
		replaced = r.ReplaceAll(content, replacementBytes)
	}
	if c.keepOld {
		core.CheckErr(registry.MoveFile(abs, fmt.Sprintf("%s.old", abs)), "cannot move file")
	}
	core.CheckErr(os.WriteFile(replaced, abs, ""), "cannot write file")
}
