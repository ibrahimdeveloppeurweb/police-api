package controle

import (
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/modules/verification"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles controle routes
type Controller struct {
	service             Service
	verificationService verification.Service
}

// NewController creates a new controle controller
func NewController(service Service, verificationService verification.Service) interfaces.Controller {
	return &Controller{
		service:             service,
		verificationService: verificationService,
	}
}

// RegisterRoutes registers controle routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/controles")

	// List all controles
	group.GET("", c.ListControles)

	// Additional endpoints with fixed paths (must be before /:id)
	group.GET("/agent/:agentId", c.GetControlesByAgent)
	group.GET("/vehicule/:vehiculeId", c.GetControlesByVehicule)
	group.GET("/conducteur/:conducteurId", c.GetControlesByConducteur)
	group.GET("/statistics", c.GetStatistics)

	// Nested routes under /:id (must be before single /:id GET)
	group.GET("/:id/verifications", c.GetVerifications)
	group.POST("/:id/verifications", c.SaveVerifications)
	group.POST("/:id/pv", c.GeneratePV)
	group.PATCH("/:id/statut", c.ChangerStatut)
	group.POST("/:id/archive", c.Archive)
	group.POST("/:id/unarchive", c.Unarchive)

	// Single controle by ID (should be last among GET routes with :id)
	group.GET("/:id", c.GetControle)

	// Create/Update/Delete
	group.POST("", c.CreateControle)
	group.PUT("/:id", c.UpdateControle)
	group.DELETE("/:id", c.DeleteControle)
}

// ListControles lists controles with filters
func (c *Controller) ListControles(ctx echo.Context) error {
	// Parse query parameters
	request := &ListControlesRequest{}
	
	// Date filters
	if dateDebut := ctx.QueryParam("date_debut"); dateDebut != "" {
		if t, err := time.Parse("2006-01-02", dateDebut); err == nil {
			request.DateDebut = &t
		}
	}
	
	if dateFin := ctx.QueryParam("date_fin"); dateFin != "" {
		if t, err := time.Parse("2006-01-02", dateFin); err == nil {
			request.DateFin = &t
		}
	}
	
	// Other filters
	if agentID := ctx.QueryParam("agent_id"); agentID != "" {
		request.AgentID = &agentID
	}
	
	if vehiculeID := ctx.QueryParam("vehicule_id"); vehiculeID != "" {
		request.VehiculeID = &vehiculeID
	}
	
	if conducteurID := ctx.QueryParam("conducteur_id"); conducteurID != "" {
		request.ConducteurID = &conducteurID
	}
	
	if typeControle := ctx.QueryParam("type_controle"); typeControle != "" {
		request.TypeControle = &typeControle
	}
	
	if statut := ctx.QueryParam("statut"); statut != "" {
		request.Statut = &statut
	}

	// Archive filter
	if isArchived := ctx.QueryParam("is_archived"); isArchived != "" {
		archived := isArchived == "true"
		request.IsArchived = &archived
	}

	// Pagination
	if limit := ctx.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			request.Limit = l
		}
	} else {
		request.Limit = 50 // Default limit
	}
	
	if offset := ctx.QueryParam("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			request.Offset = o
		}
	}
	
	result, err := c.service.List(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to list controles: " + err.Error())
	}
	
	return responses.Success(ctx, result)
}

// GetControle gets a controle by ID
func (c *Controller) GetControle(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}
	
	controle, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to get controle")
	}
	
	return responses.Success(ctx, controle)
}

// CreateControle creates a new controle
func (c *Controller) CreateControle(ctx echo.Context) error {
	var request CreateControleRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}
	
	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}
	
	controle, err := c.service.Create(ctx.Request().Context(), &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to create controle")
	}
	
	return responses.Created(ctx, controle)
}

// UpdateControle updates a controle
func (c *Controller) UpdateControle(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}
	
	var request UpdateControleRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}
	
	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}
	
	controle, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to update controle")
	}
	
	return responses.Success(ctx, controle)
}

// DeleteControle deletes a controle
func (c *Controller) DeleteControle(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}
	
	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to delete controle")
	}
	
	return responses.Success(ctx, nil)
}

// GetControlesByAgent gets controles by agent
func (c *Controller) GetControlesByAgent(ctx echo.Context) error {
	agentID := ctx.Param("agentId")
	if agentID == "" {
		return responses.BadRequest(ctx, "Agent ID is required")
	}

	result, err := c.service.GetByAgent(ctx.Request().Context(), agentID, nil)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get controles")
	}

	return responses.Success(ctx, result)
}

// GetControlesByVehicule gets controles by vehicule
func (c *Controller) GetControlesByVehicule(ctx echo.Context) error {
	vehiculeID := ctx.Param("vehiculeId")
	if vehiculeID == "" {
		return responses.BadRequest(ctx, "Vehicule ID is required")
	}
	
	result, err := c.service.GetByVehicule(ctx.Request().Context(), vehiculeID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get controles")
	}
	
	return responses.Success(ctx, result)
}

// GetControlesByConducteur gets controles by conducteur
func (c *Controller) GetControlesByConducteur(ctx echo.Context) error {
	conducteurID := ctx.Param("conducteurId")
	if conducteurID == "" {
		return responses.BadRequest(ctx, "Conducteur ID is required")
	}
	
	result, err := c.service.GetByConducteur(ctx.Request().Context(), conducteurID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get controles")
	}
	
	return responses.Success(ctx, result)
}

// GetStatistics gets controle statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	filters := &StatisticsFilters{}

	if id := ctx.QueryParam("agent_id"); id != "" {
		filters.AgentID = &id
	}

	if dateDebut := ctx.QueryParam("date_debut"); dateDebut != "" {
		if t, err := time.Parse("2006-01-02", dateDebut); err == nil {
			filters.DateDebut = &t
		}
	}

	if dateFin := ctx.QueryParam("date_fin"); dateFin != "" {
		if t, err := time.Parse("2006-01-02", dateFin); err == nil {
			filters.DateFin = &t
		}
	}

	stats, err := c.service.GetStatistics(ctx.Request().Context(), filters)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get statistics")
	}

	return responses.Success(ctx, stats)
}


// GeneratePV generates a PV from a controle with specified infractions
func (c *Controller) GeneratePV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "Controle ID is required")
	}

	var request GeneratePVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if len(request.Infractions) == 0 {
		return responses.BadRequest(ctx, "At least one infraction is required")
	}

	pv, err := c.service.GeneratePV(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		if err.Error() == "no valid infractions to generate PV" {
			return responses.BadRequest(ctx, "No valid infractions to generate PV")
		}
		return responses.InternalServerError(ctx, "Failed to generate PV: "+err.Error())
	}

	return responses.Created(ctx, pv)
}

// ChangerStatut handles PATCH /controles/:id/statut
func (c *Controller) ChangerStatut(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request ChangerStatutRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: statut is required")
	}

	controle, err := c.service.ChangerStatut(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, controle)
}

// GetVerifications returns all verifications for a controle
func (c *Controller) GetVerifications(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "Controle ID is required")
	}

	// Verify the controle exists
	_, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to get controle")
	}

	result, err := c.verificationService.GetVerifications(ctx.Request().Context(), "CONTROL", id)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get verifications")
	}

	return responses.Success(ctx, result)
}

// SaveVerifications saves verifications for a controle
func (c *Controller) SaveVerifications(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "Controle ID is required")
	}

	// Verify the controle exists
	_, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to get controle")
	}

	var request verification.BatchCheckOptionsRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if len(request.Verifications) == 0 {
		return responses.BadRequest(ctx, "At least one verification is required")
	}

	result, err := c.verificationService.SaveBatchVerifications(ctx.Request().Context(), "CONTROL", id, &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to save verifications: "+err.Error())
	}

	return responses.Success(ctx, result)
}

// Archive archives a controle
func (c *Controller) Archive(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	controle, err := c.service.Archive(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to archive controle: "+err.Error())
	}

	return responses.Success(ctx, controle)
}

// Unarchive unarchives a controle
func (c *Controller) Unarchive(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	controle, err := c.service.Unarchive(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "controle not found" {
			return responses.NotFound(ctx, "Controle not found")
		}
		return responses.InternalServerError(ctx, "Failed to unarchive controle: "+err.Error())
	}

	return responses.Success(ctx, controle)
}