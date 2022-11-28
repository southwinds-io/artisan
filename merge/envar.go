/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package merge

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"southwinds.dev/artisan/conf"
	"southwinds.dev/artisan/core"
	"strconv"
	"strings"
)

type Envar struct {
	vars map[string]string
}

func (e *Envar) Get(key string) string {
	if val, ok := e.vars[key]; ok {
		return val
	}
	return ""
}

func (e *Envar) Set(key, value string) {
	e.vars[key] = value
}

// Group used by golang text.Template to return a map of key / values for vars that whose base name is the same
// but have been suffixed with an incremental index number
func (e *Envar) Group(groupName reflect.Value) reflect.Value {
	result := make(map[string]string)
	for name, value := range e.vars {
		i := strings.LastIndex(name, "_")
		if i > 0 {
			prefix := name[0:i]
			suffix := name[i+1:]
			_, err := strconv.ParseInt(suffix, 10, 16)
			// if the parsing works it is an index
			if err == nil && prefix == groupName.String() {
				result[name] = value
			}
		}
	}
	return reflect.ValueOf(result)
}

func NewEnVarFromMap(v map[string]string) *Envar {
	return &Envar{
		vars: v,
	}
}

func NewEnVarFromFile(envFile string) (*Envar, error) {
	var outMap = make(map[string]string)
	file := core.ToAbs(envFile)
	data, err := os.ReadFile(file)
	// if it managed to find the env file load it
	// otherwise skip it
	content := strings.Split(string(data), "\n")
	if err == nil {
		for _, line := range content {
			// skips comments
			if strings.HasPrefix(strings.Trim(line, " "), "#") ||
				len(strings.Trim(line, " ")) == 0 ||
				strings.HasPrefix(strings.Trim(line, " "), "\r") ||
				strings.HasPrefix(strings.Trim(line, " "), "\n") {
				continue
			}
			// Splitting exactly on 2 strings
			// example: VAR=test= Result: val[0] is VAR val[1] is test=
			// Required for cases where value contains = sign like base64 values
			keyValue := strings.SplitN(line, "=", 2)

			outMap[keyValue[0]] = removeTrail(keyValue[1])
		}
	} else {
		core.Debug("cannot load env file: %s", err.Error())
	}
	core.Debug("loaded environment file: %s\n", envFile)
	return &Envar{
		vars: outMap,
	}, nil
}

// remove trailing \r or \n or \r\n
func removeTrail(value string) string {
	// case 1 => \r
	// case 2 => \n
	// case 3 => \r\n
	value = strings.Trim(value, "\r")
	value = strings.Trim(value, "\n")
	value = strings.Trim(value, "\r")
	return value
}

func NewEnVarFromSlice(v []string) *Envar {
	ev := &Envar{
		vars: make(map[string]string),
	}
	for _, s := range v {
		kv := strings.SplitN(s, "=", 2)
		ev.Add(kv[0], kv[1])
	}
	return ev
}

func NewEnVarEmpty() *Envar {
	return NewEnVarFromSlice([]string{})
}
func (e *Envar) Add(key, value string) {
	e.vars[key] = value
}

func (e *Envar) Slice() []string {
	var result []string
	for k, v := range e.vars {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func (e *Envar) Append(v map[string]string) conf.Configuration {
	var result = make(map[string]string)
	result = e.vars
	for k, v := range v {
		result[k] = v
	}
	e.Replace()
	return NewEnVarFromMap(result)
}

func (e *Envar) Merge(env conf.Configuration) {
	for key, value := range env.Vars() {
		e.vars[key] = value
	}
	e.Replace()
}

func (e *Envar) MergeMap(env map[string]string) {
	for key, value := range env {
		e.vars[key] = value
	}
	e.Replace()
}

func (e *Envar) Vars() map[string]string {
	return e.vars
}

func (e *Envar) String() string {
	buffer := bytes.Buffer{}
	for key, value := range e.vars {
		buffer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}
	return buffer.String()
}

// Replace any env variable in the internal map with the value
func (e *Envar) Replace() {
	for key, value := range e.vars {
		e.vars[key] = conf.ReplaceVar(value, e)
	}
}
