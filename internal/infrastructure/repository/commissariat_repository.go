package repository

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/commissariat"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CommissariatRepository defines commissariat repository interface
type CommissariatRepository interface {
	Create(ctx context.Context, input *CreateCommissariatInput) (*ent.Commissariat, error)
	GetByID(ctx context.Context, id string) (*ent.Commissariat, error)
	GetByCode(ctx context.Context, code string) (*ent.Commissariat, error)
	List(ctx context.Context, filters *CommissariatFilters) ([]*ent.Commissariat, error)
	Count(ctx context.Context, filters *CommissariatFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateCommissariatInput) (*ent.Commissariat, error)
	Delete(ctx context.Context, id string) error
	GetByRegion(ctx context.Context, region string) ([]*ent.Commissariat, error)
	GetByVille(ctx context.Context, ville string) ([]*ent.Commissariat, error)
}

// CreateCommissariatInput represents input for creating commissariat
type CreateCommissariatInput struct {
	ID        string
	Nom       string
	Code      string
	Adresse   string
	Ville     string
	Region    string
	Telephone string
	Email     *string
	Latitude  *float64
	Longitude *float64
}

// UpdateCommissariatInput represents input for updating commissariat
type UpdateCommissariatInput struct {
	Nom       *string
	Adresse   *string
	Ville     *string
	Region    *string
	Telephone *string
	Email     *string
	Latitude  *float64
	Longitude *float64
	Actif     *bool
}

// CommissariatFilters represents filters for listing commissariats
type CommissariatFilters struct {
	Region string
	Ville  string
	Actif  *bool
	Limit  int
	Offset int
}

// commissariatRepository implements CommissariatRepository
type commissariatRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewCommissariatRepository creates a new commissariat repository
func NewCommissariatRepository(client *ent.Client, logger *zap.Logger) CommissariatRepository {
	return &commissariatRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new commissariat
func (r *commissariatRepository) Create(ctx context.Context, input *CreateCommissariatInput) (*ent.Commissariat, error) {
	r.logger.Info("Creating commissariat", zap.String("code", input.Code))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Commissariat.Create().
		SetID(id).
		SetNom(input.Nom).
		SetCode(input.Code).
		SetAdresse(input.Adresse).
		SetVille(input.Ville).
		SetRegion(input.Region).
		SetTelephone(input.Telephone)

	if input.Email != nil {
		create = create.SetEmail(*input.Email)
	}
	if input.Latitude != nil {
		create = create.SetLatitude(*input.Latitude)
	}
	if input.Longitude != nil {
		create = create.SetLongitude(*input.Longitude)
	}

	comm, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create commissariat", zap.Error(err))
		return nil, fmt.Errorf("failed to create commissariat: %w", err)
	}

	return comm, nil
}

// GetByID gets commissariat by ID
func (r *commissariatRepository) GetByID(ctx context.Context, id string) (*ent.Commissariat, error) {
	uid, _ := uuid.Parse(id)
	comm, err := r.client.Commissariat.
		Query().
		Where(commissariat.ID(uid)).
		WithAgents().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("commissariat not found")
		}
		r.logger.Error("Failed to get commissariat by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get commissariat: %w", err)
	}

	return comm, nil
}

// GetByCode gets commissariat by code
func (r *commissariatRepository) GetByCode(ctx context.Context, code string) (*ent.Commissariat, error) {
	comm, err := r.client.Commissariat.
		Query().
		Where(commissariat.Code(code)).
		WithAgents().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("commissariat not found")
		}
		r.logger.Error("Failed to get commissariat by code", zap.String("code", code), zap.Error(err))
		return nil, fmt.Errorf("failed to get commissariat: %w", err)
	}

	return comm, nil
}

// List gets commissariats with filters
func (r *commissariatRepository) List(ctx context.Context, filters *CommissariatFilters) ([]*ent.Commissariat, error) {
	query := r.client.Commissariat.Query()

	if filters != nil {
		if filters.Region != "" {
			query = query.Where(commissariat.Region(filters.Region))
		}
		if filters.Ville != "" {
			query = query.Where(commissariat.Ville(filters.Ville))
		}
		if filters.Actif != nil {
			query = query.Where(commissariat.Actif(*filters.Actif))
		}
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	commList, err := query.
		WithAgents().
		Order(ent.Asc(commissariat.FieldNom)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list commissariats", zap.Error(err))
		return nil, fmt.Errorf("failed to list commissariats: %w", err)
	}

	return commList, nil
}

// Count counts commissariats with filters
func (r *commissariatRepository) Count(ctx context.Context, filters *CommissariatFilters) (int, error) {
	query := r.client.Commissariat.Query()

	if filters != nil {
		if filters.Region != "" {
			query = query.Where(commissariat.Region(filters.Region))
		}
		if filters.Ville != "" {
			query = query.Where(commissariat.Ville(filters.Ville))
		}
		if filters.Actif != nil {
			query = query.Where(commissariat.Actif(*filters.Actif))
		}
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count commissariats", zap.Error(err))
		return 0, fmt.Errorf("failed to count commissariats: %w", err)
	}

	return count, nil
}

// Update updates commissariat
func (r *commissariatRepository) Update(ctx context.Context, id string, input *UpdateCommissariatInput) (*ent.Commissariat, error) {
	r.logger.Info("Updating commissariat", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Commissariat.UpdateOneID(uid)

	if input.Nom != nil {
		update = update.SetNom(*input.Nom)
	}
	if input.Adresse != nil {
		update = update.SetAdresse(*input.Adresse)
	}
	if input.Ville != nil {
		update = update.SetVille(*input.Ville)
	}
	if input.Region != nil {
		update = update.SetRegion(*input.Region)
	}
	if input.Telephone != nil {
		update = update.SetTelephone(*input.Telephone)
	}
	if input.Email != nil {
		update = update.SetEmail(*input.Email)
	}
	if input.Latitude != nil {
		update = update.SetLatitude(*input.Latitude)
	}
	if input.Longitude != nil {
		update = update.SetLongitude(*input.Longitude)
	}
	if input.Actif != nil {
		update = update.SetActif(*input.Actif)
	}

	comm, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update commissariat", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update commissariat: %w", err)
	}

	return comm, nil
}

// Delete deletes commissariat
func (r *commissariatRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting commissariat", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Commissariat.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete commissariat", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete commissariat: %w", err)
	}

	return nil
}

// GetByRegion gets commissariats by region
func (r *commissariatRepository) GetByRegion(ctx context.Context, region string) ([]*ent.Commissariat, error) {
	commList, err := r.client.Commissariat.
		Query().
		Where(commissariat.Region(region), commissariat.Actif(true)).
		WithAgents().
		Order(ent.Asc(commissariat.FieldNom)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get commissariats by region", zap.String("region", region), zap.Error(err))
		return nil, fmt.Errorf("failed to get commissariats: %w", err)
	}

	return commList, nil
}

// GetByVille gets commissariats by ville
func (r *commissariatRepository) GetByVille(ctx context.Context, ville string) ([]*ent.Commissariat, error) {
	commList, err := r.client.Commissariat.
		Query().
		Where(commissariat.Ville(ville), commissariat.Actif(true)).
		WithAgents().
		Order(ent.Asc(commissariat.FieldNom)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get commissariats by ville", zap.String("ville", ville), zap.Error(err))
		return nil, fmt.Errorf("failed to get commissariats: %w", err)
	}

	return commList, nil
}
