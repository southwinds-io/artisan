/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
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
	// ArtRegPassword the name of the env variable that holds the artisan registry password to authenticate with a remote registry
	// when registry related commands are executed and no specific credentials are provided via command flag
	ArtRegPassword = "ART_REG_PWD"
	// ArtDefaultHome the default artisan home
	ArtDefaultHome = ""
)
