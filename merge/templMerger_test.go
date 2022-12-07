/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package merge

import (
	"fmt"
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
	for _, bytes := range m.file {
		fmt.Println(string(bytes))
	}
}

func TestMerge(t *testing.T) {
	env, err := NewEnVarFromFile("test2/.env")
	if err != nil {
		t.Fatalf(err.Error())
	}
	m, _ := NewTemplMerger()
	err = m.LoadTemplates([]string{"test2/cluster_issuer_2.yaml.art"})
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = m.Merge(env)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, bytes := range m.file {
		fmt.Println(string(bytes))
	}
}
