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
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

type TemplateContext interface {
	FuncMap() template.FuncMap
}

// Context the merge context for artisan templates .art
type Context struct {
	Env    *Envar
	Loader loader
	// the selected variable group for a range
	currentGroup string
	// a list of variable sets
	Items []Set
}

func NewContext(env *Envar) (TemplateContext, error) {
	ctx := &Context{
		Env:    env,
		Loader: NewLoader(env),
		Items:  []Set{},
	}
	return ctx, nil
}

func (c *Context) FuncMap() template.FuncMap {
	return template.FuncMap{
		"select":  c.Select,
		"item":    c.Item,
		"itemEq":  c.ItemEq,
		"itemNeq": c.ItemNeq,
		"var":     c.Var,
		"having":  c.GroupExists,
		"exists":  c.Exists,
	}
}

func (c *Context) Exists(variableName reflect.Value) reflect.Value {
	exists := len(c.Env.Get(variableName.String())) > 0
	return reflect.ValueOf(exists)
}

// Var return the value of a variable
func (c *Context) Var(name reflect.Value) reflect.Value {
	v := reflect.ValueOf(c.Loader.vars[name.String()])
	return v
}

// Select a specific variable group and populate all variable sets within the group
func (c *Context) Select(group reflect.Value) reflect.Value {
	c.currentGroup = group.String()
	ii := c.Loader.indices(c.currentGroup)
	c.Items = []Set{}
	for i := 1; i <= ii; i++ {
		c.Items = append(c.Items, c.Loader.set(c.currentGroup, fmt.Sprint(i), c))
	}
	return reflect.ValueOf("")
}

// Item return a grouped variable value using its name and the current iteration set
func (c *Context) Item(name reflect.Value, set reflect.Value) reflect.Value {
	s, ok := set.Interface().(Set)
	if !ok {
		panic("Item function requires a set for the first parameter\n")
	}
	return reflect.ValueOf(s.Value[name.String()])
}

// ItemEq return a boolean indicating whether the value of a variable identified by key is equals to the passed-in value
func (c *Context) ItemEq(name reflect.Value, set reflect.Value, value reflect.Value) reflect.Value {
	s, ok := set.Interface().(Set)
	if !ok {
		panic("Item function requires a set for the first parameter\n")
	}
	v := reflect.ValueOf(s.Value[name.String()])
	return reflect.ValueOf(strings.EqualFold(v.String(), value.String()))
}

// ItemNeq return a boolean indicating whether the value of a variable identified by key is not equal to the passed-in value
func (c *Context) ItemNeq(name reflect.Value, set reflect.Value, value reflect.Value) reflect.Value {
	s, ok := set.Interface().(Set)
	if !ok {
		panic("Item function requires a set for the first parameter\n")
	}
	v := reflect.ValueOf(s.Value[name.String()])
	return reflect.ValueOf(!strings.EqualFold(v.String(), value.String()))
}

func (c *Context) GroupExists(group reflect.Value) bool {
	ix := c.Loader.indices(group.String())
	return ix != 0
}
