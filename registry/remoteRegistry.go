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

package registry

import (
	"fmt"
	"os"
	"regexp"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/i18n"
	"strings"
	"text/tabwriter"
)

// RemoteRegistry enables admin operations on a remote registry
type RemoteRegistry struct {
	domain  string
	user    string
	pwd     string
	api     *Api
	ArtHome string `json:"-"`
}

// NewRemoteRegistry creates an object to manage a remote registry
func NewRemoteRegistry(domain, user, pwd, artHome string) (*RemoteRegistry, error) {
	if strings.HasPrefix(domain, "http") {
		return nil, fmt.Errorf("remote registry domain '%s' should not specify protocol scheme", domain)
	}
	if strings.Contains(domain, "/") {
		return nil, fmt.Errorf("remote registry domain '%s' should not contain slashes", domain)
	}
	return &RemoteRegistry{
		domain:  domain,
		user:    user,
		pwd:     pwd,
		api:     newGenericAPI(domain, artHome),
		ArtHome: artHome,
	}, nil
}

// List all packages in the remote registry
func (r *RemoteRegistry) List(quiet bool) {
	showWarnings := !quiet
	// get a reference to the remote registry
	repos, err, _, _ := r.api.GetAllRepositoryInfo(r.user, r.pwd, showWarnings)
	core.CheckErr(err, "cannot list remote registry packages")
	if quiet {
		// get a table writer for the stdout
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', 0)
		defer func(w *tabwriter.Writer) {
			_ = w.Flush()
		}(w)
		// repository, tag, package id, created, size
		for _, repo := range repos {
			for _, a := range repo.Packages {
				_, err = fmt.Fprintln(w, fmt.Sprintf("%s", a.Id[0:12]))
				core.CheckErr(err, "failed to write package Id")
			}
		}
	} else {
		// get a table writer for the stdout
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
		defer func(w *tabwriter.Writer) {
			_ = w.Flush()
		}(w)
		// print the header row
		_, err = fmt.Fprintln(w, i18n.String(r.ArtHome, i18n.LBL_LS_HEADER))
		core.CheckErr(err, "failed to write table header")
		// repository, tag, package id, created, size
		for _, repo := range repos {
			for _, a := range repo.Packages {
				for _, tag := range a.Tags {
					_, err = fmt.Fprintln(w, fmt.Sprintf("%s\t %s\t %s\t %s\t %s\t %s\t",
						fmt.Sprintf("%s/%s", r.domain, repo.Repository),
						tag,
						a.Id[0:12],
						a.Type,
						core.ToElapsedLabel(a.Created),
						a.Size),
					)
					core.CheckErr(err, "failed to write output")
				}
			}
		}
	}
}

// RemoveByNameFilter remove one or more packages whose name matches the filter regex
func (r *RemoteRegistry) RemoveByNameFilter(filter string, dryRun bool) error {
	repos, err, _, tls := r.api.GetAllRepositoryInfo(r.user, r.pwd, true)
	if err != nil {
		return err
	}
	if dryRun {
		core.InfoLogger.Printf("searching candidates for removal:\n")
	}
	for _, repo := range repos {
		for _, p := range repo.Packages {
			var tagCount = len(p.Tags)
			for _, tag := range p.Tags {
				name, err1 := core.ParseName(fmt.Sprintf("%s/%s:%s", r.domain, repo.Repository, tag))
				if err1 != nil {
					return err1
				}
				matched, err2 := regexp.MatchString(filter, name.String())
				if err2 != nil {
					return fmt.Errorf("invalid filter expression '%s': %s", filter, err2)
				}
				if matched {
					if dryRun {
						core.InfoLogger.Printf("=> %s\n", name.FullyQualifiedNameTag())
						continue
					}
					// if more than one tag exist, remove the tag
					if tagCount > 1 {
						// get the package metadata
						pInfo, err3 := r.api.GetPackageInfo(name.Group, name.Name, p.Id, r.user, r.pwd, tls)
						if err3 != nil {
							return err3
						}
						// remove the tag
						pInfo.RemoveTag(tag)
						// push the metadata back to the remote
						err3 = r.api.UpsertPackageInfo(name, pInfo, r.user, r.pwd, tls)
						if err3 != nil {
							return err3
						}
					}
					// if we are hitting the last tag
					if tagCount == 1 {
						// remove the package files
						if err = r.api.DeletePackage(name.Group, name.Name, tag, r.user, r.pwd, tls); err != nil {
							return err
						}
						if err = r.api.DeletePackageInfo(name.Group, name.Name, p.Id, r.user, r.pwd, tls); err != nil {
							return err
						}
					}
					tagCount--
				}
			}
		}
	}
	return nil
}

// RemoveByNameOrId remove any package matching the passed in name or id
func (r *RemoteRegistry) RemoveByNameOrId(nameOrId []string) error {
	repos, err, _, tls := r.api.GetAllRepositoryInfo(r.user, r.pwd, true)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		for _, p := range repo.Packages {
			var tagCount = len(p.Tags)
			// for each name or Id provided
			for _, nameId := range nameOrId {
				for _, tag := range p.Tags {
					// construct a package name:tag
					name, err1 := core.ParseName(fmt.Sprintf("%s/%s:%s", r.domain, repo.Repository, tag))
					if err1 != nil {
						return err1
					}
					// check a match for package Id first
					if strings.Contains(p.Id, nameId) {
						// remove the package files by package Id
						if err = r.api.DeletePackage(name.Group, name.Name, name.Tag, r.user, r.pwd, tls); err != nil {
							return err
						}
						// this would delete any tags with the package
						if err = r.api.DeletePackageInfo(name.Group, name.Name, p.Id, r.user, r.pwd, tls); err != nil {
							return err
						}
						// so if more than 1 tag exist, no need to loop again
						if len(p.Tags) > 1 {
							break
						}
					} else { // check for a match on package name:tag
						// parses the nameId string so that it can check for not present latest tag
						n, e := core.ParseName(nameId)
						if e == nil {
							// puts the package name back together from structured name object so that if there is a
							// not present latest tag, it is made explicit
							nameId = n.FullyQualifiedNameTag()
						}
						if nameId == name.FullyQualifiedNameTag() {
							// delete by name:tag here
							// if more than one tag exist, remove the tag
							if tagCount > 1 {
								// get the package metadata
								pInfo, err3 := r.api.GetPackageInfo(name.Group, name.Name, p.Id, r.user, r.pwd, tls)
								if err3 != nil {
									return err3
								}
								// remove the tag
								pInfo.RemoveTag(tag)
								// push the metadata back to the remote
								err3 = r.api.UpsertPackageInfo(name, pInfo, r.user, r.pwd, tls)
								if err3 != nil {
									return err3
								}
							}
							// if we are hitting the last tag
							if tagCount == 1 {
								// remove the package files
								if err = r.api.DeletePackage(name.Group, name.Name, tag, r.user, r.pwd, tls); err != nil {
									return err
								}
								if err = r.api.DeletePackageInfo(name.Group, name.Name, p.Id, r.user, r.pwd, tls); err != nil {
									return err
								}
							}
							tagCount--
						}
					}
				}
			}
		}
	}
	return nil
}

// GetDigest get the digest of the package in the remote registry
func (r *RemoteRegistry) GetDigest(name *core.PackageName) (*DigestInfo, error, int) {
	var useTls = true
	digest, err, status := r.api.GetDigest(name.Group, name.Name, name.Tag, r.user, r.pwd, useTls)
	if err != nil {
		useTls = false
		digest, err, status = r.api.GetDigest(name.Group, name.Name, name.Tag, r.user, r.pwd, useTls)
		if err != nil {
			return nil, fmt.Errorf("cannot get remote repository information"), status
		}
	}
	return digest, nil, -1
}

func printPackages(name []string) {
	core.InfoLogger.Printf("dry-run found %d candidates for removal:\n", len(name))
	for i, n := range name {
		core.InfoLogger.Printf("%d => %s\n", i, n)
	}
}
