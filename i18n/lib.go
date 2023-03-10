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
	"fmt"
	"github.com/pelletier/go-toml"
	"southwinds.dev/artisan/core"

	"os"
	"path"
	"strings"
)

func String(artHome string, key I18NKey) string {
	return get(artHome, key)
}

// Printf prints a localised message
func Printf(artHome string, key I18NKey, a ...interface{}) {
	fmt.Printf(get(artHome, key), a...)
}

// Sprintf formats according to a format specifier and returns the resulting string
func Sprintf(artHome string, key I18NKey, a ...interface{}) string {
	return fmt.Sprintf(get(artHome, key), a...)
}

// Err checks for the  error and if it exists prints a localised error
func Err(artHome string, err error, key I18NKey, a ...interface{}) {
	if err != nil {
		fmt.Printf("%s - %s\n", fmt.Sprintf(get(artHome, key), a...), err)
		os.Exit(1)
	}
}

// raise an error
func Raise(artHome string, key I18NKey, a ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(get(artHome, key), a...))
	os.Exit(1)
}

// updates a specific i18n file by adding missing keys but keeping their value in english
func Update(i18nFile string) error {
	file := core.ToAbs(i18nFile)
	f, err := toml.LoadFile(file)
	if err != nil {
		return err
	}
	for key, value := range msg_en {
		if !f.Has(string(key)) {
			f.Set(string(key), value)
		}
	}
	data, err := f.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, os.ModePerm)
}

func get(artHome string, key I18NKey) string {
	var language string
	// check if internationalisation is enabled
	inter := os.Getenv("ARTISAN_I18N")
	// if not then use english
	if len(inter) == 0 {
		core.Debug("i18n is disabled, to enable set ARTISAN_I18N")
		return msg_en[key]
	}
	// check if an overriding language has been set
	overrideLang := os.Getenv("ARTISAN_LANG")
	// if not
	if len(overrideLang) == 0 {
		// use the current user language
		language = lang()
	} else {
		// use the overriding language
		language = overrideLang
	}
	// load the dictionary from file
	t, err := toml.LoadFile(path.Join(core.LangPath(artHome), fmt.Sprintf("%s_i18n.toml", language)))
	var value interface{}
	if err == nil {
		value = t.Get(string(key))
		if value == nil {
			// set value in english
			value = msg_en[key]
		}
	} else {
		value = msg_en[key]
	}
	return value.(string)
}

func splitLocale(locale string) (language string, territory string) {
	formattedLocale := strings.Split(locale, ".")[0]
	formattedLocale = strings.Replace(formattedLocale, "-", "_", -1)

	pieces := strings.Split(formattedLocale, "_")
	language = pieces[0]
	territory = ""
	if len(pieces) > 1 {
		territory = strings.Split(formattedLocale, "_")[1]
	}
	return language, territory
}
