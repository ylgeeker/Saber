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
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Usage examples for the sbnet gin server. Simple binding: use srv.GET(path, handler)
// for public routes and srv.AuthGroup(prefix).GET(path, handler) for auth; or
// NewServer(WithRoutes(Get("/ping", h), AuthGet("/api", "/me", meH))). See
// ExampleNewGinServer, ExampleNewServer, ExampleNewServer_withRegistrars and the
// TestServer_* / TestRun* tests below.

// ExampleNewGinServer shows the minimal usage: create a default Gin engine with
// /test and serve a request via httptest.
func ExampleNewGinServer() {
	router := NewGinServer()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		panic("expected 200")
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		panic(err)
	}
	if body["status"] != "ok" {
		panic("expected status ok")
	}
	// Example runs without Output; go test verifies it does not panic.
}

// ExampleNewServer shows the configurable server: NewServer() with default options
// and serving /test through Engine().
func ExampleNewServer() {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		panic("expected 200")
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		panic(err)
	}
	if body["status"] != "ok" {
		panic("expected status ok")
	}
}

// exampleRegistrar implements APIRegistrar for ExampleNewServer_withRegistrars.
type exampleRegistrar struct{}

func (e *exampleRegistrar) Register(s *Server) {
	s.Engine().GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pong": true})
	})
	s.AuthGroup("/api").GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": "example"})
	})
}

// ExampleNewServer_withRegistrars shows full usage: APIRegistrar for public and
// auth routes, WithAuthMiddleware and WithRegistrars, then requesting /test,
// /ping, and /api/me.
func ExampleNewServer_withRegistrars() {
	authHeader := "X-Authenticated"
	srv := NewServer(
		WithAuthMiddleware(func(c *gin.Context) {
			c.Header(authHeader, "true")
			c.Next()
		}),
		WithRegistrars(&exampleRegistrar{}),
	)

	// GET /test
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		panic("test: expected 200")
	}

	// GET /ping (public)
	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec = httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		panic("ping: expected 200")
	}
	var pingBody map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&pingBody); err != nil {
		panic(err)
	}
	if pingBody["pong"] != true {
		panic("expected pong true")
	}

	// GET /api/me (auth group)
	req = httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec = httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		panic("api/me: expected 200")
	}
	if rec.Header().Get(authHeader) != "true" {
		panic("expected auth header")
	}
}

func TestServer_GET_bindsPublicRoute(t *testing.T) {
	srv := NewServer()
	srv.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pong": true})
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /ping status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["pong"] != true {
		t.Errorf("body pong = %v, want true", body["pong"])
	}
}

func TestNewServer_WithRoutes_registersPublicAndAuthRoutes(t *testing.T) {
	authHeader := "X-Test-Auth"
	srv := NewServer(
		WithAuthMiddleware(func(c *gin.Context) {
			c.Header(authHeader, "ok")
			c.Next()
		}),
		WithRoutes(
			Get("/ping", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"pong": true})
			}),
			AuthGet("/api", "/me", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"user": "test"})
			}),
		),
	)
	// Public route
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /ping status = %d, want %d", rec.Code, http.StatusOK)
	}
	var pingBody map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&pingBody); err != nil {
		t.Fatalf("decode /ping body: %v", err)
	}
	if pingBody["pong"] != true {
		t.Errorf("GET /ping body pong = %v, want true", pingBody["pong"])
	}
	// Auth route
	req = httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec = httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/me status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(authHeader); got != "ok" {
		t.Errorf("header %s = %q, want ok", authHeader, got)
	}
	var meBody map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&meBody); err != nil {
		t.Fatalf("decode /api/me body: %v", err)
	}
	if meBody["user"] != "test" {
		t.Errorf("GET /api/me body user = %q, want test", meBody["user"])
	}
}

func TestNewGinServer_test(t *testing.T) {
	router := NewGinServer()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /test status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("body status = %q, want %q", body["status"], "ok")
	}
}

func TestNewServer_test(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /test status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("body status = %q, want %q", body["status"], "ok")
	}
}

func TestNewServer_WithRegistrars(t *testing.T) {
	var received *Server
	fake := &fakeRegistrar{
		register: func(s *Server) {
			received = s
		},
	}
	_ = NewServer(WithRegistrars(fake))
	if received == nil {
		t.Fatal("Register was not called or received nil Server")
	}
	if received.Engine() == nil {
		t.Error("Server.Engine() is nil")
	}
}

type fakeRegistrar struct {
	register func(s *Server)
}

func (f *fakeRegistrar) Register(s *Server) {
	if f.register != nil {
		f.register(s)
	}
}

func TestNewServer_AuthGroup_appliesMiddlewares(t *testing.T) {
	authHeader := "X-Test-Auth"
	registrar := &fakeRegistrar{
		register: func(s *Server) {
			s.AuthGroup("/api").GET("/ping", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"pong": true})
			})
		},
	}
	srv := NewServer(
		WithAuthMiddleware(func(c *gin.Context) {
			c.Header(authHeader, "ok")
			c.Next()
		}),
		WithRegistrars(registrar),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/ping status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(authHeader); got != "ok" {
		t.Errorf("header %s = %q, want %q", authHeader, got, "ok")
	}
}

func TestServer_Engine_returnsEngine(t *testing.T) {
	srv := NewServer()
	engine := srv.Engine()
	if engine == nil {
		t.Fatal("Engine() returned nil")
	}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /test via Engine() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServer_Run_invalidAddress(t *testing.T) {
	srv := NewServer()
	err := srv.Run("invalid-address-no-port")
	if err == nil {
		t.Error("Run(invalid-address-no-port) expected error, got nil")
	}
}

func TestRun_invalidAddress(t *testing.T) {
	router := NewGinServer()
	err := Run(router, "invalid-address-no-port")
	if err == nil {
		t.Error("Run(router, invalid-address-no-port) expected error, got nil")
	}
}

func TestRunListener_servesHealth(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := lis.Addr().String()
	router := NewGinServer()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = router.RunListener(lis)
	}()

	resp, err := http.Get("http://" + addr + "/test")
	if err != nil {
		lis.Close()
		<-done
		t.Fatalf("GET /test: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		lis.Close()
		<-done
		t.Errorf("GET /test status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	lis.Close()
	<-done
}

func TestNewServer_WithRequestLogging_disabled(t *testing.T) {
	srv := NewServer(WithRequestLogging(false))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	srv.Engine().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /test with request logging disabled status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// echoRegistrar registers GET /api/echo for integration test.
type echoRegistrar struct{}

func (e *echoRegistrar) Register(s *Server) {
	s.Engine().GET("/api/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"echo": "ok"})
	})
}

func TestServer_Run_servesRegisteredRoutes(t *testing.T) {
	srv := NewServer(
		WithRegistrars(&echoRegistrar{}),
		WithRequestLogging(false),
	)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := lis.Addr().String()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = srv.Engine().RunListener(lis)
	}()

	// GET /test
	resp, err := http.Get("http://" + addr + "/test")
	if err != nil {
		lis.Close()
		<-done
		t.Fatalf("GET /test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		lis.Close()
		<-done
		t.Errorf("GET /test status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var testBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&testBody); err != nil {
		resp.Body.Close()
		lis.Close()
		<-done
		t.Fatalf("decode /test body: %v", err)
	}
	resp.Body.Close()
	if testBody["status"] != "ok" {
		lis.Close()
		<-done
		t.Errorf("GET /test body status = %q, want ok", testBody["status"])
	}

	// GET /api/echo
	resp, err = http.Get("http://" + addr + "/api/echo")
	if err != nil {
		lis.Close()
		<-done
		t.Fatalf("GET /api/echo: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		lis.Close()
		<-done
		t.Errorf("GET /api/echo status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var echoBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&echoBody); err != nil {
		resp.Body.Close()
		lis.Close()
		<-done
		t.Fatalf("decode /api/echo body: %v", err)
	}
	resp.Body.Close()
	if echoBody["echo"] != "ok" {
		lis.Close()
		<-done
		t.Errorf("GET /api/echo body echo = %q, want ok", echoBody["echo"])
	}

	lis.Close()
	<-done
}
