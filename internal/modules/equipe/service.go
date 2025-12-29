package equipe

import (
	"context"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines equipe service interface
type Service interface {
	Create(ctx context.Context, req *CreateEquipeRequest) (*EquipeResponse, error)
	GetByID(ctx context.Context, id string) (*EquipeResponse, error)
	GetByCode(ctx context.Context, code string) (*EquipeResponse, error)
	List(ctx context.Context, filters *ListEquipesFilters) ([]EquipeResponse, error)
	Update(ctx context.Context, id string, req *UpdateEquipeRequest) (*EquipeResponse, error)
	Delete(ctx context.Context, id string) error
	AddMembre(ctx context.Context, equipeID string, req *AddMembreRequest) error
	RemoveMembre(ctx context.Context, equipeID, userID string) error
	SetChefEquipe(ctx context.Context, equipeID string, req *SetChefEquipeRequest) error
}

type service struct {
	repo   repository.EquipeRepository
	logger *zap.Logger
}

// NewService creates a new equipe service
func NewService(repo repository.EquipeRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new equipe
func (s *service) Create(ctx context.Context, req *CreateEquipeRequest) (*EquipeResponse, error) {
	input := &repository.CreateEquipeInput{
		ID:             uuid.New().String(),
		Nom:            req.Nom,
		Code:           req.Code,
		Zone:           req.Zone,
		Description:    req.Description,
		CommissariatID: req.CommissariatID,
	}

	equipe, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(equipe), nil
}

// GetByID gets an equipe by ID
func (s *service) GetByID(ctx context.Context, id string) (*EquipeResponse, error) {
	equipe, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(equipe), nil
}

// GetByCode gets an equipe by code
func (s *service) GetByCode(ctx context.Context, code string) (*EquipeResponse, error) {
	equipe, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return s.toResponse(equipe), nil
}

// List lists equipes with filters
func (s *service) List(ctx context.Context, filters *ListEquipesFilters) ([]EquipeResponse, error) {
	repoFilters := &repository.EquipeFilters{}

	if filters != nil {
		if filters.CommissariatID != "" {
			repoFilters.CommissariatID = &filters.CommissariatID
		}
		if filters.Active != "" {
			active := filters.Active == "true"
			repoFilters.Active = &active
		}
		if filters.Search != "" {
			repoFilters.Search = &filters.Search
		}
	}

	equipes, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	responses := make([]EquipeResponse, len(equipes))
	for i, eq := range equipes {
		responses[i] = *s.toResponse(eq)
	}

	return responses, nil
}

// Update updates an equipe
func (s *service) Update(ctx context.Context, id string, req *UpdateEquipeRequest) (*EquipeResponse, error) {
	input := &repository.UpdateEquipeInput{
		Nom:         req.Nom,
		Zone:        req.Zone,
		Description: req.Description,
		Active:      req.Active,
	}

	equipe, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	return s.toResponse(equipe), nil
}

// Delete deletes an equipe
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// AddMembre adds a member to the equipe
func (s *service) AddMembre(ctx context.Context, equipeID string, req *AddMembreRequest) error {
	return s.repo.AddMembre(ctx, equipeID, req.UserID)
}

// RemoveMembre removes a member from the equipe
func (s *service) RemoveMembre(ctx context.Context, equipeID, userID string) error {
	return s.repo.RemoveMembre(ctx, equipeID, userID)
}

// SetChefEquipe sets the team leader
func (s *service) SetChefEquipe(ctx context.Context, equipeID string, req *SetChefEquipeRequest) error {
	return s.repo.SetChefEquipe(ctx, equipeID, req.UserID)
}

// toResponse converts ent.Equipe to EquipeResponse
func (s *service) toResponse(eq *ent.Equipe) *EquipeResponse {
	resp := &EquipeResponse{
		ID:          eq.ID.String(),
		Nom:         eq.Nom,
		Code:        eq.Code,
		Zone:        eq.Zone,
		Description: eq.Description,
		Active:      eq.Active,
		CreatedAt:   eq.CreatedAt,
		UpdatedAt:   eq.UpdatedAt,
	}

	// Map commissariat
	if eq.Edges.Commissariat != nil {
		resp.Commissariat = &CommissariatResponse{
			ID:   eq.Edges.Commissariat.ID.String(),
			Nom:  eq.Edges.Commissariat.Nom,
			Code: eq.Edges.Commissariat.Code,
		}
	}

	// Map chef d'equipe
	if eq.Edges.ChefEquipe != nil {
		resp.ChefEquipe = &MembreResponse{
			ID:        eq.Edges.ChefEquipe.ID.String(),
			Nom:       eq.Edges.ChefEquipe.Nom,
			Prenom:    eq.Edges.ChefEquipe.Prenom,
			Matricule: eq.Edges.ChefEquipe.Matricule,
			Grade:     eq.Edges.ChefEquipe.Grade,
			Role:      eq.Edges.ChefEquipe.Role,
		}
	}

	// Map membres
	if eq.Edges.Membres != nil {
		resp.Membres = make([]MembreResponse, len(eq.Edges.Membres))
		for i, m := range eq.Edges.Membres {
			resp.Membres[i] = MembreResponse{
				ID:        m.ID.String(),
				Nom:       m.Nom,
				Prenom:    m.Prenom,
				Matricule: m.Matricule,
				Grade:     m.Grade,
				Role:      m.Role,
			}
		}
		resp.NombreMembres = len(eq.Edges.Membres)
	}

	// Map missions
	if eq.Edges.Missions != nil {
		resp.Missions = make([]MissionSummaryResponse, len(eq.Edges.Missions))
		missionsActives := 0
		for i, m := range eq.Edges.Missions {
			resp.Missions[i] = MissionSummaryResponse{
				ID:        m.ID.String(),
				Type:      m.Type,
				Titre:     m.Titre,
				DateDebut: m.DateDebut,
				Statut:    m.Statut,
			}
			if m.Statut == "EN_COURS" {
				missionsActives++
			}
		}
		resp.MissionsActives = missionsActives
	}

	return resp
}
