package repository

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/observation"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ObservationRepository defines observation repository interface
type ObservationRepository interface {
	Create(ctx context.Context, input *CreateObservationInput) (*ent.Observation, error)
	GetByID(ctx context.Context, id string) (*ent.Observation, error)
	List(ctx context.Context, filters *ObservationFilters) ([]*ent.Observation, error)
	Update(ctx context.Context, id string, input *UpdateObservationInput) (*ent.Observation, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string, visibleOnly bool) ([]*ent.Observation, error)
	GetByAuteur(ctx context.Context, auteurID string) ([]*ent.Observation, error)
}

// ObservationFilters represents filters for listing observations
type ObservationFilters struct {
	AgentID      *string
	AuteurID     *string
	Type         *string
	Categorie    *string
	VisibleAgent *bool
}

// CreateObservationInput represents input for creating observation
type CreateObservationInput struct {
	ID           string
	Contenu      string
	Type         string
	Categorie    string
	VisibleAgent bool
	AgentID      string
	AuteurID     string
}

// UpdateObservationInput represents input for updating observation
type UpdateObservationInput struct {
	Contenu      *string
	Type         *string
	Categorie    *string
	VisibleAgent *bool
}

// observationRepository implements ObservationRepository
type observationRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewObservationRepository creates a new observation repository
func NewObservationRepository(client *ent.Client, logger *zap.Logger) ObservationRepository {
	return &observationRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new observation
func (r *observationRepository) Create(ctx context.Context, input *CreateObservationInput) (*ent.Observation, error) {
	r.logger.Info("Creating observation", zap.String("type", input.Type))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Observation.
		Create().
		SetID(id).
		SetContenu(input.Contenu).
		SetType(input.Type).
		SetVisibleAgent(input.VisibleAgent)

	if input.Categorie != "" {
		create = create.SetCategorie(input.Categorie)
	}
	if input.AgentID != "" {
		agentID, _ := uuid.Parse(input.AgentID)
		create = create.SetAgentID(agentID)
	}
	if input.AuteurID != "" {
		auteurID, _ := uuid.Parse(input.AuteurID)
		create = create.SetAuteurID(auteurID)
	}

	obs, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create observation", zap.Error(err))
		return nil, fmt.Errorf("failed to create observation: %w", err)
	}

	return r.GetByID(ctx, obs.ID.String())
}

// GetByID gets observation by ID with all relations
func (r *observationRepository) GetByID(ctx context.Context, id string) (*ent.Observation, error) {
	uid, _ := uuid.Parse(id)
	obs, err := r.client.Observation.
		Query().
		Where(observation.ID(uid)).
		WithAgent().
		WithAuteur().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("observation not found")
		}
		r.logger.Error("Failed to get observation by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	return obs, nil
}

// List gets observations with filters
func (r *observationRepository) List(ctx context.Context, filters *ObservationFilters) ([]*ent.Observation, error) {
	query := r.client.Observation.Query().
		WithAgent().
		WithAuteur()

	if filters != nil {
		if filters.Type != nil && *filters.Type != "" {
			query = query.Where(observation.Type(*filters.Type))
		}
		if filters.Categorie != nil && *filters.Categorie != "" {
			query = query.Where(observation.Categorie(*filters.Categorie))
		}
		if filters.VisibleAgent != nil {
			query = query.Where(observation.VisibleAgent(*filters.VisibleAgent))
		}
		if filters.AgentID != nil && *filters.AgentID != "" {
			query = query.Where(observation.HasAgentWith())
		}
		if filters.AuteurID != nil && *filters.AuteurID != "" {
			query = query.Where(observation.HasAuteurWith())
		}
	}

	observations, err := query.Order(ent.Desc(observation.FieldCreatedAt)).All(ctx)
	if err != nil {
		r.logger.Error("Failed to list observations", zap.Error(err))
		return nil, fmt.Errorf("failed to list observations: %w", err)
	}

	return observations, nil
}

// Update updates observation
func (r *observationRepository) Update(ctx context.Context, id string, input *UpdateObservationInput) (*ent.Observation, error) {
	r.logger.Info("Updating observation", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Observation.UpdateOneID(uid)

	if input.Contenu != nil {
		update = update.SetContenu(*input.Contenu)
	}
	if input.Type != nil {
		update = update.SetType(*input.Type)
	}
	if input.Categorie != nil {
		update = update.SetCategorie(*input.Categorie)
	}
	if input.VisibleAgent != nil {
		update = update.SetVisibleAgent(*input.VisibleAgent)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update observation", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update observation: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes observation
func (r *observationRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting observation", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Observation.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete observation", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete observation: %w", err)
	}

	return nil
}

// GetByAgent gets observations for an agent
func (r *observationRepository) GetByAgent(ctx context.Context, agentID string, visibleOnly bool) ([]*ent.Observation, error) {
	query := r.client.Observation.Query().
		Where(observation.HasAgentWith()).
		WithAuteur().
		Order(ent.Desc(observation.FieldCreatedAt))

	if visibleOnly {
		query = query.Where(observation.VisibleAgent(true))
	}

	observations, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to get observations by agent", zap.Error(err))
		return nil, fmt.Errorf("failed to get observations: %w", err)
	}

	return observations, nil
}

// GetByAuteur gets observations created by an auteur
func (r *observationRepository) GetByAuteur(ctx context.Context, auteurID string) ([]*ent.Observation, error) {
	observations, err := r.client.Observation.Query().
		Where(observation.HasAuteurWith()).
		WithAgent().
		Order(ent.Desc(observation.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get observations by auteur", zap.Error(err))
		return nil, fmt.Errorf("failed to get observations: %w", err)
	}

	return observations, nil
}
