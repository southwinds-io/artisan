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

package registry

import (
	"archive/zip"
	"fmt"
	"io"

	"log"
	"os"
	"path"
	"path/filepath"
	"southwinds.dev/artisan/core"
	"strings"
)

// MoveFile use instead of os.Rename() to avoid issues moving a file whose source and destination paths are
// on different file systems or drive
// e.g. when running in Kubernetes by Tekton
func MoveFile(src string, dst string) (err error) {
	err = CopyFile(src, dst)
	if err != nil {
		return fmt.Errorf("failed to copy source file %s to %s: %s", src, dst, err)
	}
	err = os.RemoveAll(src)
	if err != nil {
		return fmt.Errorf("failed to cleanup source file %s: %s", src, err)
	}
	return nil
}

// CopyFile credit https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}
	_, err = os.Stat(dst)
	// if the destination does not exist
	if os.IsNotExist(err) {
		// create the destination folder
		err = os.MkdirAll(dst, si.Mode())
		if err != nil {
			return
		}
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			i, infoErr := entry.Info()
			if infoErr != nil {
				return infoErr
			}
			if i.Mode()&os.ModeSymlink != 0 {
				continue
			}
			copyErr := CopyFile(srcPath, dstPath)
			if copyErr != nil {
				return
			}
		}
	}

	return
}

func MoveFolderContent(srcFolder, dstFolder string) error {
	srcFolder = core.ToAbs(srcFolder)
	dstFolder = core.ToAbs(dstFolder)
	file, err := os.Open(srcFolder)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	files, err := file.Readdir(-1)
	if err != nil {
		return err
	}
	for _, info := range files {
		if info.IsDir() {
			err = CopyDir(path.Join(srcFolder, info.Name()), path.Join(dstFolder, info.Name()))
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(path.Join(srcFolder, info.Name()), path.Join(dstFolder, info.Name()))
			if err != nil {
				return err
			}
		}
	}
	return os.RemoveAll(srcFolder)
}

func openFile(path string) *os.File {
	r, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return r
}

// unzip a package
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}
	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()
		zipPath := filepath.Join(dest, f.Name)
		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(zipPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", zipPath)
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(zipPath, f.Mode())
		} else {
			_ = os.MkdirAll(filepath.Dir(zipPath), f.Mode())
			f, err := os.OpenFile(zipPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

// remove an item from a slice
func removeItem(slice []string, item string) []string {
	var ix = -1
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			ix = i
			break
		}
	}
	if ix > -1 {
		return remove(slice, ix)
	}
	return slice
}

func removePackage(slice []*Package, p *Package) []*Package {
	var ix = -1
	for i := 0; i < len(slice); i++ {
		if slice[i] == p {
			ix = i
			break
		}
	}
	if ix > -1 {
		return removeP(slice, ix)
	}
	return slice
}

func remove(slice []string, ix int) []string {
	slice[ix] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func removeP(slice []*Package, ix int) []*Package {
	slice[ix] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func cleanFolder(pathToClean string) error {
	names, err := os.ReadDir(pathToClean)
	if err != nil {
		return err
	}
	for _, entry := range names {
		_ = os.RemoveAll(path.Join([]string{pathToClean, entry.Name()}...))
	}
	return nil
}
