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
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"southwinds.dev/artisan/core"
	"strings"
)

// UtilExtractCmd client url issues http requests within a retry framework
type UtilExtractCmd struct {
	Cmd     *cobra.Command
	matches int
	prefix  string
	suffix  string
}

func NewUtilExtractCmd() *UtilExtractCmd {
	c := &UtilExtractCmd{
		Cmd: &cobra.Command{
			Use:     "extract [flags]",
			Short:   "extracts text between specified prefix and suffix, it should be used only with pipes",
			Long:    `extracts text between specified prefix and suffix, it should be used only with pipes`,
			Example: "cat your-file.txt | art u extract --prefix AAA --suffix $ -n 1",
			Args:    cobra.ExactArgs(0),
		},
	}
	c.Cmd.Flags().StringVarP(&c.prefix, "prefix", "p", "", "-p \"the prefix\"")
	c.Cmd.Flags().StringVarP(&c.suffix, "suffix", "s", "$", "-s \"the suffix\", if not specified an end of line marker is assumed")
	c.Cmd.Flags().IntVarP(&c.matches, "matches", "n", -1, "the maximum number of matches to retrieve")
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilExtractCmd) Run(cmd *cobra.Command, args []string) {
	// captures information from the standard input
	info, err := os.Stdin.Stat()
	core.CheckErr(err, "cannot read from stdin")
	// check that the standard input is not a character device file - i.e. one with which the Driver communicates
	// by sending and receiving single characters (bytes, octets)
	if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage:")
		fmt.Println("  cat your-file.txt | extract --prefix AAA --suffix $")
	} else if info.Size() > 0 {
		input := new(strings.Builder)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input.WriteString(fmt.Sprintf("%s\n", scanner.Text()))
		}
		output := core.Extract(input.String(), c.prefix, c.suffix, c.matches)
		if len(output) == 1 {
			fmt.Printf(output[0])
		} else {
			fmt.Printf(strings.Join(output, ","))
		}
	}
}
