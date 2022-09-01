/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package release

import (
	// "fmt"
	// "os/exec"
	"testing"
)

func TestExportPackage(t *testing.T) {
	s, err := NewSpec("spec.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	opt := &ExportOptions{s, "/tmp/spec-test", "", "", "", ""}
	pkgs := []string{"dnsutils"}
	if err := BuildDebianPackage(pkgs, opt); err != nil {
		t.Fatal(err)
	}
}

func TestSpecImportSpec(t *testing.T) {
	_, err := ImportSpec(ImportOptions{"/tmp/spec-test", "", "", "", nil})
	if err != nil {
		t.Fatal(err)
	}
}

func TestExportSpec4Package(t *testing.T) {
	s, err := NewSpec("/home/ubuntu/deb-pkgs/just_pkgs/spec.yaml", "")
	if err != nil {
		t.Fatal(err)
	}
	err = ExportSpec(ExportOptions{s, "/tmp/spec-test", "", "", "", ""})
	if err != nil {
		t.Fatal(err)
	}
}
