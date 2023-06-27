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

package data

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Manifest struct {
	// the author of the package
	Author string `json:"author,omitempty"`
	// the signing and verification authority for the package
	Authority []string `json:"authority,omitempty"`
	// the package type
	Type string `json:"type,omitempty"`
	// the license associated to the package
	License string `json:"license,omitempty"`
	// the target OS for the package
	OS string `json:"os"`
	// the name of the package file
	Ref string `json:"ref"`
	// the build profile used
	Profile string `json:"profile"`
	// runtime image that should be used to execute exported functions in the package
	Runtime string `json:"runtime,omitempty"`
	// the labels assigned to the package
	Labels map[string]string `json:"labels,omitempty"`
	// the URI of the package source
	Source string `json:"source,omitempty"`
	// the path within the source where the project is (for uber repos)
	SourcePath string `json:"source_path,omitempty"`
	// the commit hash
	Commit string `json:"commit,omitempty"`
	// repo branch
	Branch string `json:"branch,omitempty"`
	// the name of the file or folder that has been packaged
	Target string `json:"target,omitempty"`
	// the timestamp
	Time string `json:"time"`
	// the size of the package
	Size string `json:"size"`
	// the Stock Keeping Unit code
	SKU string `json:"SKU,omitempty"`
	// what functions are available to call?
	Functions  []*FxInfo `json:"functions,omitempty"`
	OpenPolicy string    `json:"open_policy,omitempty"`
	RunPolicy  string    `json:"run_policy,omitempty"`
	SignPolicy string    `json:"sign_policy,omitempty"`
}

func (m Manifest) Fx(name string) *FxInfo {
	for _, fx := range m.Functions {
		if fx.Name == name {
			return fx
		}
	}
	return nil
}

type Network struct {
	Groups []string `yaml:"groups"`
	Rules  []string `yaml:"rules"`
}

type GroupInfo struct {
	Group string
	Tags  []string
	Min   int
	Max   int
	IPs   []string
}

type GroupsInfo []GroupInfo

func (gs *GroupsInfo) GroupIx(name string) int {
	var gsi []GroupInfo = *gs
	for ix, info := range gsi {
		if strings.EqualFold(info.Group, name) {
			return ix
		}
	}
	return -1
}

func (n *Network) parseGroups() (GroupsInfo, error) {
	result := make(GroupsInfo, 0)
	for _, group := range n.Groups {
		parts := strings.Split(group, ":")
		groupName := parts[0]
		tags := strings.Split(parts[1], ",")
		var min, max int
		min, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid minimum value in network group '%s'", groupName)
		}
		if strings.EqualFold(parts[3], "*") {
			max = 1000
		} else {
			max, err = strconv.Atoi(parts[3])
			if err != nil {
				return nil, fmt.Errorf("invalid maximum value in network group '%s'", groupName)
			}
		}
		result = append(result, GroupInfo{
			Group: groupName,
			Tags:  tags,
			Min:   min,
			Max:   max,
		})
	}
	return result, nil
}

func parseGroup(group string) (groupName string, tags []string, min, max int, err error) {
	parts := strings.Split(group, ":")
	groupName = parts[0]
	tags = strings.Split(parts[1], ",")
	min, err = strconv.Atoi(parts[2])
	if err != nil {
		return groupName, tags, min, max, fmt.Errorf("invalid minimum value in network group '%s'", groupName)
	}
	if strings.EqualFold(parts[3], "*") {
		max = 1000 // upper limit is 1000 nodes
	} else {
		max, err = strconv.Atoi(parts[3])
		if err != nil {
			return groupName, tags, min, max, fmt.Errorf("invalid maximum value in network group '%s'", groupName)
		}
	}
	return
}
func (n *Network) AllocateIPs(ipList ...string) (GroupsInfo, error) {
	if hasDuplicates(ipList) {
		return nil, fmt.Errorf("IPs in list must be unique")
	}
	var ipIx int
	g, err := n.parseGroups()
	if err != nil {
		return nil, err
	}
	// first tries and allocates the minimum requirement
	for _, group := range n.Groups {
		name, _, min, _, parseErr := parseGroup(group)
		if parseErr != nil {
			return nil, err
		}
		infoIx := g.GroupIx(name)
		for i := 0; i < min; i++ {
			g[infoIx].IPs = append(g[infoIx].IPs, ipList[ipIx])
			ipIx++
		}
	}
	// now tries and allocates the maximum requirement
	for _, group := range n.Groups {
		name, _, min, max, parseErr := parseGroup(group)
		if parseErr != nil {
			return nil, err
		}
		infoIx := g.GroupIx(name)
		topLimit := max
		if max > (len(ipList) - ipIx) {
			topLimit = len(ipList) - ipIx
		}
		for i := 0; i < topLimit; i++ {
			if max > min {
				g[infoIx].IPs = append(g[infoIx].IPs, ipList[ipIx])
				ipIx++
			}
		}
	}
	return g, nil
}

type Group struct {
	Tags []string `yaml:"tags"`
	IPs  []string `yaml:"ips"`
}

type Role struct {
	Name string `yaml:"name"`
	Min  int    `yaml:"min,omitempty"`
	Max  int    `yaml:"max,omitempty"`
	Tag  []Tag  `yaml:"tag,omitempty"`
}

type Tag struct {
	Name string `yaml:"name"`
	Min  int    `yaml:"min,omitempty"`
	Max  int    `yaml:"max,omitempty"`
}

// FxInfo exported function list
type FxInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Input       *Input   `json:"input,omitempty"`
	Credits     int      `json:"credits,omitempty"`
	Runtime     string   `json:"runtime,omitempty"` // runtime image that should be used to execute functions in the package
	Network     *Network `json:"network,omitempty"`
}

func (m *Manifest) ToMarkDownBytes(name string) []byte {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("# Package %s Manifest\n", name))
	b.WriteString(fmt.Sprintf("*autogenerated using [Artisan CLI](https://github.com/southwinds-io/artisan) on %s*\n", time.Now().Format(time.RFC822Z)))
	for _, fx := range m.Functions {
		b.WriteString(fmt.Sprintf("## Function: %s\n", fx.Name))
		b.WriteString(fmt.Sprintf("%s\n", fx.Description))
		if len(fx.Input.Var) > 0 {
			b.WriteString(fmt.Sprintf("### Variables:\n"))
			b.WriteString(fmt.Sprintf("|name|description|default|\n"))
			b.WriteString(fmt.Sprintf("|---|---|---|\n"))
			for _, v := range fx.Input.Var {
				b.WriteString(fmt.Sprintf("|%s|%s|%s|\n", v.Name, format(v.Description), v.Default))
			}
		}
		if len(fx.Input.Secret) > 0 {
			b.WriteString(fmt.Sprintf("### Secrets:\n"))
			b.WriteString(fmt.Sprintf("|name|description|\n"))
			b.WriteString(fmt.Sprintf("|---|---|\n"))
			for _, s := range fx.Input.Secret {
				b.WriteString(fmt.Sprintf("|%s|%s|\n", s.Name, format(s.Description)))
			}
		}
		if len(fx.Input.File) > 0 {
			b.WriteString(fmt.Sprintf("### Files:\n"))
			b.WriteString(fmt.Sprintf("|name|description|path|\n"))
			b.WriteString(fmt.Sprintf("|---|---|---|\n"))
			for _, f := range fx.Input.File {
				b.WriteString(fmt.Sprintf("|%s|%s|%s|\n", f.Name, format(f.Description), f.Path))
			}
		}
	}
	return b.Bytes()
}

func format(content string) string {
	return strings.Replace(content, "\n", "<br>", -1)
}

func removeDuplicateValues(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func hasDuplicates(s []string) bool {
	s2 := removeDuplicateValues(s)
	return len(s) != len(s2)
}
