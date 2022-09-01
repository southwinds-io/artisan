/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
	Cmd *cobra.Command
}

func NewLangFetchCmd() *LangFetchCmd {
	c := &LangFetchCmd{
		Cmd: &cobra.Command{
			Use:   "fetch [language code]",
			Short: "fetches a language dictionary and installs it in the local registry",
			Long:  `fetches a language dictionary and installs it in the local registry`,
		},
	}
	c.Cmd.Run = c.Run
	return c
}

func (c *LangFetchCmd) Run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		i18n.Raise("", i18n.ERR_INSUFFICIENT_ARGS)
	}
	if len(args) > 1 {
		i18n.Raise("", i18n.ERR_TOO_MANY_ARGS)
	}
	// checks the lang path exists within the registry
	core.LangExists("")
	// try and fetch the language dictionary
	url := fmt.Sprintf("https://raw.githubusercontent.com/southwinds-io/artlib/master/lang/%s_i18n.toml", args[0])
	resp, err := http.Get(url)
	i18n.Err("", err, i18n.ERR_CANT_DOWNLOAD_LANG, url)
	if resp.StatusCode != 200 {
		i18n.Err("", fmt.Errorf(resp.Status), i18n.ERR_CANT_DOWNLOAD_LANG, url)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	i18n.Err("", err, i18n.ERR_CANT_READ_RESPONSE)
	err = os.WriteFile(path.Join(core.LangPath(""), fmt.Sprintf("%s_i18n.toml", args[0])), bodyBytes, os.ModePerm)
	i18n.Err("", err, i18n.ERR_CANT_SAVE_FILE)
}
