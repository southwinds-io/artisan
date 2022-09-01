//go:build darwin || freebsd || linux || netbsd || openbsd
// +build darwin freebsd linux netbsd openbsd

/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package i18n

import (
	"os"
	"strings"
)

func lang() string {
	language, _ := splitLocale(getLocale())
	return strings.ToLower(language)
}

func getLocale() (locale string) {
	env := os.Environ()
	for _, e := range env {
		if strings.Contains(e, "LC_") {
			parts := strings.Split(e, "=")
			locale = os.Getenv(parts[0])
			if locale != "" {
				return locale
			}
		}
	}
	if locale == "" {
		locale = os.Getenv("LANG")
	}
	return locale
}
