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
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"southwinds.dev/artisan/core"
)

// UtilBase64Cmd generates passwords
type UtilBase64Cmd struct {
	Cmd    *cobra.Command
	decode *bool
}

func NewUtilBase64Cmd() *UtilBase64Cmd {
	c := &UtilBase64Cmd{
		Cmd: &cobra.Command{
			Use:   "b64 [flags] STRING",
			Short: "base 64 encode (or alternatively decode) a string",
			Long:  `base 64 encode (or alternatively decode) a string`,
			Args:  cobra.ExactArgs(1),
		},
	}
	c.Cmd.Run = c.Run
	c.decode = c.Cmd.Flags().BoolP("decode", "d", false, "if sets, decodes the string instead of encoding it")
	return c
}

func (c *UtilBase64Cmd) Run(_ *cobra.Command, args []string) {
	if *c.decode {
		decoded, err := base64.StdEncoding.DecodeString(args[0])
		core.CheckErr(err, "cannot decode string")
		fmt.Printf("%s", string(decoded[:]))
	} else {
		fmt.Printf("%s", base64.StdEncoding.EncodeToString([]byte(args[0])))
	}
}
