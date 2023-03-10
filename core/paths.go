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

package core

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

const AppName = "artisan"

// HomeDir gets the user home directory
func HomeDir() string {
	// if ARTISAN_HOME is defined use it
	if artHome := os.Getenv("ARTISAN_HOME"); len(artHome) > 0 {
		return artHome
	}
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func WorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}

// RegistryPath gets the root path of the local registry
func RegistryPath(path string) string {
	if len(path) > 0 {
		path, _ = filepath.Abs(path)
		return filepath.Join(path, fmt.Sprintf(".%s", AppName))
	}
	return filepath.Join(HomeDir(), fmt.Sprintf(".%s", AppName))
}

func FilesPath(path string) string {
	return filepath.Join(RegistryPath(path), "files")
}

func LangPath(path string) string {
	return filepath.Join(RegistryPath(path), "lang")
}

// TmpPath temporary path for file operations
func TmpPath(path string) string {
	return filepath.Join(RegistryPath(path), "tmp")
}

func TmpExists(path string) {
	tmp := TmpPath(path)
	// ensure tmp folder exists for temp file operations
	_, err := os.Stat(tmp)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(tmp, os.ModePerm)
	}
}

func LangExists(path string) {
	lang := LangPath(path)
	// ensure lang folder exists for temp file operations
	_, err := os.Stat(lang)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(lang, os.ModePerm)
	}
}

// RunPath temporary path for running package functions
func RunPath(path string) string {
	return filepath.Join(RegistryPath(path), "tmp", "run")
}

func RunPathExists(path string) {
	runPath := RunPath(path)
	// ensure tmp folder exists for  running package functions
	_, err := os.Stat(runPath)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(runPath, os.ModePerm)
	}
}

// EnsureRegistryPath check the local registry directory exists and if not creates it
func EnsureRegistryPath(path string) error {
	// check the home directory exists
	_, err := os.Stat(RegistryPath(path))
	// if it does not
	if os.IsNotExist(err) {
		if runtime.GOOS == "linux" && os.Geteuid() == 0 {
			WarningLogger.Printf("if the root user creates the local registry then runc commands will fail\n" +
				"as the runtime user will not be able to access its content when it is bind mounted\n" +
				"ensure the local registry path is not owned by the root user\n")
		}
		err = os.MkdirAll(RegistryPath(path), os.ModePerm)
		if err != nil {
			return fmt.Errorf("cannot create registry folder: %s\n", err)
		}
	}
	filesPath := FilesPath(path)
	// check the files' directory exists
	_, err = os.Stat(filesPath)
	// if it does not
	if os.IsNotExist(err) {
		err = os.Mkdir(filesPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("cannot create local registry files folder: %s\n", err)
		}
	}
	return nil
}
