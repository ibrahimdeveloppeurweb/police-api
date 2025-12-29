package rbac

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// Permission represents a specific permission
type Permission string

// Resource types
const (
	// User management
	PermReadUsers    Permission = "users:read"
	PermCreateUsers  Permission = "users:create"
	PermUpdateUsers  Permission = "users:update"
	PermDeleteUsers  Permission = "users:delete"

	// Traffic controls
	PermReadControles    Permission = "controles:read"
	PermCreateControles  Permission = "controles:create"
	PermUpdateControles  Permission = "controles:update"
	PermDeleteControles  Permission = "controles:delete"

	// Infractions
	PermReadInfractions    Permission = "infractions:read"
	PermCreateInfractions  Permission = "infractions:create"
	PermUpdateInfractions  Permission = "infractions:update"
	PermDeleteInfractions  Permission = "infractions:delete"

	// Verbaux (PV)
	PermReadPV    Permission = "pv:read"
	PermCreatePV  Permission = "pv:create"
	PermUpdatePV  Permission = "pv:update"
	PermDeletePV  Permission = "pv:delete"
	PermApprovePV Permission = "pv:approve"

	// Alerts
	PermReadAlertes    Permission = "alertes:read"
	PermCreateAlertes  Permission = "alertes:create"
	PermUpdateAlertes  Permission = "alertes:update"
	PermDeleteAlertes  Permission = "alertes:delete"

	// Commissariats
	PermReadCommissariats    Permission = "commissariats:read"
	PermCreateCommissariats  Permission = "commissariats:create"
	PermUpdateCommissariats  Permission = "commissariats:update"
	PermDeleteCommissariats  Permission = "commissariats:delete"

	// Admin permissions
	PermReadAdmin     Permission = "admin:read"
	PermManageSystem  Permission = "admin:system"
	PermViewReports   Permission = "admin:reports"
	PermManageConfig  Permission = "admin:config"
)

// Role represents a user role
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleSupervisor Role = "supervisor"
	RoleAgent      Role = "agent"
)

// RolePermissions defines which permissions each role has
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// Full access to everything
		PermReadUsers, PermCreateUsers, PermUpdateUsers, PermDeleteUsers,
		PermReadControles, PermCreateControles, PermUpdateControles, PermDeleteControles,
		PermReadInfractions, PermCreateInfractions, PermUpdateInfractions, PermDeleteInfractions,
		PermReadPV, PermCreatePV, PermUpdatePV, PermDeletePV, PermApprovePV,
		PermReadAlertes, PermCreateAlertes, PermUpdateAlertes, PermDeleteAlertes,
		PermReadCommissariats, PermCreateCommissariats, PermUpdateCommissariats, PermDeleteCommissariats,
		PermReadAdmin, PermManageSystem, PermViewReports, PermManageConfig,
	},
	RoleSupervisor: {
		// Can read users but not delete, approve PV
		PermReadUsers, PermUpdateUsers,
		PermReadControles, PermCreateControles, PermUpdateControles, PermDeleteControles,
		PermReadInfractions, PermCreateInfractions, PermUpdateInfractions, PermDeleteInfractions,
		PermReadPV, PermCreatePV, PermUpdatePV, PermApprovePV,
		PermReadAlertes, PermCreateAlertes, PermUpdateAlertes, PermDeleteAlertes,
		PermReadCommissariats, PermUpdateCommissariats,
		PermViewReports,
	},
	RoleAgent: {
		// Basic operations, cannot delete or approve
		PermReadUsers,
		PermReadControles, PermCreateControles, PermUpdateControles,
		PermReadInfractions, PermCreateInfractions, PermUpdateInfractions,
		PermReadPV, PermCreatePV, PermUpdatePV,
		PermReadAlertes, PermCreateAlertes, PermUpdateAlertes,
		PermReadCommissariats,
	},
}

// Service provides role-based access control functionality
type Service interface {
	HasPermission(role string, permission Permission) bool
	GetUserPermissions(role string) []Permission
	CanAccessResource(role string, resource string, action string) bool
	ValidateRole(role string) bool
}

type service struct {
	logger *zap.Logger
}

// NewRBACService creates a new RBAC service
func NewRBACService(logger *zap.Logger) Service {
	return &service{
		logger: logger,
	}
}

// HasPermission checks if a role has a specific permission
func (s *service) HasPermission(role string, permission Permission) bool {
	roleEnum := Role(role)
	permissions, exists := RolePermissions[roleEnum]
	if !exists {
		s.logger.Warn("Unknown role", zap.String("role", role))
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// GetUserPermissions returns all permissions for a user role
func (s *service) GetUserPermissions(role string) []Permission {
	roleEnum := Role(role)
	permissions, exists := RolePermissions[roleEnum]
	if !exists {
		s.logger.Warn("Unknown role", zap.String("role", role))
		return []Permission{}
	}

	return permissions
}

// CanAccessResource checks if a role can perform an action on a resource
// resource format: "users", "controles", etc.
// action format: "read", "create", "update", "delete", "approve"
func (s *service) CanAccessResource(role string, resource string, action string) bool {
	permission := Permission(fmt.Sprintf("%s:%s", resource, action))
	return s.HasPermission(role, permission)
}

// ValidateRole checks if a role is valid
func (s *service) ValidateRole(role string) bool {
	roleEnum := Role(role)
	_, exists := RolePermissions[roleEnum]
	return exists
}

// GetPermissionFromEndpoint extracts permission requirement from HTTP endpoint
// Format: GET /api/v1/users -> users:read
//         POST /api/v1/users -> users:create
//         PUT /api/v1/users/123 -> users:update
//         DELETE /api/v1/users/123 -> users:delete
func GetPermissionFromEndpoint(method, path string) (Permission, error) {
	// Remove /api/v1 prefix if present
	path = strings.TrimPrefix(path, "/api/v1")
	path = strings.TrimPrefix(path, "/api")
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid path")
	}

	// Filter out empty parts
	var nonEmptyParts []string
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}

	if len(nonEmptyParts) == 0 {
		return "", fmt.Errorf("invalid path: no resource specified")
	}

	resource := nonEmptyParts[0]
	
	// Special cases
	if resource == "auth" {
		// Auth endpoints don't require permissions (they handle their own auth)
		return "", fmt.Errorf("auth endpoints don't require permissions")
	}

	var action string
	switch strings.ToUpper(method) {
	case "GET":
		action = "read"
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		return "", fmt.Errorf("unsupported HTTP method: %s", method)
	}

	permission := Permission(fmt.Sprintf("%s:%s", resource, action))
	return permission, nil
}