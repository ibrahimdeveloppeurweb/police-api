package objectif

import (
	"context"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines objectif service interface
type Service interface {
	Create(ctx context.Context, req *CreateObjectifRequest) (*ObjectifResponse, error)
	GetByID(ctx context.Context, id string) (*ObjectifResponse, error)
	List(ctx context.Context, filters *ListObjectifsFilters) ([]ObjectifResponse, error)
	Update(ctx context.Context, id string, req *UpdateObjectifRequest) (*ObjectifResponse, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string) ([]ObjectifResponse, error)
	UpdateProgression(ctx context.Context, id string, req *UpdateProgressionRequest) (*ObjectifResponse, error)
}

type service struct {
	repo   repository.ObjectifRepository
	logger *zap.Logger
}

// NewService creates a new objectif service
func NewService(repo repository.ObjectifRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new objectif
func (s *service) Create(ctx context.Context, req *CreateObjectifRequest) (*ObjectifResponse, error) {
	input := &repository.CreateObjectifInput{
		ID:          uuid.New().String(),
		Titre:       req.Titre,
		Description: req.Description,
		Periode:     req.Periode,
		DateDebut:   req.DateDebut,
		DateFin:     req.DateFin,
		ValeurCible: req.ValeurCible,
		AgentID:     req.AgentID,
	}

	objectif, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(objectif), nil
}

// GetByID gets an objectif by ID
func (s *service) GetByID(ctx context.Context, id string) (*ObjectifResponse, error) {
	objectif, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(objectif), nil
}

// List lists objectifs with filters
func (s *service) List(ctx context.Context, filters *ListObjectifsFilters) ([]ObjectifResponse, error) {
	repoFilters := &repository.ObjectifFilters{}

	if filters != nil {
		if filters.AgentID != "" {
			repoFilters.AgentID = &filters.AgentID
		}
		if filters.Periode != "" {
			repoFilters.Periode = &filters.Periode
		}
		if filters.Statut != "" {
			repoFilters.Statut = &filters.Statut
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

	objectifs, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	responses := make([]ObjectifResponse, len(objectifs))
	for i, obj := range objectifs {
		responses[i] = *s.toResponse(obj)
	}

	return responses, nil
}

// Update updates an objectif
func (s *service) Update(ctx context.Context, id string, req *UpdateObjectifRequest) (*ObjectifResponse, error) {
	input := &repository.UpdateObjectifInput{
		Titre:          req.Titre,
		Description:    req.Description,
		Statut:         req.Statut,
		ValeurCible:    req.ValeurCible,
		ValeurActuelle: req.ValeurActuelle,
		DateFin:        req.DateFin,
	}

	objectif, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(objectif), nil
}

// Delete deletes an objectif
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetByAgent gets objectifs for an agent
func (s *service) GetByAgent(ctx context.Context, agentID string) ([]ObjectifResponse, error) {
	objectifs, err := s.repo.GetByAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	responses := make([]ObjectifResponse, len(objectifs))
	for i, obj := range objectifs {
		responses[i] = *s.toResponse(obj)
	}

	return responses, nil
}

// UpdateProgression updates the progression of an objectif
func (s *service) UpdateProgression(ctx context.Context, id string, req *UpdateProgressionRequest) (*ObjectifResponse, error) {
	objectif, err := s.repo.UpdateProgression(ctx, id, req.ValeurActuelle)
	if err != nil {
		return nil, err
	}

	return s.toResponse(objectif), nil
}

// toResponse converts ent.Objectif to ObjectifResponse
func (s *service) toResponse(obj *ent.Objectif) *ObjectifResponse {
	resp := &ObjectifResponse{
		ID:             obj.ID.String(),
		Titre:          obj.Titre,
		Description:    obj.Description,
		Periode:        obj.Periode,
		DateDebut:      obj.DateDebut,
		DateFin:        obj.DateFin,
		Statut:         obj.Statut,
		ValeurCible:    obj.ValeurCible,
		ValeurActuelle: obj.ValeurActuelle,
		Progression:    obj.Progression,
		CreatedAt:      obj.CreatedAt,
		UpdatedAt:      obj.UpdatedAt,
	}

	// Map agent
	if obj.Edges.Agent != nil {
		resp.Agent = &AgentResponse{
			ID:        obj.Edges.Agent.ID.String(),
			Nom:       obj.Edges.Agent.Nom,
			Prenom:    obj.Edges.Agent.Prenom,
			Matricule: obj.Edges.Agent.Matricule,
		}
	}

	return resp
}
