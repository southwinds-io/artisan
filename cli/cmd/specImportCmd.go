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
	. "southwinds.dev/artisan/release"
)

// SpecImportCmd Import the contents from a tarball to create an artisan package in the local registry
type SpecImportCmd struct {
	Cmd    *cobra.Command
	home   string
	creds  string
	filter string
}

func NewSpecImportCmd(artHome string) *SpecImportCmd {
	c := &SpecImportCmd{
		Cmd: &cobra.Command{
			Use:   "import [FLAGS] URI",
			Short: "imports an application release specification (e.g. one or more tarball files) into the local registry",
			Long: `Usage: art spec import [FLAGS] URI

Import one or more tarball files into the local registry using a specification (spec.yaml file).

Examples:
   # import a specification from a file system folder
   art spec import ./test
   
   # import a specification from an S3 bucket folder
   art spec import s3s://my-s3-service.com/my-app/v1.0 -c S3_ID:S3_SECRET
`,
		},
		home: artHome,
	}
	c.Cmd.Run = c.Run
	c.Cmd.Flags().StringVarP(&c.creds, "creds", "c", "", "the credentials used to retrieve the specification from an endpoint")
	c.Cmd.Flags().StringVarP(&c.filter, "filter", "f", "", "a regular expression used to select spec artefacts to be imported; any artefacts not matched by the filter are skipped (e.g. -f \"^quay.*$\")")
	return c
}

func (c *SpecImportCmd) Run(cmd *cobra.Command, args []string) {
	// check a package name has been provided
	if args != nil && len(args) < 1 {
		core.RaiseErr("the URI of the specification is required")
	}
	// import the tar archive(s)
	_, err := ImportSpec(ImportOptions{
		TargetUri:   args[0],
		TargetCreds: c.creds,
		Filter:      c.filter,
		ArtHome:     c.home,
	})
	core.CheckErr(err, "cannot import spec")
}
