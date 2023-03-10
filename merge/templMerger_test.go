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
	"html/template"
	"os"
	"testing"
)

func TestMergeUsingFunctions(t *testing.T) {
	e := &Envar{vars: map[string]string{}}
	// this is an ordinary variable
	e.Set("TITLE", "Example of merging Grouped Variables")

	// these are grouped variables
	// note the naming convention: GROUP-NAME__VARIABLE-NAME__VARIABLE-INDEX
	e.Set("PORT__NAME__1", "Standard TCP")
	e.Set("PORT__DESC__1", "The standard port")
	e.Set("PORT__VALUE__1", "80")

	e.Set("PORT__NAME__2", "Alternative TCP")
	e.Set("PORT__DESC__2", "An alternative http port")
	e.Set("PORT__VALUE__2", "8080")

	e.Set("PORT__NAME__3", "Standard Encrypted")
	e.Set("PORT__DESC__3", "HTTPS port")
	e.Set("PORT__VALUE__3", "443")

	e.Set("URI__NAME__1", "URI 1")
	e.Set("URI__VALUE__1", "www.hhhh.com")

	e.Set("URI__NAME__2", "URI 2")
	e.Set("URI__VALUE__2", "www.hhwedwdehh.com")

	m, _ := NewTemplMerger()
	err := m.LoadTemplates([]string{"test/sample_using_functions.yaml.art"})
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = m.Merge(e)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, bytes := range m.file {
		fmt.Println(string(bytes))
	}
}

func TestMergeUsingOperators(t *testing.T) {
	e := &Envar{vars: map[string]string{}}
	// this is an ordinary variable
	e.Set("TITLE", "Example of merging Grouped Variables")

	// these are grouped variables
	// note the naming convention: GROUP-NAME__VARIABLE-NAME__VARIABLE-INDEX
	e.Set("PORT__NAME__1", "Standard TCP")
	e.Set("PORT__DESC__1", "The standard port")
	e.Set("PORT__VALUE__1", "80")

	e.Set("PORT__NAME__2", "Alternative TCP")
	e.Set("PORT__DESC__2", "An alternative http port")
	e.Set("PORT__VALUE__2", "8080")

	e.Set("PORT__NAME__3", "Standard Encrypted")
	e.Set("PORT__DESC__3", "HTTPS port")
	e.Set("PORT__VALUE__3", "443")

	e.Set("URI__NAME__1", "URI 1")
	e.Set("URI__VALUE__1", "www.hhhh.com")

	e.Set("URI__NAME__2", "URI 2")
	e.Set("URI__VALUE__2", "www.hhwedwdehh.com")

	m, _ := NewTemplMerger()
	err := m.LoadTemplates([]string{"test/sample_using_operators.yaml.art"})
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = m.Merge(e)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, bytes := range m.Files() {
		fmt.Println(string(bytes))
	}
}

func TestMerge(t *testing.T) {
	env, err := NewEnVarFromFile("test/.env")
	if err != nil {
		t.Fatalf(err.Error())
	}
	m, _ := NewTemplMerger()
	err = m.LoadTemplates([]string{"test/sample_using_functions.yaml.art"})
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = m.Merge(env)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, bytes := range m.Files() {
		fmt.Println(string(bytes))
	}
}

func TestMergeUsingCustomCtx(t *testing.T) {
	env, err := NewEnVarFromFile("test/.env")
	if err != nil {
		t.Fatalf(err.Error())
	}
	m, _ := NewTemplMerger()
	temp, err := os.ReadFile("test/sample_using_functions.yaml.art")
	if err != nil {
		t.Fatalf(err.Error())
	}
	templs := map[string]string{}
	templs["temp1.art"] = string(temp[:])
	m.LoadStringTemplates(templs)
	ctx, err := NewCustomContext(env)
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = m.MergeWithCtx(ctx)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, bytes := range m.Files() {
		fmt.Println(string(bytes))
	}
}

func TestMergeFromString(t *testing.T) {
	m, _ := NewTemplMerger()
	templs := map[string]string{}
	os.Setenv("NAME_1", "AAA")
	os.Setenv("NAME_2", "BBB")
	templs["temp1.art"] = `HELLO {{$"NAME_1"}}`
	templs["temp2.art"] = `HELLO {{$"NAME_2"}}`
	if err := m.LoadStringTemplates(templs); err != nil {
		t.Fatalf(err.Error())
	}
	ctx, err := NewCustomContext(NewEnVarFromSlice([]string{"NAME_1=AAA", "NAME_2=BBB"}))
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = m.MergeWithCtx(ctx)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, bytes := range m.Files() {
		fmt.Println(string(bytes))
	}
}

type CustomContext struct {
	Context
}

func NewCustomContext(env *Envar) (TemplateContext, error) {
	ctx := &CustomContext{}
	ctx.Env = env
	ctx.Loader = NewLoader(env)
	ctx.Items = []Set{}
	return ctx, nil
}

func (c *CustomContext) FuncMap() template.FuncMap {
	fm := c.Context.FuncMap()
	// other functions here
	return fm
}
