package mission

import (
	"net/http"
	"strconv"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles mission HTTP requests
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController creates a new mission controller
func NewController(service Service, logger *zap.Logger) interfaces.Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// Create creates a new mission
// @Summary Create mission
// @Tags missions
// @Accept json
// @Produce json
// @Param request body CreateMissionRequest true "Create mission request"
// @Success 201 {object} MissionResponse
// @Router /api/missions [post]
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateMissionRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	mission, err := c.service.Create(ctx.Request().Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create mission", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, mission)
}

// GetByID gets a mission by ID
// @Summary Get mission by ID
// @Tags missions
// @Produce json
// @Param id path string true "Mission ID"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id} [get]
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	mission, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to get mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// List lists missions with filters
// @Summary List missions
// @Tags missions
// @Produce json
// @Param agentId query string false "Agent ID"
// @Param equipeId query string false "Equipe ID"
// @Param commissariatId query string false "Commissariat ID"
// @Param statut query string false "Statut"
// @Param type query string false "Type"
// @Param dateDebut query string false "Date debut (YYYY-MM-DD)"
// @Param dateFin query string false "Date fin (YYYY-MM-DD)"
// @Success 200 {array} MissionResponse
// @Router /api/missions [get]
func (c *Controller) List(ctx echo.Context) error {
	var filters ListMissionsFilters
	if err := ctx.Bind(&filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	missions, err := c.service.List(ctx.Request().Context(), &filters)
	if err != nil {
		c.logger.Error("Failed to list missions", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{"data": missions})
}

// Update updates a mission
// @Summary Update mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path string true "Mission ID"
// @Param request body UpdateMissionRequest true "Update mission request"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id} [put]
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req UpdateMissionRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	mission, err := c.service.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to update mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// Delete deletes a mission
// @Summary Delete mission
// @Tags missions
// @Param id path string true "Mission ID"
// @Success 204
// @Router /api/missions/{id} [delete]
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	if err := c.service.Delete(ctx.Request().Context(), id); err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to delete mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetByAgent gets missions for an agent
// @Summary Get missions by agent
// @Tags missions
// @Produce json
// @Param agentId path string true "Agent ID"
// @Param limit query int false "Limit"
// @Success 200 {array} MissionResponse
// @Router /api/agents/{agentId}/missions [get]
func (c *Controller) GetByAgent(ctx echo.Context) error {
	agentID := ctx.Param("agentId")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Agent ID is required")
	}

	limit := 0
	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	missions, err := c.service.GetByAgent(ctx.Request().Context(), agentID, limit)
	if err != nil {
		c.logger.Error("Failed to get missions by agent", zap.String("agentID", agentID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, missions)
}

// GetByEquipe gets missions for an equipe
// @Summary Get missions by equipe
// @Tags missions
// @Produce json
// @Param equipeId path string true "Equipe ID"
// @Success 200 {array} MissionResponse
// @Router /api/equipes/{equipeId}/missions [get]
func (c *Controller) GetByEquipe(ctx echo.Context) error {
	equipeID := ctx.Param("equipeId")
	if equipeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Equipe ID is required")
	}

	missions, err := c.service.GetByEquipe(ctx.Request().Context(), equipeID)
	if err != nil {
		c.logger.Error("Failed to get missions by equipe", zap.String("equipeID", equipeID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, missions)
}

// StartMission starts a mission
// @Summary Start mission
// @Tags missions
// @Produce json
// @Param id path string true "Mission ID"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id}/start [post]
func (c *Controller) StartMission(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	mission, err := c.service.StartMission(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to start mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// EndMission ends a mission
// @Summary End mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path string true "Mission ID"
// @Param request body EndMissionRequest true "End mission request"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id}/end [post]
func (c *Controller) EndMission(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req EndMissionRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	mission, err := c.service.EndMission(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to end mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// CancelMission cancels a mission
// @Summary Cancel mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path string true "Mission ID"
// @Param request body CancelMissionRequest false "Cancel mission request"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id}/cancel [post]
func (c *Controller) CancelMission(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req CancelMissionRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	mission, err := c.service.CancelMission(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to cancel mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// AddAgents adds agents to a mission
// @Summary Add agents to mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path string true "Mission ID"
// @Param request body AddAgentsRequest true "Add agents request"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id}/agents [post]
func (c *Controller) AddAgents(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req AddAgentsRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	mission, err := c.service.AddAgents(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to add agents to mission", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// RemoveAgent removes an agent from a mission
// @Summary Remove agent from mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path string true "Mission ID"
// @Param agentId path string true "Agent ID"
// @Success 200 {object} MissionResponse
// @Router /api/missions/{id}/agents/{agentId} [delete]
func (c *Controller) RemoveAgent(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Mission ID is required")
	}

	agentID := ctx.Param("agentId")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Agent ID is required")
	}

	req := &RemoveAgentRequest{AgentID: agentID}
	mission, err := c.service.RemoveAgent(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "mission not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to remove agent from mission", zap.String("id", id), zap.String("agentId", agentID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mission)
}

// RegisterRoutes registers mission routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	missions := g.Group("/missions")
	missions.POST("", c.Create)
	missions.GET("", c.List)
	missions.GET("/:id", c.GetByID)
	missions.PUT("/:id", c.Update)
	missions.DELETE("/:id", c.Delete)
	missions.POST("/:id/start", c.StartMission)
	missions.POST("/:id/end", c.EndMission)
	missions.POST("/:id/cancel", c.CancelMission)
	missions.POST("/:id/agents", c.AddAgents)
	missions.DELETE("/:id/agents/:agentId", c.RemoveAgent)

	// Nested routes
	g.GET("/agents/:agentId/missions", c.GetByAgent)
	g.GET("/equipes/:equipeId/missions", c.GetByEquipe)
}
