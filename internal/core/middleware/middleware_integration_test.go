package middleware

import (
	"testing"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"
	"police-trafic-api-frontend-aligned/internal/infrastructure/rbac"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMiddleware_Integration focuses on integration testing of middleware components
func TestMiddleware_Integration(t *testing.T) {
	logger := zap.NewNop()
	
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key-for-middleware-integration",
			AccessExpiration: 15 * time.Minute,
		},
	}
	
	jwtService := jwt.NewJWTService(cfg, logger)
	rbacService := rbac.NewRBACService(logger)
	authMiddleware := NewAuthMiddleware(jwtService, rbacService, logger)

	t.Run("middleware creation", func(t *testing.T) {
		assert.NotNil(t, authMiddleware, "AuthMiddleware should be created successfully")
		
		// Test that middleware methods exist and don't panic
		requireAuthMiddleware := authMiddleware.RequireAuth()
		assert.NotNil(t, requireAuthMiddleware, "RequireAuth middleware should be created")
		
		requireRoleMiddleware := authMiddleware.RequireRole("admin")
		assert.NotNil(t, requireRoleMiddleware, "RequireRole middleware should be created")
		
		requirePermissionMiddleware := authMiddleware.RequirePermission()
		assert.NotNil(t, requirePermissionMiddleware, "RequirePermission middleware should be created")
		
		optionalAuthMiddleware := authMiddleware.OptionalAuth()
		assert.NotNil(t, optionalAuthMiddleware, "OptionalAuth middleware should be created")
	})

	t.Run("token validation integration", func(t *testing.T) {
		// Generate test tokens for different roles
		adminToken, err := jwtService.GenerateToken("1", "67890", "admin")
		require.NoError(t, err, "Should generate admin token")
		assert.NotEmpty(t, adminToken, "Admin token should not be empty")

		supervisorToken, err := jwtService.GenerateToken("2", "11111", "supervisor")
		require.NoError(t, err, "Should generate supervisor token")
		assert.NotEmpty(t, supervisorToken, "Supervisor token should not be empty")

		agentToken, err := jwtService.GenerateToken("3", "12345", "agent")
		require.NoError(t, err, "Should generate agent token")
		assert.NotEmpty(t, agentToken, "Agent token should not be empty")

		// Validate tokens
		adminClaims, err := jwtService.ValidateToken(adminToken)
		require.NoError(t, err, "Admin token should be valid")
		assert.Equal(t, "admin", adminClaims.Role)

		supervisorClaims, err := jwtService.ValidateToken(supervisorToken)
		require.NoError(t, err, "Supervisor token should be valid")
		assert.Equal(t, "supervisor", supervisorClaims.Role)

		agentClaims, err := jwtService.ValidateToken(agentToken)
		require.NoError(t, err, "Agent token should be valid")
		assert.Equal(t, "agent", agentClaims.Role)
	})

	t.Run("RBAC permission integration", func(t *testing.T) {
		// Test permission checking for different roles
		testCases := []struct {
			role       string
			permission rbac.Permission
			expected   bool
		}{
			{"admin", rbac.PermDeleteUsers, true},
			{"supervisor", rbac.PermDeleteUsers, false},
			{"agent", rbac.PermDeleteUsers, false},
			
			{"admin", rbac.PermReadUsers, true},
			{"supervisor", rbac.PermReadUsers, true},
			{"agent", rbac.PermReadUsers, true},
			
			{"admin", rbac.PermApprovePV, true},
			{"supervisor", rbac.PermApprovePV, true},
			{"agent", rbac.PermApprovePV, false},
			
			{"admin", rbac.PermCreateControles, true},
			{"supervisor", rbac.PermCreateControles, true},
			{"agent", rbac.PermCreateControles, true},
			
			{"admin", rbac.PermDeleteControles, true},
			{"supervisor", rbac.PermDeleteControles, true},
			{"agent", rbac.PermDeleteControles, false},
		}

		for _, tc := range testCases {
			result := rbacService.HasPermission(tc.role, tc.permission)
			assert.Equal(t, tc.expected, result, 
				"Role %s should %s have permission %s", 
				tc.role, 
				map[bool]string{true: "", false: "NOT"}[tc.expected], 
				tc.permission,
			)
		}
	})

	t.Run("endpoint permission mapping", func(t *testing.T) {
		testCases := []struct {
			method           string
			path             string
			expectedError    bool
			expectedPermission rbac.Permission
		}{
			{"GET", "/users", false, "users:read"},
			{"POST", "/users", false, "users:create"},
			{"PUT", "/users/123", false, "users:update"},
			{"DELETE", "/users/123", false, "users:delete"},
			{"GET", "/controles", false, "controles:read"},
			{"POST", "/controles", false, "controles:create"},
			{"PATCH", "/infractions/456", false, "infractions:update"},
			{"POST", "/auth/login", true, ""}, // Auth endpoints should error
			{"GET", "/", true, ""}, // Invalid paths should error
		}

		for _, tc := range testCases {
			permission, err := rbac.GetPermissionFromEndpoint(tc.method, tc.path)
			
			if tc.expectedError {
				assert.Error(t, err, "Should error for %s %s", tc.method, tc.path)
			} else {
				assert.NoError(t, err, "Should not error for %s %s", tc.method, tc.path)
				assert.Equal(t, tc.expectedPermission, permission, 
					"Permission should match for %s %s", tc.method, tc.path)
			}
		}
	})

	t.Run("user context utilities", func(t *testing.T) {
		// Test UserContext creation and utility methods
		// This would need a mock Echo context, but we can test the logic
		
		// Test role validation
		validRoles := []string{"admin", "supervisor", "agent"}
		for _, role := range validRoles {
			isValid := rbacService.ValidateRole(role)
			assert.True(t, isValid, "Role %s should be valid", role)
		}

		invalidRoles := []string{"invalid", "", "ADMIN", "user"}
		for _, role := range invalidRoles {
			isValid := rbacService.ValidateRole(role)
			assert.False(t, isValid, "Role %s should be invalid", role)
		}
	})

	t.Run("token expiration handling", func(t *testing.T) {
		// Test with very short expiration
		expiredCfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:           "test-secret-expired",
				AccessExpiration: -1 * time.Second, // Already expired
			},
		}
		
		expiredJWTService := jwt.NewJWTService(expiredCfg, logger)
		
		// Generate already expired token
		expiredToken, err := expiredJWTService.GenerateToken("1", "12345", "agent")
		require.NoError(t, err, "Should generate token even if it will be expired")
		
		// Try to validate expired token
		_, err = expiredJWTService.ValidateToken(expiredToken)
		assert.Error(t, err, "Expired token should be invalid")
		assert.Contains(t, err.Error(), "expired", "Error should mention token expiration")
	})

	t.Run("middleware configuration edge cases", func(t *testing.T) {
		// Test middleware with different configurations
		
		// Test RequireRole with multiple roles
		multiRoleMiddleware := authMiddleware.RequireRole("admin", "supervisor")
		assert.NotNil(t, multiRoleMiddleware, "Multi-role middleware should be created")
		
		// Test RequireRole with single role
		singleRoleMiddleware := authMiddleware.RequireRole("admin")
		assert.NotNil(t, singleRoleMiddleware, "Single role middleware should be created")
		
		// Test RequireRole with empty roles (edge case)
		emptyRoleMiddleware := authMiddleware.RequireRole()
		assert.NotNil(t, emptyRoleMiddleware, "Empty role middleware should be created")
	})

	t.Run("permission-role consistency", func(t *testing.T) {
		// Test that permissions are consistently assigned to roles
		adminPermissions := rbacService.GetUserPermissions("admin")
		supervisorPermissions := rbacService.GetUserPermissions("supervisor")
		agentPermissions := rbacService.GetUserPermissions("agent")

		// Admin should have the most permissions
		assert.True(t, len(adminPermissions) > len(supervisorPermissions), 
			"Admin should have more permissions than supervisor")
		assert.True(t, len(supervisorPermissions) > len(agentPermissions), 
			"Supervisor should have more permissions than agent")

		// All should have basic read permissions
		basicPermissions := []rbac.Permission{
			rbac.PermReadUsers,
			rbac.PermReadControles,
			rbac.PermReadInfractions,
		}

		for _, perm := range basicPermissions {
			assert.True(t, rbacService.HasPermission("admin", perm), 
				"Admin should have basic permission %s", perm)
			assert.True(t, rbacService.HasPermission("supervisor", perm), 
				"Supervisor should have basic permission %s", perm)
			assert.True(t, rbacService.HasPermission("agent", perm), 
				"Agent should have basic permission %s", perm)
		}

		// Only admin should have admin-only permissions
		adminOnlyPermissions := []rbac.Permission{
			rbac.PermDeleteUsers,
			rbac.PermManageSystem,
		}

		for _, perm := range adminOnlyPermissions {
			assert.True(t, rbacService.HasPermission("admin", perm), 
				"Admin should have admin-only permission %s", perm)
			assert.False(t, rbacService.HasPermission("supervisor", perm), 
				"Supervisor should NOT have admin-only permission %s", perm)
			assert.False(t, rbacService.HasPermission("agent", perm), 
				"Agent should NOT have admin-only permission %s", perm)
		}
	})
}