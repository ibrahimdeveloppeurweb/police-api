package document

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines document service interface
type Service interface {
	Upload(ctx context.Context, file *multipart.FileHeader, input *UploadDocumentRequest, userID string) (*DocumentResponse, error)
	GetByID(ctx context.Context, id string) (*DocumentResponse, error)
	List(ctx context.Context, filters *ListDocumentsRequest) (*ListDocumentsResponse, error)
	Update(ctx context.Context, id string, input *UpdateDocumentRequest) (*DocumentResponse, error)
	Delete(ctx context.Context, id string) error
	GetByControle(ctx context.Context, controleID string) (*ListDocumentsResponse, error)
	GetByInfraction(ctx context.Context, infractionID string) (*ListDocumentsResponse, error)
	GetByProcesVerbal(ctx context.Context, pvID string) (*ListDocumentsResponse, error)
	GetByRecours(ctx context.Context, recoursID string) (*ListDocumentsResponse, error)
	GetByUploader(ctx context.Context, userID string) (*ListDocumentsResponse, error)
	GetStatistics(ctx context.Context) (*DocumentStatisticsResponse, error)
	GetFilePath(ctx context.Context, id string) (string, error)
}

// service implements Service interface
type service struct {
	documentRepo repository.DocumentRepository
	logger       *zap.Logger
	uploadDir    string
	baseURL      string
}

// NewDocumentService creates a new document service
func NewDocumentService(
	documentRepo repository.DocumentRepository,
	logger *zap.Logger,
) Service {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Créer le répertoire d'upload s'il n'existe pas
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory", zap.Error(err))
	}

	return &service{
		documentRepo: documentRepo,
		logger:       logger,
		uploadDir:    uploadDir,
		baseURL:      baseURL,
	}
}

// Upload uploads a new document
func (s *service) Upload(ctx context.Context, file *multipart.FileHeader, input *UploadDocumentRequest, userID string) (*DocumentResponse, error) {
	if err := s.validateUploadInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Ouvrir le fichier
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Générer un nom de fichier unique
	ext := filepath.Ext(file.Filename)
	nomFichier := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Créer le sous-répertoire basé sur la date
	subDir := time.Now().Format("2006/01")
	fullDir := filepath.Join(s.uploadDir, subDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Chemin complet du fichier
	cheminStockage := filepath.Join(subDir, nomFichier)
	fullPath := filepath.Join(s.uploadDir, cheminStockage)

	// Créer le fichier destination
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Calculer le hash et copier en même temps
	hash := sha256.New()
	writer := io.MultiWriter(dst, hash)

	if _, err := io.Copy(writer, src); err != nil {
		// Supprimer le fichier en cas d'erreur
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	hashFichier := hex.EncodeToString(hash.Sum(nil))

	// Créer l'entrée en base
	repoInput := &repository.CreateDocumentInput{
		ID:             uuid.New().String(),
		NomFichier:     nomFichier,
		NomOriginal:    file.Filename,
		TypeMime:       file.Header.Get("Content-Type"),
		Taille:         file.Size,
		CheminStockage: cheminStockage,
		TypeDocument:   input.TypeDocument,
		Description:    input.Description,
		HashFichier:    &hashFichier,
		Public:         input.Public,
		UploadedByID:   userID,
		ControleID:     input.ControleID,
		InfractionID:   input.InfractionID,
		ProcesVerbalID: input.ProcesVerbalID,
		RecoursID:      input.RecoursID,
	}

	documentEnt, err := s.documentRepo.Create(ctx, repoInput)
	if err != nil {
		// Supprimer le fichier en cas d'erreur
		os.Remove(fullPath)
		s.logger.Error("Failed to create document", zap.Error(err))
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Recharger avec les relations
	documentEnt, err = s.documentRepo.GetByID(ctx, documentEnt.ID.String())
	if err != nil {
		s.logger.Error("Failed to reload document", zap.Error(err))
		return nil, fmt.Errorf("failed to reload document: %w", err)
	}

	return s.entityToResponse(documentEnt), nil
}

// GetByID gets document by ID
func (s *service) GetByID(ctx context.Context, id string) (*DocumentResponse, error) {
	documentEnt, err := s.documentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(documentEnt), nil
}

// List gets documents with filters
func (s *service) List(ctx context.Context, input *ListDocumentsRequest) (*ListDocumentsResponse, error) {
	filters := s.buildFilters(input)

	documentsEnt, err := s.documentRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	total, err := s.documentRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	documents := make([]*DocumentResponse, len(documentsEnt))
	for i, d := range documentsEnt {
		documents[i] = s.entityToResponse(d)
	}

	return &ListDocumentsResponse{
		Documents: documents,
		Total:     total,
	}, nil
}

// Update updates document
func (s *service) Update(ctx context.Context, id string, input *UpdateDocumentRequest) (*DocumentResponse, error) {
	if err := s.validateUpdateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	repoInput := &repository.UpdateDocumentInput{
		NomOriginal:  input.NomOriginal,
		TypeDocument: input.TypeDocument,
		Description:  input.Description,
		Public:       input.Public,
	}

	documentEnt, err := s.documentRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	documentEnt, err = s.documentRepo.GetByID(ctx, documentEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(documentEnt), nil
}

// Delete deletes document
func (s *service) Delete(ctx context.Context, id string) error {
	// Récupérer le document pour avoir le chemin
	doc, err := s.documentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Supprimer le fichier physique
	fullPath := filepath.Join(s.uploadDir, doc.CheminStockage)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to delete physical file", zap.String("path", fullPath), zap.Error(err))
	}

	// Supprimer l'entrée en base
	return s.documentRepo.Delete(ctx, id)
}

// GetByControle gets documents by controle ID
func (s *service) GetByControle(ctx context.Context, controleID string) (*ListDocumentsResponse, error) {
	documentsEnt, err := s.documentRepo.GetByControle(ctx, controleID)
	if err != nil {
		return nil, err
	}

	documents := make([]*DocumentResponse, len(documentsEnt))
	for i, d := range documentsEnt {
		documents[i] = s.entityToResponse(d)
	}

	return &ListDocumentsResponse{
		Documents: documents,
		Total:     len(documents),
	}, nil
}

// GetByInfraction gets documents by infraction ID
func (s *service) GetByInfraction(ctx context.Context, infractionID string) (*ListDocumentsResponse, error) {
	documentsEnt, err := s.documentRepo.GetByInfraction(ctx, infractionID)
	if err != nil {
		return nil, err
	}

	documents := make([]*DocumentResponse, len(documentsEnt))
	for i, d := range documentsEnt {
		documents[i] = s.entityToResponse(d)
	}

	return &ListDocumentsResponse{
		Documents: documents,
		Total:     len(documents),
	}, nil
}

// GetByProcesVerbal gets documents by proces verbal ID
func (s *service) GetByProcesVerbal(ctx context.Context, pvID string) (*ListDocumentsResponse, error) {
	documentsEnt, err := s.documentRepo.GetByProcesVerbal(ctx, pvID)
	if err != nil {
		return nil, err
	}

	documents := make([]*DocumentResponse, len(documentsEnt))
	for i, d := range documentsEnt {
		documents[i] = s.entityToResponse(d)
	}

	return &ListDocumentsResponse{
		Documents: documents,
		Total:     len(documents),
	}, nil
}

// GetByRecours gets documents by recours ID
func (s *service) GetByRecours(ctx context.Context, recoursID string) (*ListDocumentsResponse, error) {
	documentsEnt, err := s.documentRepo.GetByRecours(ctx, recoursID)
	if err != nil {
		return nil, err
	}

	documents := make([]*DocumentResponse, len(documentsEnt))
	for i, d := range documentsEnt {
		documents[i] = s.entityToResponse(d)
	}

	return &ListDocumentsResponse{
		Documents: documents,
		Total:     len(documents),
	}, nil
}

// GetByUploader gets documents by uploader ID
func (s *service) GetByUploader(ctx context.Context, userID string) (*ListDocumentsResponse, error) {
	documentsEnt, err := s.documentRepo.GetByUploader(ctx, userID)
	if err != nil {
		return nil, err
	}

	documents := make([]*DocumentResponse, len(documentsEnt))
	for i, d := range documentsEnt {
		documents[i] = s.entityToResponse(d)
	}

	return &ListDocumentsResponse{
		Documents: documents,
		Total:     len(documents),
	}, nil
}

// GetStatistics gets statistics for documents
func (s *service) GetStatistics(ctx context.Context) (*DocumentStatisticsResponse, error) {
	documentsEnt, err := s.documentRepo.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	stats := &DocumentStatisticsResponse{
		TotalDocuments: len(documentsEnt),
		ParType:        make(map[string]int),
		ParMois:        make(map[string]int),
	}

	for _, d := range documentsEnt {
		stats.TailleTotale += d.Taille
		stats.ParType[d.TypeDocument]++
		stats.ParMois[d.CreatedAt.Format("2006-01")]++

		if d.Public {
			stats.DocumentsPublics++
		} else {
			stats.DocumentsPrives++
		}
	}

	stats.TailleTotaleText = formatFileSize(stats.TailleTotale)

	return stats, nil
}

// GetFilePath gets the file path for a document
func (s *service) GetFilePath(ctx context.Context, id string) (string, error) {
	doc, err := s.documentRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.uploadDir, doc.CheminStockage), nil
}

// Private helper methods

func (s *service) validateUploadInput(input *UploadDocumentRequest) error {
	if input.TypeDocument == "" {
		return fmt.Errorf("type_document is required")
	}

	validTypes := map[string]bool{
		"PHOTO":       true,
		"PERMIS":      true,
		"CARTE_GRISE": true,
		"ASSURANCE":   true,
		"CONSTAT":     true,
		"PV":          true,
		"AUTRE":       true,
	}
	if !validTypes[input.TypeDocument] {
		return fmt.Errorf("invalid type_document: %s", input.TypeDocument)
	}

	return nil
}

func (s *service) validateUpdateInput(input *UpdateDocumentRequest) error {
	if input.TypeDocument != nil {
		validTypes := map[string]bool{
			"PHOTO":       true,
			"PERMIS":      true,
			"CARTE_GRISE": true,
			"ASSURANCE":   true,
			"CONSTAT":     true,
			"PV":          true,
			"AUTRE":       true,
		}
		if !validTypes[*input.TypeDocument] {
			return fmt.Errorf("invalid type_document: %s", *input.TypeDocument)
		}
	}
	return nil
}

func (s *service) buildFilters(input *ListDocumentsRequest) *repository.DocumentFilters {
	if input == nil {
		return nil
	}

	return &repository.DocumentFilters{
		TypeDocument:   input.TypeDocument,
		Public:         input.Public,
		UploadedByID:   input.UploadedByID,
		ControleID:     input.ControleID,
		InfractionID:   input.InfractionID,
		ProcesVerbalID: input.ProcesVerbalID,
		RecoursID:      input.RecoursID,
		DateDebut:      input.DateDebut,
		DateFin:        input.DateFin,
		Limit:          input.Limit,
		Offset:         input.Offset,
	}
}

func (s *service) entityToResponse(documentEnt *ent.Document) *DocumentResponse {
	response := &DocumentResponse{
		ID:             documentEnt.ID.String(),
		NomFichier:     documentEnt.NomFichier,
		NomOriginal:    documentEnt.NomOriginal,
		TypeMime:       documentEnt.TypeMime,
		Taille:         documentEnt.Taille,
		TailleFormatee: formatFileSize(documentEnt.Taille),
		TypeDocument:   documentEnt.TypeDocument,
		Description:    documentEnt.Description,
		Public:         documentEnt.Public,
		URL:            fmt.Sprintf("%s/api/v1/documents/%s/download", s.baseURL, documentEnt.ID),
		CreatedAt:      documentEnt.CreatedAt,
		UpdatedAt:      documentEnt.UpdatedAt,
	}

	// Ajouter les relations
	if documentEnt.Edges.UploadedBy != nil {
		u := documentEnt.Edges.UploadedBy
		response.UploadedBy = &UploaderSummary{
			ID:        u.ID.String(),
			Matricule: u.Matricule,
			Nom:       u.Nom,
			Prenom:    u.Prenom,
		}
	}

	if documentEnt.Edges.Controle != nil {
		idStr := documentEnt.Edges.Controle.ID.String()
		response.ControleID = &idStr
	}

	if documentEnt.Edges.Infraction != nil {
		idStr := documentEnt.Edges.Infraction.ID.String()
		response.InfractionID = &idStr
	}

	if documentEnt.Edges.ProcesVerbal != nil {
		idStr := documentEnt.Edges.ProcesVerbal.ID.String()
		response.ProcesVerbalID = &idStr
	}

	if documentEnt.Edges.Recours != nil {
		idStr := documentEnt.Edges.Recours.ID.String()
		response.RecoursID = &idStr
	}

	return response
}

// formatFileSize formats file size to human readable string
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f Go", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f Mo", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f Ko", float64(size)/KB)
	default:
		return fmt.Sprintf("%d octets", size)
	}
}
