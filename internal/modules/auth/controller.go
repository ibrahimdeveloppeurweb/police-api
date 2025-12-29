package auth

import (
	"strings"

	"police-trafic-api-frontend-aligned/internal/shared/responses"
	"police-trafic-api-frontend-aligned/internal/shared/utils"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles auth HTTP requests
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController creates a new auth controller
func NewController(service Service, logger *zap.Logger) *Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers auth routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	auth := e.Group("/auth")
	auth.POST("/register", ctrl.Register)
	auth.POST("/login", ctrl.Login)
	auth.POST("/logout", ctrl.Logout)
	auth.POST("/refresh", ctrl.RefreshToken)
	auth.GET("/me", ctrl.GetCurrentUser)
	auth.GET("/sessions", ctrl.GetSessions)
	auth.DELETE("/sessions/:id", ctrl.RevokeSession)
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} User
// @Failure 400 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Router /auth/register [post]
func (ctrl *Controller) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		ctrl.logger.Error("Failed to bind register request", zap.Error(err))
		return responses.BadRequest(c, "Invalid request format")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		ctrl.logger.Error("Registration validation failed", zap.Error(err))
		return responses.BadRequest(c, "Validation failed: "+err.Error())
	}

	user, err := ctrl.service.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return responses.Conflict(c, err.Error())
		}
		return responses.Error(c, err)
	}

	return c.JSON(201, responses.SuccessResponse{
		Message: "User registered successfully",
		Data:    user,
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with matricule and password. For mobile apps, include device info to get refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials with optional device info"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Router /auth/login [post]
func (ctrl *Controller) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		ctrl.logger.Error("Failed to bind login request", zap.Error(err))
		return responses.BadRequest(c, "Invalid request format")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		ctrl.logger.Error("Login validation failed", zap.Error(err))
		return responses.BadRequest(c, "Validation failed: "+err.Error())
	}

	// Use GetIdentifier to support both matricule and email fields
	identifier := req.GetIdentifier()
	if identifier == "" {
		return responses.BadRequest(c, "Email or matricule is required")
	}

	// Get client IP address
	ipAddress := c.RealIP()

	response, err := ctrl.service.Login(req, ipAddress)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, response)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user and invalidate session. Can logout from specific session, current session, or all devices.
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body LogoutRequest false "Logout options"
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Router /auth/logout [post]
func (ctrl *Controller) Logout(c echo.Context) error {
	token := ctrl.extractToken(c)
	if token == "" {
		return responses.Unauthorized(c, "Authorization token required")
	}

	var req LogoutRequest
	// Bind is optional - if body is empty, defaults will be used
	_ = c.Bind(&req)

	if err := ctrl.service.Logout(req, token); err != nil {
		return responses.Error(c, err)
	}

	return responses.SuccessWithMessage(c, "Logout successful", nil)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh user access token. For mobile apps, use refresh_token and device_id in body. For web, use Bearer token.
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RefreshTokenRequest false "Refresh token (for mobile apps)"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} responses.ErrorResponse
// @Router /auth/refresh [post]
func (ctrl *Controller) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err == nil && req.RefreshToken != "" && req.DeviceID != "" {
		// Mobile app: use refresh token
		ipAddress := c.RealIP()
		response, err := ctrl.service.RefreshToken(req, ipAddress)
		if err != nil {
			return responses.Error(c, err)
		}
		return responses.Success(c, response)
	}

	// Web app: use Bearer token (legacy)
	token := ctrl.extractToken(c)
	if token == "" {
		return responses.Unauthorized(c, "Authorization token or refresh_token required")
	}

	response, err := ctrl.service.RefreshTokenLegacy(token)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, response)
}

// GetCurrentUser handles getting current user info
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} User
// @Failure 401 {object} responses.ErrorResponse
// @Router /auth/me [get]
func (ctrl *Controller) GetCurrentUser(c echo.Context) error {
	token := ctrl.extractToken(c)
	if token == "" {
		return responses.Unauthorized(c, "Authorization token required")
	}

	user, err := ctrl.service.GetCurrentUser(token)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, user)
}

// GetSessions returns all active sessions for the current user
// @Summary Get user sessions
// @Description Get all active sessions for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} []SessionDTO
// @Failure 401 {object} responses.ErrorResponse
// @Router /auth/sessions [get]
func (ctrl *Controller) GetSessions(c echo.Context) error {
	token := ctrl.extractToken(c)
	if token == "" {
		return responses.Unauthorized(c, "Authorization token required")
	}

	sessions, err := ctrl.service.GetUserSessions(token)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, sessions)
}

// RevokeSession revokes a specific session
// @Summary Revoke session
// @Description Revoke a specific session (logout from a device)
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Session ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /auth/sessions/{id} [delete]
func (ctrl *Controller) RevokeSession(c echo.Context) error {
	token := ctrl.extractToken(c)
	if token == "" {
		return responses.Unauthorized(c, "Authorization token required")
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		return responses.BadRequest(c, "Session ID required")
	}

	if err := ctrl.service.RevokeSession(token, sessionID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return responses.NotFound(c, "Session not found")
		}
		return responses.Error(c, err)
	}

	return responses.SuccessWithMessage(c, "Session revoked", nil)
}

func (ctrl *Controller) extractToken(c echo.Context) string {
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
