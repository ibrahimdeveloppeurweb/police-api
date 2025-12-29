package auth

import "time"

// DeviceInfo contains device information from the mobile app
type DeviceInfo struct {
	DeviceID   string `json:"device_id" validate:"required"`
	DeviceName string `json:"device_name,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // ios, android, web
	DeviceOS   string `json:"device_os,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
}

// LoginRequest represents a login request
// Accepts either "matricule" or "email" field from the frontend
type LoginRequest struct {
	Matricule string      `json:"matricule"`
	Email     string      `json:"email"` // Alias for matricule (frontend compatibility)
	Password  string      `json:"password" validate:"required"`
	Device    *DeviceInfo `json:"device,omitempty"` // Device info for mobile apps
}

// GetIdentifier returns the login identifier (matricule or email)
func (r *LoginRequest) GetIdentifier() string {
	if r.Matricule != "" {
		return r.Matricule
	}
	return r.Email
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Matricule string `json:"matricule" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
	Nom       string `json:"nom" validate:"required"`
	Prenom    string `json:"prenom" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Role      string `json:"role" validate:"required,oneof=admin agent supervisor"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	DeviceID     string `json:"device_id" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token,omitempty"` // Only for mobile apps with session
	SessionID    string       `json:"session_id,omitempty"`    // Session ID for mobile apps
	ExpiresAt    *time.Time   `json:"expires_at,omitempty"`    // Token expiration
	User         User         `json:"user"`
}

// User represents a user
type User struct {
	ID                string     `json:"id"`
	Matricule         string     `json:"matricule"`
	Nom               string     `json:"nom"`
	Prenom            string     `json:"prenom"`
	Email             string     `json:"email"`
	Role              string     `json:"role"`
	Grade             string     `json:"grade"`
	Commissariat      string     `json:"commissariat,omitempty"`
	CommissariatID    string     `json:"commissariat_id,omitempty"`
	Telephone         string     `json:"telephone"`
	Statut            string     `json:"statut"`
	Active            bool       `json:"active"`
	DerniereConnexion *time.Time `json:"derniere_connexion,omitempty"`
	PhotoURL          string     `json:"photo_url,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// SessionDTO represents a user session
type SessionDTO struct {
	ID             string    `json:"id"`
	DeviceID       string    `json:"device_id"`
	DeviceName     string    `json:"device_name,omitempty"`
	DeviceType     string    `json:"device_type,omitempty"`
	DeviceOS       string    `json:"device_os,omitempty"`
	AppVersion     string    `json:"app_version,omitempty"`
	LastActivityAt time.Time `json:"last_activity_at"`
	LastIPAddress  string    `json:"last_ip_address,omitempty"`
	IsCurrent      bool      `json:"is_current"`
	CreatedAt      time.Time `json:"created_at"`
}

// LogoutRequest represents logout request with optional session specification
type LogoutRequest struct {
	SessionID  string `json:"session_id,omitempty"`  // Specific session to revoke
	AllDevices bool   `json:"all_devices,omitempty"` // Logout from all devices
	DeviceID   string `json:"device_id,omitempty"`   // Device ID for mobile logout
}
