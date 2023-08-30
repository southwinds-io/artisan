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

const (
	// ArtDebug the global flag used to switch on debug / verbose mode
	ArtDebug = "ART_DEBUG"
	// ArtPackageFQDN the env var that defines the name of the artisan package to run
	ArtPackageFQDN = "ART_PACKAGE_FQDN"
	// ArtPackageDomain the domain portion specified in the artisan package name
	ArtPackageDomain = "ART_PACKAGE_DOMAIN"
	// ArtPackageGroup the group portion specified in the artisan package name
	ArtPackageGroup = "ART_PACKAGE_GROUP"
	// ArtPackageName the name portion specified in the artisan package name
	ArtPackageName = "ART_PACKAGE_NAME"
	// ArtPackageTag the tag portion specified in the artisan package name
	ArtPackageTag = "ART_PACKAGE_TAG"
	// ArtFxName the env var that defines the name of the package function to run
	ArtFxName = "ART_FX_NAME"
	// ArtPackageSource the env var that defines the type of source a runner pipeline should use
	ArtPackageSource = "ART_PACKAGE_SOURCE"
	// ArtRegUser the name of the env variable that holds the artisan registry user to authenticate with a remote registry
	// when registry related commands are executed and no specific credentials are provided via command flag
	ArtRegUser = "ART_REG_USER"
	// ArtRegPassword1 the name of the env variable that holds the artisan registry password to authenticate with a remote registry
	// when registry related commands are executed and no specific credentials are provided via command flag
	ArtRegPassword1 = "ART_REG_PWD"
	ArtRegPassword2 = "ART_REG_PASS"
	// ArtDefaultHome the default artisan home
	ArtDefaultHome = ""

	ArtReference = "ART_REF"
	ArtBuildPath = "ART_BUILD_PATH"
	ArtGitCommit = "ART_GIT_COMMIT"
	ArtWorkDir   = "ART_WORK_DIR"
	ArtFromUri   = "ART_FROM_URI"
	ArtOS        = "ART_OS"
	ArtArch      = "ART_ARCH"
	ArtShell     = "ART_SHELL"
	// ArtExeWd the path from where a package was run
	ArtExeWd = "ART_EXE_WD"
)
