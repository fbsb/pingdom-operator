/*
Copyright 2019 Fabian Sabau <fabian.sabau@gmail.com>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpcheck

import (
	"errors"
	"fmt"
	neturl "net/url"
	"strings"

	"math"
	"strconv"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

var (
	ErrEmptyURL    = errors.New("the url should not be empty string")
	ErrEmptyName   = errors.New("the name should not be empty string")
	ErrNoHost      = errors.New("the url should define at least a host")
	ErrInvalidPort = errors.New("the port is invalid")
)

func SimpleHttpCheck(name string, url string) (*pingdom.HttpCheck, error) {
	if name == "" {
		return nil, ErrEmptyName
	}

	if url == "" {
		return nil, ErrEmptyURL
	}

	parsedUrl, err := parseUrl(url)
	if err != nil {
		return nil, err
	}

	var port int
	if parsedUrl.Port() != "" {
		port, err = strconv.Atoi(parsedUrl.Port())
		if err != nil {
			return nil, ErrInvalidPort
		}
		if port < 0 || port == 0 || port > math.MaxUint16 {
			return nil, ErrInvalidPort
		}
	}

	password, _ := parsedUrl.User.Password()

	var uri = parsedUrl.RequestURI()
	if uri == "/" {
		uri = ""
	}

	check := &pingdom.HttpCheck{
		Name:       name,
		Encryption: parsedUrl.Scheme == "https",
		Username:   parsedUrl.User.Username(),
		Password:   password,
		Hostname:   parsedUrl.Hostname(),
		Port:       port,
		Url:        uri,
		Resolution: 5,
	}

	err = check.Valid()
	if err != nil {
		return nil, err
	}

	return check, nil
}

func parseUrl(raw string) (*neturl.URL, error) {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = fmt.Sprintf("http://%s", raw)
	}

	url, err := neturl.Parse(raw)
	if err != nil {
		return nil, err
	}

	// Default scheme is http
	if url.Scheme == "" {
		url.Scheme = "http"
	}

	// Parsing "example.com" results in a url with no host part and "example.com" as the path.
	// Since we need at least a host part we swap these around.
	if url.Host == "" && url.Path != "" {
		url.Host = url.Path
		url.Path = ""
	}

	if url.Host == "" {
		return nil, ErrNoHost
	}

	return url, nil
}
