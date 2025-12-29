package equipe

import (
	"net/http"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles equipe HTTP requests
type Controller struct {
	service Service
	logger  *zap.Logger
}

// NewController creates a new equipe controller
func NewController(service Service, logger *zap.Logger) interfaces.Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

// Create creates a new equipe
// @Summary Create equipe
// @Tags equipes
// @Accept json
// @Produce json
// @Param request body CreateEquipeRequest true "Create equipe request"
// @Success 201 {object} EquipeResponse
// @Router /api/equipes [post]
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateEquipeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	equipe, err := c.service.Create(ctx.Request().Context(), &req)
	if err != nil {
		c.logger.Error("Failed to create equipe", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, equipe)
}

// GetByID gets an equipe by ID
// @Summary Get equipe by ID
// @Tags equipes
// @Produce json
// @Param id path string true "Equipe ID"
// @Success 200 {object} EquipeResponse
// @Router /api/equipes/{id} [get]
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	equipe, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "equipe not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to get equipe", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, equipe)
}

// List lists equipes with filters
// @Summary List equipes
// @Tags equipes
// @Produce json
// @Param commissariatId query string false "Commissariat ID"
// @Param active query string false "Active status (true/false)"
// @Param search query string false "Search term"
// @Success 200 {array} EquipeResponse
// @Router /api/equipes [get]
func (c *Controller) List(ctx echo.Context) error {
	var filters ListEquipesFilters
	if err := ctx.Bind(&filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}

	equipes, err := c.service.List(ctx.Request().Context(), &filters)
	if err != nil {
		c.logger.Error("Failed to list equipes", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, equipes)
}

// Update updates an equipe
// @Summary Update equipe
// @Tags equipes
// @Accept json
// @Produce json
// @Param id path string true "Equipe ID"
// @Param request body UpdateEquipeRequest true "Update equipe request"
// @Success 200 {object} EquipeResponse
// @Router /api/equipes/{id} [put]
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req UpdateEquipeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	equipe, err := c.service.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "equipe not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to update equipe", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, equipe)
}

// Delete deletes an equipe
// @Summary Delete equipe
// @Tags equipes
// @Param id path string true "Equipe ID"
// @Success 204
// @Router /api/equipes/{id} [delete]
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	if err := c.service.Delete(ctx.Request().Context(), id); err != nil {
		if err.Error() == "equipe not found" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		c.logger.Error("Failed to delete equipe", zap.String("id", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// AddMembre adds a member to the equipe
// @Summary Add member to equipe
// @Tags equipes
// @Accept json
// @Param id path string true "Equipe ID"
// @Param request body AddMembreRequest true "Add member request"
// @Success 204
// @Router /api/equipes/{id}/membres [post]
func (c *Controller) AddMembre(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req AddMembreRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.service.AddMembre(ctx.Request().Context(), id, &req); err != nil {
		c.logger.Error("Failed to add membre", zap.String("equipeID", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// RemoveMembre removes a member from the equipe
// @Summary Remove member from equipe
// @Tags equipes
// @Param id path string true "Equipe ID"
// @Param userId path string true "User ID"
// @Success 204
// @Router /api/equipes/{id}/membres/{userId} [delete]
func (c *Controller) RemoveMembre(ctx echo.Context) error {
	id := ctx.Param("id")
	userID := ctx.Param("userId")
	if id == "" || userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID and userId are required")
	}

	if err := c.service.RemoveMembre(ctx.Request().Context(), id, userID); err != nil {
		c.logger.Error("Failed to remove membre", zap.String("equipeID", id), zap.String("userID", userID), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// SetChefEquipe sets the team leader
// @Summary Set team leader
// @Tags equipes
// @Accept json
// @Param id path string true "Equipe ID"
// @Param request body SetChefEquipeRequest true "Set chef equipe request"
// @Success 204
// @Router /api/equipes/{id}/chef [put]
func (c *Controller) SetChefEquipe(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	var req SetChefEquipeRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.service.SetChefEquipe(ctx.Request().Context(), id, &req); err != nil {
		c.logger.Error("Failed to set chef equipe", zap.String("equipeID", id), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// RegisterRoutes registers equipe routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	equipes := g.Group("/equipes")
	equipes.POST("", c.Create)
	equipes.GET("", c.List)
	equipes.GET("/:id", c.GetByID)
	equipes.PUT("/:id", c.Update)
	equipes.DELETE("/:id", c.Delete)
	equipes.POST("/:id/membres", c.AddMembre)
	equipes.DELETE("/:id/membres/:userId", c.RemoveMembre)
	equipes.PUT("/:id/chef", c.SetChefEquipe)
}
