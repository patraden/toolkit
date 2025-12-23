// Copyright 2025 The Toolkit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vertica

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type ConnParams struct {
	Host     string
	Port     int
	password string
	User     string
	Database string
	Params   map[string]string
}

func NewConnString(host string, port int, user, password, database string) ConnParams {
	return ConnParams{
		Host:     host,
		Port:     port,
		User:     user,
		password: password,
		Database: database,
		Params:   make(map[string]string),
	}
}

func (c ConnParams) GetString() string {
	// base URI
	hostPort := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	uri := fmt.Sprintf(
		"vertica://%s:%s@%s/%s",
		url.QueryEscape(c.User),
		url.QueryEscape(c.password),
		hostPort,
		c.Database,
	)

	if len(c.Params) > 0 {
		// sort keys for deterministic output (good for tests/logging)
		keys := make([]string, 0, len(c.Params))
		for k := range c.Params {
			keys = append(keys, k)
		}

		var sb strings.Builder

		sort.Strings(keys)
		sb.WriteString("?")

		for i, key := range keys {
			if i > 0 {
				sb.WriteString("&")
			}

			sb.WriteString(url.QueryEscape(key))
			sb.WriteString("=")
			sb.WriteString(url.QueryEscape(c.Params[key]))
		}

		uri += sb.String()
	}

	return uri
}

func (c ConnParams) ConnString() string {
	hostPort := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	return fmt.Sprintf("vertica://%s:***@%s/%s", c.User, hostPort, c.Database)
}

func (c ConnParams) GetPassword() string {
	return c.password
}
