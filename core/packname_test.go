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

package core

import (
	"testing"
)

func TestParseName(t *testing.T) {
	data := map[string]bool{
		// correct
		"localhost/my-group/my-name":           false,
		"localhost:8082/my-group/my-name":      true,
		"localhost:8082/my-group/my-name:v1-0": true,
		"my-group/my-name":                     true,
		"my-name":                              true,
		"my-name:v1":                           true,
		// invalid character in domain
		"loc%$alhost/my-group/my-name": false,
		// domain cannot start with hyphen
		"-localhost/my-group/my-name": false,
		"localhost/my-group/:ggd":     false,
		":ggd":                        true,
		// domain cannot start with colon
		":localhost/my-group/:ggd": false,
		// missing group and name
		"127.0.0.1:884": false,
		// missing name
		"127.0.0.1:884/my-group":         false,
		"127.0.0.1:884/my-group/my-name": true,
		// no protocol scheme allowed
		"http://127.0.0.1:884/my-group/my-name":  false,
		"https://127.0.0.1:884/my-group/my-name": false,
		"tcp://127.0.0.1:884/my-group/my-name":   false,
		"ws://127.0.0.1:884/my-group/my-name":    false,
		"127.0.0.1:8f84/my-group/my-name":        false,
	}
	for name, valid := range data {
		_, err := ParseName(name)
		if valid && err != nil {
			t.Errorf(err.Error())
		}
		if !valid && err == nil {
			t.Errorf("name %s should be invalid", name)
		}
	}
}
