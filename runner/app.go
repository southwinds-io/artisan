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

package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/registry"
	"strconv"
	"strings"
	"time"
)

func RunApp(name *core.PackageName, credentials string, detached, clean bool, path string, artHome string, v data.VerifyHandler, rh data.RunHandler) error {
	// gets a handle to the local registry
	r := registry.NewLocalRegistry(artHome)
	// check if the package is there
	pkg := r.FindPackageByName(name)
	// if it is not, try and pull it
	if pkg == nil {
		var err error
		pkg, err = r.Pull(name, credentials, false)
		if err != nil {
			return err
		}
		// if the package is still nil
		if pkg == nil {
			return fmt.Errorf("cannot find package '%s' in remote registry", name.FullyQualifiedNameTag())
		}
	}
	// inspects the manifest
	seal, err := r.GetSeal(pkg)
	if err != nil {
		return err
	}
	// the package type must be declared as "content/app" for the runner to attempt to run the app
	if !strings.EqualFold(seal.Manifest.Type, "content/app") {
		// if there is an entry point
		if len(seal.Manifest.Labels["app:entrypoint"]) > 0 {
			// then it is an app package and return error
			return fmt.Errorf("cannot run app in package '%s' as it is not of type 'content/app'", name.FullyQualifiedNameTag())
		} else {
			// if no entrypoint is found, it is an automation package so run it and return
			return runAutomationPackage(name, "fx_here", "source_here", credentials, detached, clean, path, artHome, v)
		}
	}
	// the manifest must declare an entry point for the app
	entryPoint := getEntryPoint(seal.Manifest)
	if len(entryPoint) == 0 {
		return fmt.Errorf("cannot run app as entrypoint is not defined: add an 'app:entrypoint' label to the package manifest")
	}
	// validates all environment variables
	if err = validateVars(seal.Manifest); err != nil {
		return err
	}
	if err = assignVolumeVars(seal.Manifest); err != nil {
		return err
	}
	path, _ = filepath.Abs(path)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		_ = os.MkdirAll(path, 0755)
	}
	if err = r.Open(name, credentials, path, v, rh, []string{}); err != nil {
		return err
	}
	if clean {
		err = r.Remove([]string{name.FullyQualifiedNameTag()})
		if err != nil {
			return err
		}
	}
	entryPath := filepath.Join(path, entryPoint)
	_, err = os.Stat(entryPath)
	var doesNotExist = os.IsNotExist(err)
	var count = 0
	for doesNotExist && count < 30 {
		time.Sleep(250 * time.Millisecond)
		count++
		_, err = os.Stat(entryPath)
		doesNotExist = os.IsNotExist(err)
	}
	core.Debug("entrypoint: %s", entryPath)
	core.Debug("execution path: %s", path)
	core.Debug("environment =>")
	env := os.Environ()
	for index, value := range env {
		core.Debug("  %d => %s", index, value)
	}
	if detached {
		_, err = build.ExeAsync(entryPath, path, merge.NewEnVarFromSlice(env), false)
		if err != nil {
			return err
		}
	} else {
		if err = build.ExeStream(entryPath, path, merge.NewEnVarFromSlice(env), false); err != nil {
			return err
		}
	}
	return nil
}

func getEntryPoint(manifest *data.Manifest) string {
	for key, value := range manifest.Labels {
		if strings.EqualFold(key, "app:entrypoint") {
			return value
		}
	}
	return ""
}

func validateVars(manifest *data.Manifest) error {
	for key, value := range manifest.Labels {
		if strings.HasPrefix(key, "app:var@") {
			parts := strings.Split(key, "@")
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable declaraction in manifest: '%s' must be of the format 'app:var@NAME'", key)
			}
			var keyValue = parts[1]
			parts = strings.Split(value, ",")
			var defaultValue string
			if len(parts) == 2 {
				parts2 := strings.Split(parts[1], "=")
				if !strings.EqualFold(parts2[0], "default") {
					return fmt.Errorf("invalid value for variable, must be of 'default' but found '%s'", parts2[0])
				}
				defaultValue = parts2[1]
			}
			required := strings.EqualFold(parts[0], "required")
			optional := strings.EqualFold(parts[0], "optional")
			if !(required || optional) {
				return fmt.Errorf("invalid value for variable, expecting 'required' or 'optional' but found '%s'", parts[0])
			}
			// if the variable is required
			if required {
				// try and retrieve its value from the environment
				v := os.Getenv(keyValue)
				// if no value has been found
				if len(v) == 0 {
					// but we have a default value
					if len(defaultValue) > 0 {
						// try and set it with the default value
						err := os.Setenv(keyValue, defaultValue)
						if err != nil {
							return fmt.Errorf("cannot set environment variable '%s' with default value: %s", keyValue, err)
						}
					} else {
						// otherwise, cannot continue as variable is not set in the environment
						return fmt.Errorf("missing variable '%s'", strings.ToUpper(keyValue))
					}
				}
			}
		}
	}
	return nil
}

func assignVolumeVars(manifest *data.Manifest) error {
	for key, value := range manifest.Labels {
		if strings.HasPrefix(key, "app:volume@") {
			parts := strings.Split(key, "@")
			if len(parts) != 2 {
				return fmt.Errorf("invalid volume label, expecting format 'app:volume@xxx' but found '%s'", key)
			}
			volumeVar := parts[1]
			volumeNumber, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid volume number '%s': %s", value, err)
			}
			if err = os.Setenv(volumeVar, fmt.Sprintf("/volume_%d", volumeNumber)); err != nil {
				return err
			}
		}
	}
	return nil
}

func runAutomationPackage(name *core.PackageName, packageFx, packageSource, credentials string, detached, clean bool, path string, artHome string, v data.VerifyHandler) error {
	core.RaiseErr("automation package run is not implemented")
	// if a package name has been provided
	if name != nil {
		// if a package  source  has been provided
		if len(packageSource) > 0 {
			switch strings.ToLower(packageSource) {
			case "create":
				// remove all files and subdirectories
				if err := removeSubDirs("/workspace/source"); err != nil {
					return err
				}
				builder := build.NewBuilder(core.ArtDefaultHome)
				builder.Execute(name, packageFx, credentials, false, "/workspace/source", true, nil, []string{}, false)
			case "merge":
			case "read":
			}
		}
	}
	return nil
}

func removeSubDirs(path string) error {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		path, err := filepath.Abs(dir.Name())
		if err != nil {
			return err
		}
		if dir.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		} else {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

/*
   app:entrypoint: artr
   app:var@ARTR_ADMIN_USER: required,default=admin
   app:var@ARTR_ADMIN_PWD: required,default=adm1n
   app:var@ARTR_READ_USER: optional
   app:var@ARTR_READ_PWD: optional
   app:volume@DATA_PATH: 0
*/
