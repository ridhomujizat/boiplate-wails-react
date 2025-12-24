package app

import (
	"context"
	"fmt"
	appstorage "onx-screen-record/internal/pkg/app-storage"
)

// App struct
type App struct {
	ctx     context.Context
	storage appstorage.AppStorage
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// untuk initialisasi App Storage (logs, database, dsb)
	storage, err := appstorage.NewAppStorage("onx-screen-record")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize app storage: %v", err))
	}
	a.storage = *storage

	// 2. Contoh: Ambil path untuk SQLite
	dbPath := storage.GetPath("data.sqlite")
	fmt.Println("DB Path:", dbPath)
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
