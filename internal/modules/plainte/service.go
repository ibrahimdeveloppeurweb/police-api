package plainte

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/plainte"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service interface defines plainte service methods
type Service interface {
	Create(ctx context.Context, req CreatePlainteRequest) (*PlainteResponse, error)
	GetByID(ctx context.Context, id string) (*PlainteResponse, error)
	GetByNumero(ctx context.Context, numero string) (*PlainteResponse, error)
	List(ctx context.Context, req ListPlaintesRequest) (*ListPlaintesResponse, error)
	Update(ctx context.Context, id string, req UpdatePlainteRequest) (*PlainteResponse, error)
	Delete(ctx context.Context, id string) error
	ChangerEtape(ctx context.Context, id string, req ChangerEtapeRequest) (*PlainteResponse, error)
	ChangerStatut(ctx context.Context, id string, req ChangerStatutRequest) (*PlainteResponse, error)
	AssignerAgent(ctx context.Context, id string, req AssignerAgentRequest) (*PlainteResponse, error)
	GetStatistics(ctx context.Context, req StatisticsRequest) (*PlainteStatisticsResponse, error)
	GetAlertes(ctx context.Context, commissariatID string) ([]AlerteResponse, error)
	GetTopAgents(ctx context.Context, commissariatID string) ([]TopAgentResponse, error)
	GetPreuves(ctx context.Context, plainteID string) ([]PreuveResponse, error)
	AddPreuve(ctx context.Context, plainteID string, req AddPreuveRequest) (*PreuveResponse, error)
	GetActesEnquete(ctx context.Context, plainteID string) ([]ActeEnqueteResponse, error)
	AddActeEnquete(ctx context.Context, plainteID string, req AddActeEnqueteRequest) (*ActeEnqueteResponse, error)
	GetTimeline(ctx context.Context, plainteID string) ([]TimelineEventResponse, error)
	AddTimelineEvent(ctx context.Context, plainteID string, req AddTimelineEventRequest) (*TimelineEventResponse, error)
	GetEnquetes(ctx context.Context, plainteID string) ([]EnqueteResponse, error)
	AddEnquete(ctx context.Context, plainteID string, req AddEnqueteRequest) (*EnqueteResponse, error)
	GetDecisions(ctx context.Context, plainteID string) ([]DecisionResponse, error)
	AddDecision(ctx context.Context, plainteID string, req AddDecisionRequest) (*DecisionResponse, error)
	GetHistorique(ctx context.Context, plainteID string) ([]HistoriqueResponse, error)
}

type service struct {
	client *ent.Client
	logger *zap.Logger
}

// NewService creates a new plainte service
func NewService(client *ent.Client, logger *zap.Logger) Service {
	return &service{
		client: client,
		logger: logger,
	}
}

func (s *service) Create(ctx context.Context, req CreatePlainteRequest) (*PlainteResponse, error) {
	s.logger.Info("Creating new plainte", zap.String("type", req.TypePlainte))

	// Generate unique ID and numero
	id := uuid.New()
	numero := fmt.Sprintf("PLT-%d-%s", time.Now().Year(), uuid.New().String()[:8])

	// Build create query
	create := s.client.Plainte.Create().
		SetID(id).
		SetNumero(numero).
		SetTypePlainte(req.TypePlainte).
		SetPlaignantNom(req.PlaignantNom).
		SetPlaignantPrenom(req.PlaignantPrenom).
		SetDateDepot(time.Now())

	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.PlaignantTelephone != nil {
		create.SetPlaignantTelephone(*req.PlaignantTelephone)
	}
	if req.PlaignantAdresse != nil {
		create.SetPlaignantAdresse(*req.PlaignantAdresse)
	}
	if req.PlaignantEmail != nil {
		create.SetPlaignantEmail(*req.PlaignantEmail)
	}
	if req.LieuFaits != nil {
		create.SetLieuFaits(*req.LieuFaits)
	}
	if req.DateFaits != nil {
		create.SetDateFaits(*req.DateFaits)
	}
	if req.Priorite != nil {
		create.SetPriorite(plainte.Priorite(*req.Priorite))
	}
	if req.Observations != nil {
		create.SetObservations(*req.Observations)
	}
	if req.CommissariatID != nil {
		commID, _ := uuid.Parse(*req.CommissariatID)
		create.SetCommissariatID(commID)
	}
	if req.AgentAssigneID != nil {
		agentID, _ := uuid.Parse(*req.AgentAssigneID)
		create.SetAgentAssigneID(agentID)
	}

	// Ajouter les suspects
	if len(req.Suspects) > 0 {
		suspectsData := make([]map[string]interface{}, len(req.Suspects))
		for i, suspect := range req.Suspects {
			suspectsData[i] = map[string]interface{}{
				"id":          uuid.New().String(),
				"nom":         suspect.Nom,
				"prenom":      suspect.Prenom,
				"description": suspect.Description,
				"adresse":     suspect.Adresse,
			}
		}
		create.SetSuspects(suspectsData)
	}

	// Ajouter les témoins
	if len(req.Temoins) > 0 {
		temoinsData := make([]map[string]interface{}, len(req.Temoins))
		for i, temoin := range req.Temoins {
			temoinsData[i] = map[string]interface{}{
				"id":        uuid.New().String(),
				"nom":       temoin.Nom,
				"prenom":    temoin.Prenom,
				"telephone": temoin.Telephone,
				"adresse":   temoin.Adresse,
			}
		}
		create.SetTemoins(temoinsData)
	}

	p, err := create.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create plainte", zap.Error(err))
		return nil, fmt.Errorf("failed to create plainte: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) GetByID(ctx context.Context, id string) (*PlainteResponse, error) {
	uid, _ := uuid.Parse(id)
	p, err := s.client.Plainte.Query().
		Where(plainte.ID(uid)).
		WithCommissariat().
		WithAgentAssigne().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("plainte not found")
		}
		return nil, fmt.Errorf("failed to get plainte: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) GetByNumero(ctx context.Context, numero string) (*PlainteResponse, error) {
	p, err := s.client.Plainte.Query().
		Where(plainte.Numero(numero)).
		WithCommissariat().
		WithAgentAssigne().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("plainte not found")
		}
		return nil, fmt.Errorf("failed to get plainte: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) List(ctx context.Context, req ListPlaintesRequest) (*ListPlaintesResponse, error) {
	query := s.client.Plainte.Query()

	// Apply filters
	if req.TypePlainte != nil {
		query = query.Where(plainte.TypePlainte(*req.TypePlainte))
	}
	if req.Statut != nil {
		query = query.Where(plainte.StatutEQ(plainte.Statut(*req.Statut)))
	}
	if req.Priorite != nil {
		query = query.Where(plainte.PrioriteEQ(plainte.Priorite(*req.Priorite)))
	}
	if req.EtapeActuelle != nil {
		query = query.Where(plainte.EtapeActuelleEQ(plainte.EtapeActuelle(*req.EtapeActuelle)))
	}
	if req.CommissariatID != nil {
		query = query.Where(plainte.HasCommissariat())
	}
	if req.AgentAssigneID != nil {
		agentID, _ := uuid.Parse(*req.AgentAssigneID)
		query = query.Where(plainte.HasAgentAssigneWith(user.ID(agentID)))
	}
	if req.DateDebut != nil {
		query = query.Where(plainte.DateDepotGTE(*req.DateDebut))
	}
	if req.DateFin != nil {
		query = query.Where(plainte.DateDepotLTE(*req.DateFin))
	}
	if req.Search != nil && *req.Search != "" {
		query = query.Where(
			plainte.Or(
				plainte.NumeroContains(*req.Search),
				plainte.PlaignantNomContains(*req.Search),
				plainte.PlaignantPrenomContains(*req.Search),
			),
		)
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count plaintes: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	} else {
		query = query.Limit(20)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Order by date depot descending
	query = query.Order(ent.Desc(plainte.FieldDateDepot))

	// Load with edges
	query = query.WithCommissariat().WithAgentAssigne()

	plaintes, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list plaintes: %w", err)
	}

	responses := make([]*PlainteResponse, len(plaintes))
	for i, p := range plaintes {
		resp, err := s.toResponse(ctx, p)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}

	return &ListPlaintesResponse{
		Plaintes: responses,
		Total:    total,
	}, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdatePlainteRequest) (*PlainteResponse, error) {
	uid, _ := uuid.Parse(id)
	update := s.client.Plainte.UpdateOneID(uid)

	if req.TypePlainte != nil {
		update.SetTypePlainte(*req.TypePlainte)
	}
	if req.Description != nil {
		update.SetDescription(*req.Description)
	}
	if req.PlaignantNom != nil {
		update.SetPlaignantNom(*req.PlaignantNom)
	}
	if req.PlaignantPrenom != nil {
		update.SetPlaignantPrenom(*req.PlaignantPrenom)
	}
	if req.PlaignantTelephone != nil {
		update.SetPlaignantTelephone(*req.PlaignantTelephone)
	}
	if req.PlaignantAdresse != nil {
		update.SetPlaignantAdresse(*req.PlaignantAdresse)
	}
	if req.PlaignantEmail != nil {
		update.SetPlaignantEmail(*req.PlaignantEmail)
	}
	if req.LieuFaits != nil {
		update.SetLieuFaits(*req.LieuFaits)
	}
	if req.DateFaits != nil {
		update.SetDateFaits(*req.DateFaits)
	}
	if req.Priorite != nil {
		update.SetPriorite(plainte.Priorite(*req.Priorite))
	}
	if req.Statut != nil {
		update.SetStatut(plainte.Statut(*req.Statut))
	}
	if req.EtapeActuelle != nil {
		update.SetEtapeActuelle(plainte.EtapeActuelle(*req.EtapeActuelle))
	}
	if req.Observations != nil {
		update.SetObservations(*req.Observations)
	}
	if req.DecisionFinale != nil {
		update.SetDecisionFinale(*req.DecisionFinale)
	}
	if req.CommissariatID != nil {
		commID, _ := uuid.Parse(*req.CommissariatID)
		update.SetCommissariatID(commID)
	}
	if req.AgentAssigneID != nil {
		agentID, _ := uuid.Parse(*req.AgentAssigneID)
		update.SetAgentAssigneID(agentID)
	}

	p, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("plainte not found")
		}
		return nil, fmt.Errorf("failed to update plainte: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) Delete(ctx context.Context, id string) error {
	uid, _ := uuid.Parse(id)
	err := s.client.Plainte.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("plainte not found")
		}
		return fmt.Errorf("failed to delete plainte: %w", err)
	}
	return nil
}

func (s *service) ChangerEtape(ctx context.Context, id string, req ChangerEtapeRequest) (*PlainteResponse, error) {
	uid, _ := uuid.Parse(id)
	update := s.client.Plainte.UpdateOneID(uid).
		SetEtapeActuelle(plainte.EtapeActuelle(req.Etape))

	if req.Observations != nil {
		update.SetObservations(*req.Observations)
	}

	// If moving to CLOTURE, set date resolution
	if req.Etape == "CLOTURE" {
		update.SetDateResolution(time.Now())
	}

	p, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("plainte not found")
		}
		return nil, fmt.Errorf("failed to change etape: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) ChangerStatut(ctx context.Context, id string, req ChangerStatutRequest) (*PlainteResponse, error) {
	uid, _ := uuid.Parse(id)
	update := s.client.Plainte.UpdateOneID(uid).
		SetStatut(plainte.Statut(req.Statut))

	if req.DecisionFinale != nil {
		update.SetDecisionFinale(*req.DecisionFinale)
	}

	// If resolved, set date resolution
	if req.Statut == "RESOLU" || req.Statut == "CLASSE" {
		update.SetDateResolution(time.Now())
	}

	p, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("plainte not found")
		}
		return nil, fmt.Errorf("failed to change statut: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) AssignerAgent(ctx context.Context, id string, req AssignerAgentRequest) (*PlainteResponse, error) {
	uid, _ := uuid.Parse(id)
	agentID, _ := uuid.Parse(req.AgentID)
	p, err := s.client.Plainte.UpdateOneID(uid).
		SetAgentAssigneID(agentID).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("plainte not found")
		}
		return nil, fmt.Errorf("failed to assign agent: %w", err)
	}

	return s.toResponse(ctx, p)
}

func (s *service) GetStatistics(ctx context.Context, req StatisticsRequest) (*PlainteStatisticsResponse, error) {
	// Build query with filters
	query := s.client.Plainte.Query()

	if req.CommissariatID != nil {
		// Filter by commissariat - just verify it has one
		query = query.Where(plainte.HasCommissariat())
	}
	if req.DateDebut != nil {
		query = query.Where(plainte.DateDepotGTE(*req.DateDebut))
	}
	if req.DateFin != nil {
		query = query.Where(plainte.DateDepotLTE(*req.DateFin))
	}

	plaintes, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaintes for statistics: %w", err)
	}

	stats := &PlainteStatisticsResponse{
		Total:       len(plaintes),
		ParType:     make(map[string]int),
		ParPriorite: make(map[string]int),
		ParEtape:    make(map[string]int),
	}

	var totalDelai float64
	var resolvedCount int

	for _, p := range plaintes {
		// Count by status
		switch p.Statut {
		case plainte.StatutEN_COURS:
			stats.EnCours++
		case plainte.StatutRESOLU:
			stats.Resolues++
		case plainte.StatutCLASSE:
			stats.Classees++
		case plainte.StatutTRANSFERE:
			stats.Transferees++
		}

		// Count by type
		stats.ParType[p.TypePlainte]++

		// Count by priority
		stats.ParPriorite[string(p.Priorite)]++

		// Count by etape
		stats.ParEtape[string(p.EtapeActuelle)]++

		// Count SLA exceeded
		if p.SLADepasse {
			stats.SLADepasse++
		}

		// Calculate average delay for resolved
		if p.DateResolution != nil {
			delai := p.DateResolution.Sub(p.DateDepot).Hours() / 24
			totalDelai += delai
			resolvedCount++
		}
	}

	if resolvedCount > 0 {
		stats.DelaiMoyenJours = totalDelai / float64(resolvedCount)
	}

	return stats, nil
}

func (s *service) toResponse(ctx context.Context, p *ent.Plainte) (*PlainteResponse, error) {
	resp := &PlainteResponse{
		ID:                 p.ID.String(),
		Numero:             p.Numero,
		TypePlainte:        p.TypePlainte,
		Description:        p.Description,
		PlaignantNom:       p.PlaignantNom,
		PlaignantPrenom:    p.PlaignantPrenom,
		PlaignantTelephone: p.PlaignantTelephone,
		PlaignantAdresse:   p.PlaignantAdresse,
		PlaignantEmail:     p.PlaignantEmail,
		DateDepot:          p.DateDepot,
		DateResolution:     p.DateResolution,
		EtapeActuelle:      string(p.EtapeActuelle),
		Priorite:           string(p.Priorite),
		Statut:             string(p.Statut),
		DelaiSLA:           p.DelaiSLA,
		SLADepasse:         p.SLADepasse,
		LieuFaits:          p.LieuFaits,
		DateFaits:          p.DateFaits,
		Observations:       p.Observations,
		DecisionFinale:     p.DecisionFinale,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}

	// Load edges if not already loaded
	if p.Edges.Commissariat != nil {
		resp.Commissariat = &CommissariatSummary{
			ID:   p.Edges.Commissariat.ID.String(),
			Nom:  p.Edges.Commissariat.Nom,
			Code: p.Edges.Commissariat.Code,
		}
	}

	if p.Edges.AgentAssigne != nil {
		resp.AgentAssigne = &AgentSummary{
			ID:        p.Edges.AgentAssigne.ID.String(),
			Matricule: p.Edges.AgentAssigne.Matricule,
			Nom:       p.Edges.AgentAssigne.Nom,
			Prenom:    p.Edges.AgentAssigne.Prenom,
		}
	}


	// Ajouter les suspects
	if p.Suspects != nil && len(p.Suspects) > 0 {
		resp.Suspects = make([]SuspectResponse, len(p.Suspects))
		for i, suspectMap := range p.Suspects {
			suspect := SuspectResponse{}
			if id, ok := suspectMap["id"].(string); ok {
				suspect.ID = id
			}
			if nom, ok := suspectMap["nom"].(string); ok {
				suspect.Nom = nom
			}
			if prenom, ok := suspectMap["prenom"].(string); ok {
				suspect.Prenom = prenom
			}
			if desc, ok := suspectMap["description"].(string); ok && desc != "" {
				suspect.Description = &desc
			}
			if addr, ok := suspectMap["adresse"].(string); ok && addr != "" {
				suspect.Adresse = &addr
			}
			resp.Suspects[i] = suspect
		}
	}

	// Ajouter les témoins
	if p.Temoins != nil && len(p.Temoins) > 0 {
		resp.Temoins = make([]TemoinResponse, len(p.Temoins))
		for i, temoinMap := range p.Temoins {
			temoin := TemoinResponse{}
			if id, ok := temoinMap["id"].(string); ok {
				temoin.ID = id
			}
			if nom, ok := temoinMap["nom"].(string); ok {
				temoin.Nom = nom
			}
			if prenom, ok := temoinMap["prenom"].(string); ok {
				temoin.Prenom = prenom
			}
			if tel, ok := temoinMap["telephone"].(string); ok && tel != "" {
				temoin.Telephone = &tel
			}
			if addr, ok := temoinMap["adresse"].(string); ok && addr != "" {
				temoin.Adresse = &addr
			}
			resp.Temoins[i] = temoin
		}
	}

	return resp, nil
}
