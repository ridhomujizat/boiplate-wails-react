package app

import (
	"context"
	"log"
	"onx-screen-record/internal/service/integration"
	"time"
)

// startHTTPServer initializes and starts the HTTP server for health checks
func (a *App) startHTTPServer() {
	config := &integration.Config{
		Port: 8080,
		Mode: "release",
	}

	a.httpServer = integration.NewServer(config)

	if err := a.httpServer.Start(); err != nil {
		log.Printf("[ERROR] Failed to start HTTP server: %v", err)
	}
}

// stopHTTPServer gracefully stops the HTTP server
func (a *App) stopHTTPServer() {
	if a.httpServer == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Stop(ctx); err != nil {
		log.Printf("[ERROR] Failed to stop HTTP server: %v", err)
	}
}
