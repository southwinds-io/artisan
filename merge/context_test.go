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
	_, _ = NewContext(env)
}
