package alertes

import (
	"net/http"
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles alertes HTTP requests
type Controller struct {
	service        Service
	authMiddleware *middleware.AuthMiddleware
}

// NewController creates a new alertes controller
func NewController(service Service, authMiddleware *middleware.AuthMiddleware) *Controller {
	return &Controller{
		service:        service,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers alertes routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	alertes := e.Group("/alertes")

	// Appliquer le middleware d'authentification à toutes les routes alertes
	alertes.Use(ctrl.authMiddleware.RequireAuth())

	// Routes principales
	alertes.POST("", ctrl.Create)
	alertes.GET("", ctrl.List)
	
	// Routes fixes (AVANT /:id pour éviter les conflits)
	alertes.GET("/actives", ctrl.GetActives)
	alertes.GET("/statistiques", ctrl.GetStatistiques)
	alertes.GET("/dashboard", ctrl.GetDashboard)
	alertes.POST("/generer-description", ctrl.GenererDescription)
	alertes.POST("/:id/generer-rapport", ctrl.GenererRapport)
	
	// Route avec paramètre :id (APRÈS les routes fixes)
	alertes.GET("/:id", ctrl.GetByID)
	alertes.PATCH("/:id", ctrl.Update)
	alertes.PUT("/:id", ctrl.Update)
	alertes.DELETE("/:id", ctrl.Delete)

	// Gestion du cycle de vie
	alertes.POST("/:id/suivi", ctrl.AddSuivi)
	alertes.POST("/:id/broadcast", ctrl.Broadcast)
	alertes.POST("/:id/diffuser", ctrl.Broadcast) // Alias
	alertes.POST("/:id/diffusion-interne", ctrl.DiffusionInterne)
	alertes.POST("/:id/assign", ctrl.Assign)
	alertes.PATCH("/:id/resolve", ctrl.Resolve)
	alertes.POST("/:id/resoudre", ctrl.Resolve) // Alias
	alertes.POST("/:id/archiver", ctrl.Archiver)
	alertes.POST("/:id/cloturer", ctrl.Cloturer)

	// Intervention
	alertes.POST("/:id/intervention/deploy", ctrl.DeployIntervention)
	alertes.PATCH("/:id/intervention", ctrl.UpdateIntervention)

	// Évaluation et rapport
	alertes.POST("/:id/evaluation", ctrl.AddEvaluation)
	alertes.POST("/:id/rapport", ctrl.AddRapport)

	// Témoins et documents
	alertes.POST("/:id/temoin", ctrl.AddTemoin)
	alertes.POST("/:id/document", ctrl.AddDocument)
	alertes.POST("/:id/photos", ctrl.AddPhotos)

	// Actions
	alertes.PATCH("/:id/actions", ctrl.UpdateActions)
}

// List handles GET /alertes
func (ctrl *Controller) List(c echo.Context) error {
	filters := &FilterAlertesRequest{}

	// Parse query parameters
	if niveau := c.QueryParam("niveau"); niveau != "" {
		n := NiveauAlerte(niveau)
		filters.Niveau = &n
	}
	if statut := c.QueryParam("statut"); statut != "" {
		st := StatutAlerte(statut)
		filters.Statut = &st
	}
	if typeAlerte := c.QueryParam("type"); typeAlerte != "" {
		t := TypeAlerte(typeAlerte)
		filters.Type = &t
	}
	if commissariatID := c.QueryParam("commissariatId"); commissariatID != "" {
		filters.CommissariatID = &commissariatID
	}
	if search := c.QueryParam("search"); search != "" {
		filters.Search = &search
	}
	if diffusee := c.QueryParam("diffusee"); diffusee != "" {
		d := diffusee == "true"
		filters.Diffusee = &d
	}

	// Parse dates - accepter plusieurs formats
	if dateDebut := c.QueryParam("dateDebut"); dateDebut != "" {
		t, err := parseDateTime(dateDebut)
		if err == nil {
			filters.DateDebut = &t
		}
	}
	if dateFin := c.QueryParam("dateFin"); dateFin != "" {
		t, err := parseDateTime(dateFin)
		if err == nil {
			filters.DateFin = &t
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

	// Debug logs
	ctrl.service.(*service).logger.Info("List alertes request",
		zap.Any("filters", map[string]interface{}{
			"commissariatId": filters.CommissariatID,
			"dateDebut":      filters.DateDebut,
			"dateFin":        filters.DateFin,
			"statut":         filters.Statut,
			"niveau":         filters.Niveau,
			"type":           filters.Type,
			"search":         filters.Search,
			"page":           filters.Page,
			"limit":          filters.Limit,
		}),
	)

	result, err := ctrl.service.List(c.Request().Context(), filters, roleStr, userID, commissariatFromContext)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	ctrl.service.(*service).logger.Info("List alertes response",
		zap.Int("count", len(result.Alertes)),
		zap.Int64("total", result.Total),
	)

	return responses.Success(c, result)
}

// GetByID handles GET /alertes/:id
func (ctrl *Controller) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	alerte, err := ctrl.service.GetByID(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// Create handles POST /alertes
func (ctrl *Controller) Create(c echo.Context) error {
	var req CreateAlerteRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Get agent ID from context
	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.Create(c.Request().Context(), &req, agentID)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Created(c, alerte)
}

// Update handles PUT /alertes/:id
func (ctrl *Controller) Update(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateAlerteRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	alerte, err := ctrl.service.Update(c.Request().Context(), id, &req)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// Delete handles DELETE /alertes/:id
func (ctrl *Controller) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	err := ctrl.service.Delete(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Resolve handles PATCH /alertes/:id/resolve
func (ctrl *Controller) Resolve(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.Resoudre(c.Request().Context(), id, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// Broadcast handles POST /alertes/:id/broadcast
func (ctrl *Controller) Broadcast(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req BroadcastAlerteRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.Diffuser(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// GetActives handles GET /alertes/actives
func (ctrl *Controller) GetActives(c echo.Context) error {
	var commissariatID *string
	if id := c.QueryParam("commissariatId"); id != "" {
		commissariatID = &id
	}

	alertes, err := ctrl.service.GetActives(c.Request().Context(), commissariatID)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alertes)
}

// GetStatistiques handles GET /alertes/statistiques
func (ctrl *Controller) GetStatistiques(c echo.Context) error {
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

	stats, err := ctrl.service.GetStatistiques(c.Request().Context(), commissariatID, dateDebut, dateFin, periode)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, stats)
}

// GetDashboard handles GET /alertes/dashboard
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

// AddSuivi handles POST /alertes/:id/suivi
func (ctrl *Controller) AddSuivi(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AddSuiviRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.AddSuivi(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// DiffusionInterne handles POST /alertes/:id/diffusion-interne
func (ctrl *Controller) DiffusionInterne(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AssignAlerteRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	commissariatID := getCommissariatIDFromContext(c)
	
	if agentID == "" || commissariatID == "" {
		return responses.BadRequest(c, "User context incomplete")
	}

	alerte, err := ctrl.service.DiffusionInterne(c.Request().Context(), id, &req, commissariatID, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.BadRequest(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// Assign handles POST /alertes/:id/assign
func (ctrl *Controller) Assign(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AssignAlerteRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	commissariatID := getCommissariatIDFromContext(c)
	
	if agentID == "" || commissariatID == "" {
		return responses.BadRequest(c, "User context incomplete")
	}

	alerte, err := ctrl.service.Assigner(c.Request().Context(), id, &req, commissariatID, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.BadRequest(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// Archiver handles POST /alertes/:id/archiver
func (ctrl *Controller) Archiver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.Archiver(c.Request().Context(), id, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// Cloturer handles POST /alertes/:id/cloturer
func (ctrl *Controller) Cloturer(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.Cloturer(c.Request().Context(), id, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// DeployIntervention handles POST /alertes/:id/intervention/deploy
func (ctrl *Controller) DeployIntervention(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req DeployInterventionRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.DeployIntervention(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// UpdateIntervention handles PATCH /alertes/:id/intervention
func (ctrl *Controller) UpdateIntervention(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req UpdateInterventionRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.UpdateIntervention(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// AddEvaluation handles POST /alertes/:id/evaluation
func (ctrl *Controller) AddEvaluation(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AddEvaluationRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.AddEvaluation(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// AddRapport handles POST /alertes/:id/rapport
func (ctrl *Controller) AddRapport(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AddRapportRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.AddRapport(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// AddTemoin handles POST /alertes/:id/temoin
func (ctrl *Controller) AddTemoin(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AddTemoinRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.AddTemoin(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// AddDocument handles POST /alertes/:id/document
func (ctrl *Controller) AddDocument(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req AddDocumentRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.AddDocument(c.Request().Context(), id, &req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// AddPhotos handles POST /alertes/:id/photos
func (ctrl *Controller) AddPhotos(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	var req struct {
		Photos []string `json:"photos"`
	}
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.AddPhotos(c.Request().Context(), id, req.Photos, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// UpdateActions handles PATCH /alertes/:id/actions
func (ctrl *Controller) UpdateActions(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	// Le frontend peut envoyer { actions: {...} } ou directement {...}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	// Extraire actions du wrapper si présent
	var actionsData map[string]interface{}
	if actions, ok := body["actions"].(map[string]interface{}); ok {
		actionsData = actions
	} else {
		actionsData = body
	}

	// Convertir en UpdateActionsRequest
	req := &UpdateActionsRequest{}
	if immediate, ok := actionsData["immediate"].([]interface{}); ok {
		req.Immediate = make([]string, len(immediate))
		for i, v := range immediate {
			if str, ok := v.(string); ok {
				req.Immediate[i] = str
			}
		}
	}
	if preventive, ok := actionsData["preventive"].([]interface{}); ok {
		req.Preventive = make([]string, len(preventive))
		for i, v := range preventive {
			if str, ok := v.(string); ok {
				req.Preventive[i] = str
			}
		}
	}
	if suivi, ok := actionsData["suivi"].([]interface{}); ok {
		req.Suivi = make([]string, len(suivi))
		for i, v := range suivi {
			if str, ok := v.(string); ok {
				req.Suivi[i] = str
			}
		}
	}

	agentID := getUserIDFromContext(c)
	if agentID == "" {
		return responses.BadRequest(c, "User ID not found in context")
	}

	alerte, err := ctrl.service.UpdateActions(c.Request().Context(), id, req, agentID)
	if err != nil {
		if err.Error() == "alerte not found" {
			return responses.NotFound(c, "Alerte not found")
		}
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, alerte)
}

// GenererDescription handles POST /alertes/generer-description
func (ctrl *Controller) GenererDescription(c echo.Context) error {
	var req GenerateDescriptionRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, "Invalid request body")
	}

	result, err := ctrl.service.GenererDescription(c.Request().Context(), &req)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	// Retourner directement sans wrapper (le service retourne déjà la structure complète)
	return c.JSON(200, result)
}

// GenererRapport handles POST /alertes/:id/generer-rapport
func (ctrl *Controller) GenererRapport(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return responses.BadRequest(c, "ID is required")
	}

	result, err := ctrl.service.GenererRapport(c.Request().Context(), id)
	if err != nil {
		return responses.InternalServerError(c, err.Error())
	}

	return c.JSON(200, result)
}

// Helper functions
func getUserIDFromContext(c echo.Context) string {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID
	}
	return ""
}

func getCommissariatIDFromContext(c echo.Context) string {
	if commissariatID, ok := c.Get("commissariat_id").(string); ok {
		return commissariatID
	}
	return ""
}

// parseDateTime parse une date avec différents formats possibles
func parseDateTime(dateStr string) (time.Time, error) {
	// Formats supportés (par ordre de priorité)
	formats := []string{
		"2006-01-02T15:04:05",           // Format avec heure complète
		"2006-01-02T15:04:05Z",          // Format ISO avec Z
		"2006-01-02T15:04:05-07:00",     // Format ISO avec timezone
		time.RFC3339,                     // Format RFC3339
		"2006-01-02",                     // Format date simple
	}
	
	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	
	return time.Time{}, lastErr
}
