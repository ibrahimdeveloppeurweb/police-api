package observation

import (
	"context"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines observation service interface
type Service interface {
	Create(ctx context.Context, req *CreateObservationRequest) (*ObservationResponse, error)
	GetByID(ctx context.Context, id string) (*ObservationResponse, error)
	List(ctx context.Context, filters *ListObservationsFilters) ([]ObservationResponse, error)
	Update(ctx context.Context, id string, req *UpdateObservationRequest) (*ObservationResponse, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string, visibleOnly bool) ([]ObservationResponse, error)
	GetByAuteur(ctx context.Context, auteurID string) ([]ObservationResponse, error)
}

type service struct {
	repo   repository.ObservationRepository
	logger *zap.Logger
}

// NewService creates a new observation service
func NewService(repo repository.ObservationRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new observation
func (s *service) Create(ctx context.Context, req *CreateObservationRequest) (*ObservationResponse, error) {
	input := &repository.CreateObservationInput{
		ID:           uuid.New().String(),
		Contenu:      req.Contenu,
		Type:         req.Type,
		Categorie:    req.Categorie,
		VisibleAgent: req.VisibleAgent,
		AgentID:      req.AgentID,
		AuteurID:     req.AuteurID,
	}

	observation, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(observation), nil
}

// GetByID gets an observation by ID
func (s *service) GetByID(ctx context.Context, id string) (*ObservationResponse, error) {
	observation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(observation), nil
}

// List lists observations with filters
func (s *service) List(ctx context.Context, filters *ListObservationsFilters) ([]ObservationResponse, error) {
	repoFilters := &repository.ObservationFilters{}

	if filters != nil {
		if filters.AgentID != "" {
			repoFilters.AgentID = &filters.AgentID
		}
		if filters.AuteurID != "" {
			repoFilters.AuteurID = &filters.AuteurID
		}
		if filters.Type != "" {
			repoFilters.Type = &filters.Type
		}
		if filters.Categorie != "" {
			repoFilters.Categorie = &filters.Categorie
		}
		if filters.VisibleAgent != "" {
			visible := filters.VisibleAgent == "true"
			repoFilters.VisibleAgent = &visible
		}
	}

	observations, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	responses := make([]ObservationResponse, len(observations))
	for i, obs := range observations {
		responses[i] = *s.toResponse(obs)
	}

	return responses, nil
}

// Update updates an observation
func (s *service) Update(ctx context.Context, id string, req *UpdateObservationRequest) (*ObservationResponse, error) {
	input := &repository.UpdateObservationInput{
		Contenu:      req.Contenu,
		Type:         req.Type,
		Categorie:    req.Categorie,
		VisibleAgent: req.VisibleAgent,
	}

	observation, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(observation), nil
}

// Delete deletes an observation
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetByAgent gets observations for an agent
func (s *service) GetByAgent(ctx context.Context, agentID string, visibleOnly bool) ([]ObservationResponse, error) {
	observations, err := s.repo.GetByAgent(ctx, agentID, visibleOnly)
	if err != nil {
		return nil, err
	}

	responses := make([]ObservationResponse, len(observations))
	for i, obs := range observations {
		responses[i] = *s.toResponse(obs)
	}

	return responses, nil
}

// GetByAuteur gets observations created by an auteur
func (s *service) GetByAuteur(ctx context.Context, auteurID string) ([]ObservationResponse, error) {
	observations, err := s.repo.GetByAuteur(ctx, auteurID)
	if err != nil {
		return nil, err
	}

	responses := make([]ObservationResponse, len(observations))
	for i, obs := range observations {
		responses[i] = *s.toResponse(obs)
	}

	return responses, nil
}

// toResponse converts ent.Observation to ObservationResponse
func (s *service) toResponse(obs *ent.Observation) *ObservationResponse {
	resp := &ObservationResponse{
		ID:           obs.ID.String(),
		Contenu:      obs.Contenu,
		Type:         obs.Type,
		Categorie:    obs.Categorie,
		VisibleAgent: obs.VisibleAgent,
		CreatedAt:    obs.CreatedAt,
		UpdatedAt:    obs.UpdatedAt,
	}

	// Map agent
	if obs.Edges.Agent != nil {
		resp.Agent = &AgentResponse{
			ID:        obs.Edges.Agent.ID.String(),
			Nom:       obs.Edges.Agent.Nom,
			Prenom:    obs.Edges.Agent.Prenom,
			Matricule: obs.Edges.Agent.Matricule,
			Grade:     obs.Edges.Agent.Grade,
		}
	}

	// Map auteur
	if obs.Edges.Auteur != nil {
		resp.Auteur = &AgentResponse{
			ID:        obs.Edges.Auteur.ID.String(),
			Nom:       obs.Edges.Auteur.Nom,
			Prenom:    obs.Edges.Auteur.Prenom,
			Matricule: obs.Edges.Auteur.Matricule,
			Grade:     obs.Edges.Auteur.Grade,
		}
	}

	return resp
}
