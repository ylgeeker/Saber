/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package sbdb

import (
	"fmt"
	"net/url"
)

type Option interface {
	apply(*options) error
}

var defaultOptions = options{
	host:      "127.0.0.1",
	port:      3306,
	charset:   "utf8mb4",
	parseTime: true,
}

type options struct {
	user      string
	password  string
	host      string
	port      int
	database  string
	charset   string
	parseTime bool
}

type funcOption struct {
	f func(*options) error
}

func (f *funcOption) apply(o *options) error {
	return f.f(o)
}

func OptionUser(val string) Option {
	return &funcOption{f: func(o *options) error {
		o.user = val
		return nil
	}}
}

func OptionPassword(val string) Option {
	return &funcOption{f: func(o *options) error {
		o.password = val
		return nil
	}}
}

func OptionHost(val string) Option {
	return &funcOption{f: func(o *options) error {
		o.host = val
		return nil
	}}
}

func OptionPort(val int) Option {
	return &funcOption{f: func(o *options) error {
		o.port = val
		return nil
	}}
}

func OptionDatabase(val string) Option {
	return &funcOption{f: func(o *options) error {
		o.database = val
		return nil
	}}
}

func OptionCharset(val string) Option {
	return &funcOption{f: func(o *options) error {
		o.charset = val
		return nil
	}}
}

func OptionParseTime(parseTime bool) Option {
	return &funcOption{f: func(o *options) error {
		o.parseTime = parseTime
		return nil
	}}
}

// DSN returns the MySQL DSN string for this options set.
func (o *options) DSN() string {
	user := o.user
	if user == "" {
		user = "root"
	}

	userInfo := url.UserPassword(user, o.password).String()
	query := "charset=" + url.QueryEscape(o.charset)

	if o.parseTime {
		query += "&parseTime=True"
	}

	return fmt.Sprintf("%s@tcp(%s:%d)/%s?%s",
		userInfo, o.host, o.port, o.database, query)
}
