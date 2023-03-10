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
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"

	"path/filepath"
	"strings"

	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/registry"
)

// Credential a set of authentication credentials for a package registry
type Credential struct {
	User     string
	Password string
	Domain   string
}

type Flow struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	// a list of labels to document key aspects of the flow execution
	// for example using a target namespace if running in Kubernetes
	Labels      map[string]string `yaml:"labels" json:"labels"`
	Git         *Git              `yaml:"git,omitempty" json:"git,omitempty"`
	Steps       []*Step           `yaml:"steps" json:"steps"`
	Input       *data.Input       `yaml:"input,omitempty" json:"input,omitempty"`
	UseRuntimes *bool             `yaml:"use_runtimes,omitempty" json:"use_runtimes,omitempty"`

	artHome string
}

type Git struct {
	Uri      string `yaml:"git_uri" json:"git_uri"`
	Branch   string `yaml:"git_branch" json:"git_branch"`
	Login    string `yaml:"git_login,omitempty" json:"git_login,omitempty"`
	Password string `yaml:"git_password,omitempty" json:"git_password,omitempty"`
}

func LoadFlow(path, artHome string) (*Flow, error) {
	var err error
	if len(path) == 0 {
		return nil, fmt.Errorf("flow definition is required")
	}
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("cannot get absolute path for %s: %s", path, err)
		}
	}
	flowBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read flow definition %s: %s", path, err)
	}
	flow := new(Flow)
	err = yaml.Unmarshal(flowBytes, flow)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal flow definition %s: %s", path, err)
	}
	if flow.UseRuntimes == nil {
		b := true
		flow.UseRuntimes = &b
	}
	if flow.Labels == nil {
		flow.Labels = make(map[string]string)
	}
	flow.artHome = artHome
	return flow, nil
}

func NewFlow(flowJSONBytes []byte, artHome string) (*Flow, error) {
	flow := new(Flow)
	err := json.Unmarshal(flowJSONBytes, flow)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal flow definition %s", err)
	}

	if flow.UseRuntimes == nil {
		b := true
		flow.UseRuntimes = &b
	}
	flow.artHome = artHome
	return flow, nil
}

// Map get the input in map format
func (f *Flow) Map() (map[string]interface{}, error) {
	bytes, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal input map: %s", err)
	}
	var input map[string]interface{}
	err = json.Unmarshal(bytes, &input)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal input bytes: %s", err)
	}
	return input, err
}

func (f *Flow) StepByFx(fxName string) *Step {
	for _, step := range f.Steps {
		if step.Function == fxName {
			return step
		}
	}
	return nil
}

func (f *Flow) RequiresGitSource() bool {
	var useGitSource, usePackageSource bool
	for _, step := range f.Steps {
		// function only - needs a git source
		if len(step.Package) == 0 && len(step.Function) > 0 {
			useGitSource = true
		}
		// requires a package source
		if len(step.Package) > 0 && len(step.PackageSource) > 0 && len(step.Function) > 0 {
			usePackageSource = true
		}
	}
	// git source is required if it is not using a package source
	return useGitSource && !usePackageSource
}

func (f *Flow) RequiresSecrets() bool {
	for _, step := range f.Steps {
		if step.Input != nil && step.Input.Secret != nil {
			return true
		}
	}
	return false
}

func (f *Flow) RequiresFile() bool {
	for _, step := range f.Steps {
		if step.Input != nil && step.Input.File != nil {
			return true
		}
	}
	return false
}

// GetInputDefinition retrieve all input data required by the flow without values
// interactive mode is off - gets definition only
func (f *Flow) GetInputDefinition(b *data.BuildFile, env *merge.Envar) (*data.Input, error) {
	result := &data.Input{
		Secret: make([]*data.Secret, 0),
		Var:    make([]*data.Var, 0),
	}
	local := registry.NewLocalRegistry(f.artHome)

	for _, step := range f.Steps {
		// if a function is defined without a package and the source is not a package
		if step.surveyBuildfile(f.RequiresGitSource()) {
			// check a build file has been specified
			if b == nil {
				core.RaiseErr("flow '%s' requires a build.yaml", f.Name)
			}
			// surveys the build.yaml for variables
			i, err := data.SurveyInputFromBuildFile(step.Function, b, false, true, env, f.artHome)
			if err != nil {
				return nil, err
			}
			if i == nil {
				i = &data.Input{
					Secret: make([]*data.Secret, 0),
					Var:    make([]*data.Var, 0),
					File:   make([]*data.File, 0),
				}
			}

			// add GIT variables
			addGitVariables(i)
			result.Merge(i)
		} else if step.surveyManifest() {
			// surveys the package manifest for variables
			name, err := core.ParseName(step.Package)
			i18n.Err(f.artHome, err, i18n.ERR_INVALID_PACKAGE_NAME)
			manif := local.GetManifest(name)
			if manif == nil {
				core.RaiseErr("manifest for package '%s' not found", name)
			}
			i, err := data.SurveyInputFromManifest(f.Name, step.Name, step.PackageSource, name.Domain, step.Function, manif, false, true, env, f.artHome)
			if err != nil {
				return nil, err
			}
			err = i.SurveyRegistryCreds(f.Name, step.Name, step.PackageSource, name.Domain, false, true, env)
			if err != nil {
				return nil, err
			}
			result.Merge(i)
		} else {
			flowHealthCheck(f, step)
		}
		// try augment the result with default values in the build.yaml
		if b != nil {
			// do we have an input section?
			if b.Input != nil && b.Input.Var != nil {
				for _, v := range b.Input.Var {
					for _, v2 := range result.Var {
						if v.Name == v2.Name && len(v.Default) > 0 {
							v2.Default = v.Default
						}
					}
				}
			} else {
				core.WarningLogger.Printf("cannot survey build file: no input section found\n")
			}
		}
	}
	return result, nil
}

func (f *Flow) JsonBytes() ([]byte, error) {
	data, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Flow) Step(name string) *Step {
	for _, step := range f.Steps {
		if step.Name == name {
			return step
		}
	}
	return nil
}

func (f *Flow) IsValid() error {

	if len(f.Steps) == 0 {
		return errors.New("step is missing for this flow")
	}

	if f.RequiresGitSource() {
		return f.validateGitSource()
	} else {
		return f.validateNonGitSource()
	}

	return nil
}

func (f *Flow) validateGitSource() error {

	if f.Git == nil {
		return errors.New("git env details [ 'GIT_URI', 'GIT_BRANCH' and optional 'GIT_LOGIN', 'GIT_PASSWORD' ]missing for flow with git source ")
	}

	if len(f.Git.Uri) == 0 {
		return errors.New("git env 'GIT_URI' missing for flow with git source")
	}

	if len(f.Git.Branch) == 0 {
		return errors.New("git env 'GIT_BRANCH' missing for flow with git source")
	}

	// if git source is requred then flow steps should not define a source attribute.
	for _, s := range f.Steps {
		if len(s.PackageSource) != 0 {
			return errors.New("flow with git source must not define 'source' attribute in all the step")
		}
	}
	return nil
}

func (f *Flow) validateNonGitSource() error {
	step := f.Steps[0]

	// if git source is not requred then first step in flow steps
	// must have package source set to "create"
	if !strings.EqualFold(step.PackageSource, "create") {
		return errors.New("first step within a flow must have package source type as create")
	}

	// if the step is read, then the package name should be same as
	// any previous create / merge package name.
	previousPackage := ""
	// merge should be done between different package source
	for _, s := range f.Steps {
		if strings.EqualFold(s.PackageSource, "create") || strings.EqualFold(s.PackageSource, "merge") {
			previousPackage = s.Package
		} else if strings.EqualFold(s.PackageSource, "read") && !strings.EqualFold(s.Package, previousPackage) {
			return errors.New("when step has 'read' source type, then package name must match with package name defined in previous step with source type as 'create' or 'merge'")
		}
	}
	return nil
}

func addGitVariables(i *data.Input) {
	i.Var = append(i.Var, &data.Var{
		Name:        "GIT_URI",
		Description: GitUriDesc,
		Required:    true,
		Type:        "url",
	})
	i.Var = append(i.Var, &data.Var{
		Name:        "GIT_BRANCH",
		Description: GitBranchDesc,
		Required:    false,
		Type:        "string",
	})
	i.Var = append(i.Var, &data.Var{
		Name:        "GIT_USER",
		Description: GitUserDesc,
		Required:    false,
		Type:        "string",
	})
	i.Var = append(i.Var, &data.Var{
		Name:        "GIT_PASSWORD",
		Description: GitPasswordDesc,
		Required:    false,
		Type:        "string",
	})
}
