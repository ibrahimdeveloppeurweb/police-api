package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"
	"police-trafic-api-frontend-aligned/internal/infrastructure/rbac"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestAuthMiddleware() (*AuthMiddleware, jwt.Service) {
	logger := zap.NewNop()
	
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-middleware",
			AccessExpiration: 15 * time.Minute,
		},
	}
	
	jwtService := jwt.NewJWTService(cfg, logger)
	rbacService := rbac.NewRBACService(logger)
	
	authMiddleware := NewAuthMiddleware(jwtService, rbacService, logger)
	
	return authMiddleware, jwtService
}

func createTestEchoContext(path, method string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	authMiddleware, jwtService := setupTestAuthMiddleware()

	// Generate valid tokens for testing
	adminToken, err := jwtService.GenerateToken("1", "67890", "admin")
	require.NoError(t, err)

	agentToken, err := jwtService.GenerateToken("2", "12345", "agent")
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkContext   func(*testing.T, echo.Context)
	}{
		{
			name:           "valid admin token",
			token:          adminToken,
			expectedStatus: 200,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Equal(t, "1", c.Get("user_id"))
				assert.Equal(t, "67890", c.Get("matricule"))
				assert.Equal(t, "admin", c.Get("user_role"))
				assert.NotNil(t, c.Get("jwt_claims"))
				assert.NotNil(t, c.Get("rbac_service"))
			},
		},
		{
			name:           "valid agent token",
			token:          agentToken,
			expectedStatus: 200,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Equal(t, "2", c.Get("user_id"))
				assert.Equal(t, "12345", c.Get("matricule"))
				assert.Equal(t, "agent", c.Get("user_role"))
				assert.NotNil(t, c.Get("jwt_claims"))
			},
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: 401,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Nil(t, c.Get("user_id"))
				assert.Nil(t, c.Get("matricule"))
			},
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: 401,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Nil(t, c.Get("user_id"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestEchoContext("/protected", "GET")

			if tt.token != "" {
				c.Request().Header.Set("Authorization", "Bearer "+tt.token)
			}

			middleware := authMiddleware.RequireAuth()
			handler := middleware(func(c echo.Context) error {
				return c.JSON(200, map[string]string{"status": "ok"})
			})

			err := handler(c)
			
			if tt.expectedStatus == 200 {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				tt.checkContext(t, c)
			} else {
				// For error cases, the middleware returns the error directly
				assert.Error(t, err)
				tt.checkContext(t, c)
			}
		})
	}
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	authMiddleware, jwtService := setupTestAuthMiddleware()

	// Generate tokens for different roles
	adminToken, err := jwtService.GenerateToken("1", "67890", "admin")
	require.NoError(t, err)

	supervisorToken, err := jwtService.GenerateToken("2", "11111", "supervisor")
	require.NoError(t, err)

	agentToken, err := jwtService.GenerateToken("3", "12345", "agent")
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		requiredRoles  []string
		expectedStatus int
		expectAccess   bool
	}{
		{
			name:           "admin accessing admin-only endpoint",
			token:          adminToken,
			requiredRoles:  []string{"admin"},
			expectedStatus: 200,
			expectAccess:   true,
		},
		{
			name:           "admin accessing supervisor endpoint",
			token:          adminToken,
			requiredRoles:  []string{"supervisor"},
			expectedStatus: 403,
			expectAccess:   false,
		},
		{
			name:           "supervisor accessing supervisor endpoint",
			token:          supervisorToken,
			requiredRoles:  []string{"supervisor"},
			expectedStatus: 200,
			expectAccess:   true,
		},
		{
			name:           "agent accessing admin endpoint",
			token:          agentToken,
			requiredRoles:  []string{"admin"},
			expectedStatus: 403,
			expectAccess:   false,
		},
		{
			name:           "admin accessing multiple role endpoint",
			token:          adminToken,
			requiredRoles:  []string{"admin", "supervisor"},
			expectedStatus: 200,
			expectAccess:   true,
		},
		{
			name:           "supervisor accessing multiple role endpoint",
			token:          supervisorToken,
			requiredRoles:  []string{"admin", "supervisor"},
			expectedStatus: 200,
			expectAccess:   true,
		},
		{
			name:           "agent accessing multiple role endpoint",
			token:          agentToken,
			requiredRoles:  []string{"admin", "supervisor"},
			expectedStatus: 403,
			expectAccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestEchoContext("/protected", "GET")
			c.Request().Header.Set("Authorization", "Bearer "+tt.token)

			middleware := authMiddleware.RequireRole(tt.requiredRoles...)
			handler := middleware(func(c echo.Context) error {
				return c.JSON(200, map[string]string{"status": "ok"})
			})

			err := handler(c)
			
			if tt.expectAccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.NotNil(t, c.Get("user_id"))
				assert.NotNil(t, c.Get("user_role"))
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuthMiddleware_RequirePermission(t *testing.T) {
	authMiddleware, jwtService := setupTestAuthMiddleware()

	// Generate tokens for different roles
	adminToken, err := jwtService.GenerateToken("1", "67890", "admin")
	require.NoError(t, err)

	supervisorToken, err := jwtService.GenerateToken("2", "11111", "supervisor")
	require.NoError(t, err)

	agentToken, err := jwtService.GenerateToken("3", "12345", "agent")
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		method         string
		path           string
		expectedAccess bool
	}{
		{
			name:           "admin can delete users",
			token:          adminToken,
			method:         "DELETE",
			path:           "/users/123",
			expectedAccess: true,
		},
		{
			name:           "supervisor cannot delete users",
			token:          supervisorToken,
			method:         "DELETE",
			path:           "/users/123",
			expectedAccess: false,
		},
		{
			name:           "agent cannot delete users",
			token:          agentToken,
			method:         "DELETE",
			path:           "/users/123",
			expectedAccess: false,
		},
		{
			name:           "all roles can read users",
			token:          agentToken,
			method:         "GET",
			path:           "/users",
			expectedAccess: true,
		},
		{
			name:           "agent can create controles",
			token:          agentToken,
			method:         "POST",
			path:           "/controles",
			expectedAccess: true,
		},
		{
			name:           "agent cannot delete controles",
			token:          agentToken,
			method:         "DELETE",
			path:           "/controles/123",
			expectedAccess: false,
		},
		{
			name:           "supervisor can delete controles",
			token:          supervisorToken,
			method:         "DELETE",
			path:           "/controles/123",
			expectedAccess: true,
		},
		{
			name:           "auth endpoints allow through",
			token:          agentToken,
			method:         "POST",
			path:           "/auth/login",
			expectedAccess: true, // Auth endpoints allow through
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestEchoContext(tt.path, tt.method)
			c.Request().Header.Set("Authorization", "Bearer "+tt.token)

			middleware := authMiddleware.RequirePermission()
			handler := middleware(func(c echo.Context) error {
				return c.JSON(200, map[string]string{"status": "ok"})
			})

			err := handler(c)
			
			if tt.expectedAccess {
				assert.NoError(t, err)
				assert.Equal(t, 200, rec.Code)
				assert.NotNil(t, c.Get("user_id"))
			} else {
				assert.Error(t, err)
				// For forbidden access, we expect a specific error type
				// but the exact implementation may vary
			}
		})
	}
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	authMiddleware, jwtService := setupTestAuthMiddleware()

	validToken, err := jwtService.GenerateToken("1", "12345", "agent")
	require.NoError(t, err)

	tests := []struct {
		name         string
		token        string
		expectUser   bool
		checkContext func(*testing.T, echo.Context)
	}{
		{
			name:       "valid token sets user context",
			token:      validToken,
			expectUser: true,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Equal(t, "1", c.Get("user_id"))
				assert.Equal(t, "12345", c.Get("matricule"))
				assert.Equal(t, "agent", c.Get("user_role"))
				assert.NotNil(t, c.Get("jwt_claims"))
			},
		},
		{
			name:       "invalid token does not set user context",
			token:      "invalid-token",
			expectUser: false,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Nil(t, c.Get("user_id"))
				assert.Nil(t, c.Get("matricule"))
				assert.Nil(t, c.Get("user_role"))
			},
		},
		{
			name:       "missing token does not set user context",
			token:      "",
			expectUser: false,
			checkContext: func(t *testing.T, c echo.Context) {
				assert.Nil(t, c.Get("user_id"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestEchoContext("/optional", "GET")

			if tt.token != "" {
				c.Request().Header.Set("Authorization", "Bearer "+tt.token)
			}

			middleware := authMiddleware.OptionalAuth()
			handler := middleware(func(c echo.Context) error {
				return c.JSON(200, map[string]string{"status": "ok"})
			})

			err := handler(c)
			assert.NoError(t, err) // OptionalAuth never fails
			assert.Equal(t, 200, rec.Code)
			
			tt.checkContext(t, c)
		})
	}
}

func TestAuthMiddleware_TokenExtraction(t *testing.T) {
	authMiddleware, jwtService := setupTestAuthMiddleware()

	validToken, err := jwtService.GenerateToken("1", "12345", "agent")
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "bearer token format",
			authHeader:     "Bearer " + validToken,
			expectedStatus: 200,
		},
		{
			name:           "missing bearer prefix",
			authHeader:     validToken,
			expectedStatus: 401,
		},
		{
			name:           "wrong prefix",
			authHeader:     "Basic " + validToken,
			expectedStatus: 401,
		},
		{
			name:           "multiple spaces",
			authHeader:     "Bearer  " + validToken,
			expectedStatus: 401,
		},
		{
			name:           "lowercase bearer",
			authHeader:     "bearer " + validToken,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := createTestEchoContext("/protected", "GET")
			c.Request().Header.Set("Authorization", tt.authHeader)

			middleware := authMiddleware.RequireAuth()
			handler := middleware(func(c echo.Context) error {
				return c.JSON(200, map[string]string{"status": "ok"})
			})

			err := handler(c)
			
			if tt.expectedStatus == 200 {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	logger := zap.NewNop()
	
	// Create JWT service with very short expiration
	expiredCfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-middleware",
			AccessExpiration: -1 * time.Second, // Already expired
		},
	}
	
	expiredJWTService := jwt.NewJWTService(expiredCfg, logger)
	rbacService := rbac.NewRBACService(logger)
	authMiddleware := NewAuthMiddleware(expiredJWTService, rbacService, logger)

	// Generate already expired token
	expiredToken, err := expiredJWTService.GenerateToken("1", "12345", "agent")
	require.NoError(t, err)

	c, _ := createTestEchoContext("/protected", "GET")
	c.Request().Header.Set("Authorization", "Bearer "+expiredToken)

	middleware := authMiddleware.RequireAuth()
	handler := middleware(func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	err = handler(c)
	assert.Error(t, err, "Expired token should be rejected")
	assert.Nil(t, c.Get("user_id"), "No user context should be set for expired token")
}