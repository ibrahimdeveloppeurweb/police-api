package infraction

import (
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles infraction routes
type Controller struct {
	service Service
}

// NewController creates a new infraction controller
func NewController(service Service) interfaces.Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers infraction routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/infractions")

	// Public endpoints
	group.GET("", c.ListInfractions)
	group.GET("/dashboard", c.GetDashboard)
	group.GET("/:id", c.GetInfraction)
	group.GET("/types", c.GetTypesInfractions)
	group.GET("/categories", c.GetCategories)
	group.GET("/stats", c.GetStatistics)
	
	// Protected endpoints
	group.POST("", c.CreateInfraction)
	group.PUT("/:id", c.UpdateInfraction)
	group.DELETE("/:id", c.DeleteInfraction)
	
	// Additional endpoints
	group.GET("/pv/:numeroPv", c.GetByNumeroPV)
	group.GET("/controle/:controleId", c.GetByControle)
	group.GET("/vehicule/:vehiculeId", c.GetByVehicule)
	group.GET("/conducteur/:conducteurId", c.GetByConducteur)
	group.GET("/statut/:statut", c.GetByStatut)
	group.POST("/:id/generate-pv", c.GeneratePV)
	group.POST("/:id/validate", c.ValidateInfraction)
	group.POST("/:id/archive", c.ArchiveInfraction)
	group.POST("/:id/unarchive", c.UnarchiveInfraction)
	group.POST("/:id/payment", c.RecordPayment)
	group.GET("/group-by-type", c.GroupByType)
}

// ListInfractions lists infractions with filters
func (c *Controller) ListInfractions(ctx echo.Context) error {
	request := &ListInfractionsRequest{}
	
	// Parse filters
	if controleID := ctx.QueryParam("controle_id"); controleID != "" {
		request.ControleID = &controleID
	}
	
	if vehiculeID := ctx.QueryParam("vehicule_id"); vehiculeID != "" {
		request.VehiculeID = &vehiculeID
	}
	
	if conducteurID := ctx.QueryParam("conducteur_id"); conducteurID != "" {
		request.ConducteurID = &conducteurID
	}
	
	if typeInfractionID := ctx.QueryParam("type_infraction_id"); typeInfractionID != "" {
		request.TypeInfractionID = &typeInfractionID
	}
	
	if statut := ctx.QueryParam("statut"); statut != "" {
		request.Statut = &statut
	}
	
	if lieu := ctx.QueryParam("lieu_infraction"); lieu != "" {
		request.LieuInfraction = &lieu
	}
	
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
	
	// Boolean filters
	if flagrant := ctx.QueryParam("flagrant_delit"); flagrant != "" {
		if b, err := strconv.ParseBool(flagrant); err == nil {
			request.FlagrantDelit = &b
		}
	}
	
	if accident := ctx.QueryParam("accident"); accident != "" {
		if b, err := strconv.ParseBool(accident); err == nil {
			request.Accident = &b
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
		return responses.InternalServerError(ctx, "Failed to list infractions")
	}
	
	return responses.Success(ctx, result)
}

// GetInfraction gets an infraction by ID
func (c *Controller) GetInfraction(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}
	
	infraction, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to get infraction")
	}
	
	return responses.Success(ctx, infraction)
}

// CreateInfraction creates a new infraction
func (c *Controller) CreateInfraction(ctx echo.Context) error {
	var request CreateInfractionRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}
	
	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}
	
	infraction, err := c.service.Create(ctx.Request().Context(), &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to create infraction")
	}
	
	return responses.Created(ctx, infraction)
}

// UpdateInfraction updates an infraction
func (c *Controller) UpdateInfraction(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}
	
	var request UpdateInfractionRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}
	
	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}
	
	infraction, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to update infraction")
	}
	
	return responses.Success(ctx, infraction)
}

// DeleteInfraction deletes an infraction
func (c *Controller) DeleteInfraction(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}
	
	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to delete infraction")
	}
	
	return responses.Success(ctx, nil)
}

// GetTypesInfractions gets all infraction types
func (c *Controller) GetTypesInfractions(ctx echo.Context) error {
	types, err := c.service.GetTypesInfractions(ctx.Request().Context())
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get infraction types")
	}

	return responses.Success(ctx, types)
}

// GetCategories gets all infraction categories
func (c *Controller) GetCategories(ctx echo.Context) error {
	categories, err := c.service.GetCategories(ctx.Request().Context())
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get infraction categories")
	}

	return responses.Success(ctx, categories)
}

// GetStatistics gets infraction statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	request := &ListInfractionsRequest{}
	
	// Parse date filters for statistics
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
	
	stats, err := c.service.GetStatistics(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get statistics")
	}
	
	return responses.Success(ctx, stats)
}

// GetByNumeroPV gets infraction by numero PV
func (c *Controller) GetByNumeroPV(ctx echo.Context) error {
	numeroPV := ctx.Param("numeroPv")
	if numeroPV == "" {
		return responses.BadRequest(ctx, "Numero PV is required")
	}
	
	infraction, err := c.service.GetByNumeroPV(ctx.Request().Context(), numeroPV)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to get infraction")
	}
	
	return responses.Success(ctx, infraction)
}

// GetByControle gets infractions by controle
func (c *Controller) GetByControle(ctx echo.Context) error {
	controleID := ctx.Param("controleId")
	if controleID == "" {
		return responses.BadRequest(ctx, "Controle ID is required")
	}
	
	result, err := c.service.GetByControle(ctx.Request().Context(), controleID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get infractions")
	}
	
	return responses.Success(ctx, result)
}

// GetByVehicule gets infractions by vehicule
func (c *Controller) GetByVehicule(ctx echo.Context) error {
	vehiculeID := ctx.Param("vehiculeId")
	if vehiculeID == "" {
		return responses.BadRequest(ctx, "Vehicule ID is required")
	}
	
	result, err := c.service.GetByVehicule(ctx.Request().Context(), vehiculeID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get infractions")
	}
	
	return responses.Success(ctx, result)
}

// GetByConducteur gets infractions by conducteur
func (c *Controller) GetByConducteur(ctx echo.Context) error {
	conducteurID := ctx.Param("conducteurId")
	if conducteurID == "" {
		return responses.BadRequest(ctx, "Conducteur ID is required")
	}
	
	result, err := c.service.GetByConducteur(ctx.Request().Context(), conducteurID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get infractions")
	}
	
	return responses.Success(ctx, result)
}

// GetByStatut gets infractions by statut
func (c *Controller) GetByStatut(ctx echo.Context) error {
	statut := ctx.Param("statut")
	if statut == "" {
		return responses.BadRequest(ctx, "Statut is required")
	}
	
	result, err := c.service.GetByStatut(ctx.Request().Context(), statut)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get infractions")
	}
	
	return responses.Success(ctx, result)
}

// GeneratePV generates a PV for an infraction
func (c *Controller) GeneratePV(ctx echo.Context) error {
	infractionID := ctx.Param("id")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}
	
	result, err := c.service.GeneratePV(ctx.Request().Context(), infractionID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to generate PV")
	}
	
	return responses.Success(ctx, result)
}

// ValidateInfraction validates an infraction
func (c *Controller) ValidateInfraction(ctx echo.Context) error {
	infractionID := ctx.Param("id")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}
	
	result, err := c.service.ValidateInfraction(ctx.Request().Context(), infractionID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to validate infraction")
	}
	
	return responses.Success(ctx, result)
}

// GroupByType groups infractions by type
func (c *Controller) GroupByType(ctx echo.Context) error {
	request := &ListInfractionsRequest{}

	// Parse date filters
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

	result, err := c.service.GroupByType(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to group infractions")
	}

	return responses.Success(ctx, result)
}

// ArchiveInfraction archives an infraction (changes status to ARCHIVEE)
func (c *Controller) ArchiveInfraction(ctx echo.Context) error {
	infractionID := ctx.Param("id")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}

	result, err := c.service.ArchiveInfraction(ctx.Request().Context(), infractionID)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to archive infraction")
	}

	return responses.Success(ctx, result)
}

// UnarchiveInfraction unarchives an infraction (changes status back to PAYEE)
func (c *Controller) UnarchiveInfraction(ctx echo.Context) error {
	infractionID := ctx.Param("id")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}

	result, err := c.service.UnarchiveInfraction(ctx.Request().Context(), infractionID)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to unarchive infraction")
	}

	return responses.Success(ctx, result)
}

// RecordPayment records a payment for an infraction
func (c *Controller) RecordPayment(ctx echo.Context) error {
	infractionID := ctx.Param("id")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}

	var request PaymentRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if request.ModePaiement == "" {
		return responses.BadRequest(ctx, "Mode de paiement requis")
	}

	if request.Montant <= 0 {
		return responses.BadRequest(ctx, "Montant invalide")
	}

	result, err := c.service.RecordPayment(ctx.Request().Context(), infractionID, &request)
	if err != nil {
		if err.Error() == "infraction not found" {
			return responses.NotFound(ctx, "Infraction not found")
		}
		return responses.InternalServerError(ctx, "Failed to record payment")
	}

	return responses.Success(ctx, result)
}

// GetDashboard returns dashboard data for infractions
func (c *Controller) GetDashboard(ctx echo.Context) error {
	request := &DashboardRequest{}

	// Parse periode parameter
	if periode := ctx.QueryParam("periode"); periode != "" {
		request.Periode = periode
	} else {
		request.Periode = "mois" // default
	}

	// Parse date filters
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

	result, err := c.service.GetDashboard(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get dashboard data")
	}

	return responses.Success(ctx, result)
}