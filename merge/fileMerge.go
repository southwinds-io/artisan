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
	"io"
	"log"
	"os"
	"path"
	"strings"
)

func writeToFile(filename string, data string) error {
	// create a file without the .tem extension
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	// write the merged content into the file
	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	log.Printf("'%v' bytes written to file '%s'\n", len(data), filenameWithoutExtension(filename))
	return file.Sync()
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}
