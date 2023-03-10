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

// RmCmd remove local packages
type RmCmd struct {
	Cmd      *cobra.Command
	home     string
	all      bool
	registry string
	filter   string
	creds    string
	dry      bool
}

func NewRmCmd(artHome string) *RmCmd {
	c := &RmCmd{
		Cmd: &cobra.Command{
			Use:   "rm PACKAGE [PACKAGE...]",
			Short: "removes one or more packages from the local package registry or a remote registry",
			Long:  `removes one or more packages from the local package registry or a remote registry`,
			Example: `
# delete all packages in local registry
art rm -a

# delete specific packages from the local registry
art rm package-1 package-2

# dry-run delete all packages in the remote registry at localhost:8082
# note: filter expression must be double quoted
art rm -r localhost:8082 -u admin:adm1n -xf ".*"

# actual delete all packages in the remote registry at localhost:8082
# note: filter expression must be double quoted
art rm -r localhost:8082 -u admin:adm1n -f ".*"

# remove two packages by their id
art rm -r localhost:8081 -u <user>:<pwd> 4562fr 76dt54

# remove all package from the remote registry
art rm -r localhost:8081 -u <user>:<pwd> $(art ls -r localhost:8081 -u <user>:<pwd> -q)
`,
		},
		home: artHome,
	}
	c.Cmd.Flags().BoolVarP(&c.dry, "dry-run", "x", false, "when using a filter on a remote registry, shows a list of packages that would be deleted without actually deleting them, use it to test remove operations before actually performing the delete")
	c.Cmd.Flags().BoolVarP(&c.all, "all", "a", false, "remove all packages")
	c.Cmd.Flags().StringVarP(&c.registry, "registry", "r", "", "the domain of the remote artisan registry to use")
	c.Cmd.Flags().StringVarP(&c.creds, "user", "u", "", "the credentials used to retrieve the information from the remote registry")
	c.Cmd.Flags().StringVarP(&c.filter, "filter", "f", "", "the regular expression used to find the packages to remove, only used if the remove operation if for a remote registry")
	c.Cmd.Run = c.Run
	return c
}

func (c *RmCmd) Run(cmd *cobra.Command, args []string) {
	// check one or more package names have been provided if remove all is not specified
	if len(args) == 0 && !c.all && len(c.filter) == 0 || // local or remote registry, specific package deletion but no args (packages defined)
		len(args) == 0 && len(c.filter) == 0 && len(c.registry) > 0 && !c.all { // remote registry, no filter and no packages
		core.RaiseErr("missing name(s) of the package(s) to remove")
	}
	// cannot provide all flag and package name
	if len(args) > 0 && c.all {
		core.RaiseErr("a package name %s should not be provided with the --all|-a flag", args[0])
	}
	// if no remote specified then it is a local operation
	if len(c.registry) == 0 {
		if len(c.filter) > 0 {
			core.RaiseErr("--filter flag is not valid for the local registry, can only be used when --registry is set")
		}
		//  create a local registry
		local := registry.NewLocalRegistry(c.home)
		if c.all {
			core.CheckErr(local.RemoveAll(), "cannot remove packages")
		} else {
			core.CheckErr(local.Remove(args), "cannot remove package")
		}
	} else {
		uname, pwd := core.RegUserPwd(c.creds)
		remote, err := registry.NewRemoteRegistry(c.registry, uname, pwd, c.home)
		core.CheckErr(err, "invalid remote")
		// otherwise, it is a remote operation
		if c.all {
			core.RaiseErr("--all flag is not valid for remote registries, use a filter expression instead")
		}
		if len(c.filter) > 0 {
			core.CheckErr(remote.RemoveByNameFilter(c.filter, c.dry), "cannot remove packages from remote registry using filter")
		} else {
			// creates a remote picking the domain from the filter expression
			core.CheckErr(remote.RemoveByNameOrId(args), "cannot remove packages from remote registry")
		}
	}
}
