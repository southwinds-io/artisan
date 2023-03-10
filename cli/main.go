/*
	Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
	Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
	Contributors to this project, hereby assign copyright in this code to the project,
	to be licensed under the same terms as the rest of the code.
*/

package main

import (
	"log"
	"southwinds.dev/artisan/cli/cmd"
	"southwinds.dev/artisan/core"
)

func main() {
	// ensure the registry folder structure is in place
	if err := core.EnsureRegistryPath(core.ArtDefaultHome); err != nil {
		log.Fatal("cannot run artisan without a local registry, its creation failed: %", err)
	}

	rootCmd := cmd.InitialiseRootCmd(core.ArtDefaultHome)

	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.
	rootCmd.Cmd.Execute()
}
