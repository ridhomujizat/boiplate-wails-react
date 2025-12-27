package integration

import "github.com/gin-gonic/gin"

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
