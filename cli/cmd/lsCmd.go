/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
	quiet    bool
	registry string
	creds    string
}

func NewListCmd() *ListCmd {
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
	}
	c.Cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "only show numeric IDs")
	c.Cmd.Flags().StringVarP(&c.registry, "registry", "r", "", "the domain name or IP of the remote registry (e.g. my-remote-registry); port can also be specified using a colon syntax")
	c.Cmd.Flags().StringVarP(&c.creds, "user", "u", "", "the credentials used to retrieve the information from the remote registry")
	c.Cmd.Run = c.Run
	return c
}

func (c *ListCmd) Run(_ *cobra.Command, _ []string) {
	if len(c.registry) == 0 {
		local := registry.NewLocalRegistry("")
		if c.quiet {
			local.ListQ()
		} else {
			local.List("", false)
		}
	}
	if len(c.registry) > 0 {
		uname, pwd := core.RegUserPwd(c.creds)
		remote, err := registry.NewRemoteRegistry(c.registry, uname, pwd, "")
		core.CheckErr(err, "invalid registry name")
		remote.List(c.quiet)
	}
}
