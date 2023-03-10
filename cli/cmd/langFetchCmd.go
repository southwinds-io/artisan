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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/i18n"

	"net/http"
	"os"
	"path"
)

// LangFetchCmd installs it in the local registry
type LangFetchCmd struct {
	Cmd  *cobra.Command
	home string
}

func NewLangFetchCmd(artHome string) *LangFetchCmd {
	c := &LangFetchCmd{
		Cmd: &cobra.Command{
			Use:   "fetch [language code]",
			Short: "fetches a language dictionary and installs it in the local registry",
			Long:  `fetches a language dictionary and installs it in the local registry`,
		},
		home: artHome,
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *LangFetchCmd) Run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		i18n.Raise(c.home, i18n.ERR_INSUFFICIENT_ARGS)
	}
	if len(args) > 1 {
		i18n.Raise(c.home, i18n.ERR_TOO_MANY_ARGS)
	}
	// checks the lang path exists within the registry
	core.LangExists("")
	// try and fetch the language dictionary
	url := fmt.Sprintf("https://raw.githubusercontent.com/southwinds-io/artlib/master/lang/%s_i18n.toml", args[0])
	resp, err := http.Get(url)
	i18n.Err(c.home, err, i18n.ERR_CANT_DOWNLOAD_LANG, url)
	if resp.StatusCode != 200 {
		i18n.Err(c.home, fmt.Errorf(resp.Status), i18n.ERR_CANT_DOWNLOAD_LANG, url)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	i18n.Err(c.home, err, i18n.ERR_CANT_READ_RESPONSE)
	err = os.WriteFile(path.Join(core.LangPath(c.home), fmt.Sprintf("%s_i18n.toml", args[0])), bodyBytes, os.ModePerm)
	i18n.Err(c.home, err, i18n.ERR_CANT_SAVE_FILE)
}
