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
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v2"
	"io"
	"southwinds.dev/artisan/conf"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/merge"

	"os"
	"path"
	"sort"
	"strings"
)

// Input describes exported input information required by functions or runtimes
type Input struct {
	// required by configuration files
	File Files `yaml:"file,omitempty" json:"file,omitempty"`
	// required string value secrets
	Secret Secrets `yaml:"secret,omitempty" json:"secret,omitempty"`
	// required variables
	Var Vars `yaml:"var,omitempty" json:"var,omitempty"`
}

func (i *Input) HasVarBinding(binding string) bool {
	for _, variable := range i.Var {
		if variable.Name == binding {
			return true
		}
	}
	return false
}

func (i *Input) HasSecretBinding(binding string) bool {
	for _, secret := range i.Secret {
		if secret.Name == binding {
			return true
		}
	}
	return false
}

func (i *Input) HasVar(name string) bool {
	if i.Var != nil {
		for _, v := range i.Var {
			if v.Name == name {
				return true
			}
		}
	}
	return false
}

func (i *Input) HasSecret(name string) bool {
	if i.Secret != nil {
		for _, s := range i.Secret {
			if s.Name == name {
				return true
			}
		}
	}
	return false
}

func (i *Input) SurveyRegistryCreds(flowName, stepName, packageSource, domain string, prompt, defOnly bool, env conf.Configuration) error {
	if packageSource != "read" {
		// check for art_reg_user
		userName := fmt.Sprintf("%s_%s_OXART_REG_USER", NormInputName(flowName), NormInputName(stepName))
		if !i.HasSecret(userName) {
			userSecret := &Secret{
				Name:        userName,
				Description: fmt.Sprintf("the username to authenticate with the registry at %s'", domain),
			}
			if !defOnly {
				if err := EvalSecret(userSecret, prompt, env); err != nil {
					return err
				}
			}
			i.Secret = append(i.Secret, userSecret)
		}
		// check for art_reg_pwd
		pwd := fmt.Sprintf("%s_%s_OXART_REG_PWD", NormInputName(flowName), NormInputName(stepName))
		if !i.HasSecret(pwd) {
			pwdSecret := &Secret{
				Name:        pwd,
				Description: fmt.Sprintf("the password to authenticate with the registry at '%s'", domain),
			}
			if !defOnly {
				if err := EvalSecret(pwdSecret, prompt, env); err != nil {
					return err
				}
			}
			i.Secret = append(i.Secret, pwdSecret)
		}
	}
	return nil
}

func (i *Input) Env() *merge.Envar {
	env := make(map[string]string)
	for _, v := range i.Var {
		env[v.Name] = v.Value
	}
	for _, s := range i.Secret {
		env[s.Name] = s.Value
	}
	return merge.NewEnVarFromMap(env)
}

// Merge the passed in input with the current input
func (i *Input) Merge(in *Input) {
	if in == nil {
		// nothing to merge
		return
	}
	for _, v := range in.Var {
		// dedup
		found := false
		for _, iV := range i.Var {
			// if the variable to be merged is already in the source
			if iV.Name == v.Name {
				found = true
				break
			}
		}
		if !found {
			i.Var = append(i.Var, v)
		}
	}
	sort.Sort(i.Var)
	for _, s := range in.Secret {
		// dedup
		found := false
		for _, iS := range i.Secret {
			// if the secret to be merged is already in the source
			if iS.Name == s.Name {
				found = true
				break
			}
		}
		if !found {
			i.Secret = append(i.Secret, s)
		}
	}
	sort.Sort(i.Secret)
}

func (i *Input) VarExist(name string) bool {
	for _, v := range i.Var {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (i *Input) SecretExist(name string) bool {
	for _, s := range i.Secret {
		if s.Name == name {
			return true
		}
	}
	return false
}

// SurveyInputFromBuildFile extracts the build file Input that is relevant to a function (using its bindings)
func SurveyInputFromBuildFile(fxName string, buildFile *BuildFile, prompt, defOnly bool, env conf.Configuration, artHome string) (*Input, error) {
	if buildFile == nil {
		return nil, fmt.Errorf("build file is required")
	}
	// get the build file function to inspect
	fx := buildFile.Fx(fxName)
	if fx == nil {
		return nil, fmt.Errorf("function '%s' cannot be found in build file", fxName)
	}
	return getBoundInput(fx.Input, buildFile.Input, prompt, defOnly, env, artHome)
}

// SurveyInputFromManifest extracts the package manifest Input in an exported function
func SurveyInputFromManifest(flowName, stepName, packageSource, domain string, fxName string, manifest *Manifest, prompt, defOnly bool, env conf.Configuration, artHome string) (*Input, error) {
	var input *Input
	// get the function in the manifest
	fx := manifest.Fx(fxName)
	if fx != nil {
		input = fx.Input
	} else if fx == nil && packageSource == "merge" {
		// this is the case of a package merge where there is not any need to survey inputs just perform a straight merge
		// of source
		input = &Input{
			Secret: make([]*Secret, 0),
			Var:    make([]*Var, 0),
			File:   make([]*File, 0),
		}
	} else {
		// requires a function to exist
		return nil, fmt.Errorf("function '%s' does not exist in or has not been exported", fxName)
	}
	// first evaluates the existing inputs
	input, err := evalInput(input, prompt, defOnly, env, artHome)
	if err != nil {
		return nil, err
	}
	// then add registry credential inputs
	err = input.SurveyRegistryCreds(flowName, stepName, packageSource, domain, prompt, defOnly, env)
	return input, err
}

// NormInputName ensure the passed in name is formatted as a valid environment variable name
func NormInputName(name string) string {
	result := strings.Replace(strings.ToUpper(name), "-", "_", -1)
	result = strings.Replace(result, ".", "_", -1)
	result = strings.Replace(result, "/", "_", -1)
	return result
}

func SurveyInputFromURI(uri string, prompt, defOnly bool, env conf.Configuration, artHome string) (*Input, error) {
	response, err := core.Get(uri, "", "")
	if err != nil {
		return nil, fmt.Errorf("cannot fetch runtime manifest")
	}
	body, err := io.ReadAll(response.Body)
	core.CheckErr(err, "cannot read runtime manifest http response")
	// need a wrapper object for the input for the unmarshaller to work so using buildfile
	var buildFile = new(BuildFile)
	err = yaml.Unmarshal(body, buildFile)
	return evalInput(buildFile.Input, prompt, defOnly, env, artHome)
}

func evalInput(input *Input, interactive, defOnly bool, env conf.Configuration, artHome string) (*Input, error) {
	// makes a shallow copy of the input
	result := *input
	// collect values from command line interface
	for _, v := range result.Var {
		if !defOnly {
			if err := EvalVar(v, interactive, env); err != nil {
				return nil, err
			}
		}
	}
	for _, secret := range result.Secret {
		if !defOnly {
			if err := EvalSecret(secret, interactive, env); err != nil {
				return nil, err
			}
		}
	}
	for _, file := range result.File {
		if !defOnly {
			if err := EvalFile(file, interactive, env, artHome); err != nil {
				return nil, err
			}
		}
	}
	// return pointer to new object
	return &result, nil
}

func EvalVar(inputVar *Var, prompt bool, env conf.Configuration) error {
	// do not evaluate it if there is already a value
	if len(inputVar.Value) > 0 {
		return nil
	}
	// check if there is an env variable
	varValue := env.Get(inputVar.Name)
	// if so
	if len(varValue) > 0 {
		// set the var value to the env variable's
		inputVar.Value = varValue
	} else if prompt {
		// survey the var value
		surveyVar(inputVar)
	} else {
		if inputVar.Required {
			// otherwise error
			return fmt.Errorf("%s is required", inputVar.Name)
		}
	}
	return nil
}

func EvalSecret(inputSecret *Secret, prompt bool, env conf.Configuration) error {
	// do not evaluate it if there is already a value
	if len(inputSecret.Value) > 0 {
		return nil
	}
	// check if there is an env variable
	secretValue := env.Get(inputSecret.Name)
	// if so
	if len(secretValue) > 0 {
		// set the secret value to the env variable's
		inputSecret.Value = secretValue
	} else if prompt {
		// survey the secret value
		surveySecret(inputSecret)
	} else {
		if inputSecret.Required {
			// otherwise error
			return fmt.Errorf("%s is required", inputSecret.Name)
		}
	}
	return nil
}

func EvalFile(inputFile *File, prompt bool, env conf.Configuration, artHome string) error {
	// do not evaluate it if there is already a value
	if len(inputFile.Content) > 0 {
		return nil
	}
	// check if there is an env variable
	filePath := env.Get(inputFile.Name)
	// if so
	if len(filePath) > 0 {
		// load the correct key using the provided path
		loadFileFromPath(inputFile, filePath, artHome)
	} else if len(inputFile.Path) > 0 {
		// load the correct key using the provided path
		loadFileFromPath(inputFile, inputFile.Path, artHome)
	} else if prompt {
		surveyFile(inputFile, artHome)
	} else {
		return fmt.Errorf("%s is required", inputFile.Name)
	}
	return nil
}

func (i *Input) ToEnvFile() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("# ===================================================\n")
	buf.WriteString("# environment configuration file generated by artisan\n")
	buf.WriteString("# ===================================================\n")
	for ix, v := range i.Var {
		if ix == 0 {
			buf.WriteString("# VARIABLES\n")
			buf.WriteString("# ---------\n")
		}
		buf.WriteString(toEnvComments(v.Description))
		if len(v.Default) > 0 {
			buf.WriteString(fmt.Sprintf("%s=\"%s\"\n\n", v.Name, v.Default))
		} else {
			buf.WriteString(fmt.Sprintf("%s=\"\"\n\n", v.Name))
		}
	}
	for ix, s := range i.Secret {
		if ix == 0 {
			buf.WriteString("# SECRETS\n")
			buf.WriteString("# -------\n")
		}
		buf.WriteString(toEnvComments(s.Description))
		buf.WriteString(fmt.Sprintf("%s=\"\"\n", s.Name))
	}
	return buf.Bytes()
}

func toEnvComments(value string) string {
	out := new(bytes.Buffer)
	values := strings.Split(value, "\n")
	for _, v := range values {
		if len(v) > 0 {
			out.WriteString(fmt.Sprintf("# %s\n", v))
		}
	}
	return out.String()
}

// extract any Input data from the source that have a binding
func getBoundInput(fxInput *InputBinding, sourceInput *Input, prompt, defOnly bool, env conf.Configuration, artHome string) (*Input, error) {
	result := &Input{
		Secret: make([]*Secret, 0),
		Var:    make([]*Var, 0),
		File:   make([]*File, 0),
	}
	// if no bindings then return an empty Input
	if fxInput == nil {
		return result, nil
	}
	// collects exported vars
	for _, varBinding := range fxInput.Var {
		for _, variable := range sourceInput.Var {
			if variable.Name == varBinding {
				result.Var = append(result.Var, variable)
				// if not definition only it should evaluate the variable
				if !defOnly {
					if err := EvalVar(variable, prompt, env); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	// collect exported secrets
	for _, secretBinding := range fxInput.Secret {
		for _, secret := range sourceInput.Secret {
			if secret.Name == secretBinding {
				result.Secret = append(result.Secret, secret)
				// if not definition only it should evaluate the secret
				if !defOnly {
					if err := EvalSecret(secret, prompt, env); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	for _, fileBinding := range fxInput.File {
		for _, file := range sourceInput.File {
			if file.Name == fileBinding {
				result.File = append(result.File, file)
				// if not definition only it should evaluate the file
				if !defOnly {
					if err := EvalFile(file, prompt, env, artHome); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return result, nil
}

func surveyVar(variable *Var) {
	// check if an env var has been set
	envVal := os.Getenv(variable.Name)
	// if so, skip survey
	if len(envVal) > 0 {
		return
	}
	// otherwise prompts the user to enter it
	var validator survey.Validator
	desc := ""
	// if a description is available use it
	if len(variable.Description) > 0 {
		desc = variable.Description
	}
	// prompt for the value
	prompt := &survey.Input{
		Message: fmt.Sprintf("var => %s (%s):", variable.Name, desc),
		Default: variable.Default,
	}
	// if required then add required validator
	if variable.Required {
		validator = survey.ComposeValidators(survey.Required)
	}
	// add type validators
	switch strings.ToLower(variable.Type) {
	case "path":
		validator = survey.ComposeValidators(validator, core.IsPath)
	case "uri":
		validator = survey.ComposeValidators(validator, core.IsURI)
	case "name":
		validator = survey.ComposeValidators(validator, core.IsPackageName)
	default:
		validator = nil
		if len(variable.Type) > 0 {
			core.InfoLogger.Printf("input '%s' has a type of '%s' which is not valid\n"+
				"	valid types are path, URI or name\n"+
				"	skipping type validation", variable.Name, variable.Type)
		}
	}
	var askOpts survey.AskOpt
	if validator != nil {
		askOpts = survey.WithValidator(validator)
	}
	core.HandleCtrlC(survey.AskOne(prompt, &variable.Value, askOpts))
}

func surveySecret(secret *Secret) {
	var validator survey.Validator
	// check if an env var has been set
	envVal := os.Getenv(secret.Name)
	// if so, skip survey
	if len(envVal) > 0 {
		return
	}
	desc := ""
	// if a description is available use it
	if len(secret.Description) > 0 {
		desc = secret.Description
	}
	// prompt for the value
	prompt := &survey.Password{
		Message: fmt.Sprintf("secret => %s (%s):", secret.Name, desc),
	}
	// if required then add required validator
	if secret.Required {
		validator = survey.ComposeValidators(survey.Required)
	}
	var askOpts survey.AskOpt
	if validator != nil {
		askOpts = survey.WithValidator(validator)
	}
	core.HandleCtrlC(survey.AskOne(prompt, &secret.Value, askOpts))
}

func surveyFile(file *File, artHome string) error {
	// check if an env var has been set
	envVal := os.Getenv(file.Name)
	// if so, skip survey
	if len(envVal) > 0 {
		// load the file using the env var path value specified
		return loadFileFromPath(file, envVal, artHome)
	}
	if len(file.Path) > 0 {
		// load the file using the path value specified in the manifest / buildfile
		return loadFileFromPath(file, file.Path, artHome)
	}
	desc := ""
	// if a description is available use it
	if len(file.Description) > 0 {
		desc = file.Description
	}
	// takes default path from input
	defaultPath := file.Path
	// prompt for the value
	prompt := &survey.Input{
		Message: fmt.Sprintf("File => path to %s (%s):", file.Name, desc),
		Default: defaultPath,
		Help:    "the path to the file to load from the Artisan registry",
	}
	var filePath string
	// survey the key path
	core.HandleCtrlC(survey.AskOne(prompt, &filePath, nil))
	// load the keys
	return loadFileFromPath(file, filePath, artHome)
}

// load the file content in the file object using the passed in file path
func loadFileFromPath(file *File, filePath, artHome string) error {
	var (
		contentBytes []byte
		err          error
		path         = path.Join(core.FilesPath(artHome), filePath)
	)
	contentBytes, err = os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot load file '%s' from registry", path)
	}
	file.Content = string(contentBytes)
	return nil
}
