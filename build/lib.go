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

package build

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"southwinds.dev/artisan/conf"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/merge"
	"strconv"
	"strings"
	"time"
)

// zip a file or a folder
func zipSource(source, target string, excludeSource []string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() {
		err := zipfile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	archive := zip.NewWriter(zipfile)
	defer func() {
		err := archive.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	info, err := os.Stat(source)
	if err != nil {
		return nil
	}
	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// do not add to the zip file excluded sources
		if contains(source, excludeSource) {
			return nil
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}
		if info.IsDir() {
			header.Name += string(os.PathSeparator)
		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			err := file.Close()
			if err != nil {
				log.Print(err)
				runtime.Goexit()
			}
		}()
		_, err = io.Copy(writer, file)
		return err
	})
	return err
}

// gets the error message for a shell exit status
func exitMsg(exitCode int) string {
	switch exitCode {
	case 1:
		return "error 1 - general error"
	case 2:
		return "error 2 - misuse of shell built-ins (check for permission or access problem)"
	case 126:
		return "error 126 - command invoked cannot execute (check for permission problem)"
	case 127:
		return "error 127 - command not found (check for typos or missing commands)"
	case 128:
		return "error 128 - invalid argument to exit (check when you are not returning something that is not integer args in the range 0 – 255)"
	case 130:
		return "error 130 - script terminated by CTRL-C"
	default:
		return fmt.Sprintf("exit code %d", exitCode)
	}
}

// wait a time duration for a file or folder to be created on the path
func waitForTargetToBeCreated(path string) {
	elapsed := 0
	found := false
	for {
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			found = true
			break
		}
		if elapsed > 30 {
			break
		}
		elapsed++
		time.Sleep(500 * time.Millisecond)
	}
	if !found {
		core.RaiseErr("target '%s' not found after command execution", path)
	}
}

// copy the files in a folder recursively
func copyFolder(src string, dst string) error {
	var err error
	var fds []os.DirEntry
	var srcInfo os.FileInfo
	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	if fds, err = os.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcFp := path.Join(src, fd.Name())
		dstFp := path.Join(dst, fd.Name())
		if fd.IsDir() {
			if err = copyFolder(srcFp, dstFp); err != nil {
				core.ErrorLogger.Printf(err.Error())
			}
		} else {
			if err = core.CopyFile(srcFp, dstFp); err != nil {
				core.ErrorLogger.Printf(err.Error())
			}
		}
	}
	return nil
}

// converts a byte count into a pretty label
func bytesToLabel(size int64) string {
	var suffixes [5]string
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"
	base := math.Log(float64(size)) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64) + getSuffix
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// executes a command and sends output and error streams to stdout and stderr
func execute(cmd string, dir string, env conf.Configuration, interactive bool) (err error) {
	core.Debug("executing command: '%s'\n", cmd)
	// executes the command
	_, err = ExeAsync(cmd, dir, env, interactive)
	// if there is an error return it
	if err != nil {
		return err
	}
	// return without error
	return nil
}

func contains(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func findGitPath(path string) (string, error) {
	for {
		_, err := os.Stat(filepath.Join(path, ".git"))
		if os.IsNotExist(err) {
			path = filepath.Dir(path)
			if strings.HasSuffix(path, string(os.PathSeparator)) {
				return "", fmt.Errorf("cannot find .git path")
			}
		} else {
			return path, nil
		}
	}
}

// check the specified function is in the manifest
func isExported(m *data.Manifest, fx string) bool {
	for _, function := range m.Functions {
		if function.Name == fx {
			return true
		}
	}
	return false
}

func EvalShell(statement string, env conf.Configuration) (string, error) {
	if env == nil {
		env = merge.NewEnVarEmpty()
	}
	if ok, expr, shell := core.HasShell(statement); ok {
		shell = strings.Trim(shell, " ")
		core.Debug("subshell evaluation started: '%s'\n", shell)
		usesArtisan := strings.HasPrefix(shell, "art ")
		core.Debug("=> subshell uses artisan command: %t\n", usesArtisan)
		out, err := Exe(shell, "", env, false)
		if err != nil {
			return "", fmt.Errorf("cannot execute subshell command '%s': %s", statement, err)
		}
		// ensure the subshell output does not end with newline
		out = core.TrimNewline(out)
		core.Debug("=> shell eval output: '%s'\n", out)
		// if subshell uses art command then check for safe output
		if usesArtisan && len(out) > 0 {
			core.Debug("=> found wrapped value in subshell output\n")
			r, _ := regexp.Compile("{{.*}}")
			if matched := r.MatchString(out); matched {
				out = r.FindString(out)
				// merges the output of the subshell in the original variable
				statement = strings.Replace(statement, expr, out[2:len(out)-2], 1)
				core.Debug("=> unwrapped value is: '%s'\n", out[2:len(out)-2])
			} else {
				return "", fmt.Errorf("non-empty returned value of subshell expression '%s', must be enclosed by double curly braces '{{...}}' markers to prevent potential corruption due to debug statements", shell)
			}
		} else {
			// merges the output of the subshell in the original command
			statement = strings.Replace(statement, expr, out, -1)
		}
	}
	// if it does not have a subshell returns the original statement
	return statement, nil
}
