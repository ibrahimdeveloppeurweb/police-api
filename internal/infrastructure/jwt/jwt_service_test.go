package jwt

import (
	"testing"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestJWTService_GenerateToken(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-jwt-token",
			AccessExpiration: 15 * time.Minute,
		},
	}

	service := NewJWTService(cfg, logger)

	tests := []struct {
		name      string
		userID    string
		matricule string
		role      string
		wantErr   bool
	}{
		{
			name:      "valid token generation",
			userID:    "123",
			matricule: "12345",
			role:      "admin",
			wantErr:   false,
		},
		{
			name:      "empty user id",
			userID:    "",
			matricule: "12345",
			role:      "admin",
			wantErr:   false, // Should still generate token
		},
		{
			name:      "empty matricule",
			userID:    "123",
			matricule: "",
			role:      "admin",
			wantErr:   false, // Should still generate token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.matricule, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token can be validated
				claims, err := service.ValidateToken(token)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.matricule, claims.Matricule)
				assert.Equal(t, tt.role, claims.Role)
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-jwt-token",
			AccessExpiration: 15 * time.Minute,
		},
	}

	service := NewJWTService(cfg, logger)

	// Generate a valid token
	userID := "123"
	matricule := "12345"
	role := "admin"
	token, err := service.GenerateToken(userID, matricule, role)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkFunc func(*testing.T, *Claims, error)
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: false,
			checkFunc: func(t *testing.T, claims *Claims, err error) {
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, matricule, claims.Matricule)
				assert.Equal(t, role, claims.Role)
				assert.Equal(t, "police-traffic-api", claims.Issuer)
				assert.Contains(t, claims.Audience, "police-traffic-frontend")
			},
		},
		{
			name:    "invalid token format",
			token:   "invalid-token",
			wantErr: true,
			checkFunc: func(t *testing.T, claims *Claims, err error) {
				assert.Error(t, err)
				assert.Nil(t, claims)
			},
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			checkFunc: func(t *testing.T, claims *Claims, err error) {
				assert.Error(t, err)
				assert.Nil(t, claims)
			},
		},
		{
			name:    "token with wrong secret",
			token:   generateTokenWithWrongSecret(t),
			wantErr: true,
			checkFunc: func(t *testing.T, claims *Claims, err error) {
				assert.Error(t, err)
				assert.Nil(t, claims)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			tt.checkFunc(t, claims, err)
		})
	}
}

func TestJWTService_RefreshToken(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-jwt-token",
			AccessExpiration: 15 * time.Minute,
		},
	}

	service := NewJWTService(cfg, logger)

	// Generate a valid token
	userID := "123"
	matricule := "12345"
	role := "admin"
	originalToken, err := service.GenerateToken(userID, matricule, role)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkFunc func(*testing.T, string, error)
	}{
		{
			name:    "valid token refresh",
			token:   originalToken,
			wantErr: false,
			checkFunc: func(t *testing.T, newToken string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, newToken)
				// Note: tokens might be identical if generated at same second
				// but that's acceptable for JWT refresh functionality
				
				// Verify new token has same user info
				claims, err := service.ValidateToken(newToken)
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, matricule, claims.Matricule)
				assert.Equal(t, role, claims.Role)
			},
		},
		{
			name:    "invalid token refresh",
			token:   "invalid-token",
			wantErr: true,
			checkFunc: func(t *testing.T, newToken string, err error) {
				assert.Error(t, err)
				assert.Empty(t, newToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := service.RefreshToken(tt.token)
			tt.checkFunc(t, newToken, err)
		})
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-jwt-token",
			AccessExpiration: -1 * time.Second, // Already expired
		},
	}

	service := NewJWTService(cfg, logger)

	// Generate an already expired token
	token, err := service.GenerateToken("123", "12345", "admin")
	require.NoError(t, err)

	// Try to validate expired token
	claims, err := service.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

// Helper function to generate a token with wrong secret
func generateTokenWithWrongSecret(t *testing.T) string {
	logger := zap.NewNop()
	wrongCfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "wrong-secret-key",
			AccessExpiration: 15 * time.Minute,
		},
	}

	wrongService := NewJWTService(wrongCfg, logger)
	token, err := wrongService.GenerateToken("123", "12345", "admin")
	require.NoError(t, err)
	return token
}