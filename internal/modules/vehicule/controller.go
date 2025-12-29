package vehicule

import (
	"strconv"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles vehicule routes
type Controller struct {
	service Service
}

// NewController creates a new vehicule controller
func NewController(service Service) interfaces.Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers vehicule routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/vehicules")

	// Public endpoints
	group.GET("", c.ListVehicules)
	group.GET("/:id", c.GetVehicule)
	group.GET("/immatriculation/:immat", c.GetByImmatriculation)
	group.GET("/search", c.SearchVehicules)

	// Protected endpoints
	group.POST("", c.CreateVehicule)
	group.PUT("/:id", c.UpdateVehicule)
	group.DELETE("/:id", c.DeleteVehicule)

	// Additional endpoints
	group.GET("/marque/:marque", c.GetByMarque)
	group.GET("/type/:type", c.GetByType)
}

// ListVehicules lists vehicules with filters
func (c *Controller) ListVehicules(ctx echo.Context) error {
	request := &ListVehiculesRequest{}

	// Parse filters
	if marque := ctx.QueryParam("marque"); marque != "" {
		request.Marque = &marque
	}

	if modele := ctx.QueryParam("modele"); modele != "" {
		request.Modele = &modele
	}

	if typeVehicule := ctx.QueryParam("type_vehicule"); typeVehicule != "" {
		request.TypeVehicule = &typeVehicule
	}

	if proprietaireNom := ctx.QueryParam("proprietaire_nom"); proprietaireNom != "" {
		request.ProprietaireNom = &proprietaireNom
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
		return responses.InternalServerError(ctx, "Failed to list vehicules")
	}

	return responses.Success(ctx, result)
}

// GetVehicule gets a vehicule by ID
func (c *Controller) GetVehicule(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	vehicule, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "vehicule not found" {
			return responses.NotFound(ctx, "Vehicule not found")
		}
		return responses.InternalServerError(ctx, "Failed to get vehicule")
	}

	return responses.Success(ctx, vehicule)
}

// GetByImmatriculation gets vehicule by immatriculation
func (c *Controller) GetByImmatriculation(ctx echo.Context) error {
	immat := ctx.Param("immat")
	if immat == "" {
		return responses.BadRequest(ctx, "Immatriculation is required")
	}

	vehicule, err := c.service.GetByImmatriculation(ctx.Request().Context(), immat)
	if err != nil {
		if err.Error() == "vehicule not found" {
			return responses.NotFound(ctx, "Vehicule not found")
		}
		return responses.InternalServerError(ctx, "Failed to get vehicule")
	}

	return responses.Success(ctx, vehicule)
}

// SearchVehicules searches vehicules
func (c *Controller) SearchVehicules(ctx echo.Context) error {
	query := ctx.QueryParam("q")
	if query == "" {
		return responses.BadRequest(ctx, "Search query is required")
	}

	result, err := c.service.Search(ctx.Request().Context(), query)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to search vehicules")
	}

	return responses.Success(ctx, result)
}

// CreateVehicule creates a new vehicule
func (c *Controller) CreateVehicule(ctx echo.Context) error {
	var request CreateVehiculeRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	vehicule, err := c.service.Create(ctx.Request().Context(), &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to create vehicule")
	}

	return responses.Created(ctx, vehicule)
}

// UpdateVehicule updates a vehicule
func (c *Controller) UpdateVehicule(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request UpdateVehiculeRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	vehicule, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "vehicule not found" {
			return responses.NotFound(ctx, "Vehicule not found")
		}
		return responses.InternalServerError(ctx, "Failed to update vehicule")
	}

	return responses.Success(ctx, vehicule)
}

// DeleteVehicule deletes a vehicule
func (c *Controller) DeleteVehicule(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "vehicule not found" {
			return responses.NotFound(ctx, "Vehicule not found")
		}
		return responses.InternalServerError(ctx, "Failed to delete vehicule")
	}

	return responses.Success(ctx, nil)
}

// GetByMarque gets vehicules by marque
func (c *Controller) GetByMarque(ctx echo.Context) error {
	marque := ctx.Param("marque")
	if marque == "" {
		return responses.BadRequest(ctx, "Marque is required")
	}

	result, err := c.service.GetByMarque(ctx.Request().Context(), marque)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get vehicules")
	}

	return responses.Success(ctx, result)
}

// GetByType gets vehicules by type
func (c *Controller) GetByType(ctx echo.Context) error {
	typeVehicule := ctx.Param("type")
	if typeVehicule == "" {
		return responses.BadRequest(ctx, "Type is required")
	}

	result, err := c.service.GetByType(ctx.Request().Context(), typeVehicule)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get vehicules")
	}

	return responses.Success(ctx, result)
}
