package paiement

import (
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles paiement routes
type Controller struct {
	service Service
}

// NewPaiementController creates a new paiement controller
func NewPaiementController(service Service) interfaces.Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers paiement routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/paiements")

	// CRUD endpoints
	group.GET("", c.ListPaiements)
	group.GET("/:id", c.GetPaiement)
	group.POST("", c.CreatePaiement)
	group.PUT("/:id", c.UpdatePaiement)
	group.DELETE("/:id", c.DeletePaiement)

	// Transaction endpoint
	group.GET("/transaction/:numero", c.GetByTransaction)

	// Process verbal payments
	group.GET("/pv/:pvId", c.GetByProcesVerbal)

	// Actions
	group.POST("/:id/validate", c.ValidatePaiement)
	group.POST("/:id/refuse", c.RefusePaiement)
	group.POST("/:id/refund", c.RemboursementPaiement)

	// Reçu Trésor Public
	group.POST("/:id/recu-tresor", c.GenerateRecuTresor)
	group.GET("/:id/recu-tresor", c.GetRecuTresor)

	// Statistics
	group.GET("/statistics", c.GetStatistics)
}

// ListPaiements lists paiements with filters
func (c *Controller) ListPaiements(ctx echo.Context) error {
	request := &ListPaiementsRequest{}

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
	if pvID := ctx.QueryParam("proces_verbal_id"); pvID != "" {
		request.ProcesVerbalID = &pvID
	}

	if statut := ctx.QueryParam("statut"); statut != "" {
		request.Statut = &statut
	}

	if moyenPaiement := ctx.QueryParam("moyen_paiement"); moyenPaiement != "" {
		request.MoyenPaiement = &moyenPaiement
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
		return responses.InternalServerError(ctx, "Failed to list paiements: "+err.Error())
	}

	return responses.Success(ctx, result)
}

// GetPaiement gets a paiement by ID
func (c *Controller) GetPaiement(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	paiement, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		return responses.InternalServerError(ctx, "Failed to get paiement")
	}

	return responses.Success(ctx, paiement)
}

// GetByTransaction gets a paiement by transaction number
func (c *Controller) GetByTransaction(ctx echo.Context) error {
	numero := ctx.Param("numero")
	if numero == "" {
		return responses.BadRequest(ctx, "Transaction number is required")
	}

	paiement, err := c.service.GetByNumeroTransaction(ctx.Request().Context(), numero)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		return responses.InternalServerError(ctx, "Failed to get paiement")
	}

	return responses.Success(ctx, paiement)
}

// CreatePaiement creates a new paiement
func (c *Controller) CreatePaiement(ctx echo.Context) error {
	var request CreatePaiementRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	paiement, err := c.service.Create(ctx.Request().Context(), &request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to create paiement: "+err.Error())
	}

	return responses.Created(ctx, paiement)
}

// UpdatePaiement updates a paiement
func (c *Controller) UpdatePaiement(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request UpdatePaiementRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed")
	}

	paiement, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		return responses.InternalServerError(ctx, "Failed to update paiement")
	}

	return responses.Success(ctx, paiement)
}

// DeletePaiement deletes a paiement
func (c *Controller) DeletePaiement(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		if err.Error() == "cannot delete a validated payment" {
			return responses.BadRequest(ctx, "Cannot delete a validated payment")
		}
		return responses.InternalServerError(ctx, "Failed to delete paiement")
	}

	return responses.Success(ctx, nil)
}

// GetByProcesVerbal gets paiements by proces verbal ID
func (c *Controller) GetByProcesVerbal(ctx echo.Context) error {
	pvID := ctx.Param("pvId")
	if pvID == "" {
		return responses.BadRequest(ctx, "Proces Verbal ID is required")
	}

	result, err := c.service.GetByProcesVerbal(ctx.Request().Context(), pvID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get paiements")
	}

	return responses.Success(ctx, result)
}

// ValidatePaiement validates a paiement
func (c *Controller) ValidatePaiement(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request ValidatePaiementRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: code_autorisation is required")
	}

	paiement, err := c.service.Validate(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, paiement)
}

// RefusePaiement refuses a paiement
func (c *Controller) RefusePaiement(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request RefusePaiementRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: motif_refus is required")
	}

	paiement, err := c.service.Refuse(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, paiement)
}

// RemboursementPaiement processes a refund
func (c *Controller) RemboursementPaiement(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request RemboursementRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: motif is required")
	}

	paiement, err := c.service.Rembourser(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		return responses.BadRequest(ctx, err.Error())
	}

	return responses.Success(ctx, paiement)
}

// GetStatistics gets paiement statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	request := &ListPaiementsRequest{}

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

	if pvID := ctx.QueryParam("proces_verbal_id"); pvID != "" {
		request.ProcesVerbalID = &pvID
	}

	stats, err := c.service.GetStatistics(ctx.Request().Context(), request)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get statistics")
	}

	return responses.Success(ctx, stats)
}

// GenerateRecuTresor generates a treasury receipt for a payment
func (c *Controller) GenerateRecuTresor(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request RecuTresorRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}
	request.PaiementID = id

	if err := ctx.Validate(request); err != nil {
		return responses.BadRequest(ctx, "Validation failed: agent_tresor and bureau_tresor are required")
	}

	recu, err := c.service.GenerateRecuTresor(ctx.Request().Context(), &request)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		if err.Error() == "payment must be TRESOR_PUBLIC type" {
			return responses.BadRequest(ctx, "Ce paiement n'est pas de type Trésor Public")
		}
		return responses.InternalServerError(ctx, "Failed to generate treasury receipt: "+err.Error())
	}

	return responses.Created(ctx, recu)
}

// GetRecuTresor retrieves an existing treasury receipt for a payment
func (c *Controller) GetRecuTresor(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	recu, err := c.service.GetRecuTresor(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "paiement not found" {
			return responses.NotFound(ctx, "Paiement not found")
		}
		if err.Error() == "recu tresor not found" {
			return responses.NotFound(ctx, "Reçu trésor non trouvé pour ce paiement")
		}
		return responses.InternalServerError(ctx, "Failed to get treasury receipt")
	}

	return responses.Success(ctx, recu)
}
