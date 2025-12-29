package document

import (
	"strconv"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles document routes
type Controller struct {
	service Service
}

// NewDocumentController creates a new document controller
func NewDocumentController(service Service) interfaces.Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers document routes
func (c *Controller) RegisterRoutes(g *echo.Group) {
	group := g.Group("/documents")

	// CRUD endpoints
	group.GET("", c.ListDocuments)
	group.GET("/:id", c.GetDocument)
	group.POST("", c.UploadDocument)
	group.PUT("/:id", c.UpdateDocument)
	group.DELETE("/:id", c.DeleteDocument)

	// Download endpoint
	group.GET("/:id/download", c.DownloadDocument)

	// Related documents
	group.GET("/controle/:controleId", c.GetByControle)
	group.GET("/infraction/:infractionId", c.GetByInfraction)
	group.GET("/pv/:pvId", c.GetByProcesVerbal)
	group.GET("/recours/:recoursId", c.GetByRecours)
	group.GET("/user/:userId", c.GetByUploader)

	// Statistics
	group.GET("/statistics", c.GetStatistics)
}

// ListDocuments lists documents with filters
func (c *Controller) ListDocuments(ctx echo.Context) error {
	request := &ListDocumentsRequest{}

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
	if typeDocument := ctx.QueryParam("type_document"); typeDocument != "" {
		request.TypeDocument = &typeDocument
	}

	if publicStr := ctx.QueryParam("public"); publicStr != "" {
		public := publicStr == "true"
		request.Public = &public
	}

	if controleID := ctx.QueryParam("controle_id"); controleID != "" {
		request.ControleID = &controleID
	}

	if infractionID := ctx.QueryParam("infraction_id"); infractionID != "" {
		request.InfractionID = &infractionID
	}

	if pvID := ctx.QueryParam("proces_verbal_id"); pvID != "" {
		request.ProcesVerbalID = &pvID
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
		return responses.InternalServerError(ctx, "Failed to list documents: "+err.Error())
	}

	return responses.Success(ctx, result)
}

// GetDocument gets a document by ID
func (c *Controller) GetDocument(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	document, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "document not found" {
			return responses.NotFound(ctx, "Document not found")
		}
		return responses.InternalServerError(ctx, "Failed to get document")
	}

	return responses.Success(ctx, document)
}

// UploadDocument uploads a new document
func (c *Controller) UploadDocument(ctx echo.Context) error {
	// Récupérer le fichier
	file, err := ctx.FormFile("file")
	if err != nil {
		return responses.BadRequest(ctx, "File is required")
	}

	// Vérifier la taille (max 10MB)
	if file.Size > 10*1024*1024 {
		return responses.BadRequest(ctx, "File too large (max 10MB)")
	}

	// Récupérer les paramètres
	request := &UploadDocumentRequest{
		TypeDocument: ctx.FormValue("type_document"),
		Public:       ctx.FormValue("public") == "true",
	}

	if description := ctx.FormValue("description"); description != "" {
		request.Description = &description
	}

	if controleID := ctx.FormValue("controle_id"); controleID != "" {
		request.ControleID = &controleID
	}

	if infractionID := ctx.FormValue("infraction_id"); infractionID != "" {
		request.InfractionID = &infractionID
	}

	if pvID := ctx.FormValue("proces_verbal_id"); pvID != "" {
		request.ProcesVerbalID = &pvID
	}

	if recoursID := ctx.FormValue("recours_id"); recoursID != "" {
		request.RecoursID = &recoursID
	}

	// Récupérer l'ID utilisateur du contexte
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return responses.Unauthorized(ctx, "User not authenticated")
	}

	document, err := c.service.Upload(ctx.Request().Context(), file, request, userID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to upload document: "+err.Error())
	}

	return responses.Created(ctx, document)
}

// UpdateDocument updates a document
func (c *Controller) UpdateDocument(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	var request UpdateDocumentRequest
	if err := ctx.Bind(&request); err != nil {
		return responses.BadRequest(ctx, "Invalid request")
	}

	document, err := c.service.Update(ctx.Request().Context(), id, &request)
	if err != nil {
		if err.Error() == "document not found" {
			return responses.NotFound(ctx, "Document not found")
		}
		return responses.InternalServerError(ctx, "Failed to update document")
	}

	return responses.Success(ctx, document)
}

// DeleteDocument deletes a document
func (c *Controller) DeleteDocument(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	err := c.service.Delete(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "document not found" {
			return responses.NotFound(ctx, "Document not found")
		}
		return responses.InternalServerError(ctx, "Failed to delete document")
	}

	return responses.Success(ctx, nil)
}

// DownloadDocument downloads a document
func (c *Controller) DownloadDocument(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return responses.BadRequest(ctx, "ID is required")
	}

	// Récupérer les infos du document
	document, err := c.service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err.Error() == "document not found" {
			return responses.NotFound(ctx, "Document not found")
		}
		return responses.InternalServerError(ctx, "Failed to get document")
	}

	// Récupérer le chemin du fichier
	filePath, err := c.service.GetFilePath(ctx.Request().Context(), id)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get file path")
	}

	// Renvoyer le fichier
	return ctx.Attachment(filePath, document.NomOriginal)
}

// GetByControle gets documents by controle ID
func (c *Controller) GetByControle(ctx echo.Context) error {
	controleID := ctx.Param("controleId")
	if controleID == "" {
		return responses.BadRequest(ctx, "Controle ID is required")
	}

	result, err := c.service.GetByControle(ctx.Request().Context(), controleID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get documents")
	}

	return responses.Success(ctx, result)
}

// GetByInfraction gets documents by infraction ID
func (c *Controller) GetByInfraction(ctx echo.Context) error {
	infractionID := ctx.Param("infractionId")
	if infractionID == "" {
		return responses.BadRequest(ctx, "Infraction ID is required")
	}

	result, err := c.service.GetByInfraction(ctx.Request().Context(), infractionID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get documents")
	}

	return responses.Success(ctx, result)
}

// GetByProcesVerbal gets documents by proces verbal ID
func (c *Controller) GetByProcesVerbal(ctx echo.Context) error {
	pvID := ctx.Param("pvId")
	if pvID == "" {
		return responses.BadRequest(ctx, "Proces Verbal ID is required")
	}

	result, err := c.service.GetByProcesVerbal(ctx.Request().Context(), pvID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get documents")
	}

	return responses.Success(ctx, result)
}

// GetByRecours gets documents by recours ID
func (c *Controller) GetByRecours(ctx echo.Context) error {
	recoursID := ctx.Param("recoursId")
	if recoursID == "" {
		return responses.BadRequest(ctx, "Recours ID is required")
	}

	result, err := c.service.GetByRecours(ctx.Request().Context(), recoursID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get documents")
	}

	return responses.Success(ctx, result)
}

// GetByUploader gets documents by uploader ID
func (c *Controller) GetByUploader(ctx echo.Context) error {
	userID := ctx.Param("userId")
	if userID == "" {
		return responses.BadRequest(ctx, "User ID is required")
	}

	result, err := c.service.GetByUploader(ctx.Request().Context(), userID)
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get documents")
	}

	return responses.Success(ctx, result)
}

// GetStatistics gets document statistics
func (c *Controller) GetStatistics(ctx echo.Context) error {
	stats, err := c.service.GetStatistics(ctx.Request().Context())
	if err != nil {
		return responses.InternalServerError(ctx, "Failed to get statistics")
	}

	return responses.Success(ctx, stats)
}

// getUserIDFromContext extracts user ID from the Echo context
func getUserIDFromContext(ctx echo.Context) string {
	// D'abord essayer le contexte JWT standard
	if userID, ok := ctx.Get("user_id").(string); ok && userID != "" {
		return userID
	}

	// Sinon essayer le claim "sub" du token
	if sub, ok := ctx.Get("sub").(string); ok && sub != "" {
		return sub
	}

	return ""
}
