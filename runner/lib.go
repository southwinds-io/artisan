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

package runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/merge"
	"strconv"
	"strings"
	"syscall"
)

// launch a container and mount the current directory on the host machine into the container
// the current directory must contain a build.yaml file where fxName is defined
func runBuildFileFx(runtimeName, fxName, dir, containerName, network string, env *merge.Envar, artHome string) error {
	// check the local registry path has not been created by the root user othewise the runtime will error
	registryPath := core.RegistryPath(artHome)
	if runtime.GOOS == "linux" && strings.HasPrefix(registryPath, "//") {
		// in linux if the user is not root but the local registry folder is owned by the root user, then
		// the registry path in a runtime will start with two consecutive forward slashes
		core.RaiseErr("cannot continue, the local registry folder is owned by root\n" +
			"ensure it is owned by the non root user for the runtime to work")
	}
	if env == nil {
		env = merge.NewEnVarFromSlice([]string{})
	}
	// adds debug mode
	if len(os.Getenv(core.ArtDebug)) > 0 {
		env.Add(core.ArtDebug, "true")
	}
	// determine which container tool is available in the host
	tool, err := containerCmd()
	if err != nil {
		return err
	}
	// add runtime vars
	env.Add(core.ArtFxName, fxName)
	// get the docker run arguments
	args := toContainerArgs(runtimeName, dir, containerName, network, env, artHome)
	// launch the container with an art exec command
	cmd := exec.Command(tool, args...)
	core.Debug("! launching runtime: %s %s\n", tool, strings.Join(args, " "))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed creating command stdoutpipe: %s", err)
	}
	defer func() {
		_ = stdout.Close()
	}()
	stdoutReader := bufio.NewReader(stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed creating command stderrpipe: %s", err)
	}
	defer func() {
		_ = stderr.Close()
	}()
	stderrReader := bufio.NewReader(stderr)

	if err = cmd.Start(); err != nil {
		return err
	}

	go handleReader(stdoutReader)
	go handleReader(stderrReader)

	if err = cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if _, ok = exitErr.Sys().(syscall.WaitStatus); ok {
				return exitErr
			}
		}
		return err
	}
	return nil
}

// launch a container and execute a package function
func runPackageFx(runtimeName, packageName, fxName, containerName, artRegistryUser, artRegistryPwd, network string, env *merge.Envar, artHome string) error {
	// determine which container tool is available in the host
	tool, err := containerCmd()
	if err != nil {
		return err
	}
	// add add runtime vars
	env.Add(core.ArtPackageFQDN, packageName)
	env.Add(core.ArtFxName, fxName)
	env.Add(core.ArtRegUser, artRegistryUser)
	env.Add(core.ArtRegPassword1, artRegistryPwd)
	env.Add(core.ArtRegPassword2, artRegistryPwd)
	// adds debug mode
	if len(os.Getenv(core.ArtDebug)) > 0 {
		env.Add(core.ArtDebug, "true")
	}
	// create a slice with docker run args
	args := toContainerArgs(runtimeName, "", containerName, network, env, artHome)
	// launch the container with an art exec command
	cmd := exec.Command(tool, args...)
	core.Debug("! launching runtime: %s %s\n", tool, strings.Join(args, " "))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed creating command stdoutpipe: %s", err)
	}
	defer func() {
		_ = stdout.Close()
	}()
	stdoutReader := bufio.NewReader(stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed creating command stderrpipe: %s", err)
	}
	defer func() {
		_ = stderr.Close()
	}()
	stderrReader := bufio.NewReader(stderr)

	if err = cmd.Start(); err != nil {
		return err
	}

	go handleReader(stdoutReader)
	go handleReader(stderrReader)

	if err = cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if _, ok = exitErr.Sys().(syscall.WaitStatus); ok {
				return exitErr
			}
		}
		return err
	}
	return nil
}

// return the command to run to launch a container
func containerCmd() (string, error) {
	if isCmdAvailable("docker") {
		return "docker", nil
	} else if isCmdAvailable("podman") {
		return "podman", nil
	}
	return "", fmt.Errorf("either podman or docker is required to launch a container")
}

// checks if a command is available
func isCmdAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// return an array of environment variable arguments to pass to docker
func toContainerArgs(imageName, dir, containerName, network string, env *merge.Envar, artHome string) []string {
	var result = []string{"run", "--name", containerName}
	vars := env.Slice()
	for _, v := range vars {
		result = append(result, "-e")
		result = append(result, v)
	}
	// attach to network if defined
	if len(network) > 0 {
		result = append(result, "--network", network)
	}

	// create bind mounts
	// note: in order to allow for art runc command to access host mounted files in linux with selinux enabled, a :Z label
	// is added to the volume see https://docs.docker.com/storage/bind-mounts/#configure-the-selinux-label
	// Z modify the selinux label of the host file or directory being mounted into the container indicating that the
	// bind mount content is private and unshared.
	if len(dir) > 0 {
		// add a bind mount for the current folder to the /workspace/source in the runtime
		result = append(result, "-v")
		result = append(result, fmt.Sprintf("%s:%s", dir, "/workspace/source:Z"))
	}
	// add a bind mount for the artisan registry folder
	result = append(result, "-v")
	// note: mind the location of the mount in the runtime must align with its user home!
	result = append(result, fmt.Sprintf("%s:%s", core.RegistryPath(artHome), "/home/runtime/.artisan:Z"))
	result = append(result, imageName)
	return result
}

func handleReader(reader *bufio.Reader) {
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Print(str)
	}
}

func isRunning(containerName string) bool {
	tool, err := containerCmd()
	core.CheckErr(err, "")
	cmd := exec.Command(tool, "container", "inspect", "-f", "'{{.State.Running}}'", containerName)
	out, _ := cmd.Output()
	if strings.Contains(strings.ToLower(string(out)), "error") {
		return false
	}
	running, err := strconv.ParseBool(string(out))
	if err != nil {
		return false
	}
	return running
}

// removes a docker container
func removeContainer(containerName string) {
	tool, err := containerCmd()
	core.CheckErr(err, "")
	rm := exec.Command(tool, "rm", containerName)
	out, err := rm.Output()
	if err != nil {
		core.InfoLogger.Printf("%s\n", string(out))
		core.CheckErr(err, "cannot remove temporary container %s", containerName)
	}
}

// check the specified function is in the manifest
func isExported(m *data.Manifest, fx string) bool {
	for _, function := range m.Functions {
		if function.Name == fx {
			return true
		}
	}
	return false
}
