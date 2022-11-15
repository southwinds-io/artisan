/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
	var si string = list[i].Name
	var sj string = list[j].Name
	var si_lower = strings.ToLower(si)
	var sj_lower = strings.ToLower(sj)
	if si_lower == sj_lower {
		return si < sj
	}
	return si_lower < sj_lower
}

type VProc func(name *core.PackageName, s *Seal, p string) error
type RProc func(name *core.PackageName, f string, seal *Seal) error
