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

package apm

import (
	"os-artificer/saber/pkg/sbnet"

	"github.com/gin-gonic/gin"
)

// MetricsRoute returns a GET /metrics route that serves the default registry
// for Prometheus scraping. Use with sbnet.Server so the APM server is based on
// the existing Gin server: pass it to sbnet.WithRoutes when creating the
// server, e.g. sbnet.NewServer(sbnet.WithRoutes(apm.MetricsRoute()), ...).
// You can mount it on the same server as your app or run a dedicated server
// with only this route on another port (e.g. :9090).
func MetricsRoute() sbnet.Route {
	return sbnet.Get("/metrics", gin.WrapH(Handler()))
}
