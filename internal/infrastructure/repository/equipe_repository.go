package repository

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/equipe"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EquipeRepository defines equipe repository interface
type EquipeRepository interface {
	Create(ctx context.Context, input *CreateEquipeInput) (*ent.Equipe, error)
	GetByID(ctx context.Context, id string) (*ent.Equipe, error)
	GetByCode(ctx context.Context, code string) (*ent.Equipe, error)
	List(ctx context.Context, filters *EquipeFilters) ([]*ent.Equipe, error)
	Update(ctx context.Context, id string, input *UpdateEquipeInput) (*ent.Equipe, error)
	Delete(ctx context.Context, id string) error
	AddMembre(ctx context.Context, equipeID, userID string) error
	RemoveMembre(ctx context.Context, equipeID, userID string) error
	SetChefEquipe(ctx context.Context, equipeID, userID string) error
}

// EquipeFilters represents filters for listing equipes
type EquipeFilters struct {
	CommissariatID *string
	Active         *bool
	Search         *string
}

// CreateEquipeInput represents input for creating equipe
type CreateEquipeInput struct {
	ID             string
	Nom            string
	Code           string
	Zone           string
	Description    string
	CommissariatID string
}

// UpdateEquipeInput represents input for updating equipe
type UpdateEquipeInput struct {
	Nom         *string
	Zone        *string
	Description *string
	Active      *bool
}

// equipeRepository implements EquipeRepository
type equipeRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewEquipeRepository creates a new equipe repository
func NewEquipeRepository(client *ent.Client, logger *zap.Logger) EquipeRepository {
	return &equipeRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new equipe
func (r *equipeRepository) Create(ctx context.Context, input *CreateEquipeInput) (*ent.Equipe, error) {
	r.logger.Info("Creating equipe", zap.String("code", input.Code))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Equipe.
		Create().
		SetID(id).
		SetNom(input.Nom).
		SetCode(input.Code).
		SetActive(true)

	if input.Zone != "" {
		create = create.SetZone(input.Zone)
	}
	if input.Description != "" {
		create = create.SetDescription(input.Description)
	}
	if input.CommissariatID != "" {
		commID, _ := uuid.Parse(input.CommissariatID)
		create = create.SetCommissariatID(commID)
	}

	eq, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create equipe", zap.Error(err))
		return nil, fmt.Errorf("failed to create equipe: %w", err)
	}

	return r.GetByID(ctx, eq.ID.String())
}

// GetByID gets equipe by ID with all relations
func (r *equipeRepository) GetByID(ctx context.Context, id string) (*ent.Equipe, error) {
	uid, _ := uuid.Parse(id)
	eq, err := r.client.Equipe.
		Query().
		Where(equipe.ID(uid)).
		WithCommissariat().
		WithChefEquipe().
		WithMembres().
		WithMissions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("equipe not found")
		}
		r.logger.Error("Failed to get equipe by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get equipe: %w", err)
	}

	return eq, nil
}

// GetByCode gets equipe by code
func (r *equipeRepository) GetByCode(ctx context.Context, code string) (*ent.Equipe, error) {
	eq, err := r.client.Equipe.
		Query().
		Where(equipe.Code(code)).
		WithCommissariat().
		WithChefEquipe().
		WithMembres().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("equipe not found")
		}
		r.logger.Error("Failed to get equipe by code", zap.String("code", code), zap.Error(err))
		return nil, fmt.Errorf("failed to get equipe: %w", err)
	}

	return eq, nil
}

// List gets equipes with filters
func (r *equipeRepository) List(ctx context.Context, filters *EquipeFilters) ([]*ent.Equipe, error) {
	query := r.client.Equipe.Query().
		WithCommissariat().
		WithChefEquipe().
		WithMembres()

	if filters != nil {
		if filters.Active != nil {
			query = query.Where(equipe.Active(*filters.Active))
		}
		if filters.CommissariatID != nil && *filters.CommissariatID != "" {
			query = query.Where(equipe.HasCommissariatWith())
		}
		if filters.Search != nil && *filters.Search != "" {
			query = query.Where(
				equipe.Or(
					equipe.NomContainsFold(*filters.Search),
					equipe.CodeContainsFold(*filters.Search),
				),
			)
		}
	}

	equipes, err := query.Order(ent.Asc(equipe.FieldNom)).All(ctx)
	if err != nil {
		r.logger.Error("Failed to list equipes", zap.Error(err))
		return nil, fmt.Errorf("failed to list equipes: %w", err)
	}

	return equipes, nil
}

// Update updates equipe
func (r *equipeRepository) Update(ctx context.Context, id string, input *UpdateEquipeInput) (*ent.Equipe, error) {
	r.logger.Info("Updating equipe", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Equipe.UpdateOneID(uid)

	if input.Nom != nil {
		update = update.SetNom(*input.Nom)
	}
	if input.Zone != nil {
		update = update.SetZone(*input.Zone)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Active != nil {
		update = update.SetActive(*input.Active)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update equipe", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update equipe: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes equipe
func (r *equipeRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting equipe", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Equipe.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete equipe", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete equipe: %w", err)
	}

	return nil
}

// AddMembre adds a user to the equipe
func (r *equipeRepository) AddMembre(ctx context.Context, equipeID, userID string) error {
	r.logger.Info("Adding membre to equipe", zap.String("equipeID", equipeID), zap.String("userID", userID))

	eqID, _ := uuid.Parse(equipeID)
	uID, _ := uuid.Parse(userID)
	_, err := r.client.Equipe.UpdateOneID(eqID).
		AddMembreIDs(uID).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to add membre", zap.Error(err))
		return fmt.Errorf("failed to add membre: %w", err)
	}

	return nil
}

// RemoveMembre removes a user from the equipe
func (r *equipeRepository) RemoveMembre(ctx context.Context, equipeID, userID string) error {
	r.logger.Info("Removing membre from equipe", zap.String("equipeID", equipeID), zap.String("userID", userID))

	eqID, _ := uuid.Parse(equipeID)
	uID, _ := uuid.Parse(userID)
	_, err := r.client.Equipe.UpdateOneID(eqID).
		RemoveMembreIDs(uID).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to remove membre", zap.Error(err))
		return fmt.Errorf("failed to remove membre: %w", err)
	}

	return nil
}

// SetChefEquipe sets the chef d'equipe
func (r *equipeRepository) SetChefEquipe(ctx context.Context, equipeID, userID string) error {
	r.logger.Info("Setting chef equipe", zap.String("equipeID", equipeID), zap.String("userID", userID))

	eqID, _ := uuid.Parse(equipeID)
	uID, _ := uuid.Parse(userID)
	_, err := r.client.Equipe.UpdateOneID(eqID).
		SetChefEquipeID(uID).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to set chef equipe", zap.Error(err))
		return fmt.Errorf("failed to set chef equipe: %w", err)
	}

	return nil
}
