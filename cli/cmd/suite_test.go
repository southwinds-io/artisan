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
	"github.com/zcalusic/sysinfo"
	"log"
	"os"
	"regexp"
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/i18n"
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/registry"
	"southwinds.dev/artisan/runner"
	"testing"
)

func TestExeC(t *testing.T) {
	packageName := "uri/recipe/java-quarkus"
	fxName := "setup"
	// create an instance of the runner
	run, err := runner.New()
	core.CheckErr(err, "cannot initialise runner")
	env, err := merge.NewEnVarFromFile(".env")
	if err != nil {
		fmt.Printf("cannot load env file: %s\n", err.Error())
		t.FailNow()
	}
	// launch a runtime to execute the function
	err = run.ExeC(packageName, fxName, "admin:sss", "", false, env)
	i18n.Err("", err, i18n.ERR_CANT_EXEC_FUNC_IN_PACKAGE, fxName, packageName)
}

func TestExe(t *testing.T) {
	packageName, err := core.ParseName("test")
	fxName := "test"
	builder := build.NewBuilder(core.ArtDefaultHome)
	core.CheckErr(err, "cannot initialise builder")
	env, err := merge.NewEnVarFromFile(".env")
	if err != nil {
		fmt.Printf("cannot load env file: %s\n", err.Error())
		t.FailNow()
	}
	// launch a runtime to execute the function
	err = builder.Execute(packageName, fxName, "admin:sss", true, "", false, env, []string{}, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuild(t *testing.T) {
	packageName, _ := core.ParseName("test")
	builder := build.NewBuilder(core.ArtDefaultHome)
	builder.Build(".", "", "", packageName, "", false, false, "", "", "", "")
}

func TestRunC(t *testing.T) {
	run, err := runner.NewFromPath(".", core.ArtDefaultHome)
	core.CheckErr(err, "cannot initialise runner")
	err = run.RunC("deploy", false, merge.NewEnVarFromSlice([]string{}), "")
}

func TestPush(t *testing.T) {
	reg := registry.NewLocalRegistry(core.ArtDefaultHome)
	name, err := core.ParseName("localhost:8080/lib/test")
	if err != nil {
		t.FailNow()
	}
	err = reg.Push(name, "admin:admin", false)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
}

func TestPull(t *testing.T) {
	reg := registry.NewLocalRegistry(core.ArtDefaultHome)
	name, err := core.ParseName("localhost:8082/tools/artisan")
	if err != nil {
		t.FailNow()
	}
	reg.Pull(name, "admin:admin", false)
}

func TestRLs(t *testing.T) {
	reg, _ := registry.NewRemoteRegistry("localhost:8080", "admin", "adm1n", core.ArtDefaultHome)
	reg.List(false)
}

func TestVars(t *testing.T) {
	env, _ := merge.NewEnVarFromFile(".env")
	builder := build.NewBuilder(core.ArtDefaultHome)
	builder.Run("test", ".", false, env)
}

// test the merging of .tem templates
func TestMergeTem(t *testing.T) {
	filename := "test/test.txt"
	tm, err := merge.NewTemplMerger()
	checkErr(err, t)
	err = tm.LoadTemplates([]string{filename + ".tem"})
	checkErr(err, t)
	err = tm.Merge(merge.NewEnVarFromSlice([]string{"VAR1=World"}))
	checkErr(err, t)
	tm.Save()
	_, err = os.Stat(filename)
	checkErr(err, t)
	_ = os.Remove(filename)
}

// test the merging of .art templates
func TestMergeArt(t *testing.T) {
	filename := "test/test.txt"
	tm, err := merge.NewTemplMerger()
	checkErr(err, t)
	err = tm.LoadTemplates([]string{filename + ".art"})
	checkErr(err, t)
	err = tm.Merge(merge.NewEnVarFromSlice([]string{"VAR1=World"}))
	checkErr(err, t)
	tm.Save()
	_, err = os.Stat(filename)
	checkErr(err, t)
	_ = os.Remove(filename)
}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
}

func TestRun(t *testing.T) {
	builder := build.NewBuilder(core.ArtDefaultHome)
	// add the build file level environment variables
	env := merge.NewEnVarFromSlice(os.Environ())
	// execute the function
	builder.Run("do", ".", false, env)
}

func TestCurl(t *testing.T) {
	core.Curl("http://localhost:8080/user/ONIX_PILOTCTL",
		"PUT",
		core.BasicToken("admin", "0n1x"),
		[]int{200, 201},
		"{\n  \"email\":\"a@a.com\", \"name\":\"aa\", \"pwd\":\"aaAA88!=12222\", \"service\":\"false\", \"acl\":\"*:*:*\"\n}",
		"",
		5,
		5,
		5,
		[]string{"Content-Type: application/json"},
		"", false)
}

func TestSave(t *testing.T) {
	names, err := core.ValidateNames([]string{"test", "artisan"})
	if err != nil {
		t.Error(err)
	}
	r := registry.NewLocalRegistry(core.ArtDefaultHome)
	_, err = r.ExportPackage(names, "", "./export", "")
	if err != nil {
		t.Error(err)
	}
}

func TestRemove(t *testing.T) {
	r := registry.NewLocalRegistry(core.ArtDefaultHome)
	p := r.AllPackages()
	for _, s := range p {
		fmt.Println(s)
	}
	err := r.Remove(p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveRemote(t *testing.T) {
	r, _ := registry.NewRemoteRegistry("localhost:8080", "admin", "admin", "")
	err := r.RemoveByNameOrId([]string{"cfe1761845c7", "fb7d78733eaf"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewSpecPush(t *testing.T) {
	// err := export.PushSpec()
}

func TestRemoveAll(t *testing.T) {
	l := registry.NewLocalRegistry("")
	l.RemoveAll()
}

func TestEnvPackage(t *testing.T) {
	var input *data.Input
	name, err := core.ParseName("play/installer:1.0b1")
	core.CheckErr(err, "invalid package name: %s", name)
	local := registry.NewLocalRegistry("")
	manifest := local.GetManifest(name)
	fxName := "deploy-horizon"
	fx := manifest.Fx(fxName)
	input = fx.Input
	// add the credentials to download the package
	input.SurveyRegistryCreds(name.Group, name.Name, "", name.Domain, false, true, merge.NewEnVarFromSlice([]string{}))
	input.ToEnvFile()
}

func TestUReplace(t *testing.T) {
	r, _ := regexp.Compile(`"paths":\[".*"\]`)
	content, _ := os.ReadFile("file.yaml")
	replaceString := `"paths":["/mnt/k3s-storage"]`
	if len(replaceString) > 0 {
		replaced := r.ReplaceAll(content, []byte(replaceString))
		fmt.Println(string(replaced))
	}
}

func TestOS(t *testing.T) {
	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}

func TestUpConf(t *testing.T) {
	updateConf("image.yaml", []string{"labels:dog,cat,other", "triggers:cat|0.8|5"})
}
