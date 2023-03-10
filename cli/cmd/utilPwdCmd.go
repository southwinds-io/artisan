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
	"golang.org/x/crypto/bcrypt"
	"southwinds.dev/artisan/core"
)

// UtilPwdCmd generates passwords
type UtilPwdCmd struct {
	Cmd          *cobra.Command
	len          *int
	specialChars *bool
	bcrypt       *bool
}

func NewUtilPwdCmd() *UtilPwdCmd {
	c := &UtilPwdCmd{
		Cmd: &cobra.Command{
			Use:   "pwd [flags]",
			Short: "generates a random password",
			Long:  `generates a random password`,
		},
	}
	c.len = c.Cmd.Flags().IntP("length", "l", 16, "length of the generated password")
	c.specialChars = c.Cmd.Flags().BoolP("special-chars", "s", false, "use special characters in the generated password")
	c.bcrypt = c.Cmd.Flags().BoolP("bcrypt", "b", false, "hash it using bcrypt algorithm")
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilPwdCmd) Run(_ *cobra.Command, _ []string) {
	pwd := core.RandomPwd(*c.len, *c.specialChars)
	fmt.Printf("%s", pwd)
	if *c.bcrypt {
		hash, err := pwdHash([]byte(pwd))
		core.CheckErr(err, "cannot hash password")
		fmt.Printf("\nbcrypt:%s\n", hash)
	}
}

func pwdHash(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
