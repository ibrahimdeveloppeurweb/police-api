package plainte

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Controller handles HTTP requests for plaintes
type Controller struct {
	service Service
}

// NewController creates a new plainte controller
func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// RegisterRoutes registers plainte routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	plaintes := g.Group("/plaintes")
	plaintes.GET("", c.List)
	plaintes.POST("", c.Create)
	plaintes.GET("/statistics", c.GetStatistics)
	plaintes.GET("/alertes", c.GetAlertes)
	plaintes.GET("/top-agents", c.GetTopAgents)
	plaintes.GET("/:id", c.GetByID)
	plaintes.GET("/numero/:numero", c.GetByNumero)
	plaintes.GET("/:id/preuves", c.GetPreuves)
	plaintes.POST("/:id/preuves", c.AddPreuve)
	plaintes.GET("/:id/actes-enquete", c.GetActesEnquete)
	plaintes.POST("/:id/actes-enquete", c.AddActeEnquete)
	plaintes.GET("/:id/timeline", c.GetTimeline)
	plaintes.POST("/:id/timeline", c.AddTimelineEvent)
	plaintes.GET("/:id/enquetes", c.GetEnquetes)
	plaintes.POST("/:id/enquetes", c.AddEnquete)
	plaintes.GET("/:id/decisions", c.GetDecisions)
	plaintes.POST("/:id/decisions", c.AddDecision)
	plaintes.GET("/:id/historique", c.GetHistorique)
	plaintes.PUT("/:id", c.Update)
	plaintes.DELETE("/:id", c.Delete)
	plaintes.PATCH("/:id/etape", c.ChangerEtape)
	plaintes.PATCH("/:id/statut", c.ChangerStatut)
	plaintes.PATCH("/:id/assigner", c.AssignerAgent)
}

// Create creates a new plainte
func (c *Controller) Create(ctx echo.Context) error {
	var req CreatePlainteRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	plainte, err := c.service.Create(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, plainte)
}

// GetByID retrieves a plainte by ID
func (c *Controller) GetByID(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	plainte, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, plainte)
}

// GetByNumero retrieves a plainte by numero
func (c *Controller) GetByNumero(ctx echo.Context) error {
	numero := ctx.Param("numero")
	if numero == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "numero is required"})
	}

	plainte, err := c.service.GetByNumero(ctx.Request().Context(), numero)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, plainte)
}

// List lists plaintes with filters
func (c *Controller) List(ctx echo.Context) error {
	var req ListPlaintesRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	// Parse query params
	if typePlainte := ctx.QueryParam("type_plainte"); typePlainte != "" {
		req.TypePlainte = &typePlainte
	}
	if statut := ctx.QueryParam("statut"); statut != "" {
		req.Statut = &statut
	}
	if priorite := ctx.QueryParam("priorite"); priorite != "" {
		req.Priorite = &priorite
	}
	if etape := ctx.QueryParam("etape_actuelle"); etape != "" {
		req.EtapeActuelle = &etape
	}
	if search := ctx.QueryParam("search"); search != "" {
		req.Search = &search
	}

	result, err := c.service.List(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, result)
}

// Update updates a plainte
func (c *Controller) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req UpdatePlainteRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	plainte, err := c.service.Update(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, plainte)
}

// Delete deletes a plainte
func (c *Controller) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ChangerEtape changes the workflow step of a plainte
func (c *Controller) ChangerEtape(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req ChangerEtapeRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	plainte, err := c.service.ChangerEtape(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, plainte)
}

// ChangerStatut changes the status of a plainte
func (c *Controller) ChangerStatut(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req ChangerStatutRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	plainte, err := c.service.ChangerStatut(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, plainte)
}

// AssignerAgent assigns an agent to a plainte
func (c *Controller) AssignerAgent(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req AssignerAgentRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	plainte, err := c.service.AssignerAgent(ctx.Request().Context(), id, req)
	if err != nil {
		if err.Error() == "plainte not found" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, plainte)
}

// GetStatistics returns statistics about plaintes
func (c *Controller) GetStatistics(ctx echo.Context) error {
	var req StatisticsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	// Parse query params for filters
	if commissariatID := ctx.QueryParam("commissariat_id"); commissariatID != "" {
		req.CommissariatID = &commissariatID
	}

	stats, err := c.service.GetStatistics(ctx.Request().Context(), req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, stats)
}

// GetAlertes returns active alerts for plaintes
func (c *Controller) GetAlertes(ctx echo.Context) error {
	commissariatID := ctx.QueryParam("commissariat_id")

	alertes, err := c.service.GetAlertes(ctx.Request().Context(), commissariatID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, alertes)
}

// GetTopAgents returns top performing agents
func (c *Controller) GetTopAgents(ctx echo.Context) error {
	commissariatID := ctx.QueryParam("commissariat_id")

	agents, err := c.service.GetTopAgents(ctx.Request().Context(), commissariatID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, agents)
}

// GetPreuves returns preuves for a plainte
func (c *Controller) GetPreuves(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	preuves, err := c.service.GetPreuves(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, preuves)
}

// AddPreuve adds a preuve to a plainte
func (c *Controller) AddPreuve(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req AddPreuveRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	preuve, err := c.service.AddPreuve(ctx.Request().Context(), id, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, preuve)
}

// GetActesEnquete returns actes d'enquête for a plainte
func (c *Controller) GetActesEnquete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	actes, err := c.service.GetActesEnquete(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, actes)
}

// AddActeEnquete adds an acte d'enquête to a plainte
func (c *Controller) AddActeEnquete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req AddActeEnqueteRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	acte, err := c.service.AddActeEnquete(ctx.Request().Context(), id, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, acte)
}

// GetTimeline returns timeline events for a plainte
func (c *Controller) GetTimeline(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	events, err := c.service.GetTimeline(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, events)
}

// AddTimelineEvent adds a timeline event to a plainte
func (c *Controller) AddTimelineEvent(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req AddTimelineEventRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	event, err := c.service.AddTimelineEvent(ctx.Request().Context(), id, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, event)
}


// GetEnquetes returns enquêtes for a plainte
func (c *Controller) GetEnquetes(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	enquetes, err := c.service.GetEnquetes(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, enquetes)
}

// AddEnquete adds an enquête to a plainte
func (c *Controller) AddEnquete(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req AddEnqueteRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	enquete, err := c.service.AddEnquete(ctx.Request().Context(), id, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, enquete)
}

// GetDecisions returns decisions for a plainte
func (c *Controller) GetDecisions(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	decisions, err := c.service.GetDecisions(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, decisions)
}

// AddDecision adds a decision to a plainte
func (c *Controller) AddDecision(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	var req AddDecisionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	decision, err := c.service.AddDecision(ctx.Request().Context(), id, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, decision)
}

// GetHistorique returns historique for a plainte
func (c *Controller) GetHistorique(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
	}

	historique, err := c.service.GetHistorique(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, historique)
}
