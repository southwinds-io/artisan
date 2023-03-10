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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"southwinds.dev/artisan/conf"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	t "github.com/google/uuid"
	"github.com/ohler55/ojg/jp"
	"gopkg.in/yaml.v2"
)

// ToAbs converts the path to absolute path
func ToAbs(path string) string {
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		CheckErr(err, "cannot return an absolute representation of path")
		path = abs
	}
	return path
}

// ToJsonBytes convert the passed in parameter to a Json Byte Array
func ToJsonBytes(s interface{}) []byte {
	// serialise the seal to json
	source, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	// indent the json to make it readable
	dest := new(bytes.Buffer)
	err = json.Indent(dest, source, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return dest.Bytes()
}

// RemoveElement remove an element in a slice
func RemoveElement(a []string, value string) []string {
	i := -1
	// find the value to remove
	for ix := 0; ix < len(a); ix++ {
		if a[ix] == value {
			i = ix
			break
		}
	}
	if i == -1 {
		return a
	}
	// Remove the element at index i from a.
	a[i] = a[len(a)-1] // Copy last element to index i.
	a[len(a)-1] = ""   // Erase last element (write zero value).
	a = a[:len(a)-1]   // Truncate slice.
	return a
}

func Infof(msg string, a ...interface{}) {
	InfoLogger.Printf(msg, a...)
}

func CheckErr(err error, msg string, a ...interface{}) {
	if err != nil {
		if len(msg) == 0 {
			ErrorLogger.Printf("%s\n", err)
			os.Exit(1)
		}
		ErrorLogger.Printf("%s - %s\n", fmt.Sprintf(msg, a...), err)
		os.Exit(1)
	}
}

func RaiseErr(msg string, a ...interface{}) {
	ErrorLogger.Printf(msg, a...)
	os.Exit(1)
}

func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// UserPwd returns username and password from a username:password formatted string
// if the passed-in creds string is empty then it returns empty values
// note: if the credentials are for an artisan registry then the function RegUserPwd should be used instead
func UserPwd(creds string) (user, pwd string) {
	if len(creds) == 0 {
		return "", ""
	}
	parts := strings.Split(creds, ":")
	if len(parts) == 1 {
		// tries to get password using interactive mode
		prompt := &survey.Password{
			Message: "what is the registry password? ",
		}
		var value string
		HandleCtrlC(survey.AskOne(prompt, &value, survey.WithValidator(survey.Required)))
		return parts[0], value
	}
	if len(parts) > 2 {
		RaiseErr("incorrect credentials format, it should be username[:password]")
	}
	return parts[0], parts[1]
}

// RegUserPwd returns username and password from a username:password formatted string
// if the passed-in creds string is empty then it checks if the artisan registry env variables have been set and if so,
// use their values as creds
// note: this function should be used any time remote artisan registry operations are required
func RegUserPwd(creds string) (user, pwd string) {
	// if the specified credentials are not set
	if len(creds) == 0 {
		// try and get them from the environment
		user = os.Getenv(ArtRegUser)
		pwd = os.Getenv(ArtRegPassword)
		// if successful
		if len(user) > 0 && len(pwd) > 0 {
			// return user and pwd
			return user, pwd
		} else {
			// returns no credentials
			return "", ""
		}
	}
	return UserPwd(creds)
}

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

// AbsPath return a valid absolute path that exists
func AbsPath(filePath string) (string, error) {
	var p = filePath
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", err
		}
		p = absPath
	}
	_, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	return p, nil
}

func HandleCtrlC(err error) {
	if err == terminal.InterruptErr {
		os.Exit(0)
	} else if err != nil {
		RaiseErr("run command failed in build.yaml: %s", err)
	}
}

// MergeEnvironmentVars merges environment variables in the arguments
// returns the merged command list and the updated environment variables map if interactive mode is used
func MergeEnvironmentVars(args []string, env conf.Configuration, interactive bool) ([]string, conf.Configuration) {
	var result = make([]string, len(args))
	// the updated environment if interactive mode is used
	var updatedEnv = env
	// env variable regex
	evExpression := regexp.MustCompile("\\${(.*?)}")
	// check if the args have env variables and if so merge them
	for ix, arg := range args {
		result[ix] = arg
		// find all environment variables in the argument
		matches := evExpression.FindAllString(arg, -1)
		// if we have matches
		if matches != nil {
			for _, match := range matches {
				// get the name of the environment variable i.e. the name part in "${name}"
				varName := match[2 : len(match)-1]
				// get the value of the variable
				value := env.Get(varName)
				// if not value exists and is not an embedded process variable
				if len(value) == 0 && !strings.HasPrefix(varName, "ARTISAN_") {
					// if running in interactive mode
					if interactive {
						// prompt for the value
						prompt := &survey.Input{
							Message: fmt.Sprintf("%s:", varName),
						}
						HandleCtrlC(survey.AskOne(prompt, &value, survey.WithValidator(survey.Required)))
						// add the variable to the updated environment map
						updatedEnv.Set(varName, value)
					} else {
						// changed behaviour to allow for empty variables, control of required values should be done at input level
						// RaiseErr("the environment variable '%s' is not defined, are you missing a binding? you can always run the command in interactive mode to manually input its value", name)
						WarningLogger.Printf("the environment variable '%s' is empty", varName)
					}
				}
				// merges the variable
				result[ix] = strings.Replace(result[ix], match, value, -1)
			}
		}
	}
	return result, updatedEnv
}

func HasFunction(value string) (bool, string) {
	matches := regexp.MustCompile("\\$\\((.*?)\\)").FindAllString(value, 1)
	if matches != nil {
		return true, matches[0][2 : len(matches[0])-1]
	}
	return false, ""
}

func HasShell(value string) (bool, string, string) {
	matches := regexp.MustCompile(`\$\[\s*(.*?)\s*\]`).FindAllString(value, -1)
	if matches != nil {
		return true, matches[0], matches[0][len("$[") : len(matches[0])-len("]")]
	}
	return false, "", ""
}

// RandomString gets a random string of specified length
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ToAbsPath(flowPath string) string {
	if !path.IsAbs(flowPath) {
		abs, err := filepath.Abs(flowPath)
		CheckErr(err, "cannot convert '%s' to absolute path", flowPath)
		flowPath = abs
	}
	return flowPath
}

// Encode strings to be used in tekton pipelines names
func Encode(value string) string {
	length := 30
	value = strings.ToLower(value)
	value = strings.Replace(value, " ", "-", -1)
	if len(value) > length {
		value = value[0:length]
	}
	return value
}

func Wait(uri, filter, token string, maxAttempts int) {
	var (
		filtered []interface{}
		attempts = 0
	)
	// executes the query
	filtered = httpGetFiltered(uri, token, filter)
	// if no result loop
	for len(filtered) == 0 {
		// wait for next attempt
		time.Sleep(500 * time.Millisecond)
		// executes query
		filtered = httpGetFiltered(uri, token, filter)
		// increments the number of attempts
		attempts++
		// exits if max attempts reached
		if attempts >= maxAttempts {
			RaiseErr("call to %s did not return expected value after %d attempts", uri, maxAttempts)
		}
	}
}

func httpGetFiltered(uri, token, filter string) []interface{} {
	result := httpGet(uri, token)
	var jason interface{}
	err := json.Unmarshal(result, &jason)
	CheckErr(err, "cannot unmarshal response")
	// filtered, err = jsonpath.Read(jason, filter)
	f, err := jp.ParseString(filter)
	CheckErr(err, "cannot apply filter")
	return f.Get(jason)
}

func httpGet(uri, token string) []byte {
	// create request
	req, err := http.NewRequest("GET", uri, nil)
	CheckErr(err, "cannot create new request")
	// add authorization header if there is a token defined
	if len(token) > 0 {
		req.Header.Set("Authorization", token)
	}
	// all content type should be in JSON format
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	CheckErr(err, "cannot call URI %s", uri)
	if resp.StatusCode > 299 {
		RaiseErr("http request return error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	CheckErr(err, "cannot read response body")
	// if the result is not in JSON format
	if !isJSON(body) {
		RaiseErr("the http response body was not in json format, cannot apply JSON path filter")
	}
	return body
}

func isJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

// defaults to quay.io/artisan root if not specified
func QualifyRuntime(runtime string) string {
	// container images must be in lower case
	runtime = strings.ToLower(runtime)
	// if no repository is specified then assume artisan library at quay.io/artisan
	if !strings.ContainsAny(runtime, "/") {
		return fmt.Sprintf("quay.io/artisan/%s", runtime)
	}
	return runtime
}

// IsPath requires the value conforms to a path
func IsPath(val interface{}) error {
	// the reflect value of the result
	value := reflect.ValueOf(val)

	// if the value passed in is a string
	if value.Kind() == reflect.String {
		// try and convert the value to an absolute path
		_, err := filepath.Abs(value.String())
		// if the value cannot be converted to an absolute path
		if err != nil {
			// assumes it is not a valid path
			return fmt.Errorf("value is not a valid path: %s", err)
		}
	} else {
		// if the value is not of a string type it cannot be a path
		return fmt.Errorf("value must be a string")
	}
	return nil
}

// IsURI requires the value conforms to a URI
func IsURI(val interface{}) error {
	// the reflect value of the result
	value := reflect.ValueOf(val)

	// if the value passed in is a string
	if value.Kind() == reflect.String {
		// try and parse the URI
		_, err := url.ParseRequestURI(value.String())

		// if the value cannot be converted to an absolute path
		if err != nil {
			// assumes it is not a valid path
			return fmt.Errorf("value is not a valid URI: %s", err)
		}
	} else {
		// if the value is not of a string type it cannot be a path
		return fmt.Errorf("value must be a string")
	}
	return nil
}

// IsPackageName requires the value conforms to an Artisan package name
func IsPackageName(val interface{}) error {
	// the "reflect" value of the result
	value := reflect.ValueOf(val)

	// if the value passed in is a string
	if value.Kind() == reflect.String {
		// try and parse the package name
		_, err := ParseName(value.String())
		// if the value cannot be parsed
		if err != nil {
			// it is not a valid package name
			return fmt.Errorf("value is not a valid package name: %s", err)
		}
	} else {
		// if the value is not of a string type it cannot be a path
		return fmt.Errorf("value must be a string")
	}
	return nil
}

// NewTempDir will create a temp folder with a random name and return the path
func NewTempDir(artHome string) (string, error) {
	// the working directory will be a build folder within the registry directory
	uid := t.New()
	folder := strings.Replace(uid.String(), "-", "", -1)[:12]
	tempDirPath := filepath.Join(TmpPath(artHome), folder)
	// creates a temporary working directory
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return tempDirPath, err
}

// FindFiles return a list of file names matching the specified regular expression pattern
// recursively checking all subfolders in the specified root
func FindFiles(root, extPattern string) ([]string, error) {
	regEx, err := regexp.Compile(extPattern)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(root, func(s string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := dir.Info()
		if !info.IsDir() && regEx.Match([]byte(s)) {
			abs, err := AbsPath(s)
			if err != nil {
				return err
			}
			files = append(files, abs)
		}
		return nil
	})
	return files, err
}

func TrimNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		s = s[0 : len(s)-1]
	}
	return s
}

// Extract up to a maximum of "n" occurrences of a string that has a "prefix" and a "suffix" in "content"
// prefix and suffix can be golang regex, for example using suffix = "$" means match up to the end of the line
func Extract(content, prefix, suffix string, n int) []string {
	expression := fmt.Sprintf(`(?m)%s(.*?)%s`, prefix, suffix)
	r := regexp.MustCompile(expression)
	subMatches := r.FindAllStringSubmatch(content, n)
	var result []string
	for _, subMatch := range subMatches {
		result = append(result, subMatch[1])
	}
	return result
}

// ToYamlBytes convert the passed in parameter to a Yaml Byte Array
func ToYamlBytes(s interface{}) ([]byte, error) {
	// serialise the seal to json
	data, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ToElapsedLabel returns the elapsed time until now in human friendly format
func ToElapsedLabel(rfc850time string) string {
	created, err := time.Parse(time.RFC850, rfc850time)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Now().UTC().Sub(created.UTC())
	seconds := elapsed.Seconds()
	minutes := elapsed.Minutes()
	hours := elapsed.Hours()
	days := hours / 24
	weeks := days / 7
	months := weeks / 4
	years := months / 12

	if math.Trunc(years) > 0 {
		return fmt.Sprintf("%d %s ago", int64(years), plural(int64(years), "year"))
	} else if math.Trunc(months) > 0 {
		return fmt.Sprintf("%d %s ago", int64(months), plural(int64(months), "month"))
	} else if math.Trunc(weeks) > 0 {
		return fmt.Sprintf("%d %s ago", int64(weeks), plural(int64(weeks), "week"))
	} else if math.Trunc(days) > 0 {
		return fmt.Sprintf("%d %s ago", int64(days), plural(int64(days), "day"))
	} else if math.Trunc(hours) > 0 {
		return fmt.Sprintf("%d %s ago", int64(hours), plural(int64(hours), "hour"))
	} else if math.Trunc(minutes) > 0 {
		return fmt.Sprintf("%d %s ago", int64(minutes), plural(int64(minutes), "minute"))
	}
	return fmt.Sprintf("%d %s ago", int64(seconds), plural(int64(seconds), "second"))
}

// turn label into plural if value is greater than one
func plural(value int64, label string) string {
	if value > 1 {
		return fmt.Sprintf("%ss", label)
	}
	return label
}
