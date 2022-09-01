/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package merge

// Set a group of values for a variable group identified by a group name for a specific index
type Set struct {
	Context *Context
	// a list of values associated with a set name
	// eg: [ "NAME" ] [ "port a" ]
	//     [ "DESC" ] [ "this is port a" ]
	//     [ "VALUE" ] [ "80" ]
	Value map[string]string
}

func (s *Set) Get(name string) string {
	return s.Value[name]
}
