package integration

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
	port       int
}

// Config holds server configuration
type Config struct {
	Port int
	Mode string // "debug", "release", "test"
}

func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	gin.SetMode(config.Mode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(loggerMiddleware())

	server := &Server{
		engine: engine,
		port:   config.Port,
	}

	// Register routes
	server.registerRoutes()

	return server
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/health/live", s.livenessHandler)
}
