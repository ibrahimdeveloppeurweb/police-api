package objectif

import (
	"net/http"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles objectif HTTP requests
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController creates a new objectif controller
func NewController(service Service, logger *zap.Logger) interfaces.Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// Create creates a new objectif
// @Summary Create objectif
// @Tags objectifs
// @Accept json
// @Produce json
// @Param request body CreateObjectifRequest true "Create objectif request"
// @Success 201 {object} ObjectifResponse
// @Router /api/objectifs [post]
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateObjectifRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	objectif, err := c.service.Create(ctx.Request().Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create objectif", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, objectif)
}

// GetByID gets an objectif by ID
// @Summary Get objectif by ID
// @Tags objectifs
// @Produce json
// @Param id path string true "Objectif ID"
// @Success 200 {object} ObjectifResponse
// @Router /api/objectifs/{id} [get]
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	objectif, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "objectif not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to get objectif", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, objectif)
}

// List lists objectifs with filters
// @Summary List objectifs
// @Tags objectifs
// @Produce json
// @Param agentId query string false "Agent ID"
// @Param periode query string false "Periode"
// @Param statut query string false "Statut"
// @Param dateDebut query string false "Date debut (YYYY-MM-DD)"
// @Param dateFin query string false "Date fin (YYYY-MM-DD)"
// @Success 200 {array} ObjectifResponse
// @Router /api/objectifs [get]
func (c *Controller) List(ctx echo.Context) error {
	var filters ListObjectifsFilters
	if err := ctx.Bind(&filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	objectifs, err := c.service.List(ctx.Request().Context(), &filters)
	if err != nil {
		c.logger.Error("Failed to list objectifs", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, objectifs)
}

// Update updates an objectif
// @Summary Update objectif
// @Tags objectifs
// @Accept json
// @Produce json
// @Param id path string true "Objectif ID"
// @Param request body UpdateObjectifRequest true "Update objectif request"
// @Success 200 {object} ObjectifResponse
// @Router /api/objectifs/{id} [put]
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req UpdateObjectifRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	objectif, err := c.service.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "objectif not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to update objectif", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, objectif)
}

// Delete deletes an objectif
// @Summary Delete objectif
// @Tags objectifs
// @Param id path string true "Objectif ID"
// @Success 204
// @Router /api/objectifs/{id} [delete]
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	if err := c.service.Delete(ctx.Request().Context(), id); err != nil {
		if err.Error() == "objectif not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to delete objectif", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetByAgent gets objectifs for an agent
// @Summary Get objectifs by agent
// @Tags objectifs
// @Produce json
// @Param agentId path string true "Agent ID"
// @Success 200 {array} ObjectifResponse
// @Router /api/agents/{agentId}/objectifs [get]
func (c *Controller) GetByAgent(ctx echo.Context) error {
	agentID := ctx.Param("agentId")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Agent ID is required")
	}

	objectifs, err := c.service.GetByAgent(ctx.Request().Context(), agentID)
	if err != nil {
		c.logger.Error("Failed to get objectifs by agent", zap.String("agentID", agentID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, objectifs)
}

// UpdateProgression updates the progression of an objectif
// @Summary Update objectif progression
// @Tags objectifs
// @Accept json
// @Produce json
// @Param id path string true "Objectif ID"
// @Param request body UpdateProgressionRequest true "Update progression request"
// @Success 200 {object} ObjectifResponse
// @Router /api/objectifs/{id}/progression [put]
func (c *Controller) UpdateProgression(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req UpdateProgressionRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	objectif, err := c.service.UpdateProgression(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "objectif not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to update progression", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, objectif)
}

// RegisterRoutes registers objectif routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	objectifs := g.Group("/objectifs")
	objectifs.POST("", c.Create)
	objectifs.GET("", c.List)
	objectifs.GET("/:id", c.GetByID)
	objectifs.PUT("/:id", c.Update)
	objectifs.DELETE("/:id", c.Delete)
	objectifs.PUT("/:id/progression", c.UpdateProgression)

	// Nested routes
	g.GET("/agents/:agentId/objectifs", c.GetByAgent)
}
