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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"southwinds.dev/artisan/core"
)

// Seal the digital Seal for a package
// the Seal contains information to determine if the package or its metadata has been compromised
// and therefore the Seal is broken
type Seal struct {
	// the package metadata
	Manifest *Manifest `json:"manifest"`
	// the combined checksum of the package and its metadata
	Digest string `json:"digest"`
	// the map of digital seals for this package
	Seal map[string]string `json:"seal,omitempty"`
}

// NoAuthority returns true if the seal does not have an authority
func (seal *Seal) NoAuthority() bool {
	return seal.Manifest == nil || (seal.Manifest != nil && len(seal.Manifest.Authority) == 0)
}

// DSha256 calculates the package SHA-256 digest by taking the combined checksum of the Seal information and the compressed file
func (seal *Seal) DSha256(path string) (string, error) {
	// precondition: the manifest is required
	if seal.Manifest == nil {
		return "", fmt.Errorf("seal has no manifest, cannot create checksum")
	}
	// read the compressed file
	file, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot open seal file: %s", err)
	}
	// serialise the seal info to json
	info := core.ToJsonBytes(seal.Manifest)
	core.Debug("manifest before checksum:\n>> start on next line\n%s\n>> ended on previous line", string(info))
	hash := sha256.New()
	written, err := hash.Write(file)
	if err != nil {
		return "", fmt.Errorf("cannot write package file to hash: %s", err)
	}
	core.Debug("%d bytes from package written to hash", written)
	written, err = hash.Write(info)
	if err != nil {
		return "", fmt.Errorf("cannot write manifest to hash: %s", err)
	}
	core.Debug("%d bytes from manifest written to hash", written)
	checksum := hash.Sum(nil)
	core.Debug("seal calculated base64 encoded checksum:\n>> start on next line\n%s\n>> ended on previous line", base64.StdEncoding.EncodeToString(checksum))
	return fmt.Sprintf("sha256:%s", base64.StdEncoding.EncodeToString(checksum)), nil
}

func (seal *Seal) ZipFile(registryRoot string) string {
	return path.Join(core.RegistryPath(registryRoot), fmt.Sprintf("%s.zip", seal.Manifest.Ref))
}

func (seal *Seal) SealFile(registryRoot string) string {
	return path.Join(core.RegistryPath(registryRoot), fmt.Sprintf("%s.json", seal.Manifest.Ref))
}

// PackageId the package id calculated as the hex encoded SHA-256 digest of the artefact Seal
func (seal *Seal) PackageId() (string, error) {
	// serialise the seal info to json
	info := core.ToJsonBytes(seal)
	hash := sha256.New()
	// copy the seal content into the hash
	if _, err := io.Copy(hash, bytes.NewReader(info)); err != nil {
		return "", fmt.Errorf("cannot create hash from package seal: %s", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Valid checks that the digest stored in the seal is the same as the digest generated using the passed-in zip file path
// and the seal
// path: the path to the package zip file to validate
func (seal *Seal) Valid(path string) (valid bool, err error) {
	// calculates the digest using the zip file
	digest, err := seal.DSha256(path)
	if err != nil {
		return false, err
	}
	// compare to the digest stored in the seal
	if seal.Digest == digest {
		return true, nil
	}
	return false, fmt.Errorf("downloaded package digest: %s does not match digest in manifest %s", digest, seal.Digest)
}
