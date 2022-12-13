/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
