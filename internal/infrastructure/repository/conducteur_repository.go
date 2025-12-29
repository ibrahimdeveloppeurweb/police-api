package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/conducteur"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConducteurRepository defines conducteur repository interface
type ConducteurRepository interface {
	Create(ctx context.Context, input *CreateConducteurInput) (*ent.Conducteur, error)
	GetByID(ctx context.Context, id string) (*ent.Conducteur, error)
	GetByNumeroPermis(ctx context.Context, numeroPermis string) (*ent.Conducteur, error)
	List(ctx context.Context, filters *ConducteurFilters) ([]*ent.Conducteur, error)
	Update(ctx context.Context, id string, input *UpdateConducteurInput) (*ent.Conducteur, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]*ent.Conducteur, error)
	GetByNomPrenom(ctx context.Context, nom, prenom string) ([]*ent.Conducteur, error)
	GetByEmail(ctx context.Context, email string) (*ent.Conducteur, error)
}

// CreateConducteurInput represents input for creating conducteur
type CreateConducteurInput struct {
	ID                  string
	Nom                 string
	Prenom              string
	DateNaissance       time.Time
	LieuNaissance       *string
	Adresse             *string
	CodePostal          *string
	Ville               *string
	Telephone           *string
	Email               *string
	NumeroPermis        *string
	PermisDelivreLe     *time.Time
	PermisValideJusqu   *time.Time
	CategoriesPermis    *string
	PointsPermis        int
	Nationalite         string
}

// UpdateConducteurInput represents input for updating conducteur
type UpdateConducteurInput struct {
	Nom                 *string
	Prenom              *string
	DateNaissance       *time.Time
	LieuNaissance       *string
	Adresse             *string
	CodePostal          *string
	Ville               *string
	Telephone           *string
	Email               *string
	NumeroPermis        *string
	PermisDelivreLe     *time.Time
	PermisValideJusqu   *time.Time
	CategoriesPermis    *string
	PointsPermis        *int
	Nationalite         *string
	Active              *bool
}

// ConducteurFilters represents filters for listing conducteurs
type ConducteurFilters struct {
	Nom         *string
	Prenom      *string
	Ville       *string
	Nationalite *string
	Active      *bool
	Limit       int
	Offset      int
}

// conducteurRepository implements ConducteurRepository
type conducteurRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewConducteurRepository creates a new conducteur repository
func NewConducteurRepository(client *ent.Client, logger *zap.Logger) ConducteurRepository {
	return &conducteurRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new conducteur
func (r *conducteurRepository) Create(ctx context.Context, input *CreateConducteurInput) (*ent.Conducteur, error) {
	r.logger.Info("Creating conducteur", 
		zap.String("nom", input.Nom), zap.String("prenom", input.Prenom))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Conducteur.Create().
		SetID(id).
		SetNom(input.Nom).
		SetPrenom(input.Prenom).
		SetDateNaissance(input.DateNaissance).
		SetPointsPermis(input.PointsPermis).
		SetNationalite(input.Nationalite)

	if input.LieuNaissance != nil {
		create = create.SetLieuNaissance(*input.LieuNaissance)
	}
	if input.Adresse != nil {
		create = create.SetAdresse(*input.Adresse)
	}
	if input.CodePostal != nil {
		create = create.SetCodePostal(*input.CodePostal)
	}
	if input.Ville != nil {
		create = create.SetVille(*input.Ville)
	}
	if input.Telephone != nil {
		create = create.SetTelephone(*input.Telephone)
	}
	if input.Email != nil {
		create = create.SetEmail(*input.Email)
	}
	if input.NumeroPermis != nil {
		create = create.SetNumeroPermis(*input.NumeroPermis)
	}
	if input.PermisDelivreLe != nil {
		create = create.SetPermisDelivreLe(*input.PermisDelivreLe)
	}
	if input.PermisValideJusqu != nil {
		create = create.SetPermisValideJusqu(*input.PermisValideJusqu)
	}
	if input.CategoriesPermis != nil {
		create = create.SetCategoriesPermis(*input.CategoriesPermis)
	}

	conducteurEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create conducteur", zap.Error(err))
		return nil, fmt.Errorf("failed to create conducteur: %w", err)
	}

	return conducteurEnt, nil
}

// GetByID gets conducteur by ID
func (r *conducteurRepository) GetByID(ctx context.Context, id string) (*ent.Conducteur, error) {
	uid, _ := uuid.Parse(id)
	conducteurEnt, err := r.client.Conducteur.
		Query().
		Where(conducteur.ID(uid)).
		WithControles().
		WithInfractions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("conducteur not found")
		}
		r.logger.Error("Failed to get conducteur by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get conducteur: %w", err)
	}

	return conducteurEnt, nil
}

// GetByNumeroPermis gets conducteur by numero permis
func (r *conducteurRepository) GetByNumeroPermis(ctx context.Context, numeroPermis string) (*ent.Conducteur, error) {
	conducteurEnt, err := r.client.Conducteur.
		Query().
		Where(conducteur.NumeroPermis(numeroPermis)).
		WithControles().
		WithInfractions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("conducteur not found")
		}
		r.logger.Error("Failed to get conducteur by numero permis", 
			zap.String("numeroPermis", numeroPermis), zap.Error(err))
		return nil, fmt.Errorf("failed to get conducteur: %w", err)
	}

	return conducteurEnt, nil
}

// List gets conducteurs with filters
func (r *conducteurRepository) List(ctx context.Context, filters *ConducteurFilters) ([]*ent.Conducteur, error) {
	query := r.client.Conducteur.Query()

	if filters != nil {
		if filters.Nom != nil {
			query = query.Where(conducteur.NomContains(*filters.Nom))
		}
		if filters.Prenom != nil {
			query = query.Where(conducteur.PrenomContains(*filters.Prenom))
		}
		if filters.Ville != nil {
			query = query.Where(conducteur.VilleContains(*filters.Ville))
		}
		if filters.Nationalite != nil {
			query = query.Where(conducteur.Nationalite(*filters.Nationalite))
		}
		if filters.Active != nil {
			query = query.Where(conducteur.Active(*filters.Active))
		}

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	conducteurs, err := query.
		WithControles().
		Order(ent.Desc(conducteur.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list conducteurs", zap.Error(err))
		return nil, fmt.Errorf("failed to list conducteurs: %w", err)
	}

	return conducteurs, nil
}

// Update updates conducteur
func (r *conducteurRepository) Update(ctx context.Context, id string, input *UpdateConducteurInput) (*ent.Conducteur, error) {
	r.logger.Info("Updating conducteur", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Conducteur.UpdateOneID(uid)

	if input.Nom != nil {
		update = update.SetNom(*input.Nom)
	}
	if input.Prenom != nil {
		update = update.SetPrenom(*input.Prenom)
	}
	if input.DateNaissance != nil {
		update = update.SetDateNaissance(*input.DateNaissance)
	}
	if input.LieuNaissance != nil {
		update = update.SetLieuNaissance(*input.LieuNaissance)
	}
	if input.Adresse != nil {
		update = update.SetAdresse(*input.Adresse)
	}
	if input.CodePostal != nil {
		update = update.SetCodePostal(*input.CodePostal)
	}
	if input.Ville != nil {
		update = update.SetVille(*input.Ville)
	}
	if input.Telephone != nil {
		update = update.SetTelephone(*input.Telephone)
	}
	if input.Email != nil {
		update = update.SetEmail(*input.Email)
	}
	if input.NumeroPermis != nil {
		update = update.SetNumeroPermis(*input.NumeroPermis)
	}
	if input.PermisDelivreLe != nil {
		update = update.SetPermisDelivreLe(*input.PermisDelivreLe)
	}
	if input.PermisValideJusqu != nil {
		update = update.SetPermisValideJusqu(*input.PermisValideJusqu)
	}
	if input.CategoriesPermis != nil {
		update = update.SetCategoriesPermis(*input.CategoriesPermis)
	}
	if input.PointsPermis != nil {
		update = update.SetPointsPermis(*input.PointsPermis)
	}
	if input.Nationalite != nil {
		update = update.SetNationalite(*input.Nationalite)
	}
	if input.Active != nil {
		update = update.SetActive(*input.Active)
	}

	conducteurEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update conducteur", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update conducteur: %w", err)
	}

	return conducteurEnt, nil
}

// Delete deletes conducteur
func (r *conducteurRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting conducteur", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Conducteur.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete conducteur", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete conducteur: %w", err)
	}

	return nil
}

// Search searches conducteurs by query
func (r *conducteurRepository) Search(ctx context.Context, query string) ([]*ent.Conducteur, error) {
	searchQuery := strings.ToLower(strings.TrimSpace(query))
	if searchQuery == "" {
		return []*ent.Conducteur{}, nil
	}

	conducteurs, err := r.client.Conducteur.Query().
		Where(
			conducteur.Or(
				conducteur.NomContains(searchQuery),
				conducteur.PrenomContains(searchQuery),
				conducteur.NumeroPermisContains(searchQuery),
				conducteur.EmailContains(searchQuery),
				conducteur.TelephoneContains(searchQuery),
			),
		).
		WithControles().
		WithInfractions().
		Order(ent.Desc(conducteur.FieldCreatedAt)).
		Limit(50).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to search conducteurs", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("failed to search conducteurs: %w", err)
	}

	return conducteurs, nil
}

// GetByNomPrenom gets conducteurs by nom and prenom
func (r *conducteurRepository) GetByNomPrenom(ctx context.Context, nom, prenom string) ([]*ent.Conducteur, error) {
	conducteurs, err := r.client.Conducteur.Query().
		Where(
			conducteur.And(
				conducteur.NomEqualFold(nom),
				conducteur.PrenomEqualFold(prenom),
			),
		).
		WithControles().
		WithInfractions().
		Order(ent.Desc(conducteur.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get conducteurs by nom/prenom", 
			zap.String("nom", nom), zap.String("prenom", prenom), zap.Error(err))
		return nil, fmt.Errorf("failed to get conducteurs by nom/prenom: %w", err)
	}

	return conducteurs, nil
}

// GetByEmail gets conducteur by email
func (r *conducteurRepository) GetByEmail(ctx context.Context, email string) (*ent.Conducteur, error) {
	conducteurEnt, err := r.client.Conducteur.
		Query().
		Where(conducteur.Email(email)).
		WithControles().
		WithInfractions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("conducteur not found")
		}
		r.logger.Error("Failed to get conducteur by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("failed to get conducteur: %w", err)
	}

	return conducteurEnt, nil
}