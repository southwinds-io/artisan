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
)

// EnvCmd access environment file functions
type EnvCmd struct {
	Cmd *cobra.Command
}

func NewEnvCmd() *EnvCmd {
	c := &EnvCmd{
		Cmd: &cobra.Command{
			Use:   "env",
			Short: "extract environment information from packages and flows",
			Long:  `extract environment information from packages and flows`,
		},
	}
	return c
}
