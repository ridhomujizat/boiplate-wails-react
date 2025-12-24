package app

import (
	"context"
	"fmt"
	pathHelper "onx-screen-record/internal/pkg/path-file"
)

type App struct {
	ctx  context.Context
	path *pathHelper.PathHelper
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	a.path = pathHelper.NewPathHelper("onx-screen-record")

}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents login result
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// Login simulates user authentication
func (a *App) Login(email string, password string) LoginResponse {
	// Simulated login - replace with actual authentication
	if email != "" && password != "" {
		return LoginResponse{
			Success: true,
			Message: "Login successful",
			Token:   "mock-jwt-token",
		}
	}
	return LoginResponse{
		Success: false,
		Message: "Invalid credentials",
	}
}

// Requirement represents a status requirement
type Requirement struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

// GetRequirements returns the list of requirements
func (a *App) GetRequirements() []Requirement {
	return []Requirement{
		{ID: "1", Title: "User Authentication", Status: "completed", Progress: 100},
		{ID: "2", Title: "Dashboard Layout", Status: "completed", Progress: 100},
		{ID: "3", Title: "API Integration", Status: "pending", Progress: 45},
		{ID: "4", Title: "Data Validation", Status: "warning", Progress: 20},
		{ID: "5", Title: "Testing Coverage", Status: "pending", Progress: 60},
	}
}
