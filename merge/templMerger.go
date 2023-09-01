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

package merge

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"southwinds.dev/artisan/core"

	"regexp"
	"strings"
	"text/template"
)

// TemplMerger merge artisan templates using artisan inputs
type TemplMerger struct {
	regex      *regexp.Regexp
	rexVar     *regexp.Regexp
	rexRange   *regexp.Regexp
	rexItem    *regexp.Regexp
	rexItemEq  *regexp.Regexp
	rexItemNeq *regexp.Regexp
	template   map[string][]byte
	file       map[string][]byte
}

// NewTemplMerger create a new instance of the template merger to merge files
func NewTemplMerger() (*TemplMerger, error) {
	// for tem templates:
	// parse ${NAME} vars
	regex, err := regexp.Compile("\\${(?P<NAME>[^}]*)}")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %s\n", err)
	}
	// for art templates:
	// parse {{ $ "NAME"  }} vars
	rexVar, err := regexp.Compile("{{[\\s]*\\$[\\s]*\"(?P<NAME>[\\w]+)\"[\\s]*}}")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %s\n", err)
	}
	// parse {{ range => "GROUP_NAME" }}
	rexRange, err := regexp.Compile("{{[\\s]*range[\\s]*=>[\\s]*\"(?P<GROUP>[\\w]+)\"[\\s]*}}")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %s\n", err)
	}
	// parse {{ % "NAME" }}
	rexItem, err := regexp.Compile("{{[\\s]*\\%[\\s]*\"(?P<ITEM>[\\w]+)\"[\\s]*}}")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %s\n", err)
	}
	// parse {{ if %= "NAME" "VALUE" }}
	rexItemEq, err := regexp.Compile(`{{\s*if\s%=\s*"(?P<NAME>[^"]*)"\s*"(?P<VALUE>[^"]*)"\s*}}`)
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %s\n", err)
	}
	// parse {{ if %!= "NAME" "VALUE" }}
	rexItemNeq, err := regexp.Compile(`{{\s*if\s%!=\s*"(?P<NAME>[^"]*)"\s*"(?P<VALUE>[^"]*)"\s*}}`)
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %s\n", err)
	}
	return &TemplMerger{
		regex:      regex,
		rexVar:     rexVar,
		rexItem:    rexItem,
		rexItemEq:  rexItemEq,
		rexItemNeq: rexItemNeq,
		rexRange:   rexRange,
	}, nil
}

// LoadTemplates load the template files to use
func (t *TemplMerger) LoadTemplates(files []string) error {
	m := make(map[string][]byte)
	for _, file := range files {
		// ensure the template path is absolute
		path, err := core.AbsPath(file)
		if err != nil {
			return fmt.Errorf("path '%s' cannot be converted to absolute path: %s\n", file, err)
		}
		// read the file content
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %s\n", file, err)
		}
		m[path] = t.transpileOperators(fileBytes)
	}
	t.template = m
	return nil
}

func (t *TemplMerger) LoadTemplatesBytes(files map[string][]byte) error {
	m := make(map[string][]byte)
	for path, fileBytes := range files {
		m[path] = t.transpileOperators(fileBytes)
	}
	t.template = m
	return nil
}

func (t *TemplMerger) LoadStringTemplates(templs map[string]string) error {
	m := make(map[string][]byte)
	for path, tmpl := range templs {
		m[path] = t.transpileOperators([]byte(tmpl))
	}
	t.template = m
	return nil
}

// MergeWithCtx templates with the passed in merge context
func (t *TemplMerger) MergeWithCtx(ctx TemplateContext) (err error) {
	t.file = make(map[string][]byte)
	var merged []byte
	for path, file := range t.template {
		if !strings.HasSuffix(path, ".art") {
			return fmt.Errorf("template name must have extension .art")
		}
		merged, err = t.mergeART(ctx, path, file)
		if err != nil {
			return fmt.Errorf("cannot merge template: %s\n", err)
		}
		t.file[path[0:len(path)-len(".art")]] = merged
	}
	return nil
}

// Merge templates with the passed in environment
func (t *TemplMerger) Merge(env *Envar) error {
	ctx, err := NewContext(env)
	if err != nil {
		return err
	}
	t.file = make(map[string][]byte)
	for path, file := range t.template {
		var merged []byte
		// if the template is in simple tem format
		if strings.HasSuffix(path, "tem") {
			merged, err = t.mergeTem(file, env)
			if err != nil {
				return fmt.Errorf("cannot merge template '%s': %s\n", path, err)
			}
			t.file[path[0:len(path)-len(".tem")]] = merged
		} else {
			// any other extension (or no extension) is merged as an artisan template
			merged, err = t.mergeART(ctx, path, file)
			if err != nil {
				return fmt.Errorf("cannot merge template: %s\n", err)
			}
			t.file[path[0:len(path)-len(filepath.Ext(path))]] = merged
		}
	}
	return nil
}

func (t *TemplMerger) Files() map[string][]byte {
	return t.file
}

// mergeTem merges a single template file using tem format and the passed in variables
func (t *TemplMerger) mergeTem(tem []byte, env *Envar) ([]byte, error) {
	content := string(tem)
	// find all environment variable placeholders in the content
	vars := t.regex.FindAll(tem, -1)
	// loop though the found vars to merge
	for _, v := range vars {
		defValue := ""
		// removes placeholder marks: ${...}
		vname := strings.TrimSuffix(strings.TrimPrefix(string(v), "${"), "}")
		// is a default value defined?
		cut := strings.Index(vname, ":")
		// split default value and var name
		if cut > 0 {
			// get the default value
			defValue = vname[cut+1:]
			// get the name of the var without the default value
			vname = vname[0:cut]
		}
		// check the name of the env variable is not "PWD" as it can return the current directory in some OSs
		if vname == "PWD" {
			return nil, fmt.Errorf("environment variable cannot be PWD, choose a different name\n")
		}
		// fetch the env variable value
		ev := env.Get(vname)
		// if the variable is not defined in the environment
		if len(ev) == 0 {
			// if no default value has been defined
			if len(defValue) == 0 {
				return nil, fmt.Errorf("environment variable '%s' required and not defined, cannot merge\n", vname)
			} else {
				// merge with the default value
				content = strings.Replace(content, string(v), defValue, -1)
			}
		} else {
			// merge with the env variable value
			content = strings.Replace(content, string(v), ev, -1)
		}
	}
	return []byte(content), nil
}

// mergeART merges a single template file using go template format and the passed in variables
func (t *TemplMerger) mergeART(ctx TemplateContext, path string, temp []byte) ([]byte, error) {
	tt, err := template.New(path).Funcs(ctx.FuncMap()).Parse(string(temp))
	if err != nil {
		return nil, err
	}
	var tpl bytes.Buffer
	err = tt.Execute(&tpl, ctx)
	if err != nil {
		return nil, err
	}
	mergedBytes := tpl.Bytes()
	if len(mergedBytes) == 0 {
		return nil, fmt.Errorf("template merge produced no output, possible causes are missing variables and syntactic errors")
	}
	return mergedBytes, nil
}

func (t *TemplMerger) transpileOperators(source []byte) []byte {
	names := t.rexVar.FindAllStringSubmatch(string(source), -1)
	for _, n := range names {
		str := strings.ReplaceAll(string(source), n[0], fmt.Sprintf("{{ var \"%s\" }}", n[1]))
		source = []byte(str)
	}
	names = t.rexRange.FindAllStringSubmatch(string(source), -1)
	for _, n := range names {
		str := strings.ReplaceAll(string(source), n[0], fmt.Sprintf("{{ select \"%s\"}}{{ range .Items }}", n[1]))
		source = []byte(str)
	}
	names = t.rexItem.FindAllStringSubmatch(string(source), -1)
	for _, n := range names {
		str := strings.ReplaceAll(string(source), n[0], fmt.Sprintf("{{ item \"%s\" . }}", n[1]))
		source = []byte(str)
	}
	names = t.rexItemEq.FindAllStringSubmatch(string(source), -1)
	for _, n := range names {
		str := strings.ReplaceAll(string(source), n[0], fmt.Sprintf("{{ if itemEq \"%s\" . \"%s\" }}", n[1], n[2]))
		source = []byte(str)
	}
	names = t.rexItemNeq.FindAllStringSubmatch(string(source), -1)
	for _, n := range names {
		str := strings.ReplaceAll(string(source), n[0], fmt.Sprintf("{{ if itemNeq \"%s\" . \"%s\" }}", n[1], n[2]))
		source = []byte(str)
	}
	return source
}

func (t *TemplMerger) Save() error {
	for fileName, fileBytes := range t.file {
		// override file with merged values
		err := writeToFile(fileName, string(fileBytes))
		if err != nil {
			return fmt.Errorf("cannot update config file: %s\n", err)
		}
	}
	return nil
}

func (t *TemplMerger) GetFile(name string) []byte {
	return t.file[name]
}
