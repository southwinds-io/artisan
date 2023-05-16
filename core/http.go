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
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Get make a GET HTTP request to the specified URL
func Get(url, user, pwd string) (*http.Response, error) {
	// create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// add http request headers
	if len(user) > 0 && len(pwd) > 0 {
		req.Header.Add("Authorization", BasicToken(user, pwd))
	}
	// issue http request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	// do we have a nil response?
	if resp == nil {
		return resp, errors.New(fmt.Sprintf("error: response was empty for resource: %s", url))
	}
	// check error status codes
	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("error: response returned status: %s. resource: %s", resp.Status, url))
	}
	return resp, err
}

func Curl(uri string, method string, token string, validCodes []int, payload string, file string, maxAttempts int, delaySecs int, timeoutSecs int, headers []string, outputFile string, responseHeaders bool) {
	var (
		bodyBytes []byte    = nil
		body      io.Reader = nil
		attempts            = 0
	)
	if len(payload) > 0 {
		if len(file) > 0 {
			RaiseErr("use either payload or file options, not both\n")
		}
		bodyBytes = []byte(payload)
	} else {
		if len(file) > 0 {
			abs, err := filepath.Abs(file)
			if err != nil {
				RaiseErr("cannot obtain absolute path for file using %s: %s\n", file, err)
			}
			bodyBytes, err = os.ReadFile(abs)
		}
	}
	if bodyBytes != nil {
		body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	// create request
	req, err := http.NewRequest(strings.ToUpper(method), uri, body)
	if err != nil {
		RaiseErr("cannot create http request object: %s\n", err)
	}
	// add authorization token to http request headers
	if len(token) > 0 {
		req.Header.Add("Authorization", token)
	}
	// add custom headers
	if headers != nil {
		for _, header := range headers {
			parts := strings.Split(header, ":")
			if len(parts) != 2 {
				WarningLogger.Printf("wrong format of http header '%s'; format should be 'key:value', skipping it\n", header)
				continue
			}
			req.Header.Add(parts[0], parts[1])
		}
	}
	// create http client with timeout
	client := &http.Client{
		Timeout: time.Duration(int64(timeoutSecs)) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // allows for self-signed https and http endpoints
			},
		},
	}
	// issue http request
	resp, err := client.Do(req)
	// retry if error or invalid response code
	for err != nil || !validResponse(resp.StatusCode, validCodes) {
		if err != nil {
			if resp != nil {
				ErrorLogger.Printf("unexpected error with response code '%d'; error was: '%s', retrying attempt %d of %d in %d seconds, please wait...\n", resp.StatusCode, err, attempts+1, maxAttempts, delaySecs)
			} else {
				ErrorLogger.Printf("unexpected error with no response; error was: '%s', retrying attempt %d of %d in %d seconds, please wait...\n", err, attempts+1, maxAttempts, delaySecs)
			}
		} else {
			// read http response body
			respBody, _ := io.ReadAll(resp.Body)
			ErrorLogger.Printf("invalid response code: '%d', body: '%s'; retrying attempt %d of %d in %d seconds, please wait...\n", resp.StatusCode, respBody, attempts+1, maxAttempts, delaySecs)
		}
		// wait for next attempt
		time.Sleep(time.Duration(int64(delaySecs)) * time.Second)
		// issue http request
		resp, err = client.Do(req)
		// increments the number of attempts
		attempts++
		// exits if max attempts reached
		if attempts >= maxAttempts {
			RaiseErr("%s request to '%s' failed after %d attempts\n", strings.ToUpper(method), uri, maxAttempts)
		}
	}
	// if there is a response body prints it to stdout
	if resp != nil && resp.Body != nil {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			WarningLogger.Printf("cannot print response body: %s\n", err)
		} else {
			if len(outputFile) > 0 {
				// saves the response to a file
				abs, err := filepath.Abs(outputFile)
				if err != nil {
					RaiseErr("cannot save response body to %s: %s\n", outputFile, err)
				}
				err = os.WriteFile(abs, b, 644)
				if err != nil {
					RaiseErr("cannot save response body to %s: %s\n", abs, err)
				}
			} else {
				if responseHeaders {
					r := curlResponse{
						StatusCode: resp.StatusCode,
						Status:     resp.Status,
						Headers:    map[string]string{},
						Body:       string(b[:]),
					}
					// add headers
					for key, values := range resp.Header {
						r.Headers[key] = values[0]
					}
					js, _ := json.MarshalIndent(r, "", "  ")
					// prints the response envelope to sdt out
					fmt.Println(string(js[:]))
				} else {
					// prints the response to sdt out
					fmt.Println(string(b[:]))
				}
			}
		}
	}
}

type curlResponse struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func validResponse(responseCode int, validCodes []int) bool {
	for _, validCode := range validCodes {
		if responseCode == validCode {
			return true
		}
	}
	return false
}

// BasicToken creates a basic authentication token
func BasicToken(user string, pwd string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pwd))))
}
