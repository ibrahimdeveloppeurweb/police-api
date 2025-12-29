package objetsretrouves

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles objets retrouves HTTP requests
type Controller struct {
	service        Service
	authMiddleware *middleware.AuthMiddleware
	logger         *zap.Logger
}

// NewController creates a new objets retrouves controller
func NewController(service Service, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) *Controller {
	return &Controller{
		service:        service,
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// RegisterRoutes registers objets retrouves routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	ctrl.logger.Info("Registering objets-retrouves routes")

	objetsRetrouves := e.Group("/objets-retrouves")

	// Appliquer le middleware d'authentification à toutes les routes
	objetsRetrouves.Use(ctrl.authMiddleware.RequireAuth())

	// Routes principales
	objetsRetrouves.POST("", ctrl.Create)
	objetsRetrouves.GET("", ctrl.List)

	// Routes fixes (AVANT /:id pour éviter les conflits)
	objetsRetrouves.GET("/statistiques", ctrl.GetStatistiques)
	objetsRetrouves.GET("/dashboard", ctrl.GetDashboard)

	// Route avec paramètre :id (APRÈS les routes fixes)
	objetsRetrouves.GET("/:id", ctrl.GetByID)
	objetsRetrouves.PATCH("/:id", ctrl.Update)
	objetsRetrouves.PATCH("/:id/statut", ctrl.UpdateStatut)
	objetsRetrouves.DELETE("/:id", ctrl.Delete)

	ctrl.logger.Info("Objets-retrouves routes registered successfully",
		zap.String("base_path", "/api/objets-retrouves"),
		zap.String("statistiques_route", "GET /api/objets-retrouves/statistiques"),
		zap.String("byid_route", "GET /api/objets-retrouves/:id"),
	)
}

// Create handles POST /objets-retrouves
func (ctrl *Controller) Create(c echo.Context) error {
	var req CreateObjetRetrouveRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Get agent ID and commissariat ID from context
	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	commissariatID := getCommissariatIDFromContext(c)
	if commissariatID == "" {
		return responses.BadRequest(c, "Commissariat ID not found in context")
	}

	// Log pour confirmer que les informations sont bien récupérées
	ctrl.logger.Info("Creating objet retrouve",
		zap.String("agent_id", agentID),
		zap.String("commissariat_id", commissariatID),
		zap.String("type_objet", req.TypeObjet),
	)

	objet, err := ctrl.service.Create(c.Request().Context(), &req, agentID, commissariatID)
	if err != nil {
		ctrl.logger.Error("Failed to create objet retrouve", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Created(c, objet)
}

// GetByID handles GET /objets-retrouves/:id
func (ctrl *Controller) GetByID(c echo.Context) error {
	id := c.Param("id")

	ctrl.logger.Info("GetByID called",
		zap.String("id", id),
		zap.String("path", c.Path()),
		zap.String("request_uri", c.Request().RequestURI),
		zap.String("request_method", c.Request().Method),
	)

	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	// Vérifier que l'ID n'est pas un chemin réservé
	reservedPaths := []string{"statistiques", "search", "numero", "correspondants"}
	for _, reserved := range reservedPaths {
		if id == reserved {
			ctrl.logger.Error("❌ GetByID intercepted reserved path - this should not happen!",
				zap.String("path", id),
				zap.String("request_uri", c.Request().RequestURI),
				zap.String("expected_route", "/objets-retrouves/statistiques"),
			)
			return responses.NotFound(c, fmt.Sprintf("Route not found: /objets-retrouves/%s", id))
		}
	}

	// Vérifier que l'ID est un UUID valide
	if len(id) != 36 {
		ctrl.logger.Warn("Invalid ID format: wrong length",
			zap.String("id", id),
			zap.Int("length", len(id)),
		)
		return responses.BadRequest(c, fmt.Sprintf("Invalid ID format: invalid UUID length: %d", len(id)))
	}

	_, err := uuid.Parse(id)
	if err != nil {
		ctrl.logger.Warn("Invalid ID format: not a valid UUID",
			zap.String("id", id),
			zap.Error(err),
		)
		return responses.BadRequest(c, fmt.Sprintf("Invalid ID format: %v", err))
	}

	objet, err := ctrl.service.GetByID(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return responses.NotFound(c, "Objet retrouve not found")
		}
		ctrl.logger.Error("Failed to get objet retrouve", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, objet)
}

// List handles GET /objets-retrouves
// List handles GET /objets-retrouves
func (ctrl *Controller) List(c echo.Context) error {
	filters := &FilterObjetsRetrouvesRequest{}

	// Parse query parameters
	if statut := c.QueryParam("statut"); statut != "" {
		filters.Statut = &statut
	}
	if typeObjet := c.QueryParam("typeObjet"); typeObjet != "" {
		filters.TypeObjet = &typeObjet
	}
	if commissariatID := c.QueryParam("commissariatId"); commissariatID != "" {
		filters.CommissariatID = &commissariatID
	}
	if search := c.QueryParam("search"); search != "" {
		filters.Search = &search
	}

	// CORRECTION: Ajouter le parsing de isContainer
	if isContainerStr := c.QueryParam("isContainer"); isContainerStr != "" {
		isContainer, err := strconv.ParseBool(isContainerStr)
		if err == nil {
			filters.IsContainer = &isContainer
			ctrl.logger.Info("Parsed isContainer filter", zap.Bool("isContainer", isContainer))
		} else {
			ctrl.logger.Warn("Failed to parse isContainer", zap.String("value", isContainerStr), zap.Error(err))
		}
	}

	// Parse dates en UTC (standard pour les bases de données)
	// Supporte plusieurs formats: "2006-01-02" et "2006-01-02T15:04:05"
	if dateDebut := c.QueryParam("dateDebut"); dateDebut != "" {
		var t time.Time
		var err error
		
		// Essayer le format ISO 8601 avec heure
		t, err = time.Parse("2006-01-02T15:04:05", dateDebut)
		if err != nil {
			// Essayer le format date simple
			t, err = time.Parse("2006-01-02", dateDebut)
		}
		
		if err == nil {
			// Début de journée en UTC (00:00:00)
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			filters.DateDebut = &startOfDay
			ctrl.logger.Info("✅ Parsed dateDebut",
				zap.String("input", dateDebut),
				zap.Time("parsed", startOfDay),
				zap.String("timezone", "UTC"),
			)
		} else {
			ctrl.logger.Error("❌ Failed to parse dateDebut", zap.String("dateDebut", dateDebut), zap.Error(err))
		}
	}

	if dateFin := c.QueryParam("dateFin"); dateFin != "" {
		var t time.Time
		var err error
		
		// Essayer le format ISO 8601 avec heure
		t, err = time.Parse("2006-01-02T15:04:05", dateFin)
		if err != nil {
			// Essayer le format date simple
			t, err = time.Parse("2006-01-02", dateFin)
		}
		
		if err == nil {
			// Fin de journée en UTC (23:59:59.999999999)
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
			filters.DateFin = &endOfDay
			ctrl.logger.Info("✅ Parsed dateFin",
				zap.String("input", dateFin),
				zap.Time("parsed", endOfDay),
				zap.String("timezone", "UTC"),
			)
		} else {
			ctrl.logger.Error("❌ Failed to parse dateFin", zap.String("dateFin", dateFin), zap.Error(err))
		}
	}

	// Parse pagination
	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters.Page = p
		}
	}
	if limit := c.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters.Limit = l
		}
	}

	// Get user context for filtering
	userID := getUserIDFromContext(c)
	role := c.Get("role")
	commissariatFromContext := getCommissariatIDFromContext(c)

	roleStr := "AGENT"
	if role != nil {
		roleStr = role.(string)
	}

	// CORRECTION: Utiliser le logger zap au lieu de log.Println
	ctrl.logger.Info("Controller List - Filters applied",
		zap.Any("DateDebut", filters.DateDebut),
		zap.Any("DateFin", filters.DateFin),
		zap.Stringp("Statut", filters.Statut),
		zap.Stringp("TypeObjet", filters.TypeObjet),
		zap.Stringp("CommissariatID", filters.CommissariatID),
		zap.Stringp("Search", filters.Search),
		zap.Any("IsContainer", filters.IsContainer),
		zap.Int("Page", filters.Page),
		zap.Int("Limit", filters.Limit),
	)

	result, err := ctrl.service.List(c.Request().Context(), filters, roleStr, userID, commissariatFromContext)
	if err != nil {
		ctrl.logger.Error("Failed to list objets retrouves", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	ctrl.logger.Info("Controller List - Results",
		zap.Int("count", len(result.Objets)),
		zap.Int64("total", result.Total),
	)

	return responses.Success(c, result)
}

// Update handles PATCH /objets-retrouves/:id
func (ctrl *Controller) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateObjetRetrouveRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	objet, err := ctrl.service.Update(c.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return responses.NotFound(c, "Objet retrouve not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, objet)
}

// UpdateStatut handles PATCH /objets-retrouves/:id/statut
func (ctrl *Controller) UpdateStatut(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateStatutRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return responses.BadRequest(c, err.Error())
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	objet, err := ctrl.service.UpdateStatut(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return responses.NotFound(c, "Objet retrouve not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, objet)
}

// Delete handles DELETE /objets-retrouves/:id
func (ctrl *Controller) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	err := ctrl.service.Delete(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return responses.NotFound(c, "Objet retrouve not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// GetDashboard handles GET /objets-retrouves/dashboard
func (ctrl *Controller) GetDashboard(c echo.Context) error {
	var commissariatID *string
	var dateDebut, dateFin *string
	var periode *string

	if id := c.QueryParam("commissariatId"); id != "" {
		commissariatID = &id
	}
	if dd := c.QueryParam("dateDebut"); dd != "" {
		dateDebut = &dd
	}
	if df := c.QueryParam("dateFin"); df != "" {
		dateFin = &df
	}
	if p := c.QueryParam("periode"); p != "" {
		periode = &p
	}

	dashboard, err := ctrl.service.GetDashboard(c.Request().Context(), commissariatID, dateDebut, dateFin, periode)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, dashboard)
}

// GetStatistiques handles GET /objets-retrouves/statistiques
func (ctrl *Controller) GetStatistiques(c echo.Context) error {
	ctrl.logger.Info("GetStatistiques called",
		zap.String("path", c.Path()),
		zap.String("request_uri", c.Request().RequestURI),
	)

	var commissariatID *string
	var dateDebut, dateFin, periode *string

	// Priorité 1: Paramètre de requête
	if id := c.QueryParam("commissariatId"); id != "" {
		commissariatID = &id
		ctrl.logger.Info("Using commissariatId from query param",
			zap.String("commissariatId", id),
		)
	} else {
		// Priorité 2: Commissariat du contexte utilisateur
		commissariatFromContext := getCommissariatIDFromContext(c)
		if commissariatFromContext != "" {
			commissariatID = &commissariatFromContext
			ctrl.logger.Info("Using commissariatId from context",
				zap.String("commissariatId", commissariatFromContext),
			)
		} else {
			ctrl.logger.Warn("No commissariatId provided in query param or context")
		}
	}

	if dd := c.QueryParam("dateDebut"); dd != "" {
		dateDebut = &dd
	}
	if df := c.QueryParam("dateFin"); df != "" {
		dateFin = &df
	}
	if p := c.QueryParam("periode"); p != "" {
		periode = &p
	}

	ctrl.logger.Info("GetStatistiques parameters",
		zap.Stringp("commissariatId", commissariatID),
		zap.Stringp("dateDebut", dateDebut),
		zap.Stringp("dateFin", dateFin),
		zap.Stringp("periode", periode),
	)

	stats, err := ctrl.service.GetStatistiques(c.Request().Context(), commissariatID, dateDebut, dateFin, periode)
	if err != nil {
		ctrl.logger.Error("Failed to get statistiques", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, stats)
}

// Helper functions
func getUserIDFromContext(c echo.Context) string {
	userID := c.Get("user_id")
	if userID != nil {
		return userID.(string)
	}
	return ""
}

func getCommissariatIDFromContext(c echo.Context) string {
	commissariatID := c.Get("commissariat_id")
	if commissariatID != nil {
		return commissariatID.(string)
	}
	return ""
}
