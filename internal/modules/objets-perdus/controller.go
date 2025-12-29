package objetsperdus

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

// Controller handles objets perdus HTTP requests
type Controller struct {
	service        Service
	authMiddleware *middleware.AuthMiddleware
	logger         *zap.Logger
}

// NewController creates a new objets perdus controller
func NewController(service Service, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) *Controller {
	return &Controller{
		service:        service,
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// RegisterRoutes registers objets perdus routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	ctrl.logger.Info("Registering objets-perdus routes")
	
	objetsPerdus := e.Group("/objets-perdus")

	// Appliquer le middleware d'authentification à toutes les routes
	objetsPerdus.Use(ctrl.authMiddleware.RequireAuth())

	// Routes principales
	objetsPerdus.POST("", ctrl.Create)
	objetsPerdus.GET("", ctrl.List)
	
	// Routes fixes (AVANT /:id pour éviter les conflits)
	// IMPORTANT: Ces routes doivent être définies AVANT /:id pour être matchées correctement
	objetsPerdus.POST("/check-matches", ctrl.CheckMatches)
	objetsPerdus.GET("/statistiques", ctrl.GetStatistiques)
	objetsPerdus.GET("/dashboard", ctrl.GetDashboard)
	
	// Route avec paramètre :id (APRÈS les routes fixes)
	// Cette route sera matchée en dernier pour éviter d'intercepter les routes fixes
	objetsPerdus.GET("/:id", ctrl.GetByID)
	objetsPerdus.PATCH("/:id", ctrl.Update)
	objetsPerdus.PATCH("/:id/statut", ctrl.UpdateStatut)
	objetsPerdus.DELETE("/:id", ctrl.Delete)
	
	ctrl.logger.Info("Objets-perdus routes registered successfully",
		zap.String("base_path", "/api/objets-perdus"),
		zap.String("statistiques_route", "GET /api/objets-perdus/statistiques"),
		zap.String("byid_route", "GET /api/objets-perdus/:id"),
	)
}

// Create handles POST /objets-perdus
func (ctrl *Controller) Create(c echo.Context) error {
	var req CreateObjetPerduRequest
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
	ctrl.logger.Info("Creating objet perdu",
		zap.String("agent_id", agentID),
		zap.String("commissariat_id", commissariatID),
		zap.String("type_objet", req.TypeObjet),
	)

	objet, err := ctrl.service.Create(c.Request().Context(), &req, agentID, commissariatID)
	if err != nil {
		ctrl.logger.Error("Failed to create objet perdu", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Created(c, objet)
}

// CheckMatches handles POST /objets-perdus/check-matches
// Vérifie si des objets retrouvés correspondent aux identifiants ultra-uniques fournis
func (ctrl *Controller) CheckMatches(c echo.Context) error {
	var req CheckMatchesRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	ctrl.logger.Info("Checking for matching objets retrouvés",
		zap.String("type_objet", req.TypeObjet),
		zap.Any("identifiers", req.Identifiers),
	)

	matches, err := ctrl.service.CheckMatches(c.Request().Context(), &req)
	if err != nil {
		ctrl.logger.Error("Failed to check matches", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	ctrl.logger.Info("Found matching objets retrouvés",
		zap.Int("count", len(matches)),
	)

	return responses.Success(c, CheckMatchesResponse{
		Matches: matches,
		Count:   len(matches),
	})
}

// GetByID handles GET /objets-perdus/:id
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

	// Vérifier que l'ID n'est pas un chemin réservé (AVANT toute autre validation)
	// Cette vérification est critique car Echo peut parfois matcher /:id avant les routes fixes
	reservedPaths := []string{"statistiques", "search", "numero", "correspondants"}
	for _, reserved := range reservedPaths {
		if id == reserved {
			ctrl.logger.Error("❌ GetByID intercepted reserved path - this should not happen!",
				zap.String("path", id),
				zap.String("request_uri", c.Request().RequestURI),
				zap.String("expected_route", "/objets-perdus/statistiques"),
			)
			// Retourner 404 au lieu de 400 pour indiquer que la route n'existe pas
			// Cela permet au client de comprendre que la route n'a pas été trouvée
			return responses.NotFound(c, fmt.Sprintf("Route not found: /objets-perdus/%s", id))
		}
	}

	// Vérifier que l'ID est un UUID valide (longueur et format)
	if len(id) != 36 {
		ctrl.logger.Warn("Invalid ID format: wrong length",
			zap.String("id", id),
			zap.Int("length", len(id)),
		)
		return responses.BadRequest(c, fmt.Sprintf("Invalid ID format: invalid UUID length: %d", len(id)))
	}

	// Essayer de parser l'UUID pour valider le format
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
		if err.Error() == "objet perdu not found" {
			return responses.NotFound(c, "Objet perdu not found")
		}
		ctrl.logger.Error("Failed to get objet perdu", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, objet)
}

// List handles GET /objets-perdus
func (ctrl *Controller) List(c echo.Context) error {
	filters := &FilterObjetsPerdusRequest{}

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
			ctrl.logger.Info("✅ Parsed dateDebut (objets-perdus)",
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
			ctrl.logger.Info("✅ Parsed dateFin (objets-perdus)",
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

	result, err := ctrl.service.List(c.Request().Context(), filters, roleStr, userID, commissariatFromContext)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, result)
}

// Update handles PATCH /objets-perdus/:id
func (ctrl *Controller) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateObjetPerduRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	objet, err := ctrl.service.Update(c.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "objet perdu not found" {
			return responses.NotFound(c, "Objet perdu not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, objet)
}

// UpdateStatut handles PATCH /objets-perdus/:id/statut
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
		if err.Error() == "objet perdu not found" {
			return responses.NotFound(c, "Objet perdu not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, objet)
}

// Delete handles DELETE /objets-perdus/:id
func (ctrl *Controller) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	err := ctrl.service.Delete(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "objet perdu not found" {
			return responses.NotFound(c, "Objet perdu not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

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

// GetStatistiques handles GET /objets-perdus/statistiques
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

