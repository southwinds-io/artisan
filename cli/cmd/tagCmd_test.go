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
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/registry"
	"testing"
)

// TestTagV1ToLatestExist test that a package with a V1 tag can be tagged to latest when a previous latest tag exists
// and the existing tag is renamed so that there is no dangling packages
func TestTagV1ToLatestExist(t *testing.T) {
	// pre-conditions
	reg := registry.NewLocalRegistry(core.ArtDefaultHome)
	// cleanup
	testLatest, _ := core.ParseName("test:latest")
	testV1, _ := core.ParseName("test:V1")
	// clear registry if previous packages exist
	// TODO: ensure all packages are removed
	reg.Remove([]string{"test:latest", "test:V1"})
	// build latest
	builder := build.NewBuilder(core.ArtDefaultHome)
	builder.Build(".", "", "", testLatest, "test1", false, false, "", "", "", "")
	// build V1
	builder.Build(".", "", "", testV1, "test1", false, false, "", "", "", "")
	// reload the registry
	reg.Load()
	testV1PId := reg.FindPackageByName(testV1).Id
	testLatestPId := reg.FindPackageByName(testLatest).Id
	// execute action tag
	reg.Tag("test:V1", "test:latest")
	// reload the registry
	reg.Load()
	// check post-conditions
	testLatestP := reg.FindPackageNamesById(testLatestPId)
	if testLatestP == nil {
		t.Fatalf("test:latest package not found")
	}
	// the old latest renamed to avoid dangling
	if len(testLatestP) != 1 {
		t.Fatalf("old test:latest package should have only one renamed tag")
	}
	testV1P := reg.FindPackageNamesById(testV1PId)
	// the new latest tag added on top of existing package with V1 tag
	if len(testV1P) != 2 {
		t.Fatalf("")
	}
}

// TestTagV1ToLatest test that a package with a V1 tag can be tagged to latest when a previous latest tag does not exist
func TestTagV1ToLatest(t *testing.T) {
	reg := registry.NewLocalRegistry(core.ArtDefaultHome)
	// cleanup
	testLatest, _ := core.ParseName("test:latest")
	testV1, _ := core.ParseName("test:V1")
	reg.Remove([]string{"test:latest", "test:V1"})
	// build latest
	builder := build.NewBuilder(core.ArtDefaultHome)
	// build V1
	builder.Build(".", "", "", testV1, "test1", false, false, "", "", "", "")
	// reload the registry
	reg.Load()
	// tag
	reg.Tag("test:V1", "test:latest")
	// check post-conditions
	if reg.FindPackageByName(testLatest) == nil {
		t.Fatalf("test:latest package not found")
	}
	if reg.FindPackageByName(testV1) == nil {
		t.Fatalf("test:V1 package not found")
	}
}
