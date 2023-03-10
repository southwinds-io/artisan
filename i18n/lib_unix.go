//go:build darwin || freebsd || linux || netbsd || openbsd
// +build darwin freebsd linux netbsd openbsd

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
