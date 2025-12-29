package inspection

import (
	"net/http"
	"time"

	"police-trafic-api-frontend-aligned/internal/modules/verification"

	"github.com/labstack/echo/v4"
)

// Controller handles HTTP requests for inspections
type Controller struct {
	service             Service
	verificationService verification.Service
}

// NewController creates a new inspection controller
func NewController(service Service, verificationService verification.Service) *Controller {
	return &Controller{
		service:             service,
		verificationService: verificationService,
	}
}

// RegisterRoutes registers inspection routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	inspections := g.Group("/inspections")

	// List all inspections
	inspections.GET("", c.List)
	inspections.POST("", c.Create)

	// Fixed path endpoints (must be before /:id)
	inspections.GET("/statistics", c.GetStatistics)
	inspections.GET("/numero/:numero", c.GetByNumero)
	inspections.GET("/vehicule/:vehicule_id", c.GetByVehicule)

	// Nested routes under /:id (must be before single /:id GET)
	inspections.GET("/:id/verifications", c.GetVerifications)
	inspections.POST("/:id/verifications", c.SaveVerifications)
	inspections.PATCH("/:id/statut", c.ChangerStatut)

	// Single inspection by ID (should be last among GET routes with :id)
	inspections.GET("/:id", c.GetByID)

	// Update/Delete
	inspections.PUT("/:id", c.Update)
	inspections.DELETE("/:id", c.Delete)
}

// Create creates a new inspection
func (c *Controller) Create(ctx echo.Context) error {
	var req CreateInspectionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	inspection, err := c.service.Create(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, inspection)
}

// GetByID retrieves an inspection by ID
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	inspection, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, inspection)
}

// GetByNumero retrieves an inspection by numero
func (c *Controller) GetByNumero(ctx echo.Context) error {
	numero := ctx.Param("numero")
	if numero == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "numero is required"})
	}

	inspection, err := c.service.GetByNumero(ctx.Request().Context(), numero)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, inspection)
}

// GetByVehicule retrieves inspections for a vehicule
func (c *Controller) GetByVehicule(ctx echo.Context) error {
	vehiculeID := ctx.Param("vehicule_id")
	if vehiculeID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "vehicule_id is required"})
	}

	inspections, err := c.service.GetByVehicule(ctx.Request().Context(), vehiculeID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"inspections": inspections,
		"total":       len(inspections),
	})
}

// List lists inspections with filters
func (c *Controller) List(ctx echo.Context) error {
	var req ListInspectionsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	// Parse query params
	if statut := ctx.QueryParam("statut"); statut != "" {
		req.Statut = &statut
	}
	if search := ctx.QueryParam("search"); search != "" {
		req.Search = &search
	}
	if dateDebut := ctx.QueryParam("dateDebut"); dateDebut != "" {
		if t, err := time.Parse("2006-01-02", dateDebut); err == nil {
			req.DateDebut = &t
		}
	}
	if dateFin := ctx.QueryParam("dateFin"); dateFin != "" {
		if t, err := time.Parse("2006-01-02", dateFin); err == nil {
			// Add 23:59:59 to include the entire day
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			req.DateFin = &endOfDay
		}
	}
	if inspecteurID := ctx.QueryParam("inspecteur_id"); inspecteurID != "" {
		req.InspecteurID = &inspecteurID
	}

	result, err := c.service.List(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, result)
}

// Update updates an inspection
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req UpdateInspectionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	inspection, err := c.service.Update(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, inspection)
}

// Delete deletes an inspection
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ChangerStatut changes the status of an inspection
func (c *Controller) ChangerStatut(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req ChangerStatutRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	inspection, err := c.service.ChangerStatut(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, inspection)
}

// GetStatistics returns statistics about inspections
func (c *Controller) GetStatistics(ctx echo.Context) error {
	// Parse date filters
	var dateDebut, dateFin *time.Time
	if d := ctx.QueryParam("dateDebut"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			dateDebut = &t
		}
	}
	if d := ctx.QueryParam("dateFin"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateFin = &endOfDay
		}
	}

	stats, err := c.service.GetStatisticsWithFilters(ctx.Request().Context(), dateDebut, dateFin)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, stats)
}

// GetVerifications returns all verifications for an inspection
func (c *Controller) GetVerifications(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	// Verify the inspection exists
	_, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := c.verificationService.GetVerifications(ctx.Request().Context(), "INSPECTION", id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

// SaveVerifications saves verifications for an inspection
func (c *Controller) SaveVerifications(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	// Verify the inspection exists
	_, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "inspection not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	var request verification.BatchCheckOptionsRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if len(request.Verifications) == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "at least one verification is required"})
	}

	result, err := c.verificationService.SaveBatchVerifications(ctx.Request().Context(), "INSPECTION", id, &request)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{"data": result})
}
