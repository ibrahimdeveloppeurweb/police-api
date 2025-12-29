package observation

import (
	"net/http"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles observation HTTP requests
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController creates a new observation controller
func NewController(service Service, logger *zap.Logger) interfaces.Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// Create creates a new observation
// @Summary Create observation
// @Tags observations
// @Accept json
// @Produce json
// @Param request body CreateObservationRequest true "Create observation request"
// @Success 201 {object} ObservationResponse
// @Router /api/observations [post]
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateObservationRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	observation, err := c.service.Create(ctx.Request().Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create observation", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, observation)
}

// GetByID gets an observation by ID
// @Summary Get observation by ID
// @Tags observations
// @Produce json
// @Param id path string true "Observation ID"
// @Success 200 {object} ObservationResponse
// @Router /api/observations/{id} [get]
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	observation, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "observation not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to get observation", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, observation)
}

// List lists observations with filters
// @Summary List observations
// @Tags observations
// @Produce json
// @Param agentId query string false "Agent ID"
// @Param auteurId query string false "Auteur ID"
// @Param type query string false "Type"
// @Param categorie query string false "Categorie"
// @Param visibleAgent query string false "Visible to agent (true/false)"
// @Success 200 {array} ObservationResponse
// @Router /api/observations [get]
func (c *Controller) List(ctx echo.Context) error {
	var filters ListObservationsFilters
	if err := ctx.Bind(&filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	observations, err := c.service.List(ctx.Request().Context(), &filters)
	if err != nil {
		c.logger.Error("Failed to list observations", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, observations)
}

// Update updates an observation
// @Summary Update observation
// @Tags observations
// @Accept json
// @Produce json
// @Param id path string true "Observation ID"
// @Param request body UpdateObservationRequest true "Update observation request"
// @Success 200 {object} ObservationResponse
// @Router /api/observations/{id} [put]
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req UpdateObservationRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	observation, err := c.service.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "observation not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to update observation", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, observation)
}

// Delete deletes an observation
// @Summary Delete observation
// @Tags observations
// @Param id path string true "Observation ID"
// @Success 204
// @Router /api/observations/{id} [delete]
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	if err := c.service.Delete(ctx.Request().Context(), id); err != nil {
		if err.Error() == "observation not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to delete observation", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetByAgent gets observations for an agent
// @Summary Get observations by agent
// @Tags observations
// @Produce json
// @Param agentId path string true "Agent ID"
// @Param visibleOnly query string false "Only visible observations (true/false)"
// @Success 200 {array} ObservationResponse
// @Router /api/agents/{agentId}/observations [get]
func (c *Controller) GetByAgent(ctx echo.Context) error {
	agentID := ctx.Param("agentId")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Agent ID is required")
	}

	visibleOnly := ctx.QueryParam("visibleOnly") == "true"

	observations, err := c.service.GetByAgent(ctx.Request().Context(), agentID, visibleOnly)
	if err != nil {
		c.logger.Error("Failed to get observations by agent", zap.String("agentID", agentID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, observations)
}

// GetByAuteur gets observations created by an auteur
// @Summary Get observations by auteur
// @Tags observations
// @Produce json
// @Param auteurId path string true "Auteur ID"
// @Success 200 {array} ObservationResponse
// @Router /api/auteurs/{auteurId}/observations [get]
func (c *Controller) GetByAuteur(ctx echo.Context) error {
	auteurID := ctx.Param("auteurId")
	if auteurID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Auteur ID is required")
	}

	observations, err := c.service.GetByAuteur(ctx.Request().Context(), auteurID)
	if err != nil {
		c.logger.Error("Failed to get observations by auteur", zap.String("auteurID", auteurID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, observations)
}

// RegisterRoutes registers observation routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	observations := g.Group("/observations")
	observations.POST("", c.Create)
	observations.GET("", c.List)
	observations.GET("/:id", c.GetByID)
	observations.PUT("/:id", c.Update)
	observations.DELETE("/:id", c.Delete)

	// Nested routes
	g.GET("/agents/:agentId/observations", c.GetByAgent)
	g.GET("/auteurs/:auteurId/observations", c.GetByAuteur)
}
