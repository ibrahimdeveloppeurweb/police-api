package competence

import (
	"net/http"
	"strconv"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles competence HTTP requests
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController creates a new competence controller
func NewController(service Service, logger *zap.Logger) interfaces.Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// Create creates a new competence
// @Summary Create competence
// @Tags competences
// @Accept json
// @Produce json
// @Param request body CreateCompetenceRequest true "Create competence request"
// @Success 201 {object} CompetenceResponse
// @Router /api/competences [post]
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateCompetenceRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	competence, err := c.service.Create(ctx.Request().Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create competence", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, competence)
}

// GetByID gets a competence by ID
// @Summary Get competence by ID
// @Tags competences
// @Produce json
// @Param id path string true "Competence ID"
// @Success 200 {object} CompetenceResponse
// @Router /api/competences/{id} [get]
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	competence, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "competence not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to get competence", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, competence)
}

// List lists competences with filters
// @Summary List competences
// @Tags competences
// @Produce json
// @Param type query string false "Type"
// @Param active query string false "Active status (true/false)"
// @Param search query string false "Search term"
// @Param organisme query string false "Organisme"
// @Success 200 {array} CompetenceResponse
// @Router /api/competences [get]
func (c *Controller) List(ctx echo.Context) error {
	var filters ListCompetencesFilters
	if err := ctx.Bind(&filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	competences, err := c.service.List(ctx.Request().Context(), &filters)
	if err != nil {
		c.logger.Error("Failed to list competences", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, competences)
}

// Update updates a competence
// @Summary Update competence
// @Tags competences
// @Accept json
// @Produce json
// @Param id path string true "Competence ID"
// @Param request body UpdateCompetenceRequest true "Update competence request"
// @Success 200 {object} CompetenceResponse
// @Router /api/competences/{id} [put]
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req UpdateCompetenceRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	competence, err := c.service.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "competence not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to update competence", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, competence)
}

// Delete deletes a competence
// @Summary Delete competence
// @Tags competences
// @Param id path string true "Competence ID"
// @Success 204
// @Router /api/competences/{id} [delete]
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	if err := c.service.Delete(ctx.Request().Context(), id); err != nil {
		if err.Error() == "competence not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to delete competence", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// AssignToAgent assigns a competence to an agent
// @Summary Assign competence to agent
// @Tags competences
// @Accept json
// @Param id path string true "Competence ID"
// @Param request body AssignCompetenceRequest true "Assign competence request"
// @Success 204
// @Router /api/competences/{id}/agents [post]
func (c *Controller) AssignToAgent(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req AssignCompetenceRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.service.AssignToAgent(ctx.Request().Context(), id, &req); err != nil {
		c.logger.Error("Failed to assign competence", zap.String("competenceID", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// RemoveFromAgent removes a competence from an agent
// @Summary Remove competence from agent
// @Tags competences
// @Param id path string true "Competence ID"
// @Param agentId path string true "Agent ID"
// @Success 204
// @Router /api/competences/{id}/agents/{agentId} [delete]
func (c *Controller) RemoveFromAgent(ctx echo.Context) error {
	id := ctx.Param("id")
	agentID := ctx.Param("agentId")
	if id == "" || agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID and agentId are required")
	}

	if err := c.service.RemoveFromAgent(ctx.Request().Context(), id, agentID); err != nil {
		c.logger.Error("Failed to remove competence", zap.String("competenceID", id), zap.String("agentID", agentID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetByAgent gets competences for an agent
// @Summary Get competences by agent
// @Tags competences
// @Produce json
// @Param agentId path string true "Agent ID"
// @Success 200 {array} CompetenceResponse
// @Router /api/agents/{agentId}/competences [get]
func (c *Controller) GetByAgent(ctx echo.Context) error {
	agentID := ctx.Param("agentId")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Agent ID is required")
	}

	competences, err := c.service.GetByAgent(ctx.Request().Context(), agentID)
	if err != nil {
		c.logger.Error("Failed to get competences by agent", zap.String("agentID", agentID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, competences)
}

// GetExpiring gets competences expiring soon
// @Summary Get expiring competences
// @Tags competences
// @Produce json
// @Param days query int false "Days ahead (default 30)"
// @Success 200 {array} CompetenceResponse
// @Router /api/competences/expiring [get]
func (c *Controller) GetExpiring(ctx echo.Context) error {
	daysAhead := 30
	if daysStr := ctx.QueryParam("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			daysAhead = d
		}
	}

	competences, err := c.service.GetExpiring(ctx.Request().Context(), daysAhead)
	if err != nil {
		c.logger.Error("Failed to get expiring competences", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, competences)
}

// RegisterRoutes registers competence routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	competences := g.Group("/competences")
	competences.POST("", c.Create)
	competences.GET("", c.List)
	competences.GET("/expiring", c.GetExpiring)
	competences.GET("/:id", c.GetByID)
	competences.PUT("/:id", c.Update)
	competences.DELETE("/:id", c.Delete)
	competences.POST("/:id/agents", c.AssignToAgent)
	competences.DELETE("/:id/agents/:agentId", c.RemoveFromAgent)

	// Nested routes
	g.GET("/agents/:agentId/competences", c.GetByAgent)
}
