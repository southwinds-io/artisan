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
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"runtime"
	"southwinds.dev/artisan/conf"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/merge"

	"path/filepath"
	"strings"
)

// BuildFile structure of build.yaml file
type BuildFile struct {
	// internal, the path from where the buildfile is loaded
	path string
	// the URI of the Git repo
	GitURI string `yaml:"git_uri,omitempty"`
	// the runtime to use to run functions
	Runtime string `yaml:"runtime,omitempty"`
	// the environment variables that apply to the build
	// any variables defined at this level will be available to all build profiles
	// in addition, the defined variables are added on top of the existing environment
	Env map[string]string `yaml:"env,omitempty"`
	// a list of labels to be added to the package seal
	// they should be used to document key aspects of the package in a generic way
	Labels map[string]string `yaml:"labels,omitempty"`
	// any input required by functions
	Input *Input `yaml:"input,omitempty"`
	// a list of build configurations in the form of labels, commands to run and environment variables
	Profiles []*Profile `yaml:"profiles,omitempty"`
	// a list of functions containing a list of commands to execute
	Functions []*Function `yaml:"functions"`
	// include other build files
	Includes []interface{} `yaml:"includes"`
	SKU      string        `yaml:"sku"`
}

func (b *BuildFile) GetEnv() map[string]string {
	return b.Env
}

func (b *BuildFile) ExportFxs() bool {
	for _, function := range b.Functions {
		if function.Export != nil && *function.Export {
			return true
		}
	}
	return false
}

// DefaultProfile return the default profile if exists
func (b *BuildFile) DefaultProfile() *Profile {
	for _, profile := range b.Profiles {
		if profile.Default {
			return profile
		}
	}
	return nil
}

// Fx return the function in the build file specified by its name
func (b *BuildFile) Fx(name string) *Function {
	for _, fx := range b.Functions {
		if fx.Name == name {
			return fx
		}
	}
	return nil
}

type Profile struct {
	// the name of the profile
	Name string `yaml:"name"`
	// whether this is the default profile
	Default bool `yaml:"default"`
	// the name of the application
	Application string `yaml:"application"`
	// the type of license used by the application
	// if not empty, it is added to the package seal
	License string `yaml:"license"`
	// the type of technology used by the application that can be used to determine the tool chain to use
	// e.g. java, nodejs, golang, python, php, etc
	Type string `yaml:"type"`
	// the pipeline Icon
	Icon string `yaml:"icon"`
	// a set of labels associated with the profile
	Labels map[string]string `yaml:"labels"`
	// a set of environment variables required by the run commands
	Env map[string]string `yaml:"env"`
	// the commands to be executed to build the application
	Run []string `yaml:"run"`
	// the output of the build process, namely either a file or a folder, that has to be compressed
	// as part of the packaging process
	Target string `yaml:"target"`
	// merged target if existed, internal use only
	MergedTarget string
	X            []string `json:"x,omitempty"`
}

// GetEnv gets a slice of string with each element containing key=value
func (p *Profile) GetEnv() map[string]string {
	return p.Env
}

// Profile return the build profile specified by its name
func (b *BuildFile) Profile(name string) *Profile {
	for _, profile := range b.Profiles {
		if profile.Name == name {
			return profile
		}
	}
	return nil
}

// Survey all missing variables in the profile
func (p *Profile) Survey(bf *BuildFile) conf.Configuration {
	env := bf.Env
	// merges the profile environment with the passed in environment
	for k, v := range p.Env {
		env[k] = v
	}
	// attempt to merge any environment variable in the profile run commands
	// run the merge in interactive mode so that any variables not available in the build file environment are surveyed
	_, updatedEnvironment := core.MergeEnvironmentVars(p.Run, merge.NewEnVarFromMap(env), true)
	// attempt to merge any environment variable in the functions run commands
	for _, run := range p.Run {
		// if the run line has a function
		if ok, fxName := core.HasFunction(run); ok {
			// merge any variables on the function
			env = bf.Fx(fxName).Survey(merge.NewEnVarFromMap(env)).Vars()
		}
	}
	return updatedEnvironment
}

func LoadBuildFile(path string) (*BuildFile, error) {
	return LoadBuildFileWithEnv(path, nil)
}

func LoadBuildFileWithEnv(path string, ev conf.Configuration) (*BuildFile, error) {
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("cannot get absolute path for %s", path)
		}
		path = abs
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot load build file from %s: %s", path, err)
	}
	core.Debug("loaded: '%s'\ncontent:\n%s\n", path, string(bytes))
	buildFile := &BuildFile{
		path: path,
	}
	err = yaml.Unmarshal(bytes, buildFile)
	if err != nil {
		return nil, fmt.Errorf("syntax error in build file %s: %s", path, err)
	}
	if ok, validErr := buildFile.Validate(); !ok {
		return buildFile, validErr
	}
	if buildFile.Env == nil {
		buildFile.Env = map[string]string{}
	}
	if ev == nil {
		ev = merge.NewEnvVarOS()
	}
	ev.Merge(merge.NewEnVarFromMap(buildFile.Env))
	ev.Replace()
	buildFile.Env = ev.Vars()
	buildFile.Env[core.ArtOS] = runtime.GOOS
	buildFile.Env[core.ArtArch] = runtime.GOARCH
	buildFile.Env[core.ArtShell] = os.Getenv("SHELL")
	for _, include := range buildFile.Includes {
		switch i := include.(type) {
		case string:
			file, _ := filepath.Abs(filepath.Join(filepath.Dir(buildFile.path), i))
			child, err := LoadBuildFileWithEnv(file, ev)
			if err != nil {
				return nil, fmt.Errorf("build file include not found in path: %s, %s", file, err)
			}
			buildFile.Env = conf.MergeMaps(buildFile.Env, child.Env)
			buildFile.Profiles = append(buildFile.Profiles, child.Profiles...)
			buildFile.Functions = append(buildFile.Functions, child.Functions...)
			buildFile.Labels = conf.MergeMaps(buildFile.Labels, child.Labels)
		case []interface{}:
			incl := true
			for _, condition := range i[1:] {
				neq := strings.Split(condition.(string), "!=")
				if len(neq) == 2 {
					value := buildFile.Env[neq[0]]
					incl = incl && !strings.EqualFold(value, neq[1])
				} else {
					eq := strings.Split(condition.(string), "=")
					if len(eq) == 2 {
						value := buildFile.Env[eq[0]]
						incl = incl && strings.EqualFold(value, eq[1])
					}
				}
			}
			if incl {
				file, _ := filepath.Abs(filepath.Join(filepath.Dir(buildFile.path), i[0].(string)))
				child, err := LoadBuildFileWithEnv(file, ev)
				if err != nil {
					return nil, err
				}
				buildFile.Env = conf.MergeMaps(buildFile.Env, child.Env)
				buildFile.Profiles = append(buildFile.Profiles, child.Profiles...)
				buildFile.Functions = append(buildFile.Functions, child.Functions...)
				buildFile.Labels = conf.MergeMaps(buildFile.Labels, child.Labels)
			}
		}
	}

	return buildFile, nil
}

func (b *BuildFile) Validate() (bool, error) {
	// checks any binding has a corresponding input
	for _, fx := range b.Functions {
		if fx.Input != nil {
			if fx.Input.Var != nil {
				for _, v := range fx.Input.Var {
					// if no inputs were defined whatsoever or inputs were defined, but they do not match the bindings
					if b.Input == nil || (b.Input != nil && !b.Input.HasVar(v)) {
						return false, fmt.Errorf("function '%s' in build file '%s' has a Var binding '%s' but not corresponding Var definition has been defined in the build file Input section", fx.Name, b.path, v)
					}
				}
			}
			if fx.Input.Secret != nil {
				for _, s := range fx.Input.Secret {
					if !b.Input.HasSecret(s) && !strings.Contains(s, "ART_REG_USER") && !strings.Contains(s, "ART_REG_PWD") {
						return false, fmt.Errorf("function '%s' in build file '%s' has a Secret binding '%s' but not corresponding Secret definition has been defined in the build file Input section", fx.Name, b.path, s)
					}
				}
			}
		}
		if fx.Network != nil {
			if fx.Export == nil || !*fx.Export {
				return false, fmt.Errorf("network definition found in non exported function '%s'", fx.Name)
			}
			for _, group := range fx.Network.Groups {
				if len(strings.Split(group, ":")) != 4 {
					return false, fmt.Errorf("group declaration in network definition for function %s is incorrect: '%s', it must have 4 sections separated by ':' such as 'NAME:TAGS:MIN:MAX'", fx.Name, group)
				}
			}
			for _, rule := range fx.Network.Rules {
				if len(strings.Split(rule, ":")) != 3 {
					return false, fmt.Errorf("rule declaration in network definition for function %s is incorrect: '%s', it must have 3 sections separated by ':' such as 'NAME_FROM:NAME_TO:PROTOCOL/PORT'", fx.Name, rule)
				}
			}
		}
	}
	// check that any profiles do not have targets using "."
	for _, profile := range b.Profiles {
		if profile.Target == "." {
			return false, fmt.Errorf("invalid target for profile '%s': it cannot point to the same location of the build file: "+
				"the build file you use to build the package must not be the same as the one embedded in the package", profile.Name)
		}
	}
	return true, nil
}
