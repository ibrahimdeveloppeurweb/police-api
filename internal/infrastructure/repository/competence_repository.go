package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/competence"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CompetenceRepository defines competence repository interface
type CompetenceRepository interface {
	Create(ctx context.Context, input *CreateCompetenceInput) (*ent.Competence, error)
	GetByID(ctx context.Context, id string) (*ent.Competence, error)
	GetByNom(ctx context.Context, nom string) (*ent.Competence, error)
	List(ctx context.Context, filters *CompetenceFilters) ([]*ent.Competence, error)
	Update(ctx context.Context, id string, input *UpdateCompetenceInput) (*ent.Competence, error)
	Delete(ctx context.Context, id string) error
	AssignToAgent(ctx context.Context, competenceID, agentID string) error
	RemoveFromAgent(ctx context.Context, competenceID, agentID string) error
	GetByAgent(ctx context.Context, agentID string) ([]*ent.Competence, error)
	GetExpiring(ctx context.Context, daysAhead int) ([]*ent.Competence, error)
}

// CompetenceFilters represents filters for listing competences
type CompetenceFilters struct {
	Type      *string
	Active    *bool
	Search    *string
	Organisme *string
}

// CreateCompetenceInput represents input for creating competence
type CreateCompetenceInput struct {
	ID             string
	Nom            string
	Type           string
	Description    string
	Organisme      string
	DateObtention  *time.Time
	DateExpiration *time.Time
}

// UpdateCompetenceInput represents input for updating competence
type UpdateCompetenceInput struct {
	Nom            *string
	Description    *string
	Organisme      *string
	DateExpiration *time.Time
	Active         *bool
}

// competenceRepository implements CompetenceRepository
type competenceRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewCompetenceRepository creates a new competence repository
func NewCompetenceRepository(client *ent.Client, logger *zap.Logger) CompetenceRepository {
	return &competenceRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new competence
func (r *competenceRepository) Create(ctx context.Context, input *CreateCompetenceInput) (*ent.Competence, error) {
	r.logger.Info("Creating competence", zap.String("nom", input.Nom))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Competence.
		Create().
		SetID(id).
		SetNom(input.Nom).
		SetType(input.Type).
		SetActive(true)

	if input.Description != "" {
		create = create.SetDescription(input.Description)
	}
	if input.Organisme != "" {
		create = create.SetOrganisme(input.Organisme)
	}
	if input.DateObtention != nil {
		create = create.SetDateObtention(*input.DateObtention)
	}
	if input.DateExpiration != nil {
		create = create.SetDateExpiration(*input.DateExpiration)
	}

	comp, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create competence", zap.Error(err))
		return nil, fmt.Errorf("failed to create competence: %w", err)
	}

	return r.GetByID(ctx, comp.ID.String())
}

// GetByID gets competence by ID with all relations
func (r *competenceRepository) GetByID(ctx context.Context, id string) (*ent.Competence, error) {
	uid, _ := uuid.Parse(id)
	comp, err := r.client.Competence.
		Query().
		Where(competence.ID(uid)).
		WithAgents().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("competence not found")
		}
		r.logger.Error("Failed to get competence by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get competence: %w", err)
	}

	return comp, nil
}

// GetByNom gets competence by nom
func (r *competenceRepository) GetByNom(ctx context.Context, nom string) (*ent.Competence, error) {
	comp, err := r.client.Competence.
		Query().
		Where(competence.Nom(nom)).
		WithAgents().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("competence not found")
		}
		r.logger.Error("Failed to get competence by nom", zap.String("nom", nom), zap.Error(err))
		return nil, fmt.Errorf("failed to get competence: %w", err)
	}

	return comp, nil
}

// List gets competences with filters
func (r *competenceRepository) List(ctx context.Context, filters *CompetenceFilters) ([]*ent.Competence, error) {
	query := r.client.Competence.Query().
		WithAgents()

	if filters != nil {
		if filters.Active != nil {
			query = query.Where(competence.Active(*filters.Active))
		}
		if filters.Type != nil && *filters.Type != "" {
			query = query.Where(competence.Type(*filters.Type))
		}
		if filters.Organisme != nil && *filters.Organisme != "" {
			query = query.Where(competence.Organisme(*filters.Organisme))
		}
		if filters.Search != nil && *filters.Search != "" {
			query = query.Where(competence.NomContainsFold(*filters.Search))
		}
	}

	competences, err := query.Order(ent.Asc(competence.FieldNom)).All(ctx)
	if err != nil {
		r.logger.Error("Failed to list competences", zap.Error(err))
		return nil, fmt.Errorf("failed to list competences: %w", err)
	}

	return competences, nil
}

// Update updates competence
func (r *competenceRepository) Update(ctx context.Context, id string, input *UpdateCompetenceInput) (*ent.Competence, error) {
	r.logger.Info("Updating competence", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Competence.UpdateOneID(uid)

	if input.Nom != nil {
		update = update.SetNom(*input.Nom)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Organisme != nil {
		update = update.SetOrganisme(*input.Organisme)
	}
	if input.DateExpiration != nil {
		update = update.SetDateExpiration(*input.DateExpiration)
	}
	if input.Active != nil {
		update = update.SetActive(*input.Active)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update competence", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update competence: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes competence
func (r *competenceRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting competence", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Competence.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete competence", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete competence: %w", err)
	}

	return nil
}

// AssignToAgent assigns a competence to an agent
func (r *competenceRepository) AssignToAgent(ctx context.Context, competenceID, agentID string) error {
	r.logger.Info("Assigning competence to agent", zap.String("competenceID", competenceID), zap.String("agentID", agentID))

	compID, _ := uuid.Parse(competenceID)
	aID, _ := uuid.Parse(agentID)
	_, err := r.client.Competence.UpdateOneID(compID).
		AddAgentIDs(aID).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to assign competence", zap.Error(err))
		return fmt.Errorf("failed to assign competence: %w", err)
	}

	return nil
}

// RemoveFromAgent removes a competence from an agent
func (r *competenceRepository) RemoveFromAgent(ctx context.Context, competenceID, agentID string) error {
	r.logger.Info("Removing competence from agent", zap.String("competenceID", competenceID), zap.String("agentID", agentID))

	compID, _ := uuid.Parse(competenceID)
	aID, _ := uuid.Parse(agentID)
	_, err := r.client.Competence.UpdateOneID(compID).
		RemoveAgentIDs(aID).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to remove competence", zap.Error(err))
		return fmt.Errorf("failed to remove competence: %w", err)
	}

	return nil
}

// GetByAgent gets competences for an agent
func (r *competenceRepository) GetByAgent(ctx context.Context, agentID string) ([]*ent.Competence, error) {
	competences, err := r.client.Competence.Query().
		Where(competence.HasAgentsWith()).
		Where(competence.Active(true)).
		Order(ent.Asc(competence.FieldNom)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get competences by agent", zap.Error(err))
		return nil, fmt.Errorf("failed to get competences: %w", err)
	}

	return competences, nil
}

// GetExpiring gets competences expiring within the specified number of days
func (r *competenceRepository) GetExpiring(ctx context.Context, daysAhead int) ([]*ent.Competence, error) {
	expirationDate := time.Now().AddDate(0, 0, daysAhead)

	competences, err := r.client.Competence.Query().
		Where(
			competence.Active(true),
			competence.DateExpirationLTE(expirationDate),
			competence.DateExpirationGTE(time.Now()),
		).
		WithAgents().
		Order(ent.Asc(competence.FieldDateExpiration)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get expiring competences", zap.Error(err))
		return nil, fmt.Errorf("failed to get expiring competences: %w", err)
	}

	return competences, nil
}
