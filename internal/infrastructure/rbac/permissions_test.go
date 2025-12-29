package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewRBACService(t *testing.T) {
	logger := zap.NewNop()
	service := NewRBACService(logger)

	assert.NotNil(t, service)
}

func TestRBACService_HasPermission(t *testing.T) {
	logger := zap.NewNop()
	service := NewRBACService(logger)

	tests := []struct {
		name       string
		role       string
		permission Permission
		expected   bool
	}{
		// Admin permissions (should have everything)
		{
			name:       "admin can read users",
			role:       string(RoleAdmin),
			permission: PermReadUsers,
			expected:   true,
		},
		{
			name:       "admin can delete users",
			role:       string(RoleAdmin),
			permission: PermDeleteUsers,
			expected:   true,
		},
		{
			name:       "admin can manage system",
			role:       string(RoleAdmin),
			permission: PermManageSystem,
			expected:   true,
		},
		
		// Supervisor permissions
		{
			name:       "supervisor can read users",
			role:       string(RoleSupervisor),
			permission: PermReadUsers,
			expected:   true,
		},
		{
			name:       "supervisor cannot delete users",
			role:       string(RoleSupervisor),
			permission: PermDeleteUsers,
			expected:   false,
		},
		{
			name:       "supervisor can approve PV",
			role:       string(RoleSupervisor),
			permission: PermApprovePV,
			expected:   true,
		},
		{
			name:       "supervisor cannot manage system",
			role:       string(RoleSupervisor),
			permission: PermManageSystem,
			expected:   false,
		},
		
		// Agent permissions
		{
			name:       "agent can read users",
			role:       string(RoleAgent),
			permission: PermReadUsers,
			expected:   true,
		},
		{
			name:       "agent cannot delete users",
			role:       string(RoleAgent),
			permission: PermDeleteUsers,
			expected:   false,
		},
		{
			name:       "agent cannot approve PV",
			role:       string(RoleAgent),
			permission: PermApprovePV,
			expected:   false,
		},
		{
			name:       "agent can create controles",
			role:       string(RoleAgent),
			permission: PermCreateControles,
			expected:   true,
		},
		
		// Invalid role
		{
			name:       "invalid role has no permissions",
			role:       "invalid_role",
			permission: PermReadUsers,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.HasPermission(tt.role, tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRBACService_GetUserPermissions(t *testing.T) {
	logger := zap.NewNop()
	service := NewRBACService(logger)

	tests := []struct {
		name     string
		role     string
		minPerms int // Minimum expected permissions
		checkContains []Permission // Permissions that should be included
		checkNotContains []Permission // Permissions that should NOT be included
	}{
		{
			name:     "admin has all permissions",
			role:     string(RoleAdmin),
			minPerms: 20, // Should have many permissions
			checkContains: []Permission{
				PermReadUsers, PermDeleteUsers, PermManageSystem, PermApprovePV,
			},
			checkNotContains: []Permission{}, // Admin should have everything
		},
		{
			name:     "supervisor has moderate permissions",
			role:     string(RoleSupervisor),
			minPerms: 10, // Should have several permissions
			checkContains: []Permission{
				PermReadUsers, PermApprovePV, PermViewReports,
			},
			checkNotContains: []Permission{
				PermDeleteUsers, PermManageSystem,
			},
		},
		{
			name:     "agent has limited permissions",
			role:     string(RoleAgent),
			minPerms: 5, // Should have basic permissions
			checkContains: []Permission{
				PermReadUsers, PermCreateControles, PermReadPV,
			},
			checkNotContains: []Permission{
				PermDeleteUsers, PermApprovePV, PermManageSystem,
			},
		},
		{
			name:     "invalid role has no permissions",
			role:     "invalid_role",
			minPerms: 0,
			checkContains: []Permission{},
			checkNotContains: []Permission{
				PermReadUsers, PermDeleteUsers,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := service.GetUserPermissions(tt.role)
			
			assert.GreaterOrEqual(t, len(permissions), tt.minPerms)
			
			for _, perm := range tt.checkContains {
				assert.Contains(t, permissions, perm, "Should contain permission %s", perm)
			}
			
			for _, perm := range tt.checkNotContains {
				assert.NotContains(t, permissions, perm, "Should NOT contain permission %s", perm)
			}
		})
	}
}

func TestRBACService_CanAccessResource(t *testing.T) {
	logger := zap.NewNop()
	service := NewRBACService(logger)

	tests := []struct {
		name     string
		role     string
		resource string
		action   string
		expected bool
	}{
		{
			name:     "admin can read users",
			role:     string(RoleAdmin),
			resource: "users",
			action:   "read",
			expected: true,
		},
		{
			name:     "admin can delete users",
			role:     string(RoleAdmin),
			resource: "users",
			action:   "delete",
			expected: true,
		},
		{
			name:     "supervisor can read users",
			role:     string(RoleSupervisor),
			resource: "users",
			action:   "read",
			expected: true,
		},
		{
			name:     "supervisor cannot delete users",
			role:     string(RoleSupervisor),
			resource: "users",
			action:   "delete",
			expected: false,
		},
		{
			name:     "agent can create controles",
			role:     string(RoleAgent),
			resource: "controles",
			action:   "create",
			expected: true,
		},
		{
			name:     "agent cannot delete controles",
			role:     string(RoleAgent),
			resource: "controles",
			action:   "delete",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CanAccessResource(tt.role, tt.resource, tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRBACService_ValidateRole(t *testing.T) {
	logger := zap.NewNop()
	service := NewRBACService(logger)

	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "admin role is valid",
			role:     string(RoleAdmin),
			expected: true,
		},
		{
			name:     "supervisor role is valid",
			role:     string(RoleSupervisor),
			expected: true,
		},
		{
			name:     "agent role is valid",
			role:     string(RoleAgent),
			expected: true,
		},
		{
			name:     "invalid role",
			role:     "invalid_role",
			expected: false,
		},
		{
			name:     "empty role",
			role:     "",
			expected: false,
		},
		{
			name:     "case sensitive role",
			role:     "ADMIN",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPermissionFromEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedPerm   Permission
		expectError    bool
	}{
		{
			name:         "GET users endpoint",
			method:       "GET",
			path:         "/users",
			expectedPerm: "users:read",
			expectError:  false,
		},
		{
			name:         "POST users endpoint",
			method:       "POST",
			path:         "/users",
			expectedPerm: "users:create",
			expectError:  false,
		},
		{
			name:         "PUT users endpoint",
			method:       "PUT",
			path:         "/users/123",
			expectedPerm: "users:update",
			expectError:  false,
		},
		{
			name:         "DELETE users endpoint",
			method:       "DELETE",
			path:         "/users/123",
			expectedPerm: "users:delete",
			expectError:  false,
		},
		{
			name:         "GET with /api/v1 prefix",
			method:       "GET",
			path:         "/api/v1/controles",
			expectedPerm: "controles:read",
			expectError:  false,
		},
		{
			name:         "PATCH endpoint",
			method:       "PATCH",
			path:         "/infractions/456",
			expectedPerm: "infractions:update",
			expectError:  false,
		},
		{
			name:        "auth endpoint (should error)",
			method:      "POST",
			path:        "/auth/login",
			expectError: true,
		},
		{
			name:        "unsupported HTTP method",
			method:      "HEAD",
			path:        "/users",
			expectError: true,
		},
		{
			name:        "empty path",
			method:      "GET",
			path:        "",
			expectError: true,
		},
		{
			name:        "root path",
			method:      "GET",
			path:        "/",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission, err := GetPermissionFromEndpoint(tt.method, tt.path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, permission)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPerm, permission)
			}
		})
	}
}

func TestRolePermissions_Consistency(t *testing.T) {
	// Test that role permissions are properly defined
	
	// Admin should have the most permissions
	adminPerms := RolePermissions[RoleAdmin]
	supervisorPerms := RolePermissions[RoleSupervisor]
	agentPerms := RolePermissions[RoleAgent]

	assert.True(t, len(adminPerms) > len(supervisorPerms), "Admin should have more permissions than supervisor")
	assert.True(t, len(supervisorPerms) > len(agentPerms), "Supervisor should have more permissions than agent")

	// All roles should have basic read permissions
	basicReadPermissions := []Permission{
		PermReadUsers,
		PermReadControles,
		PermReadInfractions,
		PermReadPV,
	}

	for _, perm := range basicReadPermissions {
		assert.Contains(t, adminPerms, perm, "Admin should have basic read permission: %s", perm)
		assert.Contains(t, supervisorPerms, perm, "Supervisor should have basic read permission: %s", perm)
		assert.Contains(t, agentPerms, perm, "Agent should have basic read permission: %s", perm)
	}

	// Only admin should have admin permissions
	adminOnlyPermissions := []Permission{
		PermDeleteUsers,
		PermManageSystem,
		PermManageConfig,
	}

	for _, perm := range adminOnlyPermissions {
		assert.Contains(t, adminPerms, perm, "Admin should have admin permission: %s", perm)
		assert.NotContains(t, supervisorPerms, perm, "Supervisor should NOT have admin permission: %s", perm)
		assert.NotContains(t, agentPerms, perm, "Agent should NOT have admin permission: %s", perm)
	}
}