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

// FlowCmd provides functions to manage Artisan execution flows
type FlowCmd struct {
	Cmd *cobra.Command
}

func NewFlowCmd() *FlowCmd {
	c := &FlowCmd{
		Cmd: &cobra.Command{
			Use:   "flow",
			Short: "provides functions to manage Artisan execution flows",
			Long:  `provides functions to manage Artisan execution flows`,
		},
	}
	return c
}
