package commissariat

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/zap"
)

// Service defines commissariat service interface
type Service interface {
	List(ctx context.Context, actif *bool, page, limit int) (*ListCommissariatsResponse, error)
	GetDashboard(ctx context.Context, commissariatID string) (*DashboardResponse, error)
	GetAgents(ctx context.Context, commissariatID string) ([]*AgentResponse, error)
	GetControles(ctx context.Context, commissariatID string, page, limit int) (*ListControlesResponse, error)
	GetStatistiques(ctx context.Context, commissariatID string, dateDebut, dateFin *time.Time) (*StatistiquesResponse, error)
}

// service implements commissariat service
type service struct {
	commissariatRepo repository.CommissariatRepository
	userRepo         repository.UserRepository
	controleRepo     repository.ControleRepository
	pvRepo           repository.PVRepository
	alerteRepo       repository.AlerteRepository
	logger           *zap.Logger
}

// NewService creates a new commissariat service
func NewService(
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	controleRepo repository.ControleRepository,
	pvRepo repository.PVRepository,
	alerteRepo repository.AlerteRepository,
	logger *zap.Logger,
) Service {
	return &service{
		commissariatRepo: commissariatRepo,
		userRepo:         userRepo,
		controleRepo:     controleRepo,
		pvRepo:           pvRepo,
		alerteRepo:       alerteRepo,
		logger:           logger,
	}
}

// List returns paginated list of commissariats
func (s *service) List(ctx context.Context, actif *bool, page, limit int) (*ListCommissariatsResponse, error) {
	s.logger.Info("Listing commissariats", zap.Int("page", page), zap.Int("limit", limit))

	// Build filters
	filters := &repository.CommissariatFilters{
		Actif:  actif,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Get commissariats
	commissariats, err := s.commissariatRepo.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list commissariats", zap.Error(err))
		return nil, fmt.Errorf("failed to list commissariats: %w", err)
	}

	// Get total count
	total, err := s.commissariatRepo.Count(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to count commissariats", zap.Error(err))
		total = len(commissariats)
	}

	// Convert to response
	data := make([]*CommissariatResponse, len(commissariats))
	for i, comm := range commissariats {
		// Count agents
		nbAgents := 0
		if comm.Edges.Agents != nil {
			nbAgents = len(comm.Edges.Agents)
		}

		data[i] = &CommissariatResponse{
			ID:        comm.ID.String(),
			Nom:       comm.Nom,
			Code:      comm.Code,
			Ville:     comm.Ville,
			Region:    comm.Region,
			Adresse:   comm.Adresse,
			Telephone: comm.Telephone,
			Email:     comm.Email,
			Actif:     comm.Actif,
			NbAgents:  nbAgents,
			CreatedAt: comm.CreatedAt,
		}
	}

	// Calculate total pages
	totalPages := (total + limit - 1) / limit

	return &ListCommissariatsResponse{
		Data: data,
		Pagination: &Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetDashboard returns commissariat dashboard
func (s *service) GetDashboard(ctx context.Context, commissariatID string) (*DashboardResponse, error) {
	s.logger.Info("Getting dashboard for commissariat", zap.String("id", commissariatID))

	// Get commissariat with agents
	comm, err := s.commissariatRepo.GetByID(ctx, commissariatID)
	if err != nil {
		return nil, fmt.Errorf("commissariat not found: %w", err)
	}

	// Count agents
	agentsActifs := 0
	if comm.Edges.Agents != nil {
		for _, agent := range comm.Edges.Agents {
			if agent.Active {
				agentsActifs++
			}
		}
	}

	// Get recent controles filtered by commissariat
	controles, err := s.controleRepo.GetByCommissariat(ctx, commissariatID, &repository.ControleFilters{Limit: 5})
	if err != nil {
		s.logger.Error("Failed to get recent controles", zap.Error(err))
		controles = []*ent.Controle{}
	}

	// Get active alertes for commissariat
	alertes, err := s.alerteRepo.GetByCommissariat(ctx, commissariatID)
	if err != nil {
		s.logger.Error("Failed to get alertes", zap.Error(err))
		alertes = []*ent.AlerteSecuritaire{}
	}

	// Count active alertes
	alertesEnCours := 0
	alertesRecentes := make([]*AlerteRecente, 0)
	for _, alerte := range alertes {
		if string(alerte.Statut) == "ACTIVE" {
			alertesEnCours++
		}
		if len(alertesRecentes) < 5 {
			alertesRecentes = append(alertesRecentes, &AlerteRecente{
				ID:         alerte.ID.String(),
				Titre:      alerte.Titre,
				Niveau:     string(alerte.Niveau),
				Statut:     string(alerte.Statut),
				DateAlerte: alerte.DateAlerte,
			})
		}
	}

	// Build recent controles
	controlesRecents := make([]*ControleRecent, 0, len(controles))
	for _, ctrl := range controles {
		agentNom := ""
		if ctrl.Edges.Agent != nil {
			agentNom = ctrl.Edges.Agent.Nom + " " + ctrl.Edges.Agent.Prenom
		}
		controlesRecents = append(controlesRecents, &ControleRecent{
			ID:           ctrl.ID.String(),
			DateControle: ctrl.DateControle,
			LieuControle: ctrl.LieuControle,
			TypeControle: string(ctrl.TypeControle),
			Statut:       string(ctrl.Statut),
			AgentNom:     agentNom,
		})
	}

	// Count total controles
	controlesTotal, _ := s.controleRepo.Count(ctx, nil)

	// Get PV statistics
	pvStats, err := s.pvRepo.GetStatistics(ctx, nil)
	if err != nil {
		pvStats = &repository.PVStatistics{}
	}

	return &DashboardResponse{
		Statistiques: &DashboardStats{
			ControlesTotal: controlesTotal,
			PvTotal:        pvStats.Total,
			AgentsActifs:   agentsActifs,
			AlertesEnCours: alertesEnCours,
		},
		ControleRecents: controlesRecents,
		AlertesRecentes: alertesRecentes,
	}, nil
}

// GetAgents returns commissariat agents
func (s *service) GetAgents(ctx context.Context, commissariatID string) ([]*AgentResponse, error) {
	s.logger.Info("Getting agents for commissariat", zap.String("id", commissariatID))

	comm, err := s.commissariatRepo.GetByID(ctx, commissariatID)
	if err != nil {
		return nil, fmt.Errorf("commissariat not found: %w", err)
	}

	agents := make([]*AgentResponse, 0)
	if comm.Edges.Agents != nil {
		for _, agent := range comm.Edges.Agents {
			agents = append(agents, &AgentResponse{
				ID:        agent.ID.String(),
				Matricule: agent.Matricule,
				Nom:       agent.Nom,
				Prenom:    agent.Prenom,
				Email:     agent.Email,
				Role:      agent.Role,
				Actif:     agent.Active,
				CreatedAt: agent.CreatedAt,
			})
		}
	}

	return agents, nil
}

// GetControles returns commissariat controles
func (s *service) GetControles(ctx context.Context, commissariatID string, page, limit int) (*ListControlesResponse, error) {
	s.logger.Info("Getting controles for commissariat", zap.String("id", commissariatID))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	// Filter controles by commissariat
	filters := &repository.ControleFilters{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	controles, err := s.controleRepo.GetByCommissariat(ctx, commissariatID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list controles: %w", err)
	}

	total, err := s.controleRepo.Count(ctx, &repository.ControleFilters{CommissariatID: &commissariatID})
	if err != nil {
		return nil, fmt.Errorf("failed to count controles: %w", err)
	}

	responses := make([]*ControleResponse, len(controles))
	for i, ctrl := range controles {
		responses[i] = s.controleToResponse(ctrl)
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return &ListControlesResponse{
		Data: responses,
		Pagination: &Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetStatistiques returns commissariat statistics
func (s *service) GetStatistiques(ctx context.Context, commissariatID string, dateDebut, dateFin *time.Time) (*StatistiquesResponse, error) {
	s.logger.Info("Getting statistiques for commissariat", zap.String("id", commissariatID))

	// Get controle statistics
	controleStats, err := s.controleRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get controle statistics", zap.Error(err))
		controleStats = &repository.ControleStatistics{
			ParType: make(map[string]int),
			ParJour: make(map[string]int),
		}
	}

	// Build evolution from ParJour
	evolution := make([]EvolutionEntry, 0, len(controleStats.ParJour))
	for date, count := range controleStats.ParJour {
		evolution = append(evolution, EvolutionEntry{
			Date:  date,
			Count: count,
		})
	}

	// Get PV statistics
	pvStats, err := s.pvRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get PV statistics", zap.Error(err))
		pvStats = &repository.PVStatistics{}
	}

	return &StatistiquesResponse{
		Controles: &ControleStats{
			Total:     controleStats.Total,
			ParType:   controleStats.ParType,
			Evolution: evolution,
		},
		PV: &PVStats{
			Total:        pvStats.Total,
			MontantTotal: pvStats.MontantTotal,
			TauxPaiement: pvStats.TauxRecouvrement,
		},
		Infractions: &InfractionStats{
			Total:   controleStats.InfractionsAvec,
			ParType: make(map[string]int),
		},
	}, nil
}

// Helper function
func (s *service) controleToResponse(ctrl *ent.Controle) *ControleResponse {
	resp := &ControleResponse{
		ID:            ctrl.ID.String(),
		DateControle:  ctrl.DateControle,
		LieuControle:  ctrl.LieuControle,
		TypeControle:  string(ctrl.TypeControle),
		Statut:        string(ctrl.Statut),
		Observations:  ctrl.Observations,
		NbInfractions: len(ctrl.Edges.Infractions),
		CreatedAt:     ctrl.CreatedAt,
	}

	if agent := ctrl.Edges.Agent; agent != nil {
		resp.Agent = &AgentSummary{
			ID:        agent.ID.String(),
			Matricule: agent.Matricule,
			Nom:       agent.Nom,
			Prenom:    agent.Prenom,
		}
	}

	if vehicule := ctrl.Edges.Vehicule; vehicule != nil {
		resp.Vehicule = &VehiculeSummary{
			ID:              vehicule.ID.String(),
			Immatriculation: vehicule.Immatriculation,
			Marque:          vehicule.Marque,
			Modele:          vehicule.Modele,
		}
	}

	if conducteur := ctrl.Edges.Conducteur; conducteur != nil {
		resp.Conducteur = &ConducteurSummary{
			ID:     conducteur.ID.String(),
			Nom:    conducteur.Nom,
			Prenom: conducteur.Prenom,
		}
	}

	return resp
}
