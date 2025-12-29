package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/mission"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MissionRepository defines mission repository interface
type MissionRepository interface {
	Create(ctx context.Context, input *CreateMissionInput) (*ent.Mission, error)
	GetByID(ctx context.Context, id string) (*ent.Mission, error)
	List(ctx context.Context, filters *MissionFilters) ([]*ent.Mission, error)
	Update(ctx context.Context, id string, input *UpdateMissionInput) (*ent.Mission, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string, limit int) ([]*ent.Mission, error)
	GetByEquipe(ctx context.Context, equipeID string) ([]*ent.Mission, error)
	StartMission(ctx context.Context, id string) (*ent.Mission, error)
	EndMission(ctx context.Context, id string, rapport string) (*ent.Mission, error)
	CancelMission(ctx context.Context, id string, raison string) (*ent.Mission, error)
	AddAgents(ctx context.Context, missionID string, agentIDs []string) (*ent.Mission, error)
	RemoveAgent(ctx context.Context, missionID string, agentID string) (*ent.Mission, error)
}

// MissionFilters represents filters for listing missions
type MissionFilters struct {
	AgentID        *string
	EquipeID       *string
	CommissariatID *string
	Statut         *string
	Type           *string
	DateDebut      *time.Time
	DateFin        *time.Time
}

// CreateMissionInput represents input for creating mission
type CreateMissionInput struct {
	ID             string
	Type           string
	Titre          string
	DateDebut      time.Time
	DateFin        *time.Time
	Duree          string
	Zone           string
	AgentIDs       []string
	CommissariatID string
	EquipeID       string
}

// UpdateMissionInput represents input for updating mission
type UpdateMissionInput struct {
	Titre   *string
	Zone    *string
	Duree   *string
	Statut  *string
	Rapport *string
	DateFin *time.Time
}

// missionRepository implements MissionRepository
type missionRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewMissionRepository creates a new mission repository
func NewMissionRepository(client *ent.Client, logger *zap.Logger) MissionRepository {
	return &missionRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new mission
func (r *missionRepository) Create(ctx context.Context, input *CreateMissionInput) (*ent.Mission, error) {
	r.logger.Info("Creating mission", zap.String("type", input.Type))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Mission.
		Create().
		SetID(id).
		SetType(input.Type).
		SetDateDebut(input.DateDebut).
		SetStatut("PLANIFIEE")

	if input.Titre != "" {
		create = create.SetTitre(input.Titre)
	}
	if input.Duree != "" {
		create = create.SetDuree(input.Duree)
	}
	if input.Zone != "" {
		create = create.SetZone(input.Zone)
	}
	if input.DateFin != nil {
		create = create.SetDateFin(*input.DateFin)
	}
	if len(input.AgentIDs) > 0 {
		agentUUIDs := make([]uuid.UUID, len(input.AgentIDs))
		for i, aid := range input.AgentIDs {
			agentUUIDs[i], _ = uuid.Parse(aid)
		}
		create = create.AddAgentIDs(agentUUIDs...)
	}
	if input.CommissariatID != "" {
		commID, _ := uuid.Parse(input.CommissariatID)
		create = create.SetCommissariatID(commID)
	}
	if input.EquipeID != "" {
		eqID, _ := uuid.Parse(input.EquipeID)
		create = create.SetEquipeID(eqID)
	}

	m, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create mission", zap.Error(err))
		return nil, fmt.Errorf("failed to create mission: %w", err)
	}

	return r.GetByID(ctx, m.ID.String())
}

// GetByID gets mission by ID with all relations
func (r *missionRepository) GetByID(ctx context.Context, id string) (*ent.Mission, error) {
	uid, _ := uuid.Parse(id)
	m, err := r.client.Mission.
		Query().
		Where(mission.ID(uid)).
		WithAgents().
		WithCommissariat().
		WithEquipe().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("mission not found")
		}
		r.logger.Error("Failed to get mission by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}

	return m, nil
}

// List gets missions with filters
func (r *missionRepository) List(ctx context.Context, filters *MissionFilters) ([]*ent.Mission, error) {
	query := r.client.Mission.Query().
		WithAgents().
		WithCommissariat().
		WithEquipe()

	if filters != nil {
		if filters.Statut != nil && *filters.Statut != "" {
			query = query.Where(mission.Statut(*filters.Statut))
		}
		if filters.Type != nil && *filters.Type != "" {
			query = query.Where(mission.Type(*filters.Type))
		}
		if filters.AgentID != nil && *filters.AgentID != "" {
			query = query.Where(mission.HasAgentsWith())
		}
		if filters.DateDebut != nil {
			query = query.Where(mission.DateDebutGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(mission.DateDebutLTE(*filters.DateFin))
		}
	}

	missions, err := query.Order(ent.Desc(mission.FieldDateDebut)).All(ctx)
	if err != nil {
		r.logger.Error("Failed to list missions", zap.Error(err))
		return nil, fmt.Errorf("failed to list missions: %w", err)
	}

	return missions, nil
}

// Update updates mission
func (r *missionRepository) Update(ctx context.Context, id string, input *UpdateMissionInput) (*ent.Mission, error) {
	r.logger.Info("Updating mission", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Mission.UpdateOneID(uid)

	if input.Titre != nil {
		update = update.SetTitre(*input.Titre)
	}
	if input.Zone != nil {
		update = update.SetZone(*input.Zone)
	}
	if input.Duree != nil {
		update = update.SetDuree(*input.Duree)
	}
	if input.Statut != nil {
		update = update.SetStatut(*input.Statut)
	}
	if input.Rapport != nil {
		update = update.SetRapport(*input.Rapport)
	}
	if input.DateFin != nil {
		update = update.SetDateFin(*input.DateFin)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update mission", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update mission: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes mission
func (r *missionRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting mission", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Mission.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete mission", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete mission: %w", err)
	}

	return nil
}

// GetByAgent gets missions for an agent
func (r *missionRepository) GetByAgent(ctx context.Context, agentID string, limit int) ([]*ent.Mission, error) {
	query := r.client.Mission.Query().
		Where(mission.HasAgentsWith()).
		WithAgents().
		WithEquipe().
		WithCommissariat().
		Order(ent.Desc(mission.FieldDateDebut))

	if limit > 0 {
		query = query.Limit(limit)
	}

	missions, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to get missions by agent", zap.Error(err))
		return nil, fmt.Errorf("failed to get missions: %w", err)
	}

	return missions, nil
}

// GetByEquipe gets missions for an equipe
func (r *missionRepository) GetByEquipe(ctx context.Context, equipeID string) ([]*ent.Mission, error) {
	missions, err := r.client.Mission.Query().
		Where(mission.HasEquipeWith()).
		WithAgents().
		WithCommissariat().
		Order(ent.Desc(mission.FieldDateDebut)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get missions by equipe", zap.Error(err))
		return nil, fmt.Errorf("failed to get missions: %w", err)
	}

	return missions, nil
}

// StartMission starts a mission
func (r *missionRepository) StartMission(ctx context.Context, id string) (*ent.Mission, error) {
	r.logger.Info("Starting mission", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	_, err := r.client.Mission.UpdateOneID(uid).
		SetStatut("EN_COURS").
		SetDateDebut(time.Now()).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to start mission", zap.Error(err))
		return nil, fmt.Errorf("failed to start mission: %w", err)
	}

	return r.GetByID(ctx, id)
}

// EndMission ends a mission with a rapport
func (r *missionRepository) EndMission(ctx context.Context, id string, rapport string) (*ent.Mission, error) {
	r.logger.Info("Ending mission", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	_, err := r.client.Mission.UpdateOneID(uid).
		SetStatut("TERMINEE").
		SetDateFin(time.Now()).
		SetRapport(rapport).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to end mission", zap.Error(err))
		return nil, fmt.Errorf("failed to end mission: %w", err)
	}

	return r.GetByID(ctx, id)
}

// CancelMission cancels a mission
func (r *missionRepository) CancelMission(ctx context.Context, id string, raison string) (*ent.Mission, error) {
	r.logger.Info("Cancelling mission", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Mission.UpdateOneID(uid).
		SetStatut("ANNULEE").
		SetDateFin(time.Now())

	if raison != "" {
		update = update.SetRapport("ANNULEE: " + raison)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to cancel mission", zap.Error(err))
		return nil, fmt.Errorf("failed to cancel mission: %w", err)
	}

	return r.GetByID(ctx, id)
}

// AddAgents adds agents to a mission
func (r *missionRepository) AddAgents(ctx context.Context, missionID string, agentIDs []string) (*ent.Mission, error) {
	r.logger.Info("Adding agents to mission", zap.String("missionID", missionID), zap.Strings("agentIDs", agentIDs))

	mID, _ := uuid.Parse(missionID)
	agentUUIDs := make([]uuid.UUID, len(agentIDs))
	for i, aid := range agentIDs {
		agentUUIDs[i], _ = uuid.Parse(aid)
	}
	_, err := r.client.Mission.UpdateOneID(mID).
		AddAgentIDs(agentUUIDs...).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to add agents to mission", zap.Error(err))
		return nil, fmt.Errorf("failed to add agents: %w", err)
	}

	return r.GetByID(ctx, missionID)
}

// RemoveAgent removes an agent from a mission
func (r *missionRepository) RemoveAgent(ctx context.Context, missionID string, agentID string) (*ent.Mission, error) {
	r.logger.Info("Removing agent from mission", zap.String("missionID", missionID), zap.String("agentID", agentID))

	mID, _ := uuid.Parse(missionID)
	aID, _ := uuid.Parse(agentID)
	_, err := r.client.Mission.UpdateOneID(mID).
		RemoveAgentIDs(aID).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to remove agent from mission", zap.Error(err))
		return nil, fmt.Errorf("failed to remove agent: %w", err)
	}

	return r.GetByID(ctx, missionID)
}
