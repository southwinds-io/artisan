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

package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"southwinds.dev/artisan/core"
	"strings"
)

// UtilUpConfCmd updates one or more a yaml file
type UtilUpConfCmd struct {
	Cmd      *cobra.Command
	keyValue []string
}

func NewUpConfCmd() *UtilUpConfCmd {
	c := &UtilUpConfCmd{
		Cmd: &cobra.Command{
			Use:   "upconf [flags] [config filename]",
			Short: "updates properties in a config file",
			Long:  `updates properties in a config file`,
			Example: `
$> art u upconf -v "MY_VAR_NAME1:MYVAR_VALUE1" -v "MY_VAR_NAME2:MYVAR_VALUE2" config.yaml
`,
		},
	}
	c.Cmd.Flags().StringArrayVarP(&c.keyValue, "value", "v", []string{}, "a key:value pair representing the property to update and its new value respectively")
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilUpConfCmd) Run(_ *cobra.Command, args []string) {
	if args != nil && len(args) == 0 {
		core.RaiseErr("config filename is required")
	}
	updateConf(args[0], c.keyValue)
}

func updateConf(filename string, keyValueList []string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		core.RaiseErr("config file '%s' not found", filename)
	} else {
		filename, _ = filepath.Abs(filename)
	}
	content, err := os.ReadFile(filename)
	core.CheckErr(err, "cannot read config file")
	var values map[string]interface{}
	ext := filepath.Ext(filename)
	switch ext {
	case ".yaml":
		fallthrough
	case ".yml":
		err = yaml.Unmarshal(content, &values)
		core.CheckErr(err, "cannot unmarshal config file to yaml")
	case ".json":
		err = json.Unmarshal(content, &values)
		core.CheckErr(err, "cannot unmarshal config file to json")
	default:
		core.RaiseErr("invalid file extension, must be either json or yaml/yml")
	}
	var updated bool
	for key, value := range values {
		if v, ok := value.(string); ok {
			if matched, newValue := keyMatched(key, v, keyValueList); matched {
				values[key] = newValue
				updated = true
				core.InfoLogger.Printf("updated '%s'\n", key)
			}
		}
	}
	var data []byte
	if updated {
		switch ext {
		case ".yaml":
			fallthrough
		case ".yml":
			data, err = yaml.Marshal(values)
			core.CheckErr(err, "cannot marshall updated values to yaml")
		case ".json":
			data, err = json.Marshal(values)
			core.CheckErr(err, "cannot marshall updated values to json")
		}
		core.CheckErr(os.WriteFile(filename, data, os.ModePerm), "cannot update config file")
	}
}

func keyMatched(key, value string, keyValueList []string) (bool, string) {
	for _, keyValue := range keyValueList {
		parts := strings.SplitN(keyValue, ":", 2)
		if len(parts) != 2 {
			core.RaiseErr("invalid key-value format found in '%s', must be KEY:VALUE", keyValue)
		}
		if strings.EqualFold(key, parts[0]) {
			return true, parts[1]
		}
	}
	return false, ""
}
