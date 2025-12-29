package conducteur

import (
	"strconv"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles conducteur routes
type Controller struct {
	service Service
}

// NewController creates a new conducteur controller
func NewController(service Service) interfaces.Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers conducteur routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/conducteurs")

	// Public endpoints
	group.GET("", c.ListConducteurs)
	group.GET("/:id", c.GetConducteur)
	group.GET("/search", c.SearchConducteurs)

	// Protected endpoints
	group.POST("", c.CreateConducteur)
	group.PUT("/:id", c.UpdateConducteur)
	group.DELETE("/:id", c.DeleteConducteur)

	// Additional endpoints
	group.GET("/permis/:numeroPermis", c.GetByNumeroPermis)
	group.GET("/email/:email", c.GetByEmail)
	group.GET("/nom/:nom/prenom/:prenom", c.GetByNomPrenom)
	group.GET("/:id/statistics", c.GetStatistics)
}

// ListConducteurs lists conducteurs with filters
func (c *Controller) ListConducteurs(ctx echo.Context) error {
	request := &ListConducteursRequest{}

	// Parse filters
	if nom := ctx.QueryParam("nom"); nom != "" {
		request.Nom = &nom
	}

	if prenom := ctx.QueryParam("prenom"); prenom != "" {
		request.Prenom = &prenom
	}

	if ville := ctx.QueryParam("ville"); ville != "" {
		request.Ville = &ville
	}

	if nationalite := ctx.QueryParam("nationalite"); nationalite != "" {
		request.Nationalite = &nationalite
	}

	if active := ctx.QueryParam("active"); active != "" {
		if b, err := strconv.ParseBool(active); err == nil {
			request.Active = &b
		}
	}

	// Pagination
	if limit := ctx.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			request.Limit = l
		}
	} else {
		request.Limit = 50
	}

	if offset := ctx.QueryParam("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			request.Offset = o
		}
	}

	result, err := c.service.List(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to list conducteurs")
	}

	return responses.Success(ctx, result)
}

// GetConducteur gets a conducteur by ID
func (c *Controller) GetConducteur(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	conducteur, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "conducteur not found" {
			return responses.NotFound(ctx, "Conducteur not found")
		}
		return responses.InternalServerError(ctx, "Failed to get conducteur")
	}

	return responses.Success(ctx, conducteur)
}

// SearchConducteurs searches conducteurs
func (c *Controller) SearchConducteurs(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	if query == "" {
		return responses.BadRequest(ctx, "Search query is required")
	}

	result, err := c.service.Search(ctx.Request().Context(), query)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to search conducteurs")
	}

	return responses.Success(ctx, result)
}

// CreateConducteur creates a new conducteur
func (c *Controller) CreateConducteur(ctx echo.Context) error {
	var request CreateConducteurRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	conducteur, err := c.service.Create(ctx.Request().Context(), &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to create conducteur")
	}

	return responses.Created(ctx, conducteur)
}

// UpdateConducteur updates a conducteur
func (c *Controller) UpdateConducteur(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request UpdateConducteurRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	conducteur, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "conducteur not found" {
			return responses.NotFound(ctx, "Conducteur not found")
		}
		return responses.InternalServerError(ctx, "Failed to update conducteur")
	}

	return responses.Success(ctx, conducteur)
}

// DeleteConducteur deletes a conducteur
func (c *Controller) DeleteConducteur(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "conducteur not found" {
			return responses.NotFound(ctx, "Conducteur not found")
		}
		return responses.InternalServerError(ctx, "Failed to delete conducteur")
	}

	return responses.Success(ctx, nil)
}

// GetByNumeroPermis gets conducteur by numero permis
func (c *Controller) GetByNumeroPermis(ctx echo.Context) error {
	numeroPermis := ctx.Param("numeroPermis")
	if numeroPermis == "" {
		return responses.BadRequest(ctx, "Numero permis is required")
	}

	conducteur, err := c.service.GetByNumeroPermis(ctx.Request().Context(), numeroPermis)
	if err != nil {
		if err.Error() == "conducteur not found" {
			return responses.NotFound(ctx, "Conducteur not found")
		}
		return responses.InternalServerError(ctx, "Failed to get conducteur")
	}

	return responses.Success(ctx, conducteur)
}

// GetByEmail gets conducteur by email
func (c *Controller) GetByEmail(ctx echo.Context) error {
	email := ctx.Param("email")
	if email == "" {
		return responses.BadRequest(ctx, "Email is required")
	}

	conducteur, err := c.service.GetByEmail(ctx.Request().Context(), email)
	if err != nil {
		if err.Error() == "conducteur not found" {
			return responses.NotFound(ctx, "Conducteur not found")
		}
		return responses.InternalServerError(ctx, "Failed to get conducteur")
	}

	return responses.Success(ctx, conducteur)
}

// GetByNomPrenom gets conducteurs by nom and prenom
func (c *Controller) GetByNomPrenom(ctx echo.Context) error {
	nom := ctx.Param("nom")
	prenom := ctx.Param("prenom")

	if nom == "" || prenom == "" {
		return responses.BadRequest(ctx, "Nom and prenom are required")
	}

	result, err := c.service.GetByNomPrenom(ctx.Request().Context(), nom, prenom)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get conducteurs")
	}

	return responses.Success(ctx, result)
}

// GetStatistics gets conducteur statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	conducteurID := ctx.Param("id")
	if conducteurID == "" {
		return responses.BadRequest(ctx, "Conducteur ID is required")
	}

	stats, err := c.service.GetStatistics(ctx.Request().Context(), conducteurID)
	if err != nil {
		if err.Error() == "conducteur not found" {
			return responses.NotFound(ctx, "Conducteur not found")
		}
		return responses.InternalServerError(ctx, "Failed to get statistics")
	}

	return responses.Success(ctx, stats)
}