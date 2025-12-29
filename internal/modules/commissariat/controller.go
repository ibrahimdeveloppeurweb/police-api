package commissariat

import (
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles commissariat HTTP requests
type Controller struct {
	service Service
}

// NewController creates a new commissariat controller
func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// RegisterRoutes registers commissariat routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	// Route pour lister tous les commissariats
	e.GET("/commissariats", ctrl.List)

	// Routes pour un commissariat spécifique
	commissariat := e.Group("/commissariat")
	commissariat.GET("/:id/dashboard", ctrl.GetDashboard)
	commissariat.GET("/:id/agents", ctrl.GetAgents)
	commissariat.GET("/:id/controles", ctrl.GetControles)
	commissariat.GET("/:id/statistiques", ctrl.GetStatistiques)
}

// List handles GET /commissariats
func (ctrl *Controller) List(c echo.Context) error {
	// Parse query params
	var actif *bool
	if a := c.QueryParam("actif"); a != "" {
		val := a == "true"
		actif = &val
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	commissariats, err := ctrl.service.List(c.Request().Context(), actif, page, limit)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, commissariats)
}

// GetDashboard handles GET /commissariat/:id/dashboard
func (ctrl *Controller) GetDashboard(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "Commissariat ID is required")
	}

	dashboard, err := ctrl.service.GetDashboard(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, dashboard)
}

// GetAgents handles GET /commissariat/:id/agents
func (ctrl *Controller) GetAgents(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "Commissariat ID is required")
	}

	agents, err := ctrl.service.GetAgents(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, agents)
}

// GetControles handles GET /commissariat/:id/controles
func (ctrl *Controller) GetControles(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "Commissariat ID is required")
	}

	// Parse pagination
	page := 1
	limit := 20
	if p := c.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	controles, err := ctrl.service.GetControles(c.Request().Context(), id, page, limit)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, controles)
}

// GetStatistiques handles GET /commissariat/:id/statistiques
func (ctrl *Controller) GetStatistiques(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "Commissariat ID is required")
	}

	// Parse dates
	var dateDebut, dateFin *time.Time
	if d := c.QueryParam("dateDebut"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			dateDebut = &t
		}
	}
	if d := c.QueryParam("dateFin"); d != "" {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			dateFin = &t
		}
	}

	stats, err := ctrl.service.GetStatistiques(c.Request().Context(), id, dateDebut, dateFin)
	if err != nil {
		if err.Error() == "commissariat not found" {
			return responses.NotFound(c, "Commissariat not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, stats)
}



