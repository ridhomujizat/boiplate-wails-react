package integration

import (
	"net/http"
	"onx-screen-record/internal/service/integration/dto"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// healthHandler returns the main health check endpoint
func (s *Server) healthHandler(c *gin.Context) {
	response := dto.DetailedHealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		System: dto.SystemInfo{
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			NumCPU:     runtime.NumCPU(),
			GoRoutines: runtime.NumGoroutine(),
			GoVersion:  runtime.Version(),
		},
		Components: map[string]string{
			"app":      "healthy",
			"database": "healthy",
		},
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) livenessHandler(c *gin.Context) {
	response := dto.HealthResponse{
		Status:    "alive",
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
