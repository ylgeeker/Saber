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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"os-artificer/saber/pkg/sbnet"
)

func TestHandler_servesMetrics(t *testing.T) {
	c := NewCounter("test", "count_total", "Total count", []string{"label"})
	c.WithLabelValues("a").Inc()
	c.WithLabelValues("a").Add(2)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /metrics: status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "test_count_total") {
		t.Errorf("GET /metrics: body missing test_count_total:\n%s", body)
	}
}

func TestMetricsRoute_returnsRoute(t *testing.T) {
	r := MetricsRoute()
	if r.Method != http.MethodGet || r.Path != "/metrics" || r.Handler == nil {
		t.Errorf("MetricsRoute(): got Method=%q Path=%q Handler=%v", r.Method, r.Path, r.Handler == nil)
	}
}

// TestServer_withMetricsRoute_servesMetrics verifies that apm metrics are
// exposed through sbnet.Server when using MetricsRoute(), and that updated
// metric values are reflected in the response (full define → update → expose
// → scrape flow).
func TestServer_withMetricsRoute_servesMetrics(t *testing.T) {
	// 1) Define metric (unique name to avoid duplicate registration with other tests)
	counter := NewCounter("apm_sbnet_test", "http_requests_total", "Total HTTP requests", []string{"method", "path"})
	// 2) Update metric data
	counter.WithLabelValues("GET", "/api").Inc()
	counter.WithLabelValues("GET", "/api").Add(2)
	counter.WithLabelValues("POST", "/echo").Inc()

	// 3) Create sbnet Server with apm metrics route
	srv := sbnet.NewServer(sbnet.WithRoutes(MetricsRoute()))

	// 4) GET /metrics via Engine (no real listener)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)

	// 5) Assert response
	if rec.Code != http.StatusOK {
		t.Errorf("GET /metrics status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	metricName := "apm_sbnet_test_http_requests_total"
	if !strings.Contains(body, metricName) {
		t.Errorf("GET /metrics body missing %q:\n%s", metricName, body)
	}
	// Check that our label values appear
	if !strings.Contains(body, `method="GET"`) || !strings.Contains(body, `path="/api"`) {
		t.Errorf("GET /metrics body missing expected labels (method=GET, path=/api):\n%s", body)
	}
}

// ExampleMetricsRoute_withServer demonstrates the recommended way to combine
// apm with sbnet: define metrics, update them in business code, then expose
// /metrics by creating a Server with MetricsRoute() and serving (or, in tests,
// by calling Engine().ServeHTTP). Prometheus scrapes GET /metrics to get
// refreshed data.
func ExampleMetricsRoute_withServer() {
	// 1) Define metrics (e.g. at startup)
	reqTotal := NewCounter("app", "http_requests_total", "Total HTTP requests", []string{"method", "path"})

	// 2) Update metrics in business code
	reqTotal.WithLabelValues("GET", "/api").Inc()
	reqTotal.WithLabelValues("POST", "/echo").Inc()

	// 3) Expose via sbnet.Server: mount MetricsRoute() with WithRoutes
	srv := sbnet.NewServer(sbnet.WithRoutes(MetricsRoute()))

	// 4) In production: srv.Run(":8080"); in tests, serve one request:
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		panic("expected 200")
	}
	if !strings.Contains(rec.Body.String(), "app_http_requests_total") {
		panic("expected metric in body")
	}
	// Example runs without Output; go test verifies it does not panic.
}
