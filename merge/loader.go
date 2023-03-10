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
	"strconv"
	"strings"
)

type loader struct {
	items []item
	vars  map[string]string
}

type item struct {
	group string
	name  string
	index string
	value string
}

func NewLoader(env *Envar) loader {
	l := &loader{
		items: []item{},
		vars:  map[string]string{},
	}
	for key, value := range env.vars {
		// any variable added to a key-value map
		l.vars[key] = value
		// now processes grouped variables (i.e. following naming convention GROUP__NAME__IX)
		ix1 := strings.Index(key, "__")
		ix2 := strings.LastIndex(key, "__")
		if ix1 > 0 && ix2 > 0 {
			group := key[:ix1]
			name := key[ix1+2 : ix2]
			index := key[ix2+2:]
			l.items = append(l.items, item{
				group: group,
				name:  name,
				index: index,
				value: value,
			})
		}
	}
	return *l
}

func (l *loader) set(group, index string, ctx *Context) Set {
	vars := make(map[string]string)
	for _, i := range l.items {
		if i.group == group && i.index == index {
			vars[i.name] = i.value
		}
	}
	return Set{
		Value:   vars,
		Context: ctx,
	}
}

func (l *loader) indices(group string) int {
	var result int = 0
	for _, i := range l.items {
		ii, _ := strconv.Atoi(i.index)
		if ii > result && i.group == group {
			result = ii
		}
	}
	return result
}
