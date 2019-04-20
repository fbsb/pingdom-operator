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
	"testing"

	"github.com/russellcardullo/go-pingdom/pingdom"
	"github.com/stretchr/testify/assert"
)

func TestSimpleHttpCheck(t *testing.T) {
	type args struct {
		name string
		url  string
	}
	tests := []struct {
		name string
		args args
		c    *pingdom.HttpCheck
		err  error
	}{
		{
			"empty name",
			args{name: "", url: "https://www.example.com"},
			nil,
			ErrEmptyName,
		},
		{
			"empty url",
			args{name: "example", url: ""},
			nil,
			ErrEmptyURL,
		},
		{
			"no host",
			args{name: "example", url: "http://"},
			nil,
			ErrNoHost,
		},
		{
			"no scheme",
			args{name: "example", url: "example.com"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Resolution: 5},
			nil,
		},
		{
			"encrypted",
			args{name: "example", url: "https://example.com"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Encryption: true, Resolution: 5},
			nil,
		},
		{
			"custom port",
			args{name: "example", url: "example.com:8080"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Port: 8080, Resolution: 5},
			nil,
		},
		{
			"malformed port",
			args{name: "example", url: "example.com:808a"},
			nil,
			ErrInvalidPort,
		},
		{
			"invalid port",
			args{name: "example", url: "example.com:65536"},
			nil,
			ErrInvalidPort,
		},
		{
			"negative port",
			args{name: "example", url: "example.com:-1"},
			nil,
			ErrInvalidPort,
		},
		{
			"zero port",
			args{name: "example", url: "example.com:0"},
			nil,
			ErrInvalidPort,
		},
		{
			"with user",
			args{name: "example", url: "user@example.com"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Username: "user", Resolution: 5},
			nil,
		},
		{
			"with pw",
			args{name: "example", url: ":pw@example.com"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Password: "pw", Resolution: 5},
			nil,
		},
		{
			"with user:pw",
			args{name: "example", url: "user:pw@example.com"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Username: "user", Password: "pw", Resolution: 5},
			nil,
		},
		{
			"simple path",
			args{name: "example", url: "example.com/a/path"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Url: "/a/path", Resolution: 5},
			nil,
		},
		{
			"path with query",
			args{name: "example", url: "example.com/a/path?q=uery&key=value"},
			&pingdom.HttpCheck{Name: "example", Hostname: "example.com", Url: "/a/path?q=uery&key=value", Resolution: 5},
			nil,
		},
		{
			"complex url",
			args{name: "example", url: "https://user:pw@www.example.com/a/path?q=uery&key=value"},
			&pingdom.HttpCheck{Name: "example", Hostname: "www.example.com", Url: "/a/path?q=uery&key=value", Username: "user", Password: "pw", Encryption: true, Resolution: 5},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := SimpleHttpCheck(tt.args.name, tt.args.url)

			if err != nil && tt.err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
			}
			assert.Equal(t, tt.c, c)
		})
	}
}
