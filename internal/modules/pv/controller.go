package pv

import (
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles PV routes
type Controller struct {
	service Service
}

// NewPVController creates a new PV controller
func NewPVController(service Service) interfaces.Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers PV routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/pv")

	// CRUD endpoints
	group.GET("", c.ListPVs)
	group.GET("/:id", c.GetPV)
	group.POST("", c.CreatePV)
	group.PUT("/:id", c.UpdatePV)
	group.DELETE("/:id", c.DeletePV)

	// By numero
	group.GET("/numero/:numero", c.GetByNumeroPV)

	// By infraction
	group.GET("/infraction/:infractionId", c.GetByInfraction)

	// Actions
	group.POST("/:id/payer", c.PayerPV)
	group.POST("/:id/contester", c.ContesterPV)
	group.POST("/:id/decision", c.DeciderContestation)
	group.POST("/:id/majorer", c.MajorerPV)
	group.POST("/:id/annuler", c.AnnulerPV)

	// Special queries
	group.GET("/expired", c.GetExpiredPVs)
	group.GET("/statistics", c.GetStatistics)

	// Rappels et retards
	group.POST("/:id/envoyer-rappel", c.EnvoyerRappel)
	group.PATCH("/:id/marquer-en-retard", c.MarquerEnRetard)
}

// ListPVs lists PVs with filters
func (c *Controller) ListPVs(ctx echo.Context) error {
	request := &ListPVRequest{}

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
	if infractionID := ctx.QueryParam("infraction_id"); infractionID != "" {
		request.InfractionID = &infractionID
	}

	if statut := ctx.QueryParam("statut"); statut != "" {
		request.Statut = &statut
	}

	// Amount filters
	if montantMin := ctx.QueryParam("montant_min"); montantMin != "" {
		if m, err := strconv.ParseFloat(montantMin, 64); err == nil {
			request.MontantMin = &m
		}
	}

	if montantMax := ctx.QueryParam("montant_max"); montantMax != "" {
		if m, err := strconv.ParseFloat(montantMax, 64); err == nil {
			request.MontantMax = &m
		}
	}

	if expiredStr := ctx.QueryParam("expired"); expiredStr != "" {
		expired := expiredStr == "true"
		request.Expired = &expired
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
		return responses.InternalServerError(ctx, "Failed to list PVs: "+err.Error())
	}

	return responses.Success(ctx, result)
}

// GetPV gets a PV by ID
func (c *Controller) GetPV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	pv, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.InternalServerError(ctx, "Failed to get PV")
	}

	return responses.Success(ctx, pv)
}

// GetByNumeroPV gets a PV by numero
func (c *Controller) GetByNumeroPV(ctx echo.Context) error {
	numero := ctx.Param("numero")
	if numero == "" {
		return responses.BadRequest(ctx, "Numero is required")
	}

	pv, err := c.service.GetByNumeroPV(ctx.Request().Context(), numero)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.InternalServerError(ctx, "Failed to get PV")
	}

	return responses.Success(ctx, pv)
}

// GetByInfraction gets a PV by infraction ID
func (c *Controller) GetByInfraction(ctx echo.Context) error {
	infractionID := ctx.Param("infractionId")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}

	pv, err := c.service.GetByInfraction(ctx.Request().Context(), infractionID)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.InternalServerError(ctx, "Failed to get PV")
	}

	return responses.Success(ctx, pv)
}

// CreatePV creates a new PV
func (c *Controller) CreatePV(ctx echo.Context) error {
	var request CreatePVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	pv, err := c.service.Create(ctx.Request().Context(), &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to create PV: "+err.Error())
	}

	return responses.Created(ctx, pv)
}

// UpdatePV updates a PV
func (c *Controller) UpdatePV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request UpdatePVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	pv, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.InternalServerError(ctx, "Failed to update PV")
	}

	return responses.Success(ctx, pv)
}

// DeletePV deletes a PV
func (c *Controller) DeletePV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		if err.Error() == "cannot delete a paid PV" {
			return responses.BadRequest(ctx, "Cannot delete a paid PV")
		}
		return responses.InternalServerError(ctx, "Failed to delete PV")
	}

	return responses.Success(ctx, nil)
}

// PayerPV records a payment on a PV
func (c *Controller) PayerPV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request PayerPVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	pv, err := c.service.Payer(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, pv)
}

// ContesterPV records a contestation on a PV
func (c *Controller) ContesterPV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request ContesterPVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: motif_contestation is required")
	}

	pv, err := c.service.Contester(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, pv)
}

// DeciderContestation records a decision on a PV contestation
func (c *Controller) DeciderContestation(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request DecisionContestationRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	pv, err := c.service.DeciderContestation(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, pv)
}

// MajorerPV adds a penalty to a PV
func (c *Controller) MajorerPV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request MajorerPVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	pv, err := c.service.Majorer(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, pv)
}

// AnnulerPV cancels a PV
func (c *Controller) AnnulerPV(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request AnnulerPVRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: motif is required")
	}

	pv, err := c.service.Annuler(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, pv)
}

// GetExpiredPVs gets expired PVs
func (c *Controller) GetExpiredPVs(ctx echo.Context) error {
	result, err := c.service.GetExpired(ctx.Request().Context())
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get expired PVs")
	}

	return responses.Success(ctx, result)
}

// GetStatistics gets PV statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	request := &ListPVRequest{}

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

	stats, err := c.service.GetStatistics(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get statistics")
	}

	return responses.Success(ctx, stats)
}

// EnvoyerRappel sends a payment reminder for a PV
func (c *Controller) EnvoyerRappel(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	result, err := c.service.EnvoyerRappel(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.InternalServerError(ctx, "Failed to send reminder")
	}

	return responses.Success(ctx, result)
}

// MarquerEnRetard marks a PV as late
func (c *Controller) MarquerEnRetard(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	pv, err := c.service.MarquerEnRetard(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "pv not found" {
			return responses.NotFound(ctx, "PV not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, pv)
}
