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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zcalusic/sysinfo"
	"southwinds.dev/artisan/core"
)

// UtilSysInfoCmd gets linux system information
type UtilSysInfoCmd struct {
	Cmd           *cobra.Command
	family        bool
	vendor        bool
	osVersion     bool
	kernelVersion bool
}

func NewUtilSysInfoCmd() *UtilSysInfoCmd {
	c := &UtilSysInfoCmd{
		Cmd: &cobra.Command{
			Use:   "sys-info [flags]",
			Short: "gets linux system information",
			Long:  `gets linux system information, this command only works on linux systems`,
		},
	}
	c.Cmd.Flags().BoolVar(&c.family, "family", false, "--family; returns the Operating System family (i.e. debian, rhel")
	c.Cmd.Flags().BoolVar(&c.vendor, "vendor", false, "--vendor; returns the Operating System vendor")
	c.Cmd.Flags().BoolVar(&c.kernelVersion, "kernel-version", false, "--kernel-version; returns the OS Kernel version")
	c.Cmd.Flags().BoolVar(&c.osVersion, "os-version", false, "--os-version; returns the OS version")
	c.Cmd.Run = c.Run
	return c
}

func (c *UtilSysInfoCmd) Run(_ *cobra.Command, _ []string) {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	if c.family {
		switch si.OS.Vendor {
		case "centos":
		case "rocky":
		case "rhel":
			fmt.Printf("rhel")
			return
		case "debian":
		case "ubuntu":
			fmt.Printf("debian")
			return
		}
		fmt.Printf("unknown")
		return
	}
	if c.vendor {
		fmt.Printf("%s", si.OS.Vendor)
		return
	}
	if c.osVersion {
		fmt.Printf("%s", si.OS.Version)
		return
	}
	if c.kernelVersion {
		fmt.Printf("%s", si.Kernel.Version)
		return
	}
	data, err := json.MarshalIndent(&si, "", "  ")
	core.CheckErr(err, "cannot marshal sys info")
	fmt.Println(string(data))
}
