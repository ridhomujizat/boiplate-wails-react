package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port: 8080,
		Mode: gin.ReleaseMode,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("[INFO] Starting HTTP server on port %d", s.port)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	log.Printf("[INFO] Stopping HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// GetEngine returns the gin engine for additional route registration
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// loggerMiddleware returns a gin middleware for logging
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Printf("[HTTP] %s %s %d %v", c.Request.Method, path, statusCode, latency)
	}
}
