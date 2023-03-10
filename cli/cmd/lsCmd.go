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
	"southwinds.dev/artisan/registry"
)

// ListCmd list packages
type ListCmd struct {
	Cmd      *cobra.Command
	home     string
	quiet    bool
	registry string
	creds    string
}

func NewListCmd(artHome string) *ListCmd {
	c := &ListCmd{
		Cmd: &cobra.Command{
			Use:   "ls [FLAGS]",
			Short: "list packages in the local or a remote registry",
			Long:  `list packages in the local or a remote registry`,
			Example: `
# list packages in local registry
art ls

# list packages from remote registry at localhost:8082
art ls -r localhost:8082 -u <user>:<pwd>
`,
		},
		home: artHome,
	}
	c.Cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "only show numeric IDs")
	c.Cmd.Flags().StringVarP(&c.registry, "registry", "r", "", "the domain name or IP of the remote registry (e.g. my-remote-registry); port can also be specified using a colon syntax")
	c.Cmd.Flags().StringVarP(&c.creds, "user", "u", "", "the credentials used to retrieve the information from the remote registry")
	c.Cmd.Run = c.Run
	return c
}

func (c *ListCmd) Run(_ *cobra.Command, _ []string) {
	if len(c.registry) == 0 {
		local := registry.NewLocalRegistry(c.home)
		if c.quiet {
			local.ListQ()
		} else {
			local.List(c.home, false)
		}
	}
	if len(c.registry) > 0 {
		uname, pwd := core.RegUserPwd(c.creds)
		remote, err := registry.NewRemoteRegistry(c.registry, uname, pwd, c.home)
		core.CheckErr(err, "invalid registry name")
		remote.List(c.quiet)
	}
}
