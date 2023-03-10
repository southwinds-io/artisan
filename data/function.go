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
	"southwinds.dev/artisan/conf"
	"southwinds.dev/artisan/core"
)

type Function struct {
	// the name of the function
	Name string `yaml:"name"`
	// the description for the function
	Description string `yaml:"description,omitempty"`
	// a set of environment variables required by the run commands
	Env map[string]string `yaml:"env,omitempty"`
	// the commands to be executed by the function
	Run []string `yaml:"run,omitempty"`
	// is this function to be available in the manifest
	Export *bool `yaml:"export,omitempty"`
	// defines any bindings to inputs required to run this function
	Input *InputBinding `yaml:"input,omitempty"`
	// the runtime to run this function
	Runtime string   `yaml:"runtime,omitempty"`
	Credits int      `yaml:"credits,omitempty"`
	Network *Network `json:"network,omitempty"`
}

type Access string

const (
	AccessPublic   Access = "public"
	AccessInternal Access = "internal"
	AccessPrivate  Access = "private"
)

// InputBinding list the names of the inputs required by a function
type InputBinding struct {
	Var    []string `yaml:"var"`
	Secret []string `yaml:"secret"`
	Key    []string `yaml:"key"`
	File   []string `yaml:"file"`
}

// GetEnv gets a slice of string with each element containing key=value
func (f *Function) GetEnv() map[string]string {
	return f.Env
}

// Survey all missing variables in the function
// pass in any available environment variables so that they are not surveyed
func (f *Function) Survey(env conf.Configuration) conf.Configuration {
	// merges the function environment with the passed in environment
	for k, v := range f.Env {
		env.Set(k, v)
	}
	// attempt to merge any environment variable in the run commands
	// run the merge in interactive mode so that any variables not available in the build file environment are surveyed
	_, updatedEnvironment := core.MergeEnvironmentVars(f.Run, env, true)
	return updatedEnvironment
}
