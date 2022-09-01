/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
