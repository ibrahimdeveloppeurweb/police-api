package repository

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/infractiontype"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InfractionTypeRepository defines infraction type repository interface
type InfractionTypeRepository interface {
	Create(ctx context.Context, input *CreateInfractionTypeInput) (*ent.InfractionType, error)
	GetByID(ctx context.Context, id string) (*ent.InfractionType, error)
	GetByCode(ctx context.Context, code string) (*ent.InfractionType, error)
	List(ctx context.Context, filters *InfractionTypeFilters) ([]*ent.InfractionType, error)
	Update(ctx context.Context, id string, input *UpdateInfractionTypeInput) (*ent.InfractionType, error)
	Delete(ctx context.Context, id string) error
	GetByCategorie(ctx context.Context, categorie string) ([]*ent.InfractionType, error)
	GetActive(ctx context.Context) ([]*ent.InfractionType, error)
	GetCategories(ctx context.Context) ([]string, error)
}

// CreateInfractionTypeInput represents input for creating infraction type
type CreateInfractionTypeInput struct {
	ID          string
	Code        string
	Libelle     string
	Description *string
	Amende      float64
	Points      int
	Categorie   string
	Active      bool
}

// UpdateInfractionTypeInput represents input for updating infraction type
type UpdateInfractionTypeInput struct {
	Code        *string
	Libelle     *string
	Description *string
	Amende      *float64
	Points      *int
	Categorie   *string
	Active      *bool
}

// InfractionTypeFilters represents filters for listing infraction types
type InfractionTypeFilters struct {
	Categorie *string
	Active    *bool
	Limit     int
	Offset    int
}

// infractionTypeRepository implements InfractionTypeRepository
type infractionTypeRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewInfractionTypeRepository creates a new infraction type repository
func NewInfractionTypeRepository(client *ent.Client, logger *zap.Logger) InfractionTypeRepository {
	return &infractionTypeRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new infraction type
func (r *infractionTypeRepository) Create(ctx context.Context, input *CreateInfractionTypeInput) (*ent.InfractionType, error) {
	r.logger.Info("Creating infraction type", zap.String("code", input.Code))

	id, _ := uuid.Parse(input.ID)
	create := r.client.InfractionType.Create().
		SetID(id).
		SetCode(input.Code).
		SetLibelle(input.Libelle).
		SetAmende(input.Amende).
		SetPoints(input.Points).
		SetCategorie(input.Categorie).
		SetActive(input.Active)

	if input.Description != nil {
		create = create.SetDescription(*input.Description)
	}

	typeEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create infraction type", zap.Error(err))
		return nil, fmt.Errorf("failed to create infraction type: %w", err)
	}

	return typeEnt, nil
}

// GetByID gets infraction type by ID
func (r *infractionTypeRepository) GetByID(ctx context.Context, id string) (*ent.InfractionType, error) {
	uid, _ := uuid.Parse(id)
	typeEnt, err := r.client.InfractionType.
		Query().
		Where(infractiontype.ID(uid)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("infraction type not found")
		}
		r.logger.Error("Failed to get infraction type by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get infraction type: %w", err)
	}

	return typeEnt, nil
}

// GetByCode gets infraction type by code
func (r *infractionTypeRepository) GetByCode(ctx context.Context, code string) (*ent.InfractionType, error) {
	typeEnt, err := r.client.InfractionType.
		Query().
		Where(infractiontype.Code(code)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("infraction type not found")
		}
		r.logger.Error("Failed to get infraction type by code", zap.String("code", code), zap.Error(err))
		return nil, fmt.Errorf("failed to get infraction type: %w", err)
	}

	return typeEnt, nil
}

// List gets infraction types with filters
func (r *infractionTypeRepository) List(ctx context.Context, filters *InfractionTypeFilters) ([]*ent.InfractionType, error) {
	query := r.client.InfractionType.Query()

	if filters != nil {
		if filters.Categorie != nil {
			query = query.Where(infractiontype.Categorie(*filters.Categorie))
		}
		if filters.Active != nil {
			query = query.Where(infractiontype.Active(*filters.Active))
		}
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	types, err := query.
		Order(ent.Asc(infractiontype.FieldCategorie), ent.Asc(infractiontype.FieldCode)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list infraction types", zap.Error(err))
		return nil, fmt.Errorf("failed to list infraction types: %w", err)
	}

	return types, nil
}

// Update updates infraction type
func (r *infractionTypeRepository) Update(ctx context.Context, id string, input *UpdateInfractionTypeInput) (*ent.InfractionType, error) {
	r.logger.Info("Updating infraction type", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.InfractionType.UpdateOneID(uid)

	if input.Code != nil {
		update = update.SetCode(*input.Code)
	}
	if input.Libelle != nil {
		update = update.SetLibelle(*input.Libelle)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Amende != nil {
		update = update.SetAmende(*input.Amende)
	}
	if input.Points != nil {
		update = update.SetPoints(*input.Points)
	}
	if input.Categorie != nil {
		update = update.SetCategorie(*input.Categorie)
	}
	if input.Active != nil {
		update = update.SetActive(*input.Active)
	}

	typeEnt, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("infraction type not found")
		}
		r.logger.Error("Failed to update infraction type", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update infraction type: %w", err)
	}

	return typeEnt, nil
}

// Delete deletes infraction type
func (r *infractionTypeRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting infraction type", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.InfractionType.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("infraction type not found")
		}
		r.logger.Error("Failed to delete infraction type", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete infraction type: %w", err)
	}

	return nil
}

// GetByCategorie gets infraction types by category
func (r *infractionTypeRepository) GetByCategorie(ctx context.Context, categorie string) ([]*ent.InfractionType, error) {
	types, err := r.client.InfractionType.Query().
		Where(infractiontype.Categorie(categorie)).
		Where(infractiontype.Active(true)).
		Order(ent.Asc(infractiontype.FieldCode)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get infraction types by category", zap.String("categorie", categorie), zap.Error(err))
		return nil, fmt.Errorf("failed to get infraction types by category: %w", err)
	}

	return types, nil
}

// GetActive gets all active infraction types
func (r *infractionTypeRepository) GetActive(ctx context.Context) ([]*ent.InfractionType, error) {
	types, err := r.client.InfractionType.Query().
		Where(infractiontype.Active(true)).
		Order(ent.Asc(infractiontype.FieldCategorie), ent.Asc(infractiontype.FieldCode)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get active infraction types", zap.Error(err))
		return nil, fmt.Errorf("failed to get active infraction types: %w", err)
	}

	return types, nil
}

// GetCategories gets all distinct categories
func (r *infractionTypeRepository) GetCategories(ctx context.Context) ([]string, error) {
	types, err := r.client.InfractionType.Query().
		Where(infractiontype.Active(true)).
		Select(infractiontype.FieldCategorie).
		GroupBy(infractiontype.FieldCategorie).
		Strings(ctx)

	if err != nil {
		r.logger.Error("Failed to get categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return types, nil
}
