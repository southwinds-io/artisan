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
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"path/filepath"
	"southwinds.dev/artisan/core"
)

// ServeCmd serve static files over http
type ServeCmd struct {
	Cmd         *cobra.Command
	port        int
	defaultRoot string // the site default root
	defaultPage string // the default file in a path if not specified
	fileListing bool
}

func NewServeCmd() *ServeCmd {
	c := &ServeCmd{
		Cmd: &cobra.Command{
			Use:   "serve [flags] PATH",
			Short: "serves static files over an http endpoint",
			Long:  `serves static files over an http endpoint`,
		},
	}
	c.Cmd.Flags().BoolVarP(&c.fileListing, "file-listing", "l", false, "if set, enables file listing")
	c.Cmd.Flags().StringVar(&c.defaultRoot, "default-root", "/", "the default web site root, does not include page (e.g. '/'")
	c.Cmd.Flags().StringVar(&c.defaultPage, "default-page", "index.html", "the default web page to render if the URL does not define a page")
	c.Cmd.Flags().IntVarP(&c.port, "port", "p", 8100, "the http port on which the server listens for connections")
	c.Cmd.Run = c.Run
	return c
}

func (c *ServeCmd) home() string {
	return fmt.Sprintf("%s/%s", c.defaultRoot, c.defaultPage)
}

func (c *ServeCmd) Run(cmd *cobra.Command, args []string) {
	var path string
	if len(args) == 0 {
		path = "."
	} else {
		path = args[0]
	}
	path, err := filepath.Abs(path)
	core.CheckErr(err, "cannot resolve absolute path")
	http.Handle("/", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if the path is root
			if r.URL.Path == "/" {
				// if a default page is set for the root that is not /
				if len(c.defaultRoot) > 0 && c.defaultRoot != "/" {
					// redirects to the selected page
					http.Redirect(w, r, c.home(), http.StatusSeeOther)
					core.InfoLogger.Printf("redirecting to %s\n", c.home())
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}(http.FileServer(customFS{
		fs:          http.Dir(path),
		defaultPage: c.defaultPage,
		dirListing:  c.fileListing,
	})))
	// start HTTP server with `http.DefaultServeMux` handler
	core.InfoLogger.Printf("serving the contents of '%s'", path)
	core.CheckErr(http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil), "cannot start http server")
}

// customFS a custom file system that overrides the default behaviour in the golang library
// to switch off directory listing
type customFS struct {
	fs          http.FileSystem
	defaultPage string
	dirListing  bool
}

func (cfs customFS) Open(path string) (http.File, error) {
	// prevent directory listing
	if !cfs.dirListing {
		f, err := cfs.fs.Open(path)
		if err != nil {
			return nil, err
		}
		s, err := f.Stat()
		if s.IsDir() {
			index := filepath.Join(path, cfs.defaultPage)
			if _, err = cfs.fs.Open(index); err != nil {
				closeErr := f.Close()
				if closeErr != nil {
					return nil, closeErr
				}
				return nil, err
			}
		}
		return f, nil
	} else {
		// standard behaviour
		return cfs.fs.Open(path)
	}
}
