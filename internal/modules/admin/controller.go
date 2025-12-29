package admin

import (
	"net/http"

	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles admin HTTP requests
type Controller struct {
	service Service
}

// NewController creates a new admin controller
func NewController(service Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers admin routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	admin := e.Group("/admin")

	// Statistics
	admin.GET("/statistiques", ctrl.GetStatistiquesNationales)

	// Dashboard
	admin.GET("/agents/dashboard", ctrl.GetAgentsDashboard)

	// Commissariats
	admin.GET("/commissariats", ctrl.GetCommissariats)
	admin.GET("/commissariats/:id", ctrl.GetCommissariat)
	admin.POST("/commissariats", ctrl.CreateCommissariat)
	admin.PUT("/commissariats/:id", ctrl.UpdateCommissariat)
	admin.DELETE("/commissariats/:id", ctrl.DeleteCommissariat)

	// Agents
	admin.GET("/agents", ctrl.GetAgents)
	admin.GET("/agents/:id", ctrl.GetAgent)
	admin.POST("/agents", ctrl.CreateAgent)
	admin.PUT("/agents/:id", ctrl.UpdateAgent)
	admin.DELETE("/agents/:id", ctrl.DeleteAgent)
	admin.GET("/agents/:id/statistiques", ctrl.GetAgentStatistiques)

	// Session Management (Remote Logout)
	admin.GET("/agents/:id/sessions", ctrl.GetAgentSessions)
	admin.DELETE("/agents/:id/sessions/:sessionId", ctrl.RevokeAgentSession)
	admin.DELETE("/agents/:id/sessions", ctrl.RevokeAllAgentSessions)
}

// GetStatistiquesNationales handles GET /admin/statistiques
func (ctrl *Controller) GetStatistiquesNationales(c echo.Context) error {
	stats, err := ctrl.service.GetStatistiquesNationales(c.Request().Context())
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, stats)
}

// GetAgentsDashboard handles GET /admin/agents/dashboard
func (ctrl *Controller) GetAgentsDashboard(c echo.Context) error {
	req := &AgentDashboardRequest{
		Periode:   c.QueryParam("periode"),
		DateDebut: c.QueryParam("dateDebut"),
		DateFin:   c.QueryParam("dateFin"),
	}

	// Default to "jour" if no period specified
	if req.Periode == "" {
		req.Periode = "jour"
	}

	dashboard, err := ctrl.service.GetAgentsDashboard(c.Request().Context(), req)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, dashboard)
}

// GetCommissariats handles GET /admin/commissariats
func (ctrl *Controller) GetCommissariats(c echo.Context) error {
	commissariats, err := ctrl.service.GetCommissariats(c.Request().Context())
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, commissariats)
}

// GetCommissariat handles GET /admin/commissariats/:id
func (ctrl *Controller) GetCommissariat(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	commissariat, err := ctrl.service.GetCommissariat(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, commissariat)
}

// CreateCommissariat handles POST /admin/commissariats
func (ctrl *Controller) CreateCommissariat(c echo.Context) error {
	var req CreateCommissariatRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return responses.BadRequest(c, err.Error())
	}

	commissariat, err := ctrl.service.CreateCommissariat(c.Request().Context(), &req)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Created(c, commissariat)
}

// UpdateCommissariat handles PUT /admin/commissariats/:id
func (ctrl *Controller) UpdateCommissariat(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateCommissariatRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	commissariat, err := ctrl.service.UpdateCommissariat(c.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, commissariat)
}

// DeleteCommissariat handles DELETE /admin/commissariats/:id
func (ctrl *Controller) DeleteCommissariat(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	err := ctrl.service.DeleteCommissariat(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// GetAgents handles GET /admin/agents
func (ctrl *Controller) GetAgents(c echo.Context) error {
	var commissariatID *string
	if id := c.QueryParam("commissariatId"); id != "" {
		commissariatID = &id
	}

	agents, err := ctrl.service.GetAgents(c.Request().Context(), commissariatID)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, agents)
}

// GetAgent handles GET /admin/agents/:id
func (ctrl *Controller) GetAgent(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	agent, err := ctrl.service.GetAgent(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			return responses.NotFound(c, "Agent not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, agent)
}

// UpdateAgent handles PUT /admin/agents/:id
func (ctrl *Controller) UpdateAgent(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateAgentRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agent, err := ctrl.service.UpdateAgent(c.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			return responses.NotFound(c, "Agent not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, agent)
}

// CreateAgent handles POST /admin/agents
func (ctrl *Controller) CreateAgent(c echo.Context) error {
	var req CreateAgentRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return responses.BadRequest(c, err.Error())
	}

	agent, err := ctrl.service.CreateAgent(c.Request().Context(), &req)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Created(c, agent)
}

// DeleteAgent handles DELETE /admin/agents/:id
func (ctrl *Controller) DeleteAgent(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	err := ctrl.service.DeleteAgent(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			return responses.NotFound(c, "Agent not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// GetAgentStatistiques handles GET /admin/agents/:id/statistiques
func (ctrl *Controller) GetAgentStatistiques(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	stats, err := ctrl.service.GetAgentStatistiques(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			return responses.NotFound(c, "Agent not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, stats)
}

// GetAgentSessions handles GET /admin/agents/:id/sessions
// @Summary Get agent sessions
// @Description Get all active sessions for an agent (for remote logout)
// @Tags admin
// @Produce json
// @Security Bearer
// @Param id path string true "Agent ID"
// @Success 200 {object} []AgentSessionResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /admin/agents/{id}/sessions [get]
func (ctrl *Controller) GetAgentSessions(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	sessions, err := ctrl.service.GetAgentSessions(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "agent not found" {
			return responses.NotFound(c, "Agent not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.Success(c, sessions)
}

// RevokeAgentSession handles DELETE /admin/agents/:id/sessions/:sessionId
// @Summary Revoke agent session
// @Description Revoke a specific session for an agent (remote logout)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Agent ID"
// @Param sessionId path string true "Session ID"
// @Param request body RevokeSessionRequest false "Revocation reason"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /admin/agents/{id}/sessions/{sessionId} [delete]
func (ctrl *Controller) RevokeAgentSession(c echo.Context) error {
	agentID := c.Param("id")
	sessionID := c.Param("sessionId")

	if agentID == "" || sessionID == "" {
		return responses.BadRequest(c, "Agent ID and Session ID are required")
	}

	var req RevokeSessionRequest
	_ = c.Bind(&req)

	err := ctrl.service.RevokeAgentSession(c.Request().Context(), agentID, sessionID, req.Reason)
	if err != nil {
		if err.Error() == "agent not found" {
			return responses.NotFound(c, "Agent not found")
		}
		if err.Error() == "session not found for this agent" {
			return responses.NotFound(c, "Session not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.SuccessWithMessage(c, "Session revoked successfully", nil)
}

// RevokeAllAgentSessions handles DELETE /admin/agents/:id/sessions
// @Summary Revoke all agent sessions
// @Description Revoke all sessions for an agent (remote logout from all devices)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Agent ID"
// @Param request body RevokeSessionRequest false "Revocation reason"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /admin/agents/{id}/sessions [delete]
func (ctrl *Controller) RevokeAllAgentSessions(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return responses.BadRequest(c, "Agent ID is required")
	}

	var req RevokeSessionRequest
	_ = c.Bind(&req)

	err := ctrl.service.RevokeAllAgentSessions(c.Request().Context(), agentID, req.Reason)
	if err != nil {
		if err.Error() == "agent not found" {
			return responses.NotFound(c, "Agent not found")
		}
		return responses.InternalServerError(c, err.Error())
	}
	return responses.SuccessWithMessage(c, "All sessions revoked successfully", nil)
}
