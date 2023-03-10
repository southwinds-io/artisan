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
)

// WaitCmd wait until the response payload contains a specific value
type WaitCmd struct {
	Cmd      *cobra.Command
	attempts int
	filter   string
	creds    string
}

func NewWaitCmd() *WaitCmd {
	c := &WaitCmd{
		Cmd: &cobra.Command{
			Use:   "wait [flags] URI",
			Short: "wait until either the an HTTP GET returns a value or the maximum attempts have been reached",
			Long:  `wait until either the an HTTP GET returns a value or the maximum attempts have been reached`,
			Args:  cobra.ExactArgs(1),
		},
	}
	c.Cmd.Flags().StringVarP(&c.filter, "filter", "f", "", "-f json/path/expression")
	c.Cmd.Flags().StringVarP(&c.creds, "creds", "u", "", "-u user:password")
	c.Cmd.Flags().IntVarP(&c.attempts, "attempts", "a", 5, "-a 10 (number of attempts before it fails)")
	c.Cmd.Run = c.Run
	return c
}

func (c *WaitCmd) Run(cmd *cobra.Command, args []string) {
	token := ""
	if len(c.creds) > 0 {
		uname, pwd := core.UserPwd(c.creds)
		token = core.BasicToken(uname, pwd)
	}
	core.Wait(args[0], c.filter, token, c.attempts)
}
