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

package conf

import (
	"regexp"
	"strings"
)

// Configuration represent a source of environment variables
type Configuration interface {
	Get(key string) string
	Set(key, value string)
	Merge(s Configuration)
	MergeMap(m map[string]string)
	Append(m map[string]string) Configuration
	Vars() map[string]string
	Replace()
	Slice() []string
	String() string
}

// ReplaceVar recursively replaces any variables in the text string using values provided by the config source
func ReplaceVar(text string, conf Configuration) string {
	evExpression := regexp.MustCompile("\\${(.*?)}")
	matches := evExpression.FindAllString(text, -1)
	// if we have matches
	if matches != nil {
		for _, match := range matches {
			// get the name of the environment variable i.e. the name part in "${name}"
			name := match[2 : len(match)-1]
			// get the value of the variable
			value := conf.Get(name)
			if len(value) > 0 {
				// check if the value contains env variables
				submatches := evExpression.FindAllString(value, -1)
				for range submatches {
					value = ReplaceVar(value, conf)
				}
				text = strings.Replace(text, match, value, -1)
			}
		}
	}
	return text
}
