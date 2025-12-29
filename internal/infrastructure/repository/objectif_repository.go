package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/objectif"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ObjectifRepository defines objectif repository interface
type ObjectifRepository interface {
	Create(ctx context.Context, input *CreateObjectifInput) (*ent.Objectif, error)
	GetByID(ctx context.Context, id string) (*ent.Objectif, error)
	List(ctx context.Context, filters *ObjectifFilters) ([]*ent.Objectif, error)
	Update(ctx context.Context, id string, input *UpdateObjectifInput) (*ent.Objectif, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string) ([]*ent.Objectif, error)
	UpdateProgression(ctx context.Context, id string, valeurActuelle int) (*ent.Objectif, error)
}

// ObjectifFilters represents filters for listing objectifs
type ObjectifFilters struct {
	AgentID   *string
	Periode   *string
	Statut    *string
	DateDebut *time.Time
	DateFin   *time.Time
}

// CreateObjectifInput represents input for creating objectif
type CreateObjectifInput struct {
	ID          string
	Titre       string
	Description string
	Periode     string
	DateDebut   time.Time
	DateFin     time.Time
	ValeurCible int
	AgentID     string
}

// UpdateObjectifInput represents input for updating objectif
type UpdateObjectifInput struct {
	Titre          *string
	Description    *string
	Statut         *string
	ValeurCible    *int
	ValeurActuelle *int
	DateFin        *time.Time
}

// objectifRepository implements ObjectifRepository
type objectifRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewObjectifRepository creates a new objectif repository
func NewObjectifRepository(client *ent.Client, logger *zap.Logger) ObjectifRepository {
	return &objectifRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new objectif
func (r *objectifRepository) Create(ctx context.Context, input *CreateObjectifInput) (*ent.Objectif, error) {
	r.logger.Info("Creating objectif", zap.String("titre", input.Titre))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Objectif.
		Create().
		SetID(id).
		SetTitre(input.Titre).
		SetPeriode(input.Periode).
		SetDateDebut(input.DateDebut).
		SetDateFin(input.DateFin).
		SetStatut("EN_COURS").
		SetValeurActuelle(0).
		SetProgression(0)

	if input.Description != "" {
		create = create.SetDescription(input.Description)
	}
	if input.ValeurCible > 0 {
		create = create.SetValeurCible(input.ValeurCible)
	}
	if input.AgentID != "" {
		agentID, _ := uuid.Parse(input.AgentID)
		create = create.SetAgentID(agentID)
	}

	obj, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create objectif", zap.Error(err))
		return nil, fmt.Errorf("failed to create objectif: %w", err)
	}

	return r.GetByID(ctx, obj.ID.String())
}

// GetByID gets objectif by ID with all relations
func (r *objectifRepository) GetByID(ctx context.Context, id string) (*ent.Objectif, error) {
	uid, _ := uuid.Parse(id)
	obj, err := r.client.Objectif.
		Query().
		Where(objectif.ID(uid)).
		WithAgent().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objectif not found")
		}
		r.logger.Error("Failed to get objectif by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get objectif: %w", err)
	}

	return obj, nil
}

// List gets objectifs with filters
func (r *objectifRepository) List(ctx context.Context, filters *ObjectifFilters) ([]*ent.Objectif, error) {
	query := r.client.Objectif.Query().
		WithAgent()

	if filters != nil {
		if filters.Statut != nil && *filters.Statut != "" {
			query = query.Where(objectif.Statut(*filters.Statut))
		}
		if filters.Periode != nil && *filters.Periode != "" {
			query = query.Where(objectif.Periode(*filters.Periode))
		}
		if filters.AgentID != nil && *filters.AgentID != "" {
			query = query.Where(objectif.HasAgentWith())
		}
		if filters.DateDebut != nil {
			query = query.Where(objectif.DateDebutGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(objectif.DateFinLTE(*filters.DateFin))
		}
	}

	objectifs, err := query.Order(ent.Desc(objectif.FieldCreatedAt)).All(ctx)
	if err != nil {
		r.logger.Error("Failed to list objectifs", zap.Error(err))
		return nil, fmt.Errorf("failed to list objectifs: %w", err)
	}

	return objectifs, nil
}

// Update updates objectif
func (r *objectifRepository) Update(ctx context.Context, id string, input *UpdateObjectifInput) (*ent.Objectif, error) {
	r.logger.Info("Updating objectif", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Objectif.UpdateOneID(uid)

	if input.Titre != nil {
		update = update.SetTitre(*input.Titre)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Statut != nil {
		update = update.SetStatut(*input.Statut)
	}
	if input.ValeurCible != nil {
		update = update.SetValeurCible(*input.ValeurCible)
	}
	if input.ValeurActuelle != nil {
		update = update.SetValeurActuelle(*input.ValeurActuelle)
	}
	if input.DateFin != nil {
		update = update.SetDateFin(*input.DateFin)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update objectif", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update objectif: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes objectif
func (r *objectifRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting objectif", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Objectif.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete objectif", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete objectif: %w", err)
	}

	return nil
}

// GetByAgent gets objectifs for an agent
func (r *objectifRepository) GetByAgent(ctx context.Context, agentID string) ([]*ent.Objectif, error) {
	objectifs, err := r.client.Objectif.Query().
		Where(objectif.HasAgentWith()).
		Order(ent.Desc(objectif.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get objectifs by agent", zap.Error(err))
		return nil, fmt.Errorf("failed to get objectifs: %w", err)
	}

	return objectifs, nil
}

// UpdateProgression updates the progression of an objectif
func (r *objectifRepository) UpdateProgression(ctx context.Context, id string, valeurActuelle int) (*ent.Objectif, error) {
	r.logger.Info("Updating objectif progression", zap.String("id", id), zap.Int("valeurActuelle", valeurActuelle))

	// Get current objectif to calculate progression
	obj, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	progression := float64(0)
	if obj.ValeurCible > 0 {
		progression = float64(valeurActuelle) / float64(obj.ValeurCible) * 100
		if progression > 100 {
			progression = 100
		}
	}

	// Determine status based on progression
	statut := obj.Statut
	if progression >= 100 {
		statut = "ATTEINT"
	} else if time.Now().After(obj.DateFin) {
		statut = "NON_ATTEINT"
	}

	uid, _ := uuid.Parse(id)
	_, err = r.client.Objectif.UpdateOneID(uid).
		SetValeurActuelle(valeurActuelle).
		SetProgression(progression).
		SetStatut(statut).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to update progression", zap.Error(err))
		return nil, fmt.Errorf("failed to update progression: %w", err)
	}

	return r.GetByID(ctx, id)
}
