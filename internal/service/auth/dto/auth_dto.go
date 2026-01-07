package dto

// LoginRequest represents the request payload for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

// Role represents user role
type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// User represents user data
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	DeviceID string `json:"device_id"`
	Role     Role   `json:"role"`
}

// LoginData represents login data from API
type LoginData struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	User      User   `json:"user"`
}

// LoginResponse represents login result from API
type LoginResponse struct {
	Data    LoginData `json:"data"`
	Message string    `json:"message"`
}
