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

package flow

import (
	"strings"

	"southwinds.dev/artisan/data"
)

type Step struct {
	Name          string      `yaml:"name" json:"name"`
	Description   string      `yaml:"description,omitempty" json:"description,omitempty"`
	Function      string      `yaml:"function,omitempty" json:"function,omitempty"`
	Package       string      `yaml:"package,omitempty" json:"package,omitempty"`
	PackageSource string      `yaml:"source,omitempty" json:"source,omitempty"`
	Input         *data.Input `yaml:"input,omitempty" json:"input,omitempty"`
	Privileged    bool        `yaml:"privileged" json:"privileged"`
}

func (s *Step) surveyBuildfile(requiresGitSource bool) bool {
	// requires a git source, it defines a function without package
	return requiresGitSource && len(s.Function) > 0 && len(s.Package) == 0
}

func (s *Step) surveyManifest() bool {
	// defines a function and a package or in the case of a package merge a function is not required but package and source = merge exist
	return (len(s.Function) > 0 && len(s.Package) > 0) || (len(s.Package) > 0 && len(s.Function) == 0 && strings.ToLower(s.PackageSource) == "merge")
}
