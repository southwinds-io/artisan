/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package build

import (
	"fmt"
	"southwinds.dev/artisan/merge"
	"testing"
)

func TestExe(t *testing.T) {
	out, err := Exe("printenv", ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}

func TestExeAsync(t *testing.T) {
	out, err := ExeAsync("printenv", ".", merge.NewEnVarFromSlice([]string{}), false)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(out)
}
