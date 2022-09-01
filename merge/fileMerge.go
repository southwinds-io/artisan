/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
