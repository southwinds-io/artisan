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
	"strings"
)

// Secret describes the secrets required by functions
type Secret struct {
	// the unique reference for the secret
	Name string `yaml:"name" json:"name"`
	// a description of the intended use or meaning of this secret
	Description string `yaml:"description" json:"description"`
	// the value of the secret
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
	// the value is required
	Required bool `yaml:"required,omitempty" json:"required,omitempty"`
}

type Secrets []*Secret

func (list Secrets) Get(ix int) *Secret {
	return list[ix]
}

func (list Secrets) GetByName(name string) *Secret {
	for _, v := range list {
		if strings.EqualFold(v.Name, name) {
			return v
		}
	}
	return nil
}

func (list Secrets) Len() int { return len(list) }

func (list Secrets) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list Secrets) Less(i, j int) bool {
	var si string = list[i].Name
	var sj string = list[j].Name
	var si_lower = strings.ToLower(si)
	var sj_lower = strings.ToLower(sj)
	if si_lower == sj_lower {
		return si < sj
	}
	return si_lower < sj_lower
}

func (list Secrets) Append(s *Secret) Secrets {
	list = append(list, s)
	return list
}
