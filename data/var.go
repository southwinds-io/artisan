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

type VerifyHandler func(name *core.PackageName, seal *Seal, path string, authorisedAuthors []string, sign bool) error
type RunHandler func(name *core.PackageName, fx string, seal *Seal) error
