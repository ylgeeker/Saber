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
	"net"
	"net/http"
	"time"

	"os-artificer/saber/pkg/logger"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// APIRegistrar registers RESTful routes. Register may use s.Engine() for public
// routes and s.AuthGroup(path) for routes that require authentication.
type APIRegistrar interface {
	Register(s *Server)
}

// Server is a configurable HTTP server based on Gin.
type Server struct {
	engine           *gin.Engine
	cfg              *serverConfig
	requestLogLogger logger.Logger
}

// NewServer builds a Server from options. Request logging is enabled by default.
// Registrars are run after the engine is created; auth middlewares are used by
// AuthGroup only.
func NewServer(opts ...Option) *Server {
	cfg := &serverConfig{requestLogging: true}
	for _, opt := range opts {
		opt(cfg)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &Server{engine: engine, cfg: cfg}
	if cfg.logger != nil {
		s.requestLogLogger = cfg.logger
	}

	if cfg.requestLogging {
		engine.Use(s.requestLogMiddleware())
	}

	// Test endpoint
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for _, r := range cfg.registrars {
		r.Register(s)
	}

	for _, r := range cfg.routes {
		s.applyRoute(r)
	}

	return s
}

// NewGinServer returns a default Gin engine with a /test endpoint.
// For configurable server with RESTful registration, auth, and log injection,
// use NewServer and Server.Run instead.
func NewGinServer() *gin.Engine {
	router := gin.Default()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}

// Run starts the router on address. It uses the global logger (pkg/logger) for
// lifecycle messages. For injected logging use NewServer and Server.Run.
func Run(router *gin.Engine, address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	logger.Infof("Server listening at %v", lis.Addr())
	return router.RunListener(lis)
}

// Engine returns the underlying Gin engine for direct route registration.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// AuthGroup returns a router group at relativePath with all configured auth
// middlewares applied. Use it inside APIRegistrar.Register to mount
// authenticated REST routes.
func (s *Server) AuthGroup(relativePath string) *gin.RouterGroup {
	g := s.engine.Group(relativePath)
	for _, mw := range s.cfg.authMiddlewares {
		g.Use(mw)
	}
	return g
}

// GET registers a GET handler for path (public route). For auth routes use AuthGroup(prefix).GET(path, handlers...).
func (s *Server) GET(path string, handlers ...gin.HandlerFunc) {
	s.engine.GET(path, handlers...)
}

// POST registers a POST handler for path (public route).
func (s *Server) POST(path string, handlers ...gin.HandlerFunc) {
	s.engine.POST(path, handlers...)
}

// PUT registers a PUT handler for path (public route).
func (s *Server) PUT(path string, handlers ...gin.HandlerFunc) {
	s.engine.PUT(path, handlers...)
}

// PATCH registers a PATCH handler for path (public route).
func (s *Server) PATCH(path string, handlers ...gin.HandlerFunc) {
	s.engine.PATCH(path, handlers...)
}

// DELETE registers a DELETE handler for path (public route).
func (s *Server) DELETE(path string, handlers ...gin.HandlerFunc) {
	s.engine.DELETE(path, handlers...)
}

func (s *Server) applyRoute(r Route) {
	if r.Handler == nil {
		return
	}

	register := func(g *gin.RouterGroup) {
		switch r.Method {
		case http.MethodGet:
			g.GET(r.Path, r.Handler)

		case http.MethodPost:
			g.POST(r.Path, r.Handler)

		case http.MethodPut:
			g.PUT(r.Path, r.Handler)

		case http.MethodPatch:
			g.PATCH(r.Path, r.Handler)

		case http.MethodDelete:
			g.DELETE(r.Path, r.Handler)
		}
	}
	if r.AuthPrefix != "" {
		register(s.AuthGroup(r.AuthPrefix))
	} else {
		register(&s.engine.RouterGroup)
	}
}

// Run listens on address and serves HTTP. Lifecycle logs use the injected
// logger or the global logger.
func (s *Server) Run(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	logger.Infof("Server listening at %v", lis.Addr())

	return s.engine.RunListener(lis)
}

// requestLogMiddleware returns a Gin middleware that logs method, path, client
// IP, status code, and latency using the injected logger or the global logger.
func (s *Server) requestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Infof("%s %s %s %d %v", method, path, clientIP, status, latency)
	}
}
