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
	"github.com/spf13/cobra"
	"southwinds.dev/artisan/core"
)

// UtilCurlCmd client url issues http requests within a retry framework
type UtilCurlCmd struct {
	Cmd           *cobra.Command
	maxAttempts   int
	creds         string
	method        string
	payload       string
	file          string
	validCodes    []int
	addValidCodes []int
	delaySecs     int
	timeoutSecs   int
	headers       []string
	outFile       string
	response      bool
}

func NewUtilCurlCmd() *UtilCurlCmd {
	c := &UtilCurlCmd{
		Cmd: &cobra.Command{
			Use:   "curl [flags] URI",
			Short: "issues an HTTP request and retry if a failure occurs",
			Long:  `issues an HTTP request and retry if a failure occurs`,
			Args:  cobra.ExactArgs(1),
		},
	}
	c.Cmd.Flags().StringVarP(&c.creds, "creds", "u", "", "-u user:password")
	c.Cmd.Flags().IntVarP(&c.maxAttempts, "max-attempts", "a", 5, "number of attempts before it stops retrying")
	c.Cmd.Flags().IntSliceVarP(&c.validCodes, "valid-codes", "c", []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226}, "comma separated list of HTTP status codes considered valid (e.g. no retry will be triggered)")
	c.Cmd.Flags().IntSliceVarP(&c.addValidCodes, "add-valid-codes", "C", []int{}, "comma separated list of additional HTTP status codes considered valid (e.g. no retry will be triggered)")
	c.Cmd.Flags().StringVarP(&c.method, "method", "X", "GET", "the http method to use (i.e. POST, PUT, GET, DELETE)")
	c.Cmd.Flags().StringVarP(&c.outFile, "out-file", "o", "", "the name of the file where the http response body should be saved; if not set, the response is not saved but printed to stdout")
	c.Cmd.Flags().StringVarP(&c.payload, "payload", "d", "", "a string with the payload to be sent in the body of the http request")
	c.Cmd.Flags().StringVarP(&c.file, "file", "f", "", "the location of a file which content is to be sent in the body of the http request")
	c.Cmd.Flags().StringSliceVarP(&c.headers, "headers", "H", nil, "a comma separated list of http headers (format 'key1:value1','key2:value2,...,'keyN:valueN')")
	c.Cmd.Flags().IntVarP(&c.delaySecs, "delay", "r", 5, "the retry delay (in seconds)")
	c.Cmd.Flags().IntVarP(&c.timeoutSecs, "timeout", "t", 30, "the period (in seconds) after which the http request will timeout if not response is received from the server")
	c.Cmd.Flags().BoolVarP(&c.response, "response", "v", false, "if set, shows additional response information such as status code and headers")
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilCurlCmd) Run(cmd *cobra.Command, args []string) {
	uri := args[0]
	token := ""
	if len(c.creds) > 0 {
		uname, pwd := core.UserPwd(c.creds)
		token = core.BasicToken(uname, pwd)
	}
	core.Curl(uri, c.method, token, append(c.validCodes, c.addValidCodes...), c.payload, c.file, c.maxAttempts, c.delaySecs, c.timeoutSecs, c.headers, c.outFile, c.response)
}
