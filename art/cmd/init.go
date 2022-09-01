/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cmd

import (
	"log"
	"southwinds.dev/artisan/core"
)

const ArtHome = ""

func init() {
	// ensure the registry folder structure is in place
	if err := core.EnsureRegistryPath(ArtHome); err != nil {
		log.Fatal("cannot run artisan without a local registry, its creation failed: %", err)
	}
}
