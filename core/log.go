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

package core

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Table struct {
	Header []string   `json:"h"`
	Rows   [][]string `json:"r"`
}

type Log struct {
	M    map[string]Table `json:"log"`
	home string
}

func NewLog(artHome string) (*Log, error) {
	l := &Log{
		home: artHome,
	}
	err := l.load()
	if l.M == nil {
		l.M = make(map[string]Table)
	}
	return l, err
}

func (l *Log) New(table string, header []string) error {
	_, exists := l.M[table]
	if exists {
		WarningLogger.Println("table already exists, it will be replaced")
	}
	l.M[table] = Table{
		Header: header,
		Rows:   [][]string{},
	}
	return l.save()
}
func (l *Log) Add(table string, values []string) error {
	t, exists := l.M[table]
	if !exists {
		return fmt.Errorf("table %s does not exists, use the add command to create one", table)
	}
	t.Rows = append(t.Rows, values)
	l.M[table] = t
	return l.save()
}

func (l *Log) AddFile(table string, file string) error {
	t, exists := l.M[table]
	if !exists {
		t = Table{
			Header: []string{},
			Rows:   [][]string{},
		}
	}
	abs, _ := filepath.Abs(file)
	b, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	m := make(map[string]string)
	switch filepath.Ext(abs) {
	case ".yaml":
		fallthrough
	case ".yml":
		err = yaml.Unmarshal(b, m)
		if err != nil {
			return err
		}
	case ".json":
		err = json.Unmarshal(b, m)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid format %s", filepath.Ext(abs))
	}
	var header, row []string
	for k, v := range m {
		header = append(header, k)
		row = append(row, v)
	}
	t.Rows = [][]string{}
	t.Rows = append(t.Rows, row)
	t.Header = header
	l.M[table] = t
	return l.save()
}
func (l *Log) Print() error {
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	fmt.Printf("<ARTLOG>%v</ARTLOG>\n", string(b[:]))
	return nil
}

func (l *Log) Clear() error {
	return os.Remove(logFile(l.home))
}
func (l *Log) save() error {
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(l.home, "log.art"), b, os.ModePerm)
}

func (l *Log) load() error {
	b, err := os.ReadFile(logFile(l.home))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, l)
}

func logFile(home string) string {
	path := filepath.Join(home, "log.art")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f := &Log{
			home: home,
			M:    make(map[string]Table),
		}
		if err = f.save(); err != nil {
			return ""
		}
	}
	return path
}
