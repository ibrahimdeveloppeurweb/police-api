package middleware

import (
	"fmt"

	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"
	"police-trafic-api-frontend-aligned/internal/infrastructure/rbac"

	"github.com/labstack/echo/v4"
)

// UserContext provides utilities to get user information from context
type UserContext struct {
	UserID        string
	Matricule     string
	Role          string
	Claims        *jwt.Claims
	HasPermission func(permission rbac.Permission) bool
}

// GetUserFromContext extracts user information from the Echo context
// This should be used in controllers after authentication middleware has run
func GetUserFromContext(c echo.Context) (*UserContext, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	matricule, ok := c.Get("matricule").(string)
	if !ok {
		return nil, fmt.Errorf("matricule not found in context")
	}

	role, ok := c.Get("user_role").(string)
	if !ok {
		return nil, fmt.Errorf("user_role not found in context")
	}

	claims, ok := c.Get("jwt_claims").(*jwt.Claims)
	if !ok {
		return nil, fmt.Errorf("jwt_claims not found in context")
	}

	// Get RBAC service from context to check permissions
	rbacService := c.Get("rbac_service")
	var hasPermission func(rbac.Permission) bool
	if rbacSvc, ok := rbacService.(rbac.Service); ok {
		hasPermission = func(permission rbac.Permission) bool {
			return rbacSvc.HasPermission(role, permission)
		}
	} else {
		hasPermission = func(rbac.Permission) bool { return false }
	}

	return &UserContext{
		UserID:        userID,
		Matricule:     matricule,
		Role:          role,
		Claims:        claims,
		HasPermission: hasPermission,
	}, nil
}

// IsAdmin checks if the user has admin role
func (uc *UserContext) IsAdmin() bool {
	return uc.Role == string(rbac.RoleAdmin)
}

// IsSupervisor checks if the user has supervisor role or higher
func (uc *UserContext) IsSupervisor() bool {
	return uc.Role == string(rbac.RoleSupervisor) || uc.IsAdmin()
}

// IsAgent checks if the user has agent role
func (uc *UserContext) IsAgent() bool {
	return uc.Role == string(rbac.RoleAgent)
}

// CanRead checks if user can read a resource
func (uc *UserContext) CanRead(resource string) bool {
	permission := rbac.Permission(fmt.Sprintf("%s:read", resource))
	return uc.HasPermission(permission)
}

// CanCreate checks if user can create a resource
func (uc *UserContext) CanCreate(resource string) bool {
	permission := rbac.Permission(fmt.Sprintf("%s:create", resource))
	return uc.HasPermission(permission)
}

// CanUpdate checks if user can update a resource
func (uc *UserContext) CanUpdate(resource string) bool {
	permission := rbac.Permission(fmt.Sprintf("%s:update", resource))
	return uc.HasPermission(permission)
}

// CanDelete checks if user can delete a resource
func (uc *UserContext) CanDelete(resource string) bool {
	permission := rbac.Permission(fmt.Sprintf("%s:delete", resource))
	return uc.HasPermission(permission)
}