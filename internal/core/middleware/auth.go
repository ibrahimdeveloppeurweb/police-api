package middleware

import (
	"strings"

	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"
	"police-trafic-api-frontend-aligned/internal/infrastructure/rbac"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtService  jwt.Service
	rbacService rbac.Service
	userRepo    repository.UserRepository
	logger      *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService jwt.Service, rbacService rbac.Service, userRepo repository.UserRepository, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		rbacService: rbacService,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return m.RequireAuthWithSkipper(nil)
}

// RequireAuthWithSkipper middleware that requires authentication with optional skip function
func (m *AuthMiddleware) RequireAuthWithSkipper(skipper func(path string) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this path should skip authentication
			if skipper != nil && skipper(c.Request().URL.Path) {
				return next(c)
			}

			token := m.extractToken(c)
			if token == "" {
				return responses.Unauthorized(c, "Authorization token required")
			}

			claims, err := m.jwtService.ValidateToken(token)
			if err != nil {
				m.logger.Warn("Token validation failed",
					zap.Error(err),
					zap.String("remote_addr", c.Request().RemoteAddr),
				)
				return responses.Unauthorized(c, "Invalid or expired token")
			}

			// Store user information in context for use in handlers
			c.Set("user_id", claims.UserID)
			c.Set("matricule", claims.Matricule)
			c.Set("user_role", claims.Role)
			c.Set("jwt_claims", claims)
			c.Set("rbac_service", m.rbacService)

			// Récupérer le commissariat de l'utilisateur
			user, err := m.userRepo.GetByID(c.Request().Context(), claims.UserID)
			if err == nil && user != nil && user.Edges.Commissariat != nil {
				c.Set("commissariat_id", user.Edges.Commissariat.ID.String())
			}

			m.logger.Debug("User authenticated successfully",
				zap.String("user_id", claims.UserID),
				zap.String("matricule", claims.Matricule),
				zap.String("role", claims.Role),
			)

			return next(c)
		}
	}
}

// RequireRole middleware that requires specific role(s)
func (m *AuthMiddleware) RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// First check authentication
			token := m.extractToken(c)
			if token == "" {
				return responses.Unauthorized(c, "Authorization token required")
			}

			claims, err := m.jwtService.ValidateToken(token)
			if err != nil {
				return responses.Unauthorized(c, "Invalid or expired token")
			}

			// Check if user has required role
			userRole := claims.Role
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.logger.Warn("Access denied - insufficient role",
					zap.String("user_id", claims.UserID),
					zap.String("user_role", userRole),
					zap.Strings("required_roles", roles),
				)
				return responses.Forbidden(c, "Insufficient permissions")
			}

			// Store user information in context
			c.Set("user_id", claims.UserID)
			c.Set("matricule", claims.Matricule)
			c.Set("user_role", claims.Role)
			c.Set("jwt_claims", claims)
			c.Set("rbac_service", m.rbacService)

			return next(c)
		}
	}
}

// OptionalAuth middleware that tries to authenticate but doesn't fail if no token
func (m *AuthMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := m.extractToken(c)
			if token != "" {
				claims, err := m.jwtService.ValidateToken(token)
				if err == nil {
					// Store user information in context if token is valid
					c.Set("user_id", claims.UserID)
					c.Set("matricule", claims.Matricule)
					c.Set("user_role", claims.Role)
					c.Set("jwt_claims", claims)
					c.Set("rbac_service", m.rbacService)
				}
			}

			return next(c)
		}
	}
}

// extractToken extracts the JWT token from Authorization header
func (m *AuthMiddleware) extractToken(c echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// RequirePermission middleware that requires specific permission based on endpoint
func (m *AuthMiddleware) RequirePermission() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// First check authentication
			token := m.extractToken(c)
			if token == "" {
				return responses.Unauthorized(c, "Authorization token required")
			}

			claims, err := m.jwtService.ValidateToken(token)
			if err != nil {
				return responses.Unauthorized(c, "Invalid or expired token")
			}

			// Get required permission from endpoint
			permission, err := rbac.GetPermissionFromEndpoint(c.Request().Method, c.Request().URL.Path)
			if err != nil {
				// If we can't determine permission, log but allow through (for auth endpoints, etc.)
				m.logger.Debug("Cannot determine permission for endpoint",
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
					zap.Error(err),
				)
				// Still store user info in context
				c.Set("user_id", claims.UserID)
				c.Set("matricule", claims.Matricule)
				c.Set("user_role", claims.Role)
				c.Set("jwt_claims", claims)
				c.Set("rbac_service", m.rbacService)
				return next(c)
			}

			// Check if user has required permission
			if !m.rbacService.HasPermission(claims.Role, permission) {
				m.logger.Warn("Access denied - insufficient permissions",
					zap.String("user_id", claims.UserID),
					zap.String("user_role", claims.Role),
					zap.String("required_permission", string(permission)),
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
				)
				return responses.Forbidden(c, "Insufficient permissions")
			}

			// Store user information in context
			c.Set("user_id", claims.UserID)
			c.Set("matricule", claims.Matricule)
			c.Set("user_role", claims.Role)
			c.Set("jwt_claims", claims)
			c.Set("required_permission", string(permission))
			c.Set("rbac_service", m.rbacService)

			m.logger.Debug("Access granted",
				zap.String("user_id", claims.UserID),
				zap.String("user_role", claims.Role),
				zap.String("permission", string(permission)),
			)

			return next(c)
		}
	}
}