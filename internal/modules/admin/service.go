package admin

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/crypto"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/infrastructure/session"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines admin service interface
type Service interface {
	// Statistics
	GetStatistiquesNationales(ctx context.Context) (*StatistiquesNationales, error)

	// Commissariats
	GetCommissariats(ctx context.Context) ([]*CommissariatResponse, error)
	GetCommissariat(ctx context.Context, id string) (*CommissariatResponse, error)
	CreateCommissariat(ctx context.Context, req *CreateCommissariatRequest) (*CommissariatResponse, error)
	UpdateCommissariat(ctx context.Context, id string, req *UpdateCommissariatRequest) (*CommissariatResponse, error)
	DeleteCommissariat(ctx context.Context, id string) error

	// Agents
	GetAgents(ctx context.Context, commissariatID *string) ([]*AgentResponse, error)
	GetAgent(ctx context.Context, id string) (*AgentResponse, error)
	CreateAgent(ctx context.Context, req *CreateAgentRequest) (*AgentResponse, error)
	UpdateAgent(ctx context.Context, id string, req *UpdateAgentRequest) (*AgentResponse, error)
	DeleteAgent(ctx context.Context, id string) error
	GetAgentStatistiques(ctx context.Context, id string) (*AgentStatistiquesResponse, error)

	// Dashboard
	GetAgentsDashboard(ctx context.Context, req *AgentDashboardRequest) (*AgentDashboardResponse, error)

	// Session Management (Remote Logout)
	GetAgentSessions(ctx context.Context, agentID string) ([]*AgentSessionResponse, error)
	RevokeAgentSession(ctx context.Context, agentID string, sessionID string, reason string) error
	RevokeAllAgentSessions(ctx context.Context, agentID string, reason string) error
}

// service implements admin service
type service struct {
	commissariatRepo repository.CommissariatRepository
	userRepo         repository.UserRepository
	controleRepo     repository.ControleRepository
	pvRepo           repository.PVRepository
	alerteRepo       repository.AlerteRepository
	infractionRepo   repository.InfractionRepository
	passwordService  crypto.Service
	sessionService   session.Service
	logger           *zap.Logger
}

// NewService creates a new admin service
func NewService(
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	controleRepo repository.ControleRepository,
	pvRepo repository.PVRepository,
	alerteRepo repository.AlerteRepository,
	infractionRepo repository.InfractionRepository,
	passwordService crypto.Service,
	sessionService session.Service,
	logger *zap.Logger,
) Service {
	return &service{
		commissariatRepo: commissariatRepo,
		userRepo:         userRepo,
		controleRepo:     controleRepo,
		pvRepo:           pvRepo,
		alerteRepo:       alerteRepo,
		infractionRepo:   infractionRepo,
		passwordService:  passwordService,
		sessionService:   sessionService,
		logger:           logger,
	}
}

// GetStatistiquesNationales returns national statistics
func (s *service) GetStatistiquesNationales(ctx context.Context) (*StatistiquesNationales, error) {
	s.logger.Info("Getting national statistics")

	// Get controles count
	controlesCount, err := s.controleRepo.Count(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to count controles", zap.Error(err))
		controlesCount = 0
	}

	// Get PV statistics
	pvStats, err := s.pvRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get PV statistics", zap.Error(err))
		pvStats = &repository.PVStatistics{}
	}

	// Get active alertes count
	activeAlertes, err := s.alerteRepo.GetActives(ctx)
	if err != nil {
		s.logger.Error("Failed to get active alertes", zap.Error(err))
		activeAlertes = []*ent.AlerteSecuritaire{}
	}

	// Get active commissariats count
	actif := true
	commissariatsCount, err := s.commissariatRepo.Count(ctx, &repository.CommissariatFilters{Actif: &actif})
	if err != nil {
		s.logger.Error("Failed to count commissariats", zap.Error(err))
		commissariatsCount = 0
	}

	// Get active agents count
	agents, err := s.userRepo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list agents", zap.Error(err))
		agents = []*ent.User{}
	}

	// Get controle statistics for evolution
	controleStats, err := s.controleRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get controle statistics", zap.Error(err))
		controleStats = &repository.ControleStatistics{ParJour: make(map[string]int)}
	}

	// Build evolution entries from ParJour
	evolutionControles := make([]EvolutionEntry, 0, len(controleStats.ParJour))
	for date, count := range controleStats.ParJour {
		evolutionControles = append(evolutionControles, EvolutionEntry{
			Date:  date,
			Count: count,
		})
	}

	// Get infraction statistics for top infractions
	infractionStats, err := s.infractionRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get infraction statistics", zap.Error(err))
		infractionStats = &repository.InfractionStatistics{TopInfractions: []repository.InfractionTypeStats{}}
	}

	// Build top infractions
	topInfractions := make([]InfractionCount, 0, len(infractionStats.TopInfractions))
	for _, top := range infractionStats.TopInfractions {
		topInfractions = append(topInfractions, InfractionCount{
			TypeCode:    top.TypeCode,
			TypeLibelle: top.TypeLibelle,
			Count:       top.Count,
			Montant:     top.MontantTotal,
		})
	}

	// Get statistics by region
	commissariats, err := s.commissariatRepo.List(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to list commissariats for region stats", zap.Error(err))
		commissariats = []*ent.Commissariat{}
	}

	// Group by region
	regionStats := make(map[string]*StatistiquesRegion)
	for _, comm := range commissariats {
		if _, exists := regionStats[comm.Region]; !exists {
			regionStats[comm.Region] = &StatistiquesRegion{
				Region:        comm.Region,
				Commissariats: 0,
				Agents:        0,
				Controles:     0,
			}
		}
		regionStats[comm.Region].Commissariats++

		// Count agents in this commissariat
		if comm.Edges.Agents != nil {
			regionStats[comm.Region].Agents += len(comm.Edges.Agents)
		}

		// Count controles in this commissariat
		if comm.Edges.Controles != nil {
			regionStats[comm.Region].Controles += len(comm.Edges.Controles)
		}
	}

	statistiquesParRegion := make([]StatistiquesRegion, 0, len(regionStats))
	for _, stat := range regionStats {
		statistiquesParRegion = append(statistiquesParRegion, *stat)
	}

	return &StatistiquesNationales{
		ControlesTotal:        controlesCount,
		PvTotal:               pvStats.Total,
		MontantPVTotal:        pvStats.MontantTotal,
		AlertesActives:        len(activeAlertes),
		CommissariatsActifs:   commissariatsCount,
		AgentsActifs:          len(agents),
		TauxPaiementPV:        pvStats.TauxRecouvrement,
		EvolutionControles:    evolutionControles,
		TopInfractions:        topInfractions,
		StatistiquesParRegion: statistiquesParRegion,
	}, nil
}

// GetCommissariats returns all commissariats
func (s *service) GetCommissariats(ctx context.Context) ([]*CommissariatResponse, error) {
	s.logger.Info("Getting commissariats")

	commList, err := s.commissariatRepo.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list commissariats: %w", err)
	}

	responses := make([]*CommissariatResponse, len(commList))
	for i, comm := range commList {
		responses[i] = s.commissariatToResponse(comm)
	}

	return responses, nil
}

// GetCommissariat returns a commissariat by ID
func (s *service) GetCommissariat(ctx context.Context, id string) (*CommissariatResponse, error) {
	s.logger.Info("Getting commissariat", zap.String("id", id))

	comm, err := s.commissariatRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.commissariatToResponse(comm), nil
}

// CreateCommissariat creates a new commissariat
func (s *service) CreateCommissariat(ctx context.Context, req *CreateCommissariatRequest) (*CommissariatResponse, error) {
	s.logger.Info("Creating commissariat", zap.String("code", req.Code))

	input := &repository.CreateCommissariatInput{
		ID:        uuid.New().String(),
		Nom:       req.Nom,
		Code:      req.Code,
		Adresse:   req.Adresse,
		Ville:     req.Ville,
		Region:    req.Region,
		Telephone: req.Telephone,
		Email:     req.Email,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	comm, err := s.commissariatRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create commissariat: %w", err)
	}

	return s.commissariatToResponse(comm), nil
}

// UpdateCommissariat updates a commissariat
func (s *service) UpdateCommissariat(ctx context.Context, id string, req *UpdateCommissariatRequest) (*CommissariatResponse, error) {
	s.logger.Info("Updating commissariat", zap.String("id", id))

	input := &repository.UpdateCommissariatInput{
		Nom:       req.Nom,
		Adresse:   req.Adresse,
		Ville:     req.Ville,
		Region:    req.Region,
		Telephone: req.Telephone,
		Email:     req.Email,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Actif:     req.Actif,
	}

	comm, err := s.commissariatRepo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update commissariat: %w", err)
	}

	return s.commissariatToResponse(comm), nil
}

// DeleteCommissariat deletes a commissariat
func (s *service) DeleteCommissariat(ctx context.Context, id string) error {
	s.logger.Info("Deleting commissariat", zap.String("id", id))
	return s.commissariatRepo.Delete(ctx, id)
}

// GetAgents returns agents, optionally filtered by commissariat
func (s *service) GetAgents(ctx context.Context, commissariatID *string) ([]*AgentResponse, error) {
	s.logger.Info("Getting agents")

	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	responses := make([]*AgentResponse, 0, len(users))
	for _, user := range users {
		// Filter by commissariat if specified
		if commissariatID != nil {
			if user.Edges.Commissariat == nil || user.Edges.Commissariat.ID.String() != *commissariatID {
				continue
			}
		}
		responses = append(responses, s.userToAgentResponse(user))
	}

	return responses, nil
}

// GetAgent returns an agent by ID
func (s *service) GetAgent(ctx context.Context, id string) (*AgentResponse, error) {
	s.logger.Info("Getting agent", zap.String("id", id))

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.userToAgentResponse(user), nil
}

// UpdateAgent updates an agent
func (s *service) UpdateAgent(ctx context.Context, id string, req *UpdateAgentRequest) (*AgentResponse, error) {
	s.logger.Info("Updating agent", zap.String("id", id))

	input := &repository.UpdateUserInput{
		Nom:            req.Nom,
		Prenom:         req.Prenom,
		Email:          req.Email,
		Role:           req.Role,
		Grade:          req.Grade,
		Telephone:      req.Telephone,
		StatutService:  req.StatutService,
		Localisation:   req.Localisation,
		Activite:       req.Activite,
		Active:         req.Actif,
		CommissariatID: req.CommissariatID,
	}

	user, err := s.userRepo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return s.userToAgentResponse(user), nil
}

// CreateAgent creates a new agent with hashed password
func (s *service) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*AgentResponse, error) {
	s.logger.Info("Creating agent", zap.String("matricule", req.Matricule))

	// Hash the password
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	input := &repository.CreateUserInput{
		ID:             uuid.New().String(),
		Matricule:      req.Matricule,
		Nom:            req.Nom,
		Prenom:         req.Prenom,
		Email:          req.Email,
		Password:       hashedPassword,
		Role:           req.Role,
		Grade:          req.Grade,
		Telephone:      req.Telephone,
		CommissariatID: req.CommissariatID,
	}

	user, err := s.userRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return s.userToAgentResponse(user), nil
}

// DeleteAgent deletes an agent
func (s *service) DeleteAgent(ctx context.Context, id string) error {
	s.logger.Info("Deleting agent", zap.String("id", id))
	return s.userRepo.Delete(ctx, id)
}

// GetAgentStatistiques returns statistics for an agent
func (s *service) GetAgentStatistiques(ctx context.Context, id string) (*AgentStatistiquesResponse, error) {
	s.logger.Info("Getting agent statistics", zap.String("id", id))

	// Verify agent exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get controles by agent
	controlesCount, err := s.controleRepo.Count(ctx, &repository.ControleFilters{AgentID: &id})
	if err != nil {
		s.logger.Error("Failed to count controles for agent", zap.Error(err))
		controlesCount = 0
	}

	// Get controle statistics for the agent
	controleStats, err := s.controleRepo.GetStatistics(ctx, &repository.ControleStatsFilters{AgentID: &id})
	if err != nil {
		s.logger.Error("Failed to get controle statistics", zap.Error(err))
		controleStats = &repository.ControleStatistics{ParJour: make(map[string]int)}
	}

	// Get infraction statistics for the agent
	infractionStats, err := s.infractionRepo.GetStatistics(ctx, &repository.InfractionStatsFilters{AgentID: &id})
	if err != nil {
		s.logger.Error("Failed to get infraction statistics", zap.Error(err))
		infractionStats = &repository.InfractionStatistics{}
	}

	// Get PV statistics
	pvStats, err := s.pvRepo.GetStatistics(ctx, &repository.PVFilters{AgentID: &id})
	if err != nil {
		s.logger.Error("Failed to get PV statistics", zap.Error(err))
		pvStats = &repository.PVStatistics{}
	}

	// Calculate taux infraction
	tauxInfraction := 0.0
	if controlesCount > 0 {
		tauxInfraction = float64(infractionStats.Total) / float64(controlesCount) * 100
	}

	// Calculate controles par jour (last 30 days average)
	controlesParJour := 0.0
	if len(controleStats.ParJour) > 0 {
		total := 0
		for _, count := range controleStats.ParJour {
			total += count
		}
		controlesParJour = float64(total) / float64(len(controleStats.ParJour))
	}

	// Utiliser le montant des PVs si disponible, sinon utiliser le montant des infractions
	montantTotal := pvStats.MontantTotal
	if montantTotal == 0 {
		montantTotal = infractionStats.MontantTotal
	}

	return &AgentStatistiquesResponse{
		AgentID:          id,
		TotalControles:   controlesCount,
		TotalInfractions: infractionStats.Total,
		TotalPV:          pvStats.Total,
		MontantTotalPV:   montantTotal,
		TauxInfraction:   tauxInfraction,
		ControlesParJour: controlesParJour,
		ControlesParMois: controleStats.ParJour,
	}, nil
}

// Helper functions

func (s *service) commissariatToResponse(comm *ent.Commissariat) *CommissariatResponse {
	resp := &CommissariatResponse{
		ID:        comm.ID.String(),
		Nom:       comm.Nom,
		Code:      comm.Code,
		Adresse:   comm.Adresse,
		Ville:     comm.Ville,
		Region:    comm.Region,
		Telephone: comm.Telephone,
		Email:     comm.Email,
		Actif:     comm.Actif,
		CreatedAt: comm.CreatedAt,
		UpdatedAt: comm.UpdatedAt,
	}

	if comm.Latitude != 0 {
		resp.Latitude = &comm.Latitude
	}
	if comm.Longitude != 0 {
		resp.Longitude = &comm.Longitude
	}

	return resp
}

func (s *service) userToAgentResponse(user *ent.User) *AgentResponse {
	resp := &AgentResponse{
		ID:            user.ID.String(),
		Matricule:     user.Matricule,
		Nom:           user.Nom,
		Prenom:        user.Prenom,
		Email:         user.Email,
		Role:          user.Role,
		Grade:         user.Grade,
		Telephone:     user.Telephone,
		StatutService: user.StatutService,
		Localisation:  user.Localisation,
		Activite:      user.Activite,
		Actif:         user.Active,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		GpsPrecision:  user.GpsPrecision,
		TempsService:  user.TempsService,
		CNI:           user.Cni,
		Adresse:       user.Adresse,
	}

	// Set derniere activite if available
	if !user.DerniereActivite.IsZero() {
		resp.DerniereActivite = &user.DerniereActivite
	}

	// Set date naissance if available
	if !user.DateNaissance.IsZero() {
		resp.DateNaissance = &user.DateNaissance
	}

	// Set date entree if available
	if !user.DateEntree.IsZero() {
		resp.DateEntree = &user.DateEntree
	}

	// Load commissariat if available
	if comm := user.Edges.Commissariat; comm != nil {
		resp.CommissariatID = comm.ID.String()
		resp.Commissariat = s.commissariatToResponse(comm)
	}

	// Load equipe if available
	if equipe := user.Edges.Equipe; equipe != nil {
		resp.Equipe = &EquipeResponse{
			ID:          equipe.ID.String(),
			Nom:         equipe.Nom,
			Code:        equipe.Code,
			Zone:        equipe.Zone,
			Description: equipe.Description,
			Active:      equipe.Active,
		}
	}

	// Load superieur if available
	if superieur := user.Edges.Superieur; superieur != nil {
		resp.Superieur = &SuperieurResponse{
			ID:        superieur.ID.String(),
			Nom:       superieur.Nom,
			Prenom:    superieur.Prenom,
			Grade:     superieur.Grade,
			Matricule: superieur.Matricule,
		}
	}

	// Load missions if available
	if missions := user.Edges.Missions; missions != nil {
		resp.Missions = make([]MissionResponse, len(missions))
		for i, m := range missions {
			missionResp := MissionResponse{
				ID:        m.ID.String(),
				Type:      m.Type,
				Titre:     m.Titre,
				DateDebut: m.DateDebut,
				Duree:     m.Duree,
				Zone:      m.Zone,
				Statut:    m.Statut,
				Rapport:   m.Rapport,
			}
			if !m.DateFin.IsZero() {
				missionResp.DateFin = &m.DateFin
			}
			resp.Missions[i] = missionResp
		}
	}

	// Load objectifs if available
	if objectifs := user.Edges.Objectifs; objectifs != nil {
		resp.Objectifs = make([]ObjectifResponse, len(objectifs))
		for i, o := range objectifs {
			resp.Objectifs[i] = ObjectifResponse{
				ID:             o.ID.String(),
				Titre:          o.Titre,
				Description:    o.Description,
				Periode:        o.Periode,
				DateDebut:      o.DateDebut,
				DateFin:        o.DateFin,
				Statut:         o.Statut,
				ValeurCible:    o.ValeurCible,
				ValeurActuelle: o.ValeurActuelle,
				Progression:    o.Progression,
			}
		}
	}

	// Load observations if available (only visible ones)
	if observations := user.Edges.Observations; observations != nil {
		resp.Observations = make([]ObservationResponse, 0, len(observations))
		for _, o := range observations {
			if o.VisibleAgent {
				resp.Observations = append(resp.Observations, ObservationResponse{
					ID:           o.ID.String(),
					Contenu:      o.Contenu,
					Type:         o.Type,
					Categorie:    o.Categorie,
					VisibleAgent: o.VisibleAgent,
					CreatedAt:    o.CreatedAt,
				})
			}
		}
	}

	// Load competences if available
	if competences := user.Edges.Competences; competences != nil {
		resp.Competences = make([]CompetenceResponse, len(competences))
		for i, c := range competences {
			compResp := CompetenceResponse{
				ID:          c.ID.String(),
				Nom:         c.Nom,
				Type:        c.Type,
				Description: c.Description,
				Organisme:   c.Organisme,
				Active:      c.Active,
			}
			if !c.DateObtention.IsZero() {
				compResp.DateObtention = &c.DateObtention
			}
			if !c.DateExpiration.IsZero() {
				compResp.DateExpiration = &c.DateExpiration
			}
			resp.Competences[i] = compResp
		}
	}

	return resp
}

// GetAgentsDashboard returns comprehensive dashboard data for agents management
func (s *service) GetAgentsDashboard(ctx context.Context, req *AgentDashboardRequest) (*AgentDashboardResponse, error) {
	s.logger.Info("Getting agents dashboard", zap.String("periode", req.Periode))

	// Get all users/agents
	users, err := s.userRepo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Get all commissariats
	commissariats, err := s.commissariatRepo.List(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to list commissariats", zap.Error(err))
		commissariats = []*ent.Commissariat{}
	}

	// Calculate agent statistics by status
	totalAgents := len(users)
	enService := 0
	enPause := 0
	horsService := 0

	for _, user := range users {
		switch user.StatutService {
		case "EN_SERVICE":
			enService++
		case "EN_PAUSE":
			enPause++
		default:
			horsService++
		}
	}

	// Get controles statistics
	controlesTotal, err := s.controleRepo.Count(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to count controles", zap.Error(err))
		controlesTotal = 0
	}

	// Get infraction statistics
	infractionStats, err := s.infractionRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get infraction stats", zap.Error(err))
		infractionStats = &repository.InfractionStatistics{}
	}

	// Get PV statistics for revenus
	pvStats, err := s.pvRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get PV stats", zap.Error(err))
		pvStats = &repository.PVStatistics{}
	}

	// Calculate performance moyenne (controles per agent)
	performanceMoyenne := 0.0
	if totalAgents > 0 {
		performanceMoyenne = float64(controlesTotal) / float64(totalAgents)
	}

	// Calculate taux reussite (infractions / controles)
	tauxReussite := 0.0
	if controlesTotal > 0 {
		tauxReussite = float64(infractionStats.Total) / float64(controlesTotal) * 100
	}

	// Build stats
	stats := AgentDashboardStats{
		TotalAgents:        totalAgents,
		EnService:          enService,
		EnPause:            enPause,
		HorsService:        horsService,
		ControlesTotal:     controlesTotal,
		InfractionsTotales: infractionStats.Total,
		RevenusTotal:       pvStats.MontantTotal / 1000000, // Convert to millions
		PerformanceMoyenne: performanceMoyenne,
		TempsServiceMoyen:  "8h00", // Default value - would need time tracking
		TauxReussite:       tauxReussite,
	}

	// Build pie data for status distribution
	pieData := []PieDataEntry{
		{Name: "En service", Value: calcPercent(enService, totalAgents), Color: "#10b981"},
		{Name: "En pause", Value: calcPercent(enPause, totalAgents), Color: "#f59e0b"},
		{Name: "Hors service", Value: calcPercent(horsService, totalAgents), Color: "#ef4444"},
	}

	// Build activity data based on period
	activityData := s.buildActivityData(ctx, req.Periode)

	// Build performance data by commissariat
	performanceData := s.buildPerformanceData(ctx, commissariats, users)

	// Build detailed agent list
	agents := s.buildDetailedAgentsList(ctx, users)

	// Build commissariat stats
	commissariatStats := s.buildCommissariatStats(ctx, commissariats, users)

	return &AgentDashboardResponse{
		Stats:           stats,
		ActivityData:    activityData,
		PerformanceData: performanceData,
		PieData:         pieData,
		Agents:          agents,
		Commissariats:   commissariatStats,
	}, nil
}

// Helper function to calculate percentage
func calcPercent(value, total int) int {
	if total == 0 {
		return 0
	}
	return int(float64(value) / float64(total) * 100)
}

// buildActivityData creates activity data based on period
func (s *service) buildActivityData(ctx context.Context, periode string) []ActivityDataEntry {
	// Get controle stats
	controleStats, err := s.controleRepo.GetStatistics(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to get controle statistics", zap.Error(err))
		return []ActivityDataEntry{}
	}

	// Build data based on periode - for now return data from ParJour
	data := make([]ActivityDataEntry, 0)

	switch periode {
	case "jour":
		// Hourly data - simplified
		periods := []string{"00h-04h", "04h-08h", "08h-12h", "12h-16h", "16h-20h", "20h-24h"}
		for i, p := range periods {
			// Distribute data across periods (simplified)
			total := 0
			for _, c := range controleStats.ParJour {
				total += c
			}
			estimate := total / 6
			if i == 2 || i == 3 { // Peak hours
				estimate = estimate * 2
			}
			data = append(data, ActivityDataEntry{
				Period:      p,
				Controles:   estimate,
				Agents:      10 + i*2,
				Infractions: estimate * 30 / 100,
			})
		}
	case "semaine":
		periods := []string{"Lun", "Mar", "Mer", "Jeu", "Ven", "Sam", "Dim"}
		for i, p := range periods {
			total := 0
			for _, c := range controleStats.ParJour {
				total += c
			}
			estimate := total / 7
			data = append(data, ActivityDataEntry{
				Period:      p,
				Controles:   estimate,
				Agents:      40 + i,
				Infractions: estimate * 30 / 100,
			})
		}
	case "mois":
		periods := []string{"Sem 1", "Sem 2", "Sem 3", "Sem 4"}
		for i, p := range periods {
			total := 0
			for _, c := range controleStats.ParJour {
				total += c
			}
			estimate := total / 4
			data = append(data, ActivityDataEntry{
				Period:      p,
				Controles:   estimate,
				Agents:      42 + i,
				Infractions: estimate * 30 / 100,
			})
		}
	default:
		// Use actual data from ParJour
		for date, count := range controleStats.ParJour {
			data = append(data, ActivityDataEntry{
				Period:      date,
				Controles:   count,
				Agents:      45,
				Infractions: count * 30 / 100,
			})
		}
	}

	return data
}

// buildPerformanceData creates performance data by commissariat
func (s *service) buildPerformanceData(ctx context.Context, commissariats []*ent.Commissariat, users []*ent.User) []PerformanceDataEntry {
	data := make([]PerformanceDataEntry, 0)

	for _, comm := range commissariats {
		agentCount := 0
		enServiceCount := 0

		for _, user := range users {
			if user.Edges.Commissariat != nil && user.Edges.Commissariat.ID == comm.ID {
				agentCount++
				if user.StatutService == "EN_SERVICE" {
					enServiceCount++
				}
			}
		}

		if agentCount == 0 {
			continue
		}

		tauxActivite := float64(enServiceCount) / float64(agentCount) * 100

		data = append(data, PerformanceDataEntry{
			Commissariat: comm.Nom,
			TauxActivite: tauxActivite,
			Agents:       agentCount,
		})
	}

	return data
}

// buildDetailedAgentsList creates detailed agent information
func (s *service) buildDetailedAgentsList(ctx context.Context, users []*ent.User) []AgentDetailedResponse {
	agents := make([]AgentDetailedResponse, 0, len(users))

	for _, user := range users {
		// Get agent statistics
		agentStats, err := s.GetAgentStatistiques(ctx, user.ID.String())
		if err != nil {
			agentStats = &AgentStatistiquesResponse{}
		}

		commissariatName := ""
		if user.Edges.Commissariat != nil {
			commissariatName = user.Edges.Commissariat.Nom
		}

		// Map status
		status := "HORS SERVICE"
		switch user.StatutService {
		case "EN_SERVICE":
			status = "EN SERVICE"
		case "EN_PAUSE":
			status = "EN PAUSE"
		}

		// Calculate performance
		performance := "Correcte"
		if agentStats.TauxInfraction > 40 {
			performance = "Excellente"
		} else if agentStats.TauxInfraction < 20 {
			performance = "Critique"
		}

		// Format derniere activite
		derniereActivite := "Aucune activité récente"
		if !user.DerniereActivite.IsZero() {
			derniereActivite = user.DerniereActivite.Format("02/01/2006 15:04")
		}

		agents = append(agents, AgentDetailedResponse{
			ID:               user.ID.String(),
			Nom:              user.Prenom + " " + user.Nom,
			Grade:            user.Grade,
			Commissariat:     commissariatName,
			Status:           status,
			Localisation:     user.Localisation,
			Activite:         user.Activite,
			Controles:        agentStats.TotalControles,
			Infractions:      agentStats.TotalInfractions,
			Revenus:          agentStats.MontantTotalPV,
			TauxInfractions:  agentStats.TauxInfraction,
			TempsService:     "8h00", // Would need time tracking
			Gps:              100,    // Would need GPS tracking
			DerniereActivite: derniereActivite,
			Performance:      performance,
		})
	}

	return agents
}

// buildCommissariatStats creates commissariat statistics
func (s *service) buildCommissariatStats(ctx context.Context, commissariats []*ent.Commissariat, users []*ent.User) []CommissariatStatsEntry {
	stats := make([]CommissariatStatsEntry, 0, len(commissariats))

	for _, comm := range commissariats {
		agentCount := 0
		enServiceCount := 0

		for _, user := range users {
			if user.Edges.Commissariat != nil && user.Edges.Commissariat.ID == comm.ID {
				agentCount++
				if user.StatutService == "EN_SERVICE" {
					enServiceCount++
				}
			}
		}

		// Get controles count for this commissariat
		commIDStr := comm.ID.String()
		controlesCount, _ := s.controleRepo.Count(ctx, &repository.ControleFilters{CommissariatID: &commIDStr})

		tauxActivite := 0.0
		if agentCount > 0 {
			tauxActivite = float64(enServiceCount) / float64(agentCount) * 100
		}

		stats = append(stats, CommissariatStatsEntry{
			Name:         comm.Nom,
			Agents:       agentCount,
			EnService:    enServiceCount,
			Controles:    controlesCount,
			TauxActivite: tauxActivite,
		})
	}

	return stats
}

// GetAgentSessions returns all active sessions for an agent
func (s *service) GetAgentSessions(ctx context.Context, agentID string) ([]*AgentSessionResponse, error) {
	s.logger.Info("Getting agent sessions", zap.String("agent_id", agentID))

	// Verify agent exists
	_, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	userID, err := uuid.Parse(agentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent ID")
	}

	sessions, err := s.sessionService.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	result := make([]*AgentSessionResponse, len(sessions))
	for i, sess := range sessions {
		result[i] = &AgentSessionResponse{
			ID:             sess.ID.String(),
			DeviceID:       sess.DeviceID,
			DeviceName:     sess.DeviceName,
			DeviceType:     sess.DeviceType,
			DeviceOS:       sess.DeviceOs,
			AppVersion:     sess.AppVersion,
			LastActivityAt: sess.LastActivityAt,
			LastIPAddress:  sess.LastIPAddress,
			SessionStarted: sess.SessionStartedAt,
			IsActive:       sess.IsActive && !sess.IsRevoked,
		}
	}

	return result, nil
}

// RevokeAgentSession revokes a specific session for an agent
func (s *service) RevokeAgentSession(ctx context.Context, agentID string, sessionID string, reason string) error {
	s.logger.Info("Revoking agent session",
		zap.String("agent_id", agentID),
		zap.String("session_id", sessionID),
		zap.String("reason", reason),
	)

	// Verify agent exists
	_, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found")
	}

	userID, err := uuid.Parse(agentID)
	if err != nil {
		return fmt.Errorf("invalid agent ID")
	}

	targetSessionID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID")
	}

	// Verify the session belongs to this agent
	sessions, err := s.sessionService.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	found := false
	for _, sess := range sessions {
		if sess.ID == targetSessionID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("session not found for this agent")
	}

	if reason == "" {
		reason = "admin_revoked"
	}

	return s.sessionService.RevokeSession(ctx, targetSessionID, reason)
}

// RevokeAllAgentSessions revokes all sessions for an agent
func (s *service) RevokeAllAgentSessions(ctx context.Context, agentID string, reason string) error {
	s.logger.Info("Revoking all agent sessions",
		zap.String("agent_id", agentID),
		zap.String("reason", reason),
	)

	// Verify agent exists
	_, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found")
	}

	userID, err := uuid.Parse(agentID)
	if err != nil {
		return fmt.Errorf("invalid agent ID")
	}

	if reason == "" {
		reason = "admin_revoked_all"
	}

	return s.sessionService.RevokeAllUserSessions(ctx, userID, reason)
}
