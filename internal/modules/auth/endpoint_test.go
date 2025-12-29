package auth

import (
	"testing"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/crypto"
	"police-trafic-api-frontend-aligned/internal/infrastructure/database"
	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestEndpointIntegration tests the auth endpoints with focus on happy paths and core functionality
func TestEndpointIntegration(t *testing.T) {
	// Setup services
	logger := zap.NewNop()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-endpoint-integration",
			AccessExpiration: 15 * time.Minute,
		},
	}
	
	jwtService := jwt.NewJWTService(cfg, logger)
	cryptoService := crypto.NewPasswordService(logger)
	mockDB := &database.MockDB{}
	authService := NewService(logger, mockDB, jwtService, cryptoService)

	t.Run("full authentication flow", func(t *testing.T) {
		// Test 1: Register a new user
		registerReq := RegisterRequest{
			Matricule: "55555",
			Password:  "testPassword123",
			Nom:       "Integration",
			Prenom:    "Test",
			Email:     "integration@police.gouv.fr",
			Role:      "agent",
		}

		user, err := authService.Register(registerReq)
		require.NoError(t, err, "Registration should succeed")
		assert.NotNil(t, user)
		assert.Equal(t, registerReq.Matricule, user.Matricule)
		assert.Equal(t, registerReq.Nom, user.Nom)
		assert.Equal(t, registerReq.Email, user.Email)
		assert.Equal(t, registerReq.Role, user.Role)
		assert.True(t, user.Active)

		// Test 2: Login with existing mock user (since register doesn't persist in mock)
		loginResp, err := authService.Login("12345", "any_password")
		require.NoError(t, err, "Login should succeed")
		assert.NotNil(t, loginResp)
		assert.NotEmpty(t, loginResp.Token)
		assert.Equal(t, "12345", loginResp.User.Matricule)
		assert.Equal(t, "agent", loginResp.User.Role)

		// Test 3: Validate token and get current user
		currentUser, err := authService.GetCurrentUser(loginResp.Token)
		require.NoError(t, err, "GetCurrentUser should succeed")
		assert.NotNil(t, currentUser)
		assert.Equal(t, loginResp.User.Matricule, currentUser.Matricule)
		assert.Equal(t, loginResp.User.Role, currentUser.Role)

		// Test 4: Refresh token
		refreshResp, err := authService.RefreshToken(loginResp.Token)
		require.NoError(t, err, "Token refresh should succeed")
		assert.NotNil(t, refreshResp)
		assert.NotEmpty(t, refreshResp.Token)
		assert.Equal(t, loginResp.User.Matricule, refreshResp.User.Matricule)

		// Test 5: Validate refreshed token
		refreshedUser, err := authService.GetCurrentUser(refreshResp.Token)
		require.NoError(t, err, "GetCurrentUser with refreshed token should succeed")
		assert.Equal(t, currentUser.Matricule, refreshedUser.Matricule)

		// Test 6: Logout
		err = authService.Logout(refreshResp.Token)
		assert.NoError(t, err, "Logout should succeed")
	})

	t.Run("multiple role login flow", func(t *testing.T) {
		roles := []struct {
			name      string
			matricule string
			role      string
		}{
			{"agent", "12345", "agent"},
			{"admin", "67890", "admin"},
			{"supervisor", "11111", "supervisor"},
		}

		tokens := make(map[string]string)

		// Login with all roles
		for _, roleTest := range roles {
			t.Run("login_"+roleTest.name, func(t *testing.T) {
				loginResp, err := authService.Login(roleTest.matricule, "any_password")
				require.NoError(t, err, "Login should succeed for %s", roleTest.role)
				
				assert.NotEmpty(t, loginResp.Token)
				assert.Equal(t, roleTest.matricule, loginResp.User.Matricule)
				assert.Equal(t, roleTest.role, loginResp.User.Role)
				
				tokens[roleTest.role] = loginResp.Token
			})
		}

		// Verify each token works
		for role, token := range tokens {
			t.Run("verify_token_"+role, func(t *testing.T) {
				user, err := authService.GetCurrentUser(token)
				require.NoError(t, err, "Token should be valid for %s", role)
				assert.Equal(t, role, user.Role)
			})
		}

		// Refresh each token
		for role, token := range tokens {
			t.Run("refresh_token_"+role, func(t *testing.T) {
				refreshResp, err := authService.RefreshToken(token)
				require.NoError(t, err, "Token refresh should succeed for %s", role)
				
				assert.NotEmpty(t, refreshResp.Token)
				assert.Equal(t, role, refreshResp.User.Role)
			})
		}
	})

	t.Run("error scenarios", func(t *testing.T) {
		// Invalid credentials
		_, err := authService.Login("invalid_user", "any_password")
		assert.Error(t, err, "Login with invalid user should fail")

		// Empty password
		_, err = authService.Login("12345", "")
		assert.Error(t, err, "Login with empty password should fail")

		// Invalid token
		_, err = authService.GetCurrentUser("invalid_token")
		assert.Error(t, err, "GetCurrentUser with invalid token should fail")

		// Empty token
		_, err = authService.GetCurrentUser("")
		assert.Error(t, err, "GetCurrentUser with empty token should fail")

		// Invalid token refresh
		_, err = authService.RefreshToken("invalid_token")
		assert.Error(t, err, "RefreshToken with invalid token should fail")

		// Invalid token logout
		err = authService.Logout("invalid_token")
		assert.Error(t, err, "Logout with invalid token should fail")
	})

	t.Run("token validation edge cases", func(t *testing.T) {
		// Create a token with very short expiration
		shortCfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:           "test-secret-key",
				AccessExpiration: -1 * time.Second, // Already expired
			},
		}
		
		shortJWTService := jwt.NewJWTService(shortCfg, logger)
		shortAuthService := NewService(logger, mockDB, shortJWTService, cryptoService)

		// Generate expired token
		expiredToken, err := shortJWTService.GenerateToken("123", "12345", "agent")
		require.NoError(t, err)

		// Try to use expired token
		_, err = shortAuthService.GetCurrentUser(expiredToken)
		assert.Error(t, err, "Expired token should be rejected")
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("password security validation", func(t *testing.T) {
		// Test registration with various password types
		passwords := []string{
			"simplePassword",
			"Complex!Password123",
			"very-long-password-with-many-characters-and-symbols-!@#$%^&*()",
			"motDePasseAvecAccents123éàü",
		}

		for i, password := range passwords {
			t.Run("register_with_password_"+string(rune(i+'A')), func(t *testing.T) {
				req := RegisterRequest{
					Matricule: "test" + string(rune(i+'0')),
					Password:  password,
					Nom:       "Test",
					Prenom:    "User",
					Email:     "test" + string(rune(i+'0')) + "@police.gouv.fr",
					Role:      "agent",
				}

				user, err := authService.Register(req)
				assert.NoError(t, err, "Registration should succeed with password type %d", i)
				assert.NotNil(t, user)
			})
		}

		// Test that empty password fails
		req := RegisterRequest{
			Matricule: "emptypass",
			Password:  "",
			Nom:       "Test",
			Prenom:    "User",
			Email:     "emptypass@police.gouv.fr",
			Role:      "agent",
		}

		_, err := authService.Register(req)
		assert.Error(t, err, "Registration with empty password should fail")
	})
}