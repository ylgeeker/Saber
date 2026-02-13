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

// Package apm provides a Prometheus-based APM layer. Business code only needs
// to define metrics (Counter, Gauge, Histogram, Summary) and update their
// values; that is enough to refresh monitoring data when Prometheus scrapes
// the endpoint. The default registry is exposed via Handler() and
// MetricsRoute() for Prometheus to scrape—no direct Prometheus or HTTP
// wiring is required in application code. The APM server is implemented
// using the existing sbnet.Server: create a server with
// sbnet.WithRoutes(apm.MetricsRoute()) and run it (same process or dedicated
// port).
package apm

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var defaultRegistry = prometheus.NewRegistry()

// DefaultRegistry returns the default Prometheus registry used by all apm
// metrics. Exposed for advanced use (e.g. custom collectors).
func DefaultRegistry() *prometheus.Registry {
	return defaultRegistry
}

// Register registers a custom Prometheus collector with the default registry.
// Use this to expose custom metrics that are not covered by NewCounter,
// NewGauge, NewHistogram, or NewSummary.
func Register(c prometheus.Collector) error {
	return defaultRegistry.Register(c)
}

// Handler returns an http.Handler that serves Prometheus metrics for the
// default registry. Mount it at GET /metrics (e.g. via sbnet with
// apm.MetricsRoute()) so Prometheus can scrape; updating metric values in
// business code will refresh the exposed data on each scrape.
func Handler() http.Handler {
	return promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{})
}
