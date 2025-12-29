package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/controle"
	"police-trafic-api-frontend-aligned/ent/document"
	"police-trafic-api-frontend-aligned/ent/infraction"
	"police-trafic-api-frontend-aligned/ent/procesverbal"
	"police-trafic-api-frontend-aligned/ent/recours"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentRepository defines document repository interface
type DocumentRepository interface {
	Create(ctx context.Context, input *CreateDocumentInput) (*ent.Document, error)
	GetByID(ctx context.Context, id string) (*ent.Document, error)
	List(ctx context.Context, filters *DocumentFilters) ([]*ent.Document, error)
	Count(ctx context.Context, filters *DocumentFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateDocumentInput) (*ent.Document, error)
	Delete(ctx context.Context, id string) error
	GetByControle(ctx context.Context, controleID string) ([]*ent.Document, error)
	GetByInfraction(ctx context.Context, infractionID string) ([]*ent.Document, error)
	GetByProcesVerbal(ctx context.Context, pvID string) ([]*ent.Document, error)
	GetByRecours(ctx context.Context, recoursID string) ([]*ent.Document, error)
	GetByUploader(ctx context.Context, userID string) ([]*ent.Document, error)
	GetByType(ctx context.Context, typeDocument string) ([]*ent.Document, error)
}

// CreateDocumentInput represents input for creating document
type CreateDocumentInput struct {
	ID             string
	NomFichier     string
	NomOriginal    string
	TypeMime       string
	Taille         int64
	CheminStockage string
	TypeDocument   string
	Description    *string
	HashFichier    *string
	Public         bool
	UploadedByID   string
	ControleID     *string
	InfractionID   *string
	ProcesVerbalID *string
	RecoursID      *string
}

// UpdateDocumentInput represents input for updating document
type UpdateDocumentInput struct {
	NomOriginal  *string
	TypeDocument *string
	Description  *string
	Public       *bool
}

// DocumentFilters represents filters for listing documents
type DocumentFilters struct {
	TypeDocument   *string
	Public         *bool
	UploadedByID   *string
	ControleID     *string
	InfractionID   *string
	ProcesVerbalID *string
	RecoursID      *string
	DateDebut      *time.Time
	DateFin        *time.Time
	Limit          int
	Offset         int
}

// documentRepository implements DocumentRepository
type documentRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(client *ent.Client, logger *zap.Logger) DocumentRepository {
	return &documentRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new document
func (r *documentRepository) Create(ctx context.Context, input *CreateDocumentInput) (*ent.Document, error) {
	r.logger.Info("Creating document",
		zap.String("nom_fichier", input.NomFichier),
		zap.String("type_document", input.TypeDocument))

	id, _ := uuid.Parse(input.ID)
	uploadedByID, _ := uuid.Parse(input.UploadedByID)
	create := r.client.Document.Create().
		SetID(id).
		SetNomFichier(input.NomFichier).
		SetNomOriginal(input.NomOriginal).
		SetTypeMime(input.TypeMime).
		SetTaille(input.Taille).
		SetCheminStockage(input.CheminStockage).
		SetTypeDocument(input.TypeDocument).
		SetPublic(input.Public).
		SetUploadedByID(uploadedByID)

	if input.Description != nil {
		create = create.SetDescription(*input.Description)
	}
	if input.HashFichier != nil {
		create = create.SetHashFichier(*input.HashFichier)
	}
	if input.ControleID != nil {
		ctrlID, _ := uuid.Parse(*input.ControleID)
		create = create.SetControleID(ctrlID)
	}
	if input.InfractionID != nil {
		infID, _ := uuid.Parse(*input.InfractionID)
		create = create.SetInfractionID(infID)
	}
	if input.ProcesVerbalID != nil {
		pvID, _ := uuid.Parse(*input.ProcesVerbalID)
		create = create.SetProcesVerbalID(pvID)
	}
	if input.RecoursID != nil {
		recID, _ := uuid.Parse(*input.RecoursID)
		create = create.SetRecoursID(recID)
	}

	documentEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create document", zap.Error(err))
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return documentEnt, nil
}

// GetByID gets document by ID
func (r *documentRepository) GetByID(ctx context.Context, id string) (*ent.Document, error) {
	uid, _ := uuid.Parse(id)
	documentEnt, err := r.client.Document.
		Query().
		Where(document.ID(uid)).
		WithUploadedBy().
		WithControle().
		WithInfraction().
		WithProcesVerbal().
		WithRecours().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("document not found")
		}
		r.logger.Error("Failed to get document by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return documentEnt, nil
}

// List gets documents with filters
func (r *documentRepository) List(ctx context.Context, filters *DocumentFilters) ([]*ent.Document, error) {
	query := r.client.Document.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	documents, err := query.
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list documents", zap.Error(err))
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return documents, nil
}

// Count counts documents with filters
func (r *documentRepository) Count(ctx context.Context, filters *DocumentFilters) (int, error) {
	query := r.client.Document.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count documents", zap.Error(err))
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// applyFilters applies filters to document query
func (r *documentRepository) applyFilters(query *ent.DocumentQuery, filters *DocumentFilters) *ent.DocumentQuery {
	if filters.TypeDocument != nil {
		query = query.Where(document.TypeDocument(*filters.TypeDocument))
	}
	if filters.Public != nil {
		query = query.Where(document.Public(*filters.Public))
	}
	if filters.UploadedByID != nil {
		uploaderID, _ := uuid.Parse(*filters.UploadedByID)
		query = query.Where(document.HasUploadedByWith(user.ID(uploaderID)))
	}
	if filters.ControleID != nil {
		ctrlID, _ := uuid.Parse(*filters.ControleID)
		query = query.Where(document.HasControleWith(controle.ID(ctrlID)))
	}
	if filters.InfractionID != nil {
		infID, _ := uuid.Parse(*filters.InfractionID)
		query = query.Where(document.HasInfractionWith(infraction.ID(infID)))
	}
	if filters.ProcesVerbalID != nil {
		pvID, _ := uuid.Parse(*filters.ProcesVerbalID)
		query = query.Where(document.HasProcesVerbalWith(procesverbal.ID(pvID)))
	}
	if filters.RecoursID != nil {
		recID, _ := uuid.Parse(*filters.RecoursID)
		query = query.Where(document.HasRecoursWith(recours.ID(recID)))
	}
	if filters.DateDebut != nil {
		query = query.Where(document.CreatedAtGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(document.CreatedAtLTE(*filters.DateFin))
	}
	return query
}

// Update updates document
func (r *documentRepository) Update(ctx context.Context, id string, input *UpdateDocumentInput) (*ent.Document, error) {
	r.logger.Info("Updating document", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Document.UpdateOneID(uid)

	if input.NomOriginal != nil {
		update = update.SetNomOriginal(*input.NomOriginal)
	}
	if input.TypeDocument != nil {
		update = update.SetTypeDocument(*input.TypeDocument)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Public != nil {
		update = update.SetPublic(*input.Public)
	}

	documentEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update document", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return documentEnt, nil
}

// Delete deletes document
func (r *documentRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting document", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Document.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete document", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// GetByControle gets documents by controle ID
func (r *documentRepository) GetByControle(ctx context.Context, controleID string) ([]*ent.Document, error) {
	ctrlID, _ := uuid.Parse(controleID)
	documents, err := r.client.Document.Query().
		Where(document.HasControleWith(controle.ID(ctrlID))).
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get documents by controle",
			zap.String("controleID", controleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get documents by controle: %w", err)
	}

	return documents, nil
}

// GetByInfraction gets documents by infraction ID
func (r *documentRepository) GetByInfraction(ctx context.Context, infractionID string) ([]*ent.Document, error) {
	infID, _ := uuid.Parse(infractionID)
	documents, err := r.client.Document.Query().
		Where(document.HasInfractionWith(infraction.ID(infID))).
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get documents by infraction",
			zap.String("infractionID", infractionID), zap.Error(err))
		return nil, fmt.Errorf("failed to get documents by infraction: %w", err)
	}

	return documents, nil
}

// GetByProcesVerbal gets documents by proces verbal ID
func (r *documentRepository) GetByProcesVerbal(ctx context.Context, pvID string) ([]*ent.Document, error) {
	uid, _ := uuid.Parse(pvID)
	documents, err := r.client.Document.Query().
		Where(document.HasProcesVerbalWith(procesverbal.ID(uid))).
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get documents by PV",
			zap.String("pvID", pvID), zap.Error(err))
		return nil, fmt.Errorf("failed to get documents by PV: %w", err)
	}

	return documents, nil
}

// GetByRecours gets documents by recours ID
func (r *documentRepository) GetByRecours(ctx context.Context, recoursID string) ([]*ent.Document, error) {
	uid, _ := uuid.Parse(recoursID)
	documents, err := r.client.Document.Query().
		Where(document.HasRecoursWith(recours.ID(uid))).
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get documents by recours",
			zap.String("recoursID", recoursID), zap.Error(err))
		return nil, fmt.Errorf("failed to get documents by recours: %w", err)
	}

	return documents, nil
}

// GetByUploader gets documents by uploader ID
func (r *documentRepository) GetByUploader(ctx context.Context, userID string) ([]*ent.Document, error) {
	uid, _ := uuid.Parse(userID)
	documents, err := r.client.Document.Query().
		Where(document.HasUploadedByWith(user.ID(uid))).
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get documents by uploader",
			zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get documents by uploader: %w", err)
	}

	return documents, nil
}

// GetByType gets documents by type
func (r *documentRepository) GetByType(ctx context.Context, typeDocument string) ([]*ent.Document, error) {
	documents, err := r.client.Document.Query().
		Where(document.TypeDocument(typeDocument)).
		WithUploadedBy().
		Order(ent.Desc(document.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get documents by type",
			zap.String("typeDocument", typeDocument), zap.Error(err))
		return nil, fmt.Errorf("failed to get documents by type: %w", err)
	}

	return documents, nil
}
