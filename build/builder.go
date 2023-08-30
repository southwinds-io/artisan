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

package build

import (
	"archive/zip"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"regexp"
	"southwinds.dev/artisan/conf"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/merge"
	"southwinds.dev/artisan/registry"

	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Builder struct {
	zipWriter        *zip.Writer
	workingDir       string
	uniqueIdName     string
	repoURI          string
	commit           string
	from             string
	RepoName         *core.PackageName
	buildFile        *data.BuildFile // the build file for building the package
	localReg         *registry.LocalRegistry
	shouldCopySource bool
	loadFrom         string
	env              conf.Configuration
	artHome          string
	sProc            BuildHandler
	vProc            data.VerifyHandler
	rProc            data.RunHandler
}

type BuildHandler func(b *Builder, s *data.Seal, openP, runP, signP string) error

func NewBuilder(artHome string) *Builder {
	// create the builder instance
	builder := new(Builder)
	builder.artHome = artHome
	// check the localRepo directory is there
	builder.localReg = registry.NewLocalRegistry(artHome)
	builder.sProc = sProcessor
	return builder
}

// Build the package
// from: the source to build, either http based git repository or local system git repository
// gitToken: if provided it is used to clone a remote repository that has authentication enabled
// name: the full name of the package to be built including the tag
// profileName: the name of the profile to be built. If empty then the default profile is built. If no default profile exists, the first profile is built.
// copy: indicates whether a copy should be made of the project files before packaging (only valid for from location in the file system)
// interactive: true if the console should survey for missing variables
// target: a specific target without relying on a build file (can be either relative or absolute)
func (b *Builder) Build(from, fromPath, gitToken string, name *core.PackageName, profileName string, copy bool, interactive bool, target, openP, runP, signP string) error {
	b.from = from
	// prepare the source ready for the build
	repo := b.prepareSource(from, fromPath, gitToken, name, copy, target)
	// set the unique identifier name for both the zip file and the seal file
	b.setUniqueIdName(repo)
	// run commands
	buildProfile, err := b.runProfile(profileName, b.loadFrom, interactive)
	if err != nil {
		return err
	}
	// merge env with target
	mergedTarget, _ := core.MergeEnvironmentVars([]string{buildProfile.Target}, b.env, interactive)
	// set the merged target for later use
	buildProfile.MergedTarget = mergedTarget[0]
	// wait for the target to be created in the file system
	core.Debug("waiting for build process to complete\n")
	workingTarget := mergedTarget[0]
	if strings.HasPrefix(workingTarget, "./") || workingTarget[0] != '/' {
		workingTarget = filepath.Join(b.loadFrom, workingTarget)
	}
	waitForTargetToBeCreated(workingTarget)
	// compress the target defined in the build.yaml profile
	core.Debug("zipping target path '%s'\n", workingTarget)
	b.zipPackage(workingTarget)
	// creates a seal
	core.Debug("creating package seal\n")
	s, err := b.createSeal(buildProfile, openP, runP, signP)
	if err != nil {
		return err
	}
	// save the seal
	err = os.WriteFile(b.workDirJsonFilename(), core.ToJsonBytes(s), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write package seal file")
	}
	// add the package to the local repo
	core.Debug("adding package to local registry\n")
	err = b.localReg.Add(b.WorkDirPackageFilename(), b.RepoName, s)
	if err != nil {
		return err
	}
	// cleanup all relevant folders and move package to target location
	core.Debug("performing cleanup\n")
	b.cleanUp()
	return nil
}

// Run execute the specified function
func (b *Builder) Run(function string, path string, interactive bool, env conf.Configuration) error {
	// if no path is specified use .
	if len(path) == 0 {
		path = "."
	}
	var localPath = path
	// if a relative path is passed
	if strings.HasPrefix(path, "http") {
		return fmt.Errorf("the path must not be an http resource")
	}
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") || (!strings.HasPrefix(path, "/")) {
		// turn it into an absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		localPath = absPath
	}
	bf, err := data.LoadBuildFileWithEnv(filepath.Join(localPath, "build.yaml"), env, false)
	if err != nil {
		return err
	}
	b.buildFile = bf
	return b.runFunction(function, localPath, interactive, env)
}

// either clone a remote git repo or copy a local one onto the source folder
func (b *Builder) prepareSource(from string, fromPath string, gitToken string, tagName *core.PackageName, copy bool, targetFromFlag string) *git.Repository {
	var repo *git.Repository
	b.RepoName = tagName
	// creates a temporary working directory
	b.workingDir = b.newWorkingDir()
	core.Debug("creating temporary working directory '%s'\n", b.workingDir)
	// if "from" is an http url
	if strings.HasPrefix(strings.ToLower(from), "http") {
		b.loadFrom = b.sourceDir(b.workingDir)
		// if a sub-folder was specified
		if len(fromPath) > 0 {
			// add it to the path
			b.loadFrom = filepath.Join(b.loadFrom, fromPath)
		}
		core.Debug("cloning build source repository '%s'\n", from)
		repo = b.cloneRepo(from, gitToken)
	} else
	// there is a local repo instead of a downloadable url
	{
		var localPath = from
		// if a relative path is passed
		if strings.HasPrefix(from, "./") || (!strings.HasPrefix(from, "/")) {
			// turn it into an absolute path
			absPath, err := filepath.Abs(from)
			if err != nil {
				log.Fatal(err)
			}
			localPath = absPath
		}
		// if the user requested a copy of the project before building it
		if copy {
			b.loadFrom = b.sourceDir(b.workingDir)
			// if a sub-folder was specified
			if len(fromPath) > 0 {
				// add it to the path
				b.loadFrom = filepath.Join(b.loadFrom, fromPath)
			}
			// copy the folder to the source directory
			err := copyFolder(from, b.sourceDir(b.workingDir))
			if err != nil {
				log.Fatal(err)
			}
			b.repoURI = localPath
		} else {
			// the working directory is the current directory
			b.loadFrom = localPath
			// if a sub-folder was specified
			if len(fromPath) > 0 {
				// add it to the path
				b.loadFrom = filepath.Join(b.loadFrom, fromPath)
			}
		}
		core.Debug("opening git repository '%s'", localPath)
		repo = b.openRepo(localPath)
	}
	var (
		targetBf *data.BuildFile
		err      error
	)
	// if a specific target has not been set via flag, this is the case of a target being specified in the build file
	if len(targetFromFlag) == 0 {
		buildFilePath := filepath.Join(b.loadFrom, "build.yaml")
		core.Debug("loading build file from %s\n", buildFilePath)
		targetBf, err = data.LoadBuildFile(buildFilePath)
		core.CheckErr(err, "failed to get target build file")
		b.buildFile = targetBf
	} else {
		// there is no build file as the target has been set via flag, so it creates one dynamically for the builder to work
		core.WarningLogger.Printf("no build file found at target %s, building a content only package\n", targetFromFlag)
		// dynamically creates one that packages anything on the build target
		bf := &data.BuildFile{
			Profiles: []*data.Profile{
				{
					Name:    "content-only",
					Default: true,
					Target:  targetFromFlag,
					Type:    "content/file",
				},
			},
		}
		if ok, validationErr := bf.Validate(); !ok {
			core.RaiseErr(validationErr.Error())
		}
		b.buildFile = bf
	}
	return repo
}

// compress the target
func (b *Builder) zipPackage(targetPath string) {
	ignored := b.getIgnored()
	// get the target source information
	info, err := os.Stat(targetPath)
	core.CheckErr(err, "failed to retrieve target to compress: '%s'", targetPath)
	// if the target is a directory
	if info.IsDir() {
		// then zip it
		core.CheckErr(zipSource(targetPath, b.WorkDirPackageFilename(), ignored), "failed to compress folder")
	} else {
		core.RaiseErr("build target %s must be a folder", targetPath)
	}
}

// clones a remote git LocalRegistry, it only accepts a token if authentication is required
// if the token is not provided (empty string) then no authentication is used
func (b *Builder) cloneRepo(repoUrl string, gitToken string) *git.Repository {
	b.repoURI = repoUrl
	// clone the remote repository
	opts := &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
	}
	// if authentication token has been provided
	if len(gitToken) > 0 {
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		opts.Auth = &http.BasicAuth{
			Username: "abc123", // yes, this can be anything except an empty string
			Password: gitToken,
		}
	}
	repo, err := git.PlainClone(b.sourceDir(b.workingDir), false, opts)
	if err != nil {
		_ = os.RemoveAll(b.workingDir)
		log.Fatal(err)
	}
	return repo
}

// opens a git LocalRegistry from the given path
func (b *Builder) openRepo(path string) *git.Repository {
	// find .git path in the current directory or any parents
	gitPath, _ := findGitPath(path)
	repo, _ := git.PlainOpen(gitPath)
	return repo
}

// cleanup all relevant folders and move package to target location
func (b *Builder) cleanUp() {
	// remove the working directory
	core.CheckErr(os.RemoveAll(b.workingDir), "failed to remove temporary build directory")
	// set the directory to empty
	b.workingDir = ""
}

// create a new working directory and return its path
func (b *Builder) newWorkingDir() string {
	// the working directory will be a build folder within the registry directory
	basePath := filepath.Join(core.RegistryPath(b.artHome), "build")
	uid := uuid.New()
	folder := strings.Replace(uid.String(), "-", "", -1)[:12]
	workingDirPath := filepath.Join(basePath, folder)
	// creates a temporary working directory
	err := os.MkdirAll(workingDirPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	// create a sub-folder to zip
	err = os.MkdirAll(b.sourceDir(workingDirPath), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	return workingDirPath
}

// construct a unique name for the package using the short HEAD commit hash and current time
func (b *Builder) setUniqueIdName(repo *git.Repository) {
	var hash = ""
	// if the repo is there
	if repo != nil {
		// get the commit head and add it to the unique reference
		ref, err := repo.Head()
		if err != nil {
			// the git repo has no commits yet so cannot get a valid commit hashtag
			hash = ""
			b.commit = ""
		} else {
			hash = fmt.Sprintf("-%s", ref.Hash().String()[:10])
			b.commit = ref.Hash().String()
		}
	}
	// get the current time
	t := time.Now().UTC()
	timeStamp := fmt.Sprintf("%04s%02d%02d%02d%02d%02d%s", strconv.Itoa(t.Year()), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), strconv.Itoa(t.Nanosecond())[:3])
	b.uniqueIdName = fmt.Sprintf("%s%s", timeStamp, hash)
	core.Debug("package files name is '%s'\n", b.uniqueIdName)
}

// remove files in the source folder that are specified in the .buildignore file
func (b *Builder) getIgnored() []string {
	ignoreFilename := ".buildignore"
	// retrieve the ignore file
	ignoreFileBytes, err := os.ReadFile(filepath.Join(b.loadFrom, ".buildignore"))
	if err != nil {
		// assume no ignore file exists, do nothing
		return []string{}
	}
	// get the lines in the ignore file
	lines := strings.Split(string(ignoreFileBytes), "\n")
	// adds the .ignore file
	lines = append(lines, ignoreFilename)
	// turns relative paths into absolute paths
	var output []string
	for _, line := range lines {
		if !filepath.IsAbs(line) {
			line, err = filepath.Abs(line)
			if err != nil {
				core.RaiseErr("cannot convert relation path to absolute path: %s", err)
			}
		}
		output = append(output, line)
	}
	return output
}

// run a specified function
func (b *Builder) runFunction(function string, path string, interactive bool, env conf.Configuration) error {
	// if in debug mode, print environment variables
	core.Debug(fmt.Sprintf("executing function: %s\n", function))
	core.Debug(env.String())
	// if inputs are defined for the function then survey for data
	i, err := data.SurveyInputFromBuildFile(function, b.buildFile, interactive, false, env, b.artHome)
	if err != nil {
		return err
	}
	core.Debug("function '%s' input survey complete\n", function)
	// merge the collected input with the current environment
	env.Merge(i.Env())
	// gets the function to run
	fx := b.buildFile.Fx(function)
	if fx == nil {
		return fmt.Errorf("function %s does not exist in the build file", function)
	}
	// set the unique name for the run
	b.setUniqueIdName(b.openRepo(path))
	if len(b.from) == 0 {
		b.from = path
	}
	// get the build file environment and merge any subshell command
	vars, err := b.evalSubshell(b.buildFile.GetEnv(), path, env, interactive)
	if err != nil {
		return err
	}
	// add the merged vars to the env
	env = env.Append(vars)
	// get the fx environment and merge any subshell command
	vars, err = b.evalSubshell(fx.GetEnv(), path, env, interactive)
	if err != nil {
		return err
	}
	// combine the current environment with the function environment
	buildEnv := env.Append(vars)
	// add build specific variables
	buildEnv = buildEnv.Append(b.getBuildEnv())
	// for each run statement in the function
	for _, cmd := range fx.Run {
		// add function level vars
		buildEnv = buildEnv.Append(fx.GetEnv())
		// if the statement has a function call
		if ok, _, _ := core.HasShell(cmd); ok {
			evalCmd, evalErr := EvalShell(cmd, buildEnv)
			if evalErr != nil {
				return fmt.Errorf("cannot evaluate subshell expression in '%s': %s", evalCmd, evalErr)
			}
			// execute the statement
			err = execute(evalCmd, path, buildEnv, interactive)
			if err != nil {
				return fmt.Errorf("cannot execute command %s: %s", cmd, err)
			}
		} else if hasFx, fxName := core.HasFunction(cmd); hasFx {
			// executes the function
			runErr := b.runFunction(fxName, path, interactive, env)
			if runErr != nil {
				return runErr
			}
		} else {
			// execute the statement
			err = execute(cmd, path, buildEnv, interactive)
			if err != nil {
				return fmt.Errorf("cannot execute command %s: %s", cmd, err)
			}
		}
	}
	return nil
}

// execute all commands in the specified profile
// if not profile is specified, it uses the default profile
// if a default profile has not been defined, then uses the first profile in the build file
// returns the profile used
func (b *Builder) runProfile(profileName string, execDir string, interactive bool) (*data.Profile, error) {
	var env conf.Configuration
	// construct an environment with the vars at build file level
	env = merge.NewEnVarFromSlice(os.Environ())
	// get the build file environment and merge any subshell command
	vars, err := b.evalSubshell(b.buildFile.GetEnv(), execDir, env, interactive)
	if err != nil {
		return nil, err
	}
	// add the merged vars to the env
	env = env.Append(vars)
	if b.buildFile.Profiles == nil {
		core.RaiseErr("cannot build without at least one profile in the build file")
	}
	// for each build profile
	for _, profile := range b.buildFile.Profiles {
		// if a profile name has been provided then build it
		if len(profileName) > 0 && profile.Name == profileName {
			core.Debug("using build profile '%s'\n", profile.Name)
			// get the profile environment and merge any subshell command
			vars, err = b.evalSubshell(profile.GetEnv(), execDir, env, interactive)
			if err != nil {
				return nil, err
			}
			// combine the current environment with the profile environment
			buildEnv := env.Append(vars)
			// add build specific variables
			buildEnv = buildEnv.Append(b.getBuildEnv())
			// stores the build environment
			b.env = buildEnv
			core.Debug("profile variables:\n%s\n", buildEnv.String())
			// for each run statement in the profile
			for _, cmd := range profile.Run {
				// execute the statement
				if ok, expr, shell := core.HasShell(cmd); ok {
					out, err := Exe(shell, execDir, buildEnv, interactive)
					core.CheckErr(err, "cannot execute subshell command: %s", cmd)
					// merges the output of the subshell in the original command
					cmd = strings.Replace(cmd, expr, out, -1)
					// execute the statement
					core.Debug("executing profile command: %s; @ %s\n", cmd, execDir)
					err = execute(cmd, execDir, buildEnv, interactive)
					core.CheckErr(err, "cannot execute command: %s", cmd)
				} else if ok, fx := core.HasFunction(cmd); ok {
					// executes the function
					err := b.runFunction(fx, execDir, interactive, env)
					if err != nil {
						return nil, err
					}
				} else {
					// execute the statement
					core.Debug("executing profile command: %s; @ %s\n", cmd, execDir)
					err := execute(cmd, execDir, buildEnv, interactive)
					if err != nil {
						return nil, fmt.Errorf("cannot execute command: %s", cmd)
					}
				}
			}
			return profile, nil
		}
		// if the profile has not been provided
		if len(profileName) == 0 {
			// check if a default profile has been set
			defaultProfile := b.buildFile.DefaultProfile()
			// use the default profile
			if defaultProfile != nil {
				core.Debug("using default profile: %s\n", defaultProfile.Name)
				return b.runProfile(defaultProfile.Name, execDir, interactive)
			} else {
				core.Debug("using first profile: %s\n", b.buildFile.Profiles[0].Name)
				// there is no default profile defined so use the first profile
				return b.runProfile(b.buildFile.Profiles[0].Name, execDir, interactive)
			}
		}
	}
	// if we got to this point then a specific profile was requested but not defined
	// so cannot continue
	return nil, fmt.Errorf("the requested profile '%s' is not defined in Artisan's build configuration", profileName)
}

// evaluate sub-shells and replace their values in the variables
func (b *Builder) evalSubshell(vars map[string]string, execDir string, env conf.Configuration, interactive bool) (map[string]string, error) {
	// if env is nil then create one injecting the artisan build environment variables
	if env == nil {
		env = merge.NewEnVarFromMap(b.getBuildEnv())
	} else {
		// otherwise, add the artisan build environment variables to the existing environment
		env.MergeMap(b.getBuildEnv())
	}
	// ensures env contains the variables in vars
	env.MergeMap(vars)
	for k, v := range vars {
		// merge any existing variables in the variable
		s, _ := core.MergeEnvironmentVars([]string{v}, env, false)
		// update the value with merged expression
		vars[k] = s[0]
		if ok, expr, shell := core.HasShell(v); ok {
			core.Debug("subshell evaluation started: '%s'\n", shell)
			shell = strings.Trim(shell, " ")
			usesArtisan := strings.HasPrefix(shell, "art ")
			core.Debug("=> subshell uses artisan command: %t\n", usesArtisan)
			out, err := Exe(shell, execDir, env, interactive)
			if err != nil {
				return nil, fmt.Errorf("cannot execute subshell command '%s': %s", v, err)
			}
			// ensure the subshell output does not end with newline
			out = core.TrimNewline(out)
			core.Debug("=> shell eval output: '%s'\n", out)
			// if subshell uses art command then check for safe output
			if usesArtisan && len(out) > 0 {
				core.Debug("=> found wrapped value in subshell output\n")
				r, _ := regexp.Compile("{{.*}}")
				if matched := r.MatchString(out); matched {
					out = r.FindString(out)
					// merges the output of the subshell in the original variable
					vars[k] = strings.Replace(v, expr, out[2:len(out)-2], 1)
					core.Debug("=> unwrapped value is: '%s'\n", out[2:len(out)-2])
				} else {
					return nil, fmt.Errorf("non-empty returned value of subshell expression '%s', must be enclosed by double curly braces '{{...}}' markers to prevent potential corruption due to debug statements", shell)
				}
			} else {
				// merges the output of the subshell in the original variable
				vars[k] = strings.Replace(v, expr, out, -1)
			}
		}
	}
	return vars, nil
}

// return an absolute path using the working directory as base
func (b *Builder) inWorkingDirectory(relativePath string) string {
	return filepath.Join(b.workingDir, relativePath)
}

// return an absolute path using the source directory as base
func (b *Builder) inSourceDirectory(relativePath string) string {
	return filepath.Join(b.sourceDir(b.workingDir), relativePath)
}

// create the package Seal
func (b *Builder) createSeal(profile *data.Profile, openP, runP, signP string) (*data.Seal, error) {
	filename := b.uniqueIdName
	// merge the labels in the profile with the ones at the build file level
	labels := conf.MergeMaps(b.buildFile.Labels, profile.Labels)
	// gets the size of the package
	zipInfo, err := os.Stat(b.WorkDirPackageFilename())
	if err != nil {
		return nil, err
	}
	// prepare the seal info
	info := &data.Manifest{
		Type:    profile.Type,
		License: profile.License,
		Ref:     filename,
		OS:      runtime.GOOS,
		Profile: profile.Name,
		Labels:  labels,
		Source:  b.repoURI,
		Commit:  b.commit,
		Branch:  "",
		Target:  filepath.Base(profile.MergedTarget),
		Time:    time.Now().Format(time.RFC850),
		Size:    bytesToLabel(zipInfo.Size()),
	}
	if len(info.Type) == 0 {
		core.WarningLogger.Printf("build profile '%s' does not define a package type\n", profile.Name)
	}
	// take the hash of the zip file and seal info combined
	s := new(data.Seal)
	// the seal needs the manifest to create a checksum
	s.Manifest = info
	var buildFile *data.BuildFile
	// gets the absolute path  to the target folder
	targetFolder, _ := filepath.Abs(path.Join(b.from, profile.MergedTarget))
	// check the target folder is not a file
	f, statErr := os.Stat(targetFolder)
	if statErr == nil && !f.IsDir() {
		core.RaiseErr("the build target must be a folder, if you are packaging a single file ensure it is place in a target folder on its own")
	}
	// check if target is a folder containing a build.yaml
	packageBuildFilePath := path.Join(targetFolder, "build.yaml")
	// if the package build file does not exist
	_, statErr = os.Stat(packageBuildFilePath)
	if os.IsNotExist(statErr) {
		core.Debug("cannot find a build.yaml in the target folder '%s', building content package only\n", packageBuildFilePath)
		// then it is a content only package, so creates an empty build file so the process can continue
		// without adding functions to package manifest
		buildFile = &data.BuildFile{
			Env:       map[string]string{},
			Labels:    map[string]string{},
			Input:     &data.Input{},
			Profiles:  []*data.Profile{},
			Functions: []*data.Function{},
		}
	} else {
		// load the build file
		core.Debug("loading build file from target folder '%s'\n", packageBuildFilePath)
		buildFile, err = data.LoadBuildFile(packageBuildFilePath)
		core.CheckErr(err, "cannot load build file from target folder")
	}
	// only export functions if the target contains a build.yaml
	// if the manifest contains exported functions then include the runtime
	// image that should be used to execute such functions
	if buildFile.ExportFxs() {
		core.Debug("build file exports functions\n")
		// pick the runtime at the build file level if exists
		if len(buildFile.Runtime) > 0 {
			s.Manifest.Runtime = buildFile.Runtime
		}
		s.Manifest.SKU = buildFile.SKU
	}
	// add exported functions to the manifest
	for _, fx := range buildFile.Functions {
		// if the function is exported
		if fx.Export != nil && *fx.Export {
			core.Debug("adding inputs to the manifest for exported function '%s'\n", fx.Name)
			input, err := data.SurveyInputFromBuildFile(fx.Name, buildFile, false, true, merge.NewEnVarFromSlice(os.Environ()), b.artHome)
			if err != nil {
				return nil, err
			}
			// then grab the required inputs
			f := &data.FxInfo{
				Name:        fx.Name,
				Description: fx.Description,
				Input:       input,
				Runtime:     fx.Runtime,
				Network:     fx.Network,
			}
			if fx.Credits > 0 {
				f.Credits = fx.Credits
			}
			s.Manifest.Functions = append(s.Manifest.Functions, f)
		}
	}
	if b.sProc != nil {
		err = b.sProc(b, s, openP, runP, signP)
		if err != nil {
			return nil, fmt.Errorf("cannot post process package: %s", err)
		}
	}
	return s, nil
}

func (b *Builder) sourceDir(workingDirectory string) string {
	return filepath.Join(workingDirectory, core.AppName)
}

// the fully qualified name of the json Seal in the working directory
func (b *Builder) workDirJsonFilename() string {
	return filepath.Join(b.workingDir, fmt.Sprintf("%s.json", b.uniqueIdName))
}

// WorkDirPackageFilename the fully qualified name of the zip file in the working directory
func (b *Builder) WorkDirPackageFilename() string {
	return filepath.Join(b.workingDir, fmt.Sprintf("%s.zip", b.uniqueIdName))
}

// determine if the from location is a file system path
func (b *Builder) copySource(from string, profile *data.Profile) bool {
	// location is in the file system and no target is specified for the profile
	// should only run commands where the source is
	return !(!strings.HasPrefix(from, "http") && len(profile.Target) == 0)
}

// prepares build specific environment variables
func (b *Builder) getBuildEnv() map[string]string {
	var env = make(map[string]string)
	env[core.ArtReference] = b.uniqueIdName
	env["ARTISAN_REF"] = b.uniqueIdName // backwards compatibility
	env[core.ArtBuildPath] = b.loadFrom
	env[core.ArtGitCommit] = b.commit
	env[core.ArtWorkDir] = b.workingDir
	env[core.ArtFromUri] = b.from
	return env
}

// Execute an exported function in a package
func (b *Builder) Execute(name *core.PackageName, function string, credentials string, interactive bool, path string, preserveFiles bool, env conf.Configuration, authorisedAuthors []string, ignoreExports bool) error {
	// get a local registry handle
	local := registry.NewLocalRegistry(b.artHome)
	// check the run path exist
	core.RunPathExists(b.artHome)
	// if no path is specified
	if len(path) == 0 {
		// create a temp random path to open the package
		path = filepath.Join(core.RunPath(b.artHome), core.RandomString(10))
	} else {
		// otherwise make sure the path is absolute
		path = core.ToAbs(path)
	}
	// open the package on the temp random path
	err := local.Open(
		name,
		credentials,
		path,
		b.vProc,
		nil,
		authorisedAuthors)
	if err != nil {
		return err
	}
	// ensure the package files are removed after execution
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(path)
	a := local.FindPackageByName(name)
	// get the package seal
	seal, err := local.GetSeal(a)
	if err != nil {
		return fmt.Errorf("cannot get package seal: %s", err)
	}
	m := seal.Manifest
	// stop execution if the package was built in an OS different from the executing OS
	if runtime.GOOS == "windows" && m.OS != "windows" {
		return fmt.Errorf("cannot run package, as it was built in '%s' OS and it is trying to execute in '%s' OS\n"+
			"ensure the package is built under the executing OS\n", m.OS, runtime.GOOS)
	}
	// check the function is exported
	if isExported(m, function) || ignoreExports {
		// inject runtime information
		env.Set(core.ArtPackageFQDN, name.FullyQualifiedNameTag())
		env.Set(core.ArtPackageDomain, name.Domain)
		env.Set(core.ArtPackageGroup, name.Group)
		env.Set(core.ArtPackageName, name.Name)
		env.Set(core.ArtPackageTag, name.Tag)
		env.Set(core.ArtFxName, function)
		wd, err2 := os.Getwd()
		if err2 != nil {
			core.WarningLogger.Printf("failed to retrieve working directory: %s", err)
		}
		env.Set(core.ArtExeWd, wd)
		// run the function on the open package
		err = b.Run(function, path, interactive, env)
		if err != nil {
			return err
		}
		// if there is no instruction to preserve the open files
		if !preserveFiles {
			// remove the package files
			err = os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("cannot cleanup build path: %s", err)
			}
		}
		if b.rProc != nil {
			err = b.rProc(name, function, seal)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("the function '%s' is not defined in the package manifest, check that it has been exported in the build profile\n", function)
	}
	core.Debug("builder exe '%s'=>'%s' complete\n", name.FullyQualifiedNameTag(), function)
	return nil
}

func (b *Builder) SetBProc(p BuildHandler) {
	b.sProc = p
}

func sProcessor(b *Builder, s *data.Seal, openP, runP, signP string) error {
	// calculates the package digest used to check its integrity
	digest, err := s.DSha256(b.WorkDirPackageFilename())
	if err != nil {
		return err
	}
	core.Debug("the package digest is '%s'\n", digest)
	// writes the digest to the seal
	s.Digest = digest
	return nil
}

func (b *Builder) SetVProc(p data.VerifyHandler) {
	b.vProc = p
}

func (b *Builder) SetRProc(p data.RunHandler) {
	b.rProc = p
}
