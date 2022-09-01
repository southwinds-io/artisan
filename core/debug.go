/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package core

import (
	"fmt"
	"os"
)

// Debug writes a debug message to the console
func Debug(msg string, a ...interface{}) {
	if InDebugMode() {
		DebugLogger.Printf("%s\n", fmt.Sprintf(msg, a...))
	}
}

func InDebugMode() bool {
	return len(os.Getenv(ArtDebug)) > 0
}
