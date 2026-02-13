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

package sbnet

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route defines a single HTTP route (method, path, handler). If AuthPrefix is
// non-empty, the route is registered under Server.AuthGroup(AuthPrefix).
type Route struct {
	Method     string
	Path       string
	Handler    gin.HandlerFunc
	AuthPrefix string
}

// Get returns a public GET route.
func Get(path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Path: path, Handler: h}
}

// Post returns a public POST route.
func Post(path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Path: path, Handler: h}
}

// Put returns a public PUT route.
func Put(path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodPut, Path: path, Handler: h}
}

// Patch returns a public PATCH route.
func Patch(path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodPatch, Path: path, Handler: h}
}

// Delete returns a public DELETE route.
func Delete(path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodDelete, Path: path, Handler: h}
}

// AuthGet returns a GET route under the given auth group prefix.
func AuthGet(prefix, path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Path: path, Handler: h, AuthPrefix: prefix}
}

// AuthPost returns a POST route under the given auth group prefix.
func AuthPost(prefix, path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Path: path, Handler: h, AuthPrefix: prefix}
}

// AuthPut returns a PUT route under the given auth group prefix.
func AuthPut(prefix, path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodPut, Path: path, Handler: h, AuthPrefix: prefix}
}

// AuthPatch returns a PATCH route under the given auth group prefix.
func AuthPatch(prefix, path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodPatch, Path: path, Handler: h, AuthPrefix: prefix}
}

// AuthDelete returns a DELETE route under the given auth group prefix.
func AuthDelete(prefix, path string, h gin.HandlerFunc) Route {
	return Route{Method: http.MethodDelete, Path: path, Handler: h, AuthPrefix: prefix}
}
