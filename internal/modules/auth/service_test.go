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

// setupTestServices creates test services for auth testing
func setupTestServices() (jwt.Service, crypto.Service, *zap.Logger) {
	logger := zap.NewNop()
	
	// JWT service with test config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-jwt-token",
			AccessExpiration: 15 * time.Minute,
		},
	}
	jwtService := jwt.NewJWTService(cfg, logger)
	
	// Crypto service
	cryptoService := crypto.NewPasswordService(logger)
	
	return jwtService, cryptoService, logger
}

func TestNewService(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	
	// Test with mock DB
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)
	
	assert.NotNil(t, service)
}

func TestAuthService_Login_MockData(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)

	tests := []struct {
		name      string
		matricule string
		password  string
		wantErr   bool
		checkFunc func(*testing.T, *LoginResponse, error)
	}{
		{
			name:      "valid login with mock data",
			matricule: "12345",
			password:  "any_password", // Mock data accepts any non-empty password
			wantErr:   false,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, "12345", resp.User.Matricule)
				assert.Equal(t, "Dupont", resp.User.Nom)
				assert.Equal(t, "agent", resp.User.Role)
				
				// Verify token is valid
				claims, err := jwtService.ValidateToken(resp.Token)
				assert.NoError(t, err)
				assert.Equal(t, resp.User.ID, claims.UserID)
				assert.Equal(t, resp.User.Matricule, claims.Matricule)
				assert.Equal(t, resp.User.Role, claims.Role)
			},
		},
		{
			name:      "valid supervisor login",
			matricule: "67890",
			password:  "any_password",
			wantErr:   false,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "67890", resp.User.Matricule)
				assert.Equal(t, "Martin", resp.User.Nom)
				assert.Equal(t, "admin", resp.User.Role)
			},
		},
		{
			name:      "invalid matricule",
			matricule: "99999",
			password:  "any_password",
			wantErr:   true,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "invalid credentials")
			},
		},
		{
			name:      "empty password",
			matricule: "12345",
			password:  "",
			wantErr:   true,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "invalid password")
			},
		},
		{
			name:      "empty matricule",
			matricule: "",
			password:  "any_password",
			wantErr:   true,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Login(tt.matricule, tt.password)
			tt.checkFunc(t, resp, err)
		})
	}
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)

	// First login to get a valid token
	loginResp, err := service.Login("12345", "any_password")
	require.NoError(t, err)
	validToken := loginResp.Token

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkFunc func(*testing.T, *User, error)
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
			checkFunc: func(t *testing.T, user *User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "12345", user.Matricule)
				assert.Equal(t, "Dupont", user.Nom)
				assert.Equal(t, "agent", user.Role)
			},
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: true,
			checkFunc: func(t *testing.T, user *User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)
			},
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			checkFunc: func(t *testing.T, user *User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetCurrentUser(tt.token)
			tt.checkFunc(t, user, err)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)

	// First login to get a valid token
	loginResp, err := service.Login("12345", "any_password")
	require.NoError(t, err)
	validToken := loginResp.Token

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkFunc func(*testing.T, *LoginResponse, error)
	}{
		{
			name:    "valid token refresh",
			token:   validToken,
			wantErr: false,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)
				// Note: tokens might be identical if generated at same second, but that's OK
				assert.Equal(t, "12345", resp.User.Matricule)
				assert.Equal(t, "Dupont", resp.User.Nom)
			},
		},
		{
			name:    "invalid token refresh",
			token:   "invalid-token",
			wantErr: true,
			checkFunc: func(t *testing.T, resp *LoginResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.RefreshToken(tt.token)
			tt.checkFunc(t, resp, err)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)

	// First login to get a valid token
	loginResp, err := service.Login("12345", "any_password")
	require.NoError(t, err)
	validToken := loginResp.Token

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token logout",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid token logout",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "empty token logout",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Logout(tt.token)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_Register_MockData(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)

	tests := []struct {
		name      string
		req       RegisterRequest
		wantErr   bool
		checkFunc func(*testing.T, *User, error)
	}{
		{
			name: "valid registration",
			req: RegisterRequest{
				Matricule: "99999",
				Password:  "securePassword123",
				Nom:       "Test",
				Prenom:    "User",
				Email:     "test@police.gouv.fr",
				Role:      "agent",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "99999", user.Matricule)
				assert.Equal(t, "Test", user.Nom)
				assert.Equal(t, "User", user.Prenom)
				assert.Equal(t, "test@police.gouv.fr", user.Email)
				assert.Equal(t, "agent", user.Role)
				assert.True(t, user.Active)
				assert.NotEmpty(t, user.ID)
			},
		},
		{
			name: "registration with empty password",
			req: RegisterRequest{
				Matricule: "88888",
				Password:  "",
				Nom:       "Test",
				Prenom:    "User",
				Email:     "test2@police.gouv.fr",
				Role:      "agent",
			},
			wantErr: true,
			checkFunc: func(t *testing.T, user *User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Contains(t, err.Error(), "password")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Register(tt.req)
			tt.checkFunc(t, user, err)
		})
	}
}

func TestAuthService_GetMockUsers(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	authService := NewService(logger, mockDB, jwtService, cryptoService)
	
	// Cast to access private method for testing
	service := authService.(*service)
	mockUsers := service.getMockUsers()
	
	assert.Len(t, mockUsers, 3)
	
	// Verify first user
	user1 := mockUsers[0]
	assert.Equal(t, "12345", user1.Matricule)
	assert.Equal(t, "Dupont", user1.Nom)
	assert.Equal(t, "Jean", user1.Prenom)
	assert.Equal(t, "agent", user1.Role)
	assert.True(t, user1.Active)
	
	// Verify all users have required fields
	for i, user := range mockUsers {
		assert.NotEmpty(t, user.ID, "User %d should have ID", i)
		assert.NotEmpty(t, user.Matricule, "User %d should have Matricule", i)
		assert.NotEmpty(t, user.Nom, "User %d should have Nom", i)
		assert.NotEmpty(t, user.Prenom, "User %d should have Prenom", i)
		assert.NotEmpty(t, user.Email, "User %d should have Email", i)
		assert.NotEmpty(t, user.Role, "User %d should have Role", i)
		assert.True(t, user.Active, "User %d should be active", i)
	}
}

func TestAuthService_TokenValidation_Integration(t *testing.T) {
	jwtService, cryptoService, logger := setupTestServices()
	mockDB := &database.MockDB{}
	service := NewService(logger, mockDB, jwtService, cryptoService)

	// Test complete flow: login -> get user -> refresh -> logout
	matricule := "12345"
	password := "any_password"
	
	// Step 1: Login
	loginResp, err := service.Login(matricule, password)
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.Token)
	
	// Step 2: Get current user
	user, err := service.GetCurrentUser(loginResp.Token)
	require.NoError(t, err)
	assert.Equal(t, matricule, user.Matricule)
	
	// Step 3: Refresh token
	refreshResp, err := service.RefreshToken(loginResp.Token)
	require.NoError(t, err)
	// Note: tokens might be identical if generated at same second, but that's OK
	
	// Step 4: Verify new token works
	user2, err := service.GetCurrentUser(refreshResp.Token)
	require.NoError(t, err)
	assert.Equal(t, user.Matricule, user2.Matricule)
	
	// Step 5: Logout
	err = service.Logout(refreshResp.Token)
	assert.NoError(t, err)
}