package convocations

import (
	"net/http"
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller handles convocations HTTP requests
type Controller struct {
	service        Service
	authMiddleware *middleware.AuthMiddleware
	logger         *zap.Logger
}

// NewController creates a new convocations controller
func NewController(service Service, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) *Controller {
	return &Controller{
		service:        service,
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// RegisterRoutes registers convocations routes
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	ctrl.logger.Info("Registering convocations routes")
	
	// Le groupe est déjà sous /api/v1 avec auth middleware appliqué
	// On crée juste le sous-groupe /convocations
	convocations := e.Group("/convocations")

	// Routes principales
	convocations.POST("", ctrl.Create)
	convocations.GET("", ctrl.List)
	
	// Routes fixes (AVANT /:id pour éviter les conflits)
	convocations.GET("/statistiques", ctrl.GetStatistiques)
	convocations.GET("/dashboard", ctrl.GetDashboard)
	
	// Route avec paramètre :id (APRÈS les routes fixes)
	convocations.GET("/:id", ctrl.GetByID)
	convocations.PATCH("/:id/statut", ctrl.UpdateStatut)
	convocations.PATCH("/:id/reporter", ctrl.ReporterRdv)
	convocations.POST("/:id/notifier", ctrl.Notifier)
	convocations.POST("/:id/notes", ctrl.AjouterNote)
	convocations.GET("/:id/pdf", ctrl.DownloadPDF)
	
	ctrl.logger.Info("Convocations routes registered successfully",
		zap.String("base_path", "/api/v1/convocations"),
		zap.String("statistiques_route", "GET /api/v1/convocations/statistiques"),
		zap.String("byid_route", "GET /api/v1/convocations/:id"),
	)
}

// Create handles POST /convocations
func (ctrl *Controller) Create(c echo.Context) error {
	ctrl.logger.Info("📥 [Create Convocation] Request received")
	
	var req CreateConvocationRequest
	if err := c.Bind(&req); err != nil {
		ctrl.logger.Error("❌ [Create Convocation] Bind error", zap.Error(err))
		return responses.BadRequest(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		ctrl.logger.Error("❌ [Create Convocation] Validation error", zap.Error(err))
		return responses.BadRequest(c, err.Error())
	}

	// Get agent ID and commissariat ID from context
	agentID := getUserIDFromContext(c)
	if agentID == "" {
		ctrl.logger.Error("❌ [Create Convocation] User ID not found in context")
		return responses.BadRequest(c, "User ID not found in context")
	}

	commissariatID := getCommissariatIDFromContext(c)
	if commissariatID == "" {
		ctrl.logger.Error("❌ [Create Convocation] Commissariat ID not found in context")
		return responses.BadRequest(c, "Commissariat ID not found in context")
	}

	ctrl.logger.Info("✅ [Create Convocation] Context extracted",
		zap.String("agent_id", agentID),
		zap.String("commissariat_id", commissariatID),
		zap.String("type_convocation", req.TypeConvocation),
		zap.String("nom", req.Nom),
		zap.String("prenom", req.Prenom),
	)

	convocation, err := ctrl.service.Create(c.Request().Context(), &req, agentID, commissariatID)
	if err != nil {
		ctrl.logger.Error("❌ [Create Convocation] Service error", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	ctrl.logger.Info("🎉 [Create Convocation] Success",
		zap.String("numero", convocation.Numero),
		zap.String("id", convocation.ID),
	)

	return responses.Created(c, convocation)
}

// GetByID handles GET /convocations/:id
func (ctrl *Controller) GetByID(c echo.Context) error {
	id := c.Param("id")

	convocation, err := ctrl.service.GetByID(c.Request().Context(), id)
	if err != nil {
		ctrl.logger.Error("Failed to get convocation", zap.Error(err))
		return responses.NotFound(c, "Convocation not found")
	}

	return responses.Success(c, convocation)
}

// List handles GET /convocations
func (ctrl *Controller) List(c echo.Context) error {
	// Parse filters
	filters := &FilterConvocationsRequest{
		Page:  1,
		Limit: 5,
	}

	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filters.Page = p
		}
	}

	if limit := c.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}

	if statut := c.QueryParam("statut"); statut != "" {
		filters.Statut = &statut
	}

	if typeConvocation := c.QueryParam("typeConvocation"); typeConvocation != "" {
		filters.TypeConvocation = &typeConvocation
	}

	if qualiteConvoque := c.QueryParam("qualiteConvoque"); qualiteConvoque != "" {
		filters.QualiteConvoque = &qualiteConvoque
	}

	if commissariatID := c.QueryParam("commissariatId"); commissariatID != "" {
		filters.CommissariatID = &commissariatID
	}

	if search := c.QueryParam("search"); search != "" {
		filters.Search = &search
	}

	// Parse dates
	if dateDebut := c.QueryParam("dateDebut"); dateDebut != "" {
		if t, err := time.Parse(time.RFC3339, dateDebut); err == nil {
			filters.DateDebut = &t
		}
	}

	if dateFin := c.QueryParam("dateFin"); dateFin != "" {
		if t, err := time.Parse(time.RFC3339, dateFin); err == nil {
			filters.DateFin = &t
		}
	}

	// Get user info from context
	role := getRoleFromContext(c)
	userID := getUserIDFromContext(c)
	commissariatID := getCommissariatIDFromContext(c)

	result, err := ctrl.service.List(c.Request().Context(), filters, role, userID, commissariatID)
	if err != nil {
		ctrl.logger.Error("Failed to list convocations", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	// Return with data and pagination wrapper
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"convocations": result.Convocations,
			"pagination":   result.Pagination,
		},
	})
}

// UpdateStatut handles PATCH /convocations/:id/statut
func (ctrl *Controller) UpdateStatut(c echo.Context) error {
	id := c.Param("id")

	var req UpdateStatutConvocationRequest
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

	convocation, err := ctrl.service.UpdateStatut(c.Request().Context(), id, &req, agentID)
	if err != nil {
		ctrl.logger.Error("Failed to update convocation statut", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, convocation)
}

// GetStatistiques handles GET /convocations/statistiques
func (ctrl *Controller) GetStatistiques(c echo.Context) error {
	ctrl.logger.Info("📊 GetStatistiques called")

	// Parse filters
	var commissariatID *string
	if commID := c.QueryParam("commissariatId"); commID != "" {
		commissariatID = &commID
	}

	var dateDebut *string
	if dd := c.QueryParam("dateDebut"); dd != "" {
		dateDebut = &dd
	}

	var dateFin *string
	if df := c.QueryParam("dateFin"); df != "" {
		dateFin = &df
	}

	var periode *string
	if p := c.QueryParam("periode"); p != "" {
		periode = &p
	}

	stats, err := ctrl.service.GetStatistiques(c.Request().Context(), commissariatID, dateDebut, dateFin, periode)
	if err != nil {
		ctrl.logger.Error("Failed to get statistiques", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, stats)
}

// GetDashboard handles GET /convocations/dashboard
func (ctrl *Controller) GetDashboard(c echo.Context) error {
	ctrl.logger.Info("📊 GetDashboard called")

	// Parse filters
	var commissariatID *string
	if commID := c.QueryParam("commissariatId"); commID != "" {
		commissariatID = &commID
	}

	var dateDebut *string
	if dd := c.QueryParam("dateDebut"); dd != "" {
		dateDebut = &dd
	}

	var dateFin *string
	if df := c.QueryParam("dateFin"); df != "" {
		dateFin = &df
	}

	var periode *string
	if p := c.QueryParam("periode"); p != "" {
		periode = &p
	}

	dashboard, err := ctrl.service.GetDashboard(c.Request().Context(), commissariatID, dateDebut, dateFin, periode)
	if err != nil {
		ctrl.logger.Error("Failed to get dashboard", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, dashboard)
}

// Helper functions to get user info from context
func getUserIDFromContext(c echo.Context) string {
	if userID := c.Get("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func getCommissariatIDFromContext(c echo.Context) string {
	if commissariatID := c.Get("commissariat_id"); commissariatID != nil {
		if id, ok := commissariatID.(string); ok {
			return id
		}
	}
	return ""
}

func getRoleFromContext(c echo.Context) string {
	if role := c.Get("role"); role != nil {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}

// ReporterRdv handles PATCH /convocations/:id/reporter
func (ctrl *Controller) ReporterRdv(c echo.Context) error {
	id := c.Param("id")

	var req ReporterRdvRequest
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

	convocation, err := ctrl.service.ReporterRdv(c.Request().Context(), id, &req, agentID)
	if err != nil {
		ctrl.logger.Error("Failed to reporter rendez-vous", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, convocation)
}

// Notifier handles POST /convocations/:id/notifier
func (ctrl *Controller) Notifier(c echo.Context) error {
	id := c.Param("id")

	var req NotifierRequest
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

	convocation, err := ctrl.service.Notifier(c.Request().Context(), id, &req, agentID)
	if err != nil {
		ctrl.logger.Error("Failed to send notification", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, convocation)
}

// AjouterNote handles POST /convocations/:id/notes
func (ctrl *Controller) AjouterNote(c echo.Context) error {
	id := c.Param("id")

	var req AjouterNoteRequest
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

	convocation, err := ctrl.service.AjouterNote(c.Request().Context(), id, &req, agentID)
	if err != nil {
		ctrl.logger.Error("Failed to add note", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	return responses.Success(c, convocation)
}

// DownloadPDF handles GET /convocations/:id/pdf
func (ctrl *Controller) DownloadPDF(c echo.Context) error {
	id := c.Param("id")

	pdfData, err := ctrl.service.GeneratePDF(c.Request().Context(), id)
	if err != nil {
		ctrl.logger.Error("Failed to generate PDF", zap.Error(err))
		return responses.InternalServerError(c, err.Error())
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=convocation_"+id+".pdf")
	return c.Blob(http.StatusOK, "application/pdf", pdfData)
}
