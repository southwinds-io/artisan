/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package release

import (
	"fmt"
	"southwinds.dev/artisan/build"
	"southwinds.dev/artisan/data"
)

type ImportOptions struct {
	TargetUri         string
	TargetCreds       string
	Filter            string
	ArtHome           string
	DecryptProc       data.VProc
	VerifyProc        data.VProc
	RunProc           data.RProc
	AuthorisedAuthors []string
}

func (o ImportOptions) Valid() error {
	if len(o.TargetUri) == 0 {
		return fmt.Errorf("missing target URI")
	}
	return nil
}

type ExportOptions struct {
	Specification *Spec
	TargetUri     string
	SourceCreds   string
	TargetCreds   string
	Filter        string
	ArtHome       string
	BuildProc     build.BProc
}

func (o ExportOptions) Valid() error {
	if len(o.TargetUri) == 0 {
		return fmt.Errorf("missing target URI")
	}
	if o.Specification == nil {
		return fmt.Errorf("missing specification")
	}
	return nil
}

type UpDownOptions struct {
	TargetUri   string
	TargetCreds string
	LocalPath   string
}

func (o UpDownOptions) Valid() error {
	if len(o.TargetUri) == 0 {
		return fmt.Errorf("missing target URI")
	}
	return nil
}

type PullOptions struct {
	TargetUri   string
	SourceCreds string
	TargetCreds string
	ArtHome     string
}

func (o PullOptions) Valid() error {
	if len(o.TargetUri) == 0 {
		return fmt.Errorf("missing target URI")
	}
	return nil
}

type PushOptions struct {
	SpecPath string
	Host     string
	Group    string
	User     string
	Creds    string
	Image    bool
	Clean    bool
	Logout   bool
	ArtHome  string
}

func (o PushOptions) Valid() error {
	if len(o.Host) == 0 {
		return fmt.Errorf("missing host")
	}
	return nil
}
