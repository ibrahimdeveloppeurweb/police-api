package competence

import (
	"context"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines competence service interface
type Service interface {
	Create(ctx context.Context, req *CreateCompetenceRequest) (*CompetenceResponse, error)
	GetByID(ctx context.Context, id string) (*CompetenceResponse, error)
	GetByNom(ctx context.Context, nom string) (*CompetenceResponse, error)
	List(ctx context.Context, filters *ListCompetencesFilters) ([]CompetenceResponse, error)
	Update(ctx context.Context, id string, req *UpdateCompetenceRequest) (*CompetenceResponse, error)
	Delete(ctx context.Context, id string) error
	AssignToAgent(ctx context.Context, competenceID string, req *AssignCompetenceRequest) error
	RemoveFromAgent(ctx context.Context, competenceID, agentID string) error
	GetByAgent(ctx context.Context, agentID string) ([]CompetenceResponse, error)
	GetExpiring(ctx context.Context, daysAhead int) ([]CompetenceResponse, error)
}

type service struct {
	repo   repository.CompetenceRepository
	logger *zap.Logger
}

// NewService creates a new competence service
func NewService(repo repository.CompetenceRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new competence
func (s *service) Create(ctx context.Context, req *CreateCompetenceRequest) (*CompetenceResponse, error) {
	input := &repository.CreateCompetenceInput{
		ID:             uuid.New().String(),
		Nom:            req.Nom,
		Type:           req.Type,
		Description:    req.Description,
		Organisme:      req.Organisme,
		DateObtention:  req.DateObtention,
		DateExpiration: req.DateExpiration,
	}

	competence, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(competence), nil
}

// GetByID gets a competence by ID
func (s *service) GetByID(ctx context.Context, id string) (*CompetenceResponse, error) {
	competence, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(competence), nil
}

// GetByNom gets a competence by nom
func (s *service) GetByNom(ctx context.Context, nom string) (*CompetenceResponse, error) {
	competence, err := s.repo.GetByNom(ctx, nom)
	if err != nil {
		return nil, err
	}

	return s.toResponse(competence), nil
}

// List lists competences with filters
func (s *service) List(ctx context.Context, filters *ListCompetencesFilters) ([]CompetenceResponse, error) {
	repoFilters := &repository.CompetenceFilters{}

	if filters != nil {
		if filters.Type != "" {
			repoFilters.Type = &filters.Type
		}
		if filters.Active != "" {
			active := filters.Active == "true"
			repoFilters.Active = &active
		}
		if filters.Search != "" {
			repoFilters.Search = &filters.Search
		}
		if filters.Organisme != "" {
			repoFilters.Organisme = &filters.Organisme
		}
	}

	competences, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	responses := make([]CompetenceResponse, len(competences))
	for i, comp := range competences {
		responses[i] = *s.toResponse(comp)
	}

	return responses, nil
}

// Update updates a competence
func (s *service) Update(ctx context.Context, id string, req *UpdateCompetenceRequest) (*CompetenceResponse, error) {
	input := &repository.UpdateCompetenceInput{
		Nom:            req.Nom,
		Description:    req.Description,
		Organisme:      req.Organisme,
		DateExpiration: req.DateExpiration,
		Active:         req.Active,
	}

	competence, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(competence), nil
}

// Delete deletes a competence
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// AssignToAgent assigns a competence to an agent
func (s *service) AssignToAgent(ctx context.Context, competenceID string, req *AssignCompetenceRequest) error {
	return s.repo.AssignToAgent(ctx, competenceID, req.AgentID)
}

// RemoveFromAgent removes a competence from an agent
func (s *service) RemoveFromAgent(ctx context.Context, competenceID, agentID string) error {
	return s.repo.RemoveFromAgent(ctx, competenceID, agentID)
}

// GetByAgent gets competences for an agent
func (s *service) GetByAgent(ctx context.Context, agentID string) ([]CompetenceResponse, error) {
	competences, err := s.repo.GetByAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	responses := make([]CompetenceResponse, len(competences))
	for i, comp := range competences {
		responses[i] = *s.toResponse(comp)
	}

	return responses, nil
}

// GetExpiring gets competences expiring within the specified number of days
func (s *service) GetExpiring(ctx context.Context, daysAhead int) ([]CompetenceResponse, error) {
	competences, err := s.repo.GetExpiring(ctx, daysAhead)
	if err != nil {
		return nil, err
	}

	responses := make([]CompetenceResponse, len(competences))
	for i, comp := range competences {
		responses[i] = *s.toResponse(comp)
	}

	return responses, nil
}

// toResponse converts ent.Competence to CompetenceResponse
func (s *service) toResponse(comp *ent.Competence) *CompetenceResponse {
	resp := &CompetenceResponse{
		ID:          comp.ID.String(),
		Nom:         comp.Nom,
		Type:        comp.Type,
		Description: comp.Description,
		Organisme:   comp.Organisme,
		Active:      comp.Active,
		CreatedAt:   comp.CreatedAt,
		UpdatedAt:   comp.UpdatedAt,
	}

	if !comp.DateObtention.IsZero() {
		resp.DateObtention = &comp.DateObtention
	}

	if !comp.DateExpiration.IsZero() {
		resp.DateExpiration = &comp.DateExpiration
		// Calculate days remaining
		daysRemaining := int(time.Until(comp.DateExpiration).Hours() / 24)
		if daysRemaining >= 0 {
			resp.JoursRestants = &daysRemaining
		}
	}

	// Map agents
	if comp.Edges.Agents != nil {
		resp.Agents = make([]AgentResponse, len(comp.Edges.Agents))
		for i, a := range comp.Edges.Agents {
			resp.Agents[i] = AgentResponse{
				ID:        a.ID.String(),
				Nom:       a.Nom,
				Prenom:    a.Prenom,
				Matricule: a.Matricule,
				Grade:     a.Grade,
			}
		}
		resp.NombreAgents = len(comp.Edges.Agents)
	}

	return resp
}
