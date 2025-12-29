package recours

import (
	"net/http"
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles recours HTTP requests
type Controller struct {
	service Service
}

// NewRecoursController creates a new recours controller
func NewRecoursController(service Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers recours routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	recours := g.Group("/recours")

	recours.POST("", c.Create)
	recours.GET("", c.List)
	recours.GET("/en-cours", c.GetEnCours)
	recours.GET("/statistics", c.GetStatistics)
	recours.GET("/:id", c.GetByID)
	recours.GET("/numero/:numero", c.GetByNumero)
	recours.GET("/pv/:pvId", c.GetByProcesVerbal)
	recours.PUT("/:id", c.Update)
	recours.DELETE("/:id", c.Delete)
	recours.POST("/:id/traiter", c.Traiter)
	recours.POST("/:id/assigner", c.Assigner)
	recours.POST("/:id/abandonner", c.Abandonner)
	recours.GET("/:id/etapes", c.GetEtapes)
}

// Create handles POST /recours
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateRecoursRequest
	if err := ctx.Bind(&req); err != nil {
		return responses.BadRequest(ctx, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return responses.BadRequest(ctx, err.Error())
	}

	result, err := c.service.Create(ctx.Request().Context(), &req)
	if err != nil {
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Created(ctx, result)
}

// GetByID handles GET /recours/:id
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	result, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// GetByNumero handles GET /recours/numero/:numero
func (c *Controller) GetByNumero(ctx echo.Context) error {
	numero := ctx.Param("numero")
	if numero == "" {
		return responses.BadRequest(ctx, "Numero is required")
	}

	result, err := c.service.GetByNumeroRecours(ctx.Request().Context(), numero)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// GetByProcesVerbal handles GET /recours/pv/:pvId
func (c *Controller) GetByProcesVerbal(ctx echo.Context) error {
	pvID := ctx.Param("pvId")
	if pvID == "" {
		return responses.BadRequest(ctx, "PV ID is required")
	}

	result, err := c.service.GetByProcesVerbal(ctx.Request().Context(), pvID)
	if err != nil {
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// List handles GET /recours
func (c *Controller) List(ctx echo.Context) error {
	var filters ListRecoursRequest

	// Parse query parameters
	filters.ProcesVerbalID = getStringPtr(ctx.QueryParam("proces_verbal_id"))
	filters.TypeRecours = getStringPtr(ctx.QueryParam("type_recours"))
	filters.Statut = getStringPtr(ctx.QueryParam("statut"))
	filters.TraiteParID = getStringPtr(ctx.QueryParam("traite_par_id"))

	// Parse dates
	if dateDebut := ctx.QueryParam("date_debut"); dateDebut != "" {
		if t, err := parseTime(dateDebut); err == nil {
			filters.DateDebut = &t
		}
	}
	if dateFin := ctx.QueryParam("date_fin"); dateFin != "" {
		if t, err := parseTime(dateFin); err == nil {
			filters.DateFin = &t
		}
	}

	// Parse pagination
	filters.Limit = getIntParam(ctx, "limit", 20)
	filters.Offset = getIntParam(ctx, "offset", 0)

	result, err := c.service.List(ctx.Request().Context(), &filters)
	if err != nil {
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// Update handles PUT /recours/:id
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var req UpdateRecoursRequest
	if err := ctx.Bind(&req); err != nil {
		return responses.BadRequest(ctx, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return responses.BadRequest(ctx, err.Error())
	}

	result, err := c.service.Update(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// Delete handles DELETE /recours/:id
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Traiter handles POST /recours/:id/traiter
func (c *Controller) Traiter(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var req TraiterRecoursRequest
	if err := ctx.Bind(&req); err != nil {
		return responses.BadRequest(ctx, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return responses.BadRequest(ctx, err.Error())
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return responses.BadRequest(ctx, "User ID not found in context")
	}

	result, err := c.service.Traiter(ctx.Request().Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// Assigner handles POST /recours/:id/assigner
func (c *Controller) Assigner(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var req AssignerRecoursRequest
	if err := ctx.Bind(&req); err != nil {
		return responses.BadRequest(ctx, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return responses.BadRequest(ctx, err.Error())
	}

	result, err := c.service.Assigner(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// Abandonner handles POST /recours/:id/abandonner
func (c *Controller) Abandonner(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var req AbandonnerRecoursRequest
	if err := ctx.Bind(&req); err != nil {
		return responses.BadRequest(ctx, "Invalid request body")
	}

	if err := ctx.Validate(&req); err != nil {
		return responses.BadRequest(ctx, err.Error())
	}

	result, err := c.service.Abandonner(ctx.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// GetEnCours handles GET /recours/en-cours
func (c *Controller) GetEnCours(ctx echo.Context) error {
	result, err := c.service.GetEnCours(ctx.Request().Context())
	if err != nil {
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// GetEtapes handles GET /recours/:id/etapes
func (c *Controller) GetEtapes(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	result, err := c.service.GetEtapes(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "recours not found" {
			return responses.NotFound(ctx, "Recours not found")
		}
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// GetStatistics handles GET /recours/statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	var filters ListRecoursRequest

	// Parse query parameters
	filters.ProcesVerbalID = getStringPtr(ctx.QueryParam("proces_verbal_id"))
	filters.TypeRecours = getStringPtr(ctx.QueryParam("type_recours"))
	filters.Statut = getStringPtr(ctx.QueryParam("statut"))

	// Parse dates
	if dateDebut := ctx.QueryParam("date_debut"); dateDebut != "" {
		if t, err := parseTime(dateDebut); err == nil {
			filters.DateDebut = &t
		}
	}
	if dateFin := ctx.QueryParam("date_fin"); dateFin != "" {
		if t, err := parseTime(dateFin); err == nil {
			filters.DateFin = &t
		}
	}

	result, err := c.service.GetStatistics(ctx.Request().Context(), &filters)
	if err != nil {
		return responses.InternalServerError(ctx, err.Error())
	}

	return responses.Success(ctx, result)
}

// Helper functions

func getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getIntParam(ctx echo.Context, name string, defaultVal int) int {
	val := ctx.QueryParam(name)
	if val == "" {
		return defaultVal
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}

func parseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func getUserIDFromContext(ctx echo.Context) string {
	if userID, ok := ctx.Get("user_id").(string); ok {
		return userID
	}
	return ""
}
