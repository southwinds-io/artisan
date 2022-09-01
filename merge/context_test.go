/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package merge

import (
	"testing"
)

func TestContext(t *testing.T) {
	env := NewEnVarFromMap(map[string]string{
		"PORT__NAME__1":  "port a",
		"PORT__NAME__2":  "port b",
		"PORT__NAME__3":  "port c",
		"PORT__DESC__1":  "port a description",
		"PORT__DESC__2":  "port b description",
		"PORT__DESC__3":  "port c description",
		"PORT__VALUE__1": "80",
		"PORT__VALUE__2": "8080",
		"PORT__VALUE__3": "443",
	})
	_, _ = NewContext(*env)
}
