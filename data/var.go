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

package data

import (
	"southwinds.dev/artisan/core"
	"strings"
)

type Var struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Required    bool   `yaml:"required" json:"required"`
	Type        string `yaml:"type" json:"type"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
}

type Vars []*Var

func (list Vars) Len() int { return len(list) }

func (list Vars) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list Vars) Less(i, j int) bool {
	var si = list[i].Name
	var sj = list[j].Name
	var siLower = strings.ToLower(si)
	var sjLower = strings.ToLower(sj)
	if siLower == sjLower {
		return si < sj
	}
	return siLower < sjLower
}

func (list Vars) Get(ix int) *Var {
	return list[ix]
}

func (list Vars) GetByName(name string) *Var {
	for _, v := range list {
		if strings.EqualFold(v.Name, name) {
			return v
		}
	}
	return nil
}

func (list Vars) Append(v *Var) Vars {
	list = append(list, v)
	return list
}

type VerifyHandler func(name *core.PackageName, seal *Seal, path string, authorisedAuthors []string, flags uint8) error
type RunHandler func(name *core.PackageName, fx string, seal *Seal) error
