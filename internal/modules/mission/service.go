package mission

import (
	"context"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines mission service interface
type Service interface {
	Create(ctx context.Context, req *CreateMissionRequest) (*MissionResponse, error)
	GetByID(ctx context.Context, id string) (*MissionResponse, error)
	List(ctx context.Context, filters *ListMissionsFilters) ([]MissionResponse, error)
	Update(ctx context.Context, id string, req *UpdateMissionRequest) (*MissionResponse, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string, limit int) ([]MissionResponse, error)
	GetByEquipe(ctx context.Context, equipeID string) ([]MissionResponse, error)
	StartMission(ctx context.Context, id string) (*MissionResponse, error)
	EndMission(ctx context.Context, id string, req *EndMissionRequest) (*MissionResponse, error)
	CancelMission(ctx context.Context, id string, req *CancelMissionRequest) (*MissionResponse, error)
	AddAgents(ctx context.Context, missionID string, req *AddAgentsRequest) (*MissionResponse, error)
	RemoveAgent(ctx context.Context, missionID string, req *RemoveAgentRequest) (*MissionResponse, error)
}

type service struct {
	repo   repository.MissionRepository
	logger *zap.Logger
}

// NewService creates a new mission service
func NewService(repo repository.MissionRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new mission
func (s *service) Create(ctx context.Context, req *CreateMissionRequest) (*MissionResponse, error) {
	input := &repository.CreateMissionInput{
		ID:             uuid.New().String(),
		Type:           req.Type,
		Titre:          req.Titre,
		DateDebut:      req.DateDebut,
		DateFin:        req.DateFin,
		Duree:          req.Duree,
		Zone:           req.Zone,
		AgentIDs:       req.AgentIDs,
		CommissariatID: req.CommissariatID,
		EquipeID:       req.EquipeID,
	}

	mission, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// GetByID gets a mission by ID
func (s *service) GetByID(ctx context.Context, id string) (*MissionResponse, error) {
	mission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// List lists missions with filters
func (s *service) List(ctx context.Context, filters *ListMissionsFilters) ([]MissionResponse, error) {
	repoFilters := &repository.MissionFilters{}

	if filters != nil {
		if filters.AgentID != "" {
			repoFilters.AgentID = &filters.AgentID
		}
		if filters.EquipeID != "" {
			repoFilters.EquipeID = &filters.EquipeID
		}
		if filters.CommissariatID != "" {
			repoFilters.CommissariatID = &filters.CommissariatID
		}
		if filters.Statut != "" {
			repoFilters.Statut = &filters.Statut
		}
		if filters.Type != "" {
			repoFilters.Type = &filters.Type
		}
		if filters.DateDebut != "" {
			if t, err := time.Parse("2006-01-02", filters.DateDebut); err == nil {
				repoFilters.DateDebut = &t
			}
		}
		if filters.DateFin != "" {
			if t, err := time.Parse("2006-01-02", filters.DateFin); err == nil {
				repoFilters.DateFin = &t
			}
		}
	}

	missions, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	responses := make([]MissionResponse, len(missions))
	for i, m := range missions {
		responses[i] = *s.toResponse(m)
	}

	return responses, nil
}

// Update updates a mission
func (s *service) Update(ctx context.Context, id string, req *UpdateMissionRequest) (*MissionResponse, error) {
	input := &repository.UpdateMissionInput{
		Titre:   req.Titre,
		Zone:    req.Zone,
		Duree:   req.Duree,
		Statut:  req.Statut,
		Rapport: req.Rapport,
		DateFin: req.DateFin,
	}

	mission, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// Delete deletes a mission
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetByAgent gets missions for an agent
func (s *service) GetByAgent(ctx context.Context, agentID string, limit int) ([]MissionResponse, error) {
	missions, err := s.repo.GetByAgent(ctx, agentID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]MissionResponse, len(missions))
	for i, m := range missions {
		responses[i] = *s.toResponse(m)
	}

	return responses, nil
}

// GetByEquipe gets missions for an equipe
func (s *service) GetByEquipe(ctx context.Context, equipeID string) ([]MissionResponse, error) {
	missions, err := s.repo.GetByEquipe(ctx, equipeID)
	if err != nil {
		return nil, err
	}

	responses := make([]MissionResponse, len(missions))
	for i, m := range missions {
		responses[i] = *s.toResponse(m)
	}

	return responses, nil
}

// StartMission starts a mission
func (s *service) StartMission(ctx context.Context, id string) (*MissionResponse, error) {
	mission, err := s.repo.StartMission(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// EndMission ends a mission with a rapport
func (s *service) EndMission(ctx context.Context, id string, req *EndMissionRequest) (*MissionResponse, error) {
	mission, err := s.repo.EndMission(ctx, id, req.Rapport)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// CancelMission cancels a mission
func (s *service) CancelMission(ctx context.Context, id string, req *CancelMissionRequest) (*MissionResponse, error) {
	raison := ""
	if req != nil {
		raison = req.Raison
	}
	mission, err := s.repo.CancelMission(ctx, id, raison)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// AddAgents adds agents to a mission
func (s *service) AddAgents(ctx context.Context, missionID string, req *AddAgentsRequest) (*MissionResponse, error) {
	mission, err := s.repo.AddAgents(ctx, missionID, req.AgentIDs)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// RemoveAgent removes an agent from a mission
func (s *service) RemoveAgent(ctx context.Context, missionID string, req *RemoveAgentRequest) (*MissionResponse, error) {
	mission, err := s.repo.RemoveAgent(ctx, missionID, req.AgentID)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mission), nil
}

// toResponse converts ent.Mission to MissionResponse
func (s *service) toResponse(m *ent.Mission) *MissionResponse {
	resp := &MissionResponse{
		ID:        m.ID.String(),
		Type:      m.Type,
		Titre:     m.Titre,
		DateDebut: m.DateDebut,
		Duree:     m.Duree,
		Zone:      m.Zone,
		Statut:    m.Statut,
		Rapport:   m.Rapport,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if !m.DateFin.IsZero() {
		resp.DateFin = &m.DateFin
	}

	// Map agents (many-to-many)
	if len(m.Edges.Agents) > 0 {
		resp.Agents = make([]AgentResponse, len(m.Edges.Agents))
		for i, agent := range m.Edges.Agents {
			resp.Agents[i] = AgentResponse{
				ID:        agent.ID.String(),
				Nom:       agent.Nom,
				Prenom:    agent.Prenom,
				Matricule: agent.Matricule,
				Grade:     agent.Grade,
			}
		}
	}

	// Map equipe
	if m.Edges.Equipe != nil {
		resp.Equipe = &EquipeResponse{
			ID:   m.Edges.Equipe.ID.String(),
			Nom:  m.Edges.Equipe.Nom,
			Code: m.Edges.Equipe.Code,
		}
	}

	// Map commissariat
	if m.Edges.Commissariat != nil {
		resp.Commissariat = &CommissariatResponse{
			ID:   m.Edges.Commissariat.ID.String(),
			Nom:  m.Edges.Commissariat.Nom,
			Code: m.Edges.Commissariat.Code,
		}
	}

	return resp
}
