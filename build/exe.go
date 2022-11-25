package build

/*
  Artisan - © 2018-Present SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mattn/go-shellwords"
	"os"
	"os/exec"
	"runtime"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/merge"
	"strings"
	"syscall"
)

// ExeAsync executes a command and sends output and error streams asynchronously
func ExeAsync(cmd string, dir string, env *merge.Envar, interactive bool) (string, error) {
	if cmd == "" {
		return "", errors.New("no command provided")
	}
	// create a command parser
	p := shellwords.NewParser()
	// parse the command line
	cmdArr, err := p.Parse(cmd)
	if err != nil {
		return "", err
	}
	// if we are in windows
	if runtime.GOOS == "windows" {
		// prepend "cmd /C" to the command line
		cmdArr = append([]string{"cmd", "/C"}, cmdArr...)
		core.Debug("windows cmd => %s", cmdArr)
	}
	name := cmdArr[0]

	var args []string
	if len(cmdArr) > 1 {
		args = cmdArr[1:]
	}

	args, _ = core.MergeEnvironmentVars(args, env.Vars, interactive)

	// create the command to execute
	command := exec.Command(name, args...)
	// set the command working directory
	command.Dir = dir
	// set the command environment
	command.Env = env.Slice()

	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed creating command stdoutpipe: %s", err)
	}
	defer func() {
		_ = stdout.Close()
	}()
	stdoutReader := bufio.NewReader(stdout)

	stderr, err := command.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed creating command stderrpipe: %s", err)
	}
	defer func() {
		_ = stderr.Close()
	}()
	stderrReader := bufio.NewReader(stderr)

	// start the execution of the command
	if err = command.Start(); err != nil {
		return "", err
	}

	// asynchronous print output
	sOut := &strings.Builder{}
	go printOut(stdoutReader, sOut, false)
	sErr := &strings.Builder{}
	go printOut(stderrReader, sErr, true)

	// wait for the command to complete
	if err = command.Wait(); err != nil {
		core.Debug("stdout='%s'", sOut.String())
		if err != nil || len(sErr.String()) > 0 {
			core.Debug("artisan runner exec error: '%s' - stderr='%s'", err, sErr.String())
		}
		// only happens if the command exits with os.Exit(>0)
		if exitErr, ok := err.(*exec.ExitError); ok {
			core.Debug("artisan runner exec error from os.Exit == %d: '%s'", exitErr.ExitCode(), string(exitErr.Stderr[:]))
			var v syscall.WaitStatus
			if v, ok = exitErr.Sys().(syscall.WaitStatus); ok {
				core.Debug("WaitStatus = %d", v)
				return sOut.String(), fmt.Errorf("run command failed: '%s'- %s - %s (%s)", cmd, sErr.String(), exitErr, exitMsg(exitErr.ExitCode()))
			}
		}
		return sOut.String(), err
	}
	core.Debug("artisan runner exec successful, stdout='%s'", sOut.String())
	return sOut.String(), nil
}

// Exe executes a command and sends output and error streams to stdout and stderr
func Exe(cmd string, dir string, env *merge.Envar, interactive bool) (string, error) {
	if cmd == "" {
		return "", errors.New("no command provided")
	}
	// create a command parser
	p := shellwords.NewParser()
	// parse the command line
	cmdArr, err := p.Parse(cmd)
	if err != nil {
		return "", err
	}
	// if we are in windows
	if runtime.GOOS == "windows" {
		// prepend "cmd /C" to the command line
		cmdArr = append([]string{"cmd", "/C"}, cmdArr...)
		core.Debug("windows cmd => %s", cmdArr)
	}
	name := cmdArr[0]

	var args []string
	if len(cmdArr) > 1 {
		args = cmdArr[1:]
	}

	args, _ = core.MergeEnvironmentVars(args, env.Vars, interactive)

	// create the command to execute
	command := exec.Command(name, args...)
	// set the command working directory
	command.Dir = dir
	// set the command environment
	command.Env = env.Slice()
	// capture the command output and error streams in a buffer
	var outbuf, errbuf strings.Builder // or bytes.Buffer
	command.Stdout = &outbuf
	command.Stderr = &errbuf

	// start the execution of the command
	if err := command.Start(); err != nil {
		return "", err
	}

	// wait for the command to complete
	if err := command.Wait(); err != nil {
		// only happens if the command exits with os.Exit(>0)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if _, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return "", fmt.Errorf("run command failed: '%s'- %s - %s (%s)", cmd, errbuf.String(), exitErr, exitMsg(exitErr.ExitCode()))
			}
		}
		return "", err
	}

	// NOTE: I have observed that some programs exit with no error (code 0) but write to stderr instead of stdout
	// probably due to misuse of the print() function in golang or lack mistake
	// this condition can be found if we reach this point and the errbuf contains bytes
	// at this point I have to assume that as the exit code is 0 there is no actual error and whatever is in stderr
	// it should be in stdout, therefore code below
	if len(errbuf.String()) > 0 {
		// append to stdout
		outbuf.WriteString(errbuf.String())
		if core.InDebugMode() {
			// issue a warning to alert people just in case
			core.WarningLogger.Printf("command %s returned successfully but data was found in stderr. it is assumed that it is not an error and therefore, it has been added to stdout\n", cmd)
		}
	}

	return outbuf.String(), err
}

func ExeStream(cmd string, dir string, env *merge.Envar, interactive bool) error {
	if cmd == "" {
		return errors.New("no command provided")
	}
	// create a command parser
	p := shellwords.NewParser()
	// parse the command line
	cmdArr, err := p.Parse(cmd)
	if err != nil {
		return err
	}
	// if we are in windows
	if runtime.GOOS == "windows" {
		// prepend "cmd /C" to the command line
		cmdArr = append([]string{"cmd", "/C"}, cmdArr...)
		core.Debug("windows cmd => %s", cmdArr)
	}
	name := cmdArr[0]

	var args []string
	if len(cmdArr) > 1 {
		args = cmdArr[1:]
	}

	args, _ = core.MergeEnvironmentVars(args, env.Vars, interactive)

	// create the command to execute
	command := exec.Command(name, args...)
	// set the command working directory
	command.Dir = dir
	// set the command environment
	command.Env = env.Slice()
	// sends the command output and error streams to std
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// run the command
	return command.Run()
}

// print the content of the reader to stdout
func printOut(reader *bufio.Reader, out *strings.Builder, isStdErr bool) {
	for {
		str, err := reader.ReadString('\n')
		// if we are in nested execution scenarios there might be already log headers
		if strings.Contains(str, "ART INFO") || strings.Contains(str, "ART ERROR") || strings.Contains(str, "ART WARNING") {
			// then prints directly to stdout to avoid repeating log headers
			fmt.Print(str)
		} else if len(str) > 0 {
			if isStdErr {
				core.ErrorLogger.Print(str)
			} else {
				core.InfoLogger.Print(str)
			}
		}
		// and collect the output for further use if there is one
		if len(str) > 0 {
			out.WriteString(str)
		}
		// exits after collecting output and printing to stdout to avoid swallowing the last line in case there is err == EOF
		if err != nil {
			break
		}
	}
}

// print the content of the reader to stderr
func printErr(reader *bufio.Reader) {
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		core.ErrorLogger.Print(str)
	}
}
