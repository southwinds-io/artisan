/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package build

import (
	"southwinds.dev/artisan/core"
	"testing"
)

func TestBuildContentOnly(t *testing.T) {
	builder := NewBuilder("")
	name, _ := core.ParseName("localhost:8080/lib/test1:1")
	_ = builder.Build("", "", "", name, "", false, false, "test/test")
}
