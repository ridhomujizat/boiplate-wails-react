package permission

// PermissionStatus represents the status of a system permission
type PermissionStatus struct {
	Granted bool   `json:"granted"`
	Message string `json:"message"`
}

// PermissionManager handles system permission checks and requests
type PermissionManager struct{}

// NewPermissionManager creates a new permission manager
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{}
}
