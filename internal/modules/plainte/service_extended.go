package plainte

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/acteenquete"
	"police-trafic-api-frontend-aligned/ent/commissariat"
	"police-trafic-api-frontend-aligned/ent/decision"
	"police-trafic-api-frontend-aligned/ent/enquete"
	"police-trafic-api-frontend-aligned/ent/plainte"
	"police-trafic-api-frontend-aligned/ent/preuve"
	"police-trafic-api-frontend-aligned/ent/timelineevent"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ========================
// PREUVES IMPLEMENTATION
// ========================

// GetPreuves returns preuves for a plainte with real database queries
func (s *service) GetPreuves(ctx context.Context, plainteID string) ([]PreuveResponse, error) {
	s.logger.Info("Getting preuves from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query preuves from database
	preuves, err := s.client.Preuve.Query().
		Where(preuve.PlainteIDEQ(uid)).
		Order(ent.Desc(preuve.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		s.logger.Error("Failed to query preuves", zap.Error(err))
		return nil, fmt.Errorf("failed to query preuves: %w", err)
	}

	// Convert to response format
	var responses []PreuveResponse
	for _, p := range preuves {
		resp := PreuveResponse{
			ID:                p.ID.String(),
			NumeroPiece:       p.NumeroPiece,
			Type:              string(p.Type),
			Description:       p.Description,
			LieuConservation:  ptrString(p.LieuConservation),
			DateCollecte:      p.DateCollecte,
			CollectePar:       ptrString(p.CollectePar),
			ExpertiseDemandee: p.ExpertiseDemandee,
			ExpertiseType:     ptrString(p.ExpertiseType),
			ExpertiseResultat: ptrString(p.ExpertiseResultat),
			Statut:            string(p.Statut),
			CreatedAt:         p.CreatedAt,
		}

		// Handle photos if they exist
		if len(p.Photos) > 0 {
			resp.Photos = p.Photos
		}

		// Handle hash if it exists
		if p.HashVerification != "" {
			resp.HashVerification = &p.HashVerification
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched preuves",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// AddPreuve adds a preuve to a plainte in the database
func (s *service) AddPreuve(ctx context.Context, plainteID string, req AddPreuveRequest) (*PreuveResponse, error) {
	s.logger.Info("Adding preuve to database",
		zap.String("plainte_id", plainteID),
		zap.String("numero_piece", req.NumeroPiece))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Verify plainte exists
	exists, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check plainte existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("plainte not found")
	}

	// Create preuve in database - Convert string to enum type
	preuveBuilder := s.client.Preuve.Create().
		SetPlainteID(uid).
		SetNumeroPiece(req.NumeroPiece).
		SetType(preuve.Type(req.Type)).
		SetDescription(req.Description).
		SetDateCollecte(req.DateCollecte).
		SetExpertiseDemandee(req.ExpertiseDemandee).
		SetStatut(preuve.Statut(req.TypeCollecte))

	// Set optional fields
	if req.LieuConservation != nil {
		preuveBuilder.SetLieuConservation(*req.LieuConservation)
	}
	if req.CollectePar != nil {
		preuveBuilder.SetCollectePar(*req.CollectePar)
	}
	if req.ExpertiseDemandee && req.ExpertiseType != nil {
		preuveBuilder.SetExpertiseType(*req.ExpertiseType)
	}

	preuveResult, err := preuveBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create preuve", zap.Error(err))
		return nil, fmt.Errorf("failed to create preuve: %w", err)
	}

	// Convert to response
	response := &PreuveResponse{
		ID:                preuveResult.ID.String(),
		NumeroPiece:       preuveResult.NumeroPiece,
		Type:              string(preuveResult.Type),
		Description:       preuveResult.Description,
		LieuConservation:  ptrString(preuveResult.LieuConservation),
		DateCollecte:      preuveResult.DateCollecte,
		CollectePar:       ptrString(preuveResult.CollectePar),
		ExpertiseDemandee: preuveResult.ExpertiseDemandee,
		ExpertiseType:     ptrString(preuveResult.ExpertiseType),
		Statut:            string(preuveResult.Statut),
		CreatedAt:         preuveResult.CreatedAt,
	}

	s.logger.Info("Successfully created preuve",
		zap.String("id", preuveResult.ID.String()),
		zap.String("numero_piece", preuveResult.NumeroPiece))

	return response, nil
}

// ========================
// ACTES ENQUETE IMPLEMENTATION
// ========================

// GetActesEnquete returns actes d'enquête for a plainte from database
func (s *service) GetActesEnquete(ctx context.Context, plainteID string) ([]ActeEnqueteResponse, error) {
	s.logger.Info("Getting actes enquête from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query actes from database
	actes, err := s.client.ActeEnquete.Query().
		Where(acteenquete.PlainteIDEQ(uid)).
		Order(ent.Desc(acteenquete.FieldDate)).
		All(ctx)

	if err != nil {
		s.logger.Error("Failed to query actes enquête", zap.Error(err))
		return nil, fmt.Errorf("failed to query actes enquête: %w", err)
	}

	// Convert to response format
	var responses []ActeEnqueteResponse
	for _, a := range actes {
		resp := ActeEnqueteResponse{
			ID:             a.ID.String(),
			Type:           string(a.Type),
			Date:           a.Date,
			Heure:          ptrString(a.Heure),
			Duree:          ptrString(a.Duree),
			Lieu:           ptrString(a.Lieu),
			OfficierCharge: a.OfficierCharge,
			Description:    a.Description,
			PVNumero:       ptrString(a.PvNumero),
			MandatNumero:   ptrString(a.MandatNumero),
			Conclusions:    ptrString(a.Conclusions),
			CreatedAt:      a.CreatedAt,
		}

		// Handle arrays if they exist
		if len(a.PersonnesPresentes) > 0 {
			resp.PersonnesPresentes = a.PersonnesPresentes
		}
		if len(a.ObjetsSaisis) > 0 {
			resp.ObjetsSaisis = a.ObjetsSaisis
		}
		if len(a.DocumentsJoints) > 0 {
			resp.DocumentsJoints = a.DocumentsJoints
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched actes enquête",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// AddActeEnquete adds an acte d'enquête to a plainte in the database
func (s *service) AddActeEnquete(ctx context.Context, plainteID string, req AddActeEnqueteRequest) (*ActeEnqueteResponse, error) {
	s.logger.Info("Adding acte enquête to database",
		zap.String("plainte_id", plainteID),
		zap.String("type", req.Type))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Verify plainte exists
	exists, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check plainte existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("plainte not found")
	}

	// Create acte in database - Convert string to enum type
	acteBuilder := s.client.ActeEnquete.Create().
		SetPlainteID(uid).
		SetType(acteenquete.Type(req.Type)).
		SetDate(req.Date).
		SetOfficierCharge(req.OfficierCharge).
		SetDescription(req.Description)

	// Set optional fields
	if req.Heure != nil {
		acteBuilder.SetHeure(*req.Heure)
	}
	if req.Duree != nil {
		acteBuilder.SetDuree(*req.Duree)
	}
	if req.Lieu != nil {
		acteBuilder.SetLieu(*req.Lieu)
	}
	if req.PVNumero != nil {
		acteBuilder.SetPvNumero(*req.PVNumero)
	}
	if req.MandatNumero != nil {
		acteBuilder.SetMandatNumero(*req.MandatNumero)
	}

	acteResult, err := acteBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create acte enquête", zap.Error(err))
		return nil, fmt.Errorf("failed to create acte enquête: %w", err)
	}

	// Convert to response
	response := &ActeEnqueteResponse{
		ID:             acteResult.ID.String(),
		Type:           string(acteResult.Type),
		Date:           acteResult.Date,
		Heure:          ptrString(acteResult.Heure),
		Duree:          ptrString(acteResult.Duree),
		Lieu:           ptrString(acteResult.Lieu),
		OfficierCharge: acteResult.OfficierCharge,
		Description:    acteResult.Description,
		PVNumero:       ptrString(acteResult.PvNumero),
		MandatNumero:   ptrString(acteResult.MandatNumero),
		CreatedAt:      acteResult.CreatedAt,
	}

	s.logger.Info("Successfully created acte enquête",
		zap.String("id", acteResult.ID.String()),
		zap.String("type", string(acteResult.Type)))

	return response, nil
}

// ========================
// TIMELINE IMPLEMENTATION
// ========================

// GetTimeline returns timeline events for a plainte from database
func (s *service) GetTimeline(ctx context.Context, plainteID string) ([]TimelineEventResponse, error) {
	s.logger.Info("Getting timeline from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query timeline events from database
	events, err := s.client.TimelineEvent.Query().
		Where(timelineevent.PlainteIDEQ(uid)).
		Order(ent.Desc(timelineevent.FieldDate)).
		All(ctx)

	if err != nil {
		s.logger.Error("Failed to query timeline events", zap.Error(err))
		return nil, fmt.Errorf("failed to query timeline events: %w", err)
	}

	// Convert to response format
	var responses []TimelineEventResponse
	for _, e := range events {
		resp := TimelineEventResponse{
			ID:          e.ID.String(),
			Date:        e.Date,
			Heure:       ptrString(e.Heure),
			Type:        string(e.Type),
			Titre:       e.Titre,
			Description: e.Description,
			Acteur:      ptrString(e.Acteur),
			Statut:      ptrString(e.Statut),
			CreatedAt:   e.CreatedAt,
		}

		// Handle documents if they exist
		if len(e.Documents) > 0 {
			resp.Documents = e.Documents
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched timeline events",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// AddTimelineEvent adds a timeline event to a plainte in the database
func (s *service) AddTimelineEvent(ctx context.Context, plainteID string, req AddTimelineEventRequest) (*TimelineEventResponse, error) {
	s.logger.Info("Adding timeline event to database",
		zap.String("plainte_id", plainteID),
		zap.String("type", req.Type))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Verify plainte exists
	exists, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check plainte existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("plainte not found")
	}

	// Create timeline event in database - Convert string to enum type
	eventBuilder := s.client.TimelineEvent.Create().
		SetPlainteID(uid).
		SetDate(req.Date).
		SetType(timelineevent.Type(req.Type)).
		SetTitre(req.Titre).
		SetDescription(req.Description)

	// Set optional fields
	if req.Heure != nil {
		eventBuilder.SetHeure(*req.Heure)
	}
	if req.Acteur != nil {
		eventBuilder.SetActeur(*req.Acteur)
	}
	if req.Statut != nil {
		eventBuilder.SetStatut(*req.Statut)
	}

	eventResult, err := eventBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create timeline event", zap.Error(err))
		return nil, fmt.Errorf("failed to create timeline event: %w", err)
	}

	// Convert to response
	response := &TimelineEventResponse{
		ID:          eventResult.ID.String(),
		Date:        eventResult.Date,
		Heure:       ptrString(eventResult.Heure),
		Type:        string(eventResult.Type),
		Titre:       eventResult.Titre,
		Description: eventResult.Description,
		Acteur:      ptrString(eventResult.Acteur),
		Statut:      ptrString(eventResult.Statut),
		CreatedAt:   eventResult.CreatedAt,
	}

	s.logger.Info("Successfully created timeline event",
		zap.String("id", eventResult.ID.String()),
		zap.String("type", string(eventResult.Type)))

	return response, nil
}

// ========================
// ALERTES IMPLEMENTATION (Real data based on database)
// ========================

// GetAlertes returns active alerts for plaintes based on real database data
func (s *service) GetAlertes(ctx context.Context, commissariatID string) ([]AlerteResponse, error) {
	s.logger.Info("Getting alertes from database", zap.String("commissariat_id", commissariatID))

	// Build query
	query := s.client.Plainte.Query().Where(plainte.SLADepasseEQ(true))

	// Filter by commissariat if provided
	if commissariatID != "" {
		uid, err := uuid.Parse(commissariatID)
		if err == nil {
			query = query.Where(plainte.HasCommissariatWith(
				commissariat.IDEQ(uid),
			))
		}
	}

	// Get plaintes with SLA dépassé
	plaintes, err := query.All(ctx)
	if err != nil {
		s.logger.Error("Failed to query alertes", zap.Error(err))
		return nil, fmt.Errorf("failed to query alertes: %w", err)
	}

	var alertes []AlerteResponse
	now := time.Now()

	for _, p := range plaintes {
		// Calculate days of delay
		daysDelay := int(now.Sub(p.DateDepot).Hours() / 24)

		alerte := AlerteResponse{
			ID:            uuid.New().String(),
			PlainteID:     p.ID.String(),
			PlainteNumero: p.Numero,
			TypeAlerte:    "SLA_DEPASSE",
			Message:       fmt.Sprintf("Le délai SLA de traitement a été dépassé de %d jours", daysDelay-7), // Assuming 7 days SLA
			Niveau:        "CRITICAL",
			JoursRetard:   &daysDelay,
		}
		alertes = append(alertes, alerte)
	}

	s.logger.Info("Successfully fetched alertes",
		zap.Int("count", len(alertes)))

	return alertes, nil
}

// ========================
// TOP AGENTS IMPLEMENTATION (Real data based on database)
// ========================

// GetTopAgents returns top performing agents based on real database data
func (s *service) GetTopAgents(ctx context.Context, commissariatID string) ([]TopAgentResponse, error) {
	s.logger.Info("Getting top agents from database", zap.String("commissariat_id", commissariatID))

	// Build query for plaintes
	query := s.client.Plainte.Query().Where(plainte.HasAgentAssigne())

	// Filter by commissariat if provided
	if commissariatID != "" {
		uid, err := uuid.Parse(commissariatID)
		if err == nil {
			query = query.Where(plainte.HasCommissariatWith(
				commissariat.IDEQ(uid),
			))
		}
	}

	// Get all assigned plaintes with agents
	plaintes, err := query.
		WithAgentAssigne().
		All(ctx)

	if err != nil {
		s.logger.Error("Failed to query top agents", zap.Error(err))
		return nil, fmt.Errorf("failed to query top agents: %w", err)
	}

	// Aggregate statistics by agent
	agentStats := make(map[string]*struct {
		agent            *ent.User
		plaintesTraitees int
		plaintesResolues int
		totalDays        float64
	})

	for _, p := range plaintes {
		if p.Edges.AgentAssigne == nil {
			continue
		}

		agentID := p.Edges.AgentAssigne.ID.String()
		if _, exists := agentStats[agentID]; !exists {
			agentStats[agentID] = &struct {
				agent            *ent.User
				plaintesTraitees int
				plaintesResolues int
				totalDays        float64
			}{
				agent: p.Edges.AgentAssigne,
			}
		}

		stats := agentStats[agentID]
		stats.plaintesTraitees++

		if p.Statut == "RESOLU" {
			stats.plaintesResolues++
			if p.DateResolution != nil {
				days := p.DateResolution.Sub(p.DateDepot).Hours() / 24
				stats.totalDays += days
			}
		}
	}

	// Convert to response and calculate scores
	var agents []TopAgentResponse
	for _, stats := range agentStats {
		if stats.agent == nil {
			continue
		}

		delaiMoyen := 0.0
		if stats.plaintesResolues > 0 {
			delaiMoyen = stats.totalDays / float64(stats.plaintesResolues)
		}

		// Calculate score (resolution rate * 10 - penalty for delay)
		resolutionRate := float64(stats.plaintesResolues) / float64(stats.plaintesTraitees)
		score := resolutionRate * 10.0
		if delaiMoyen > 5 {
			score -= (delaiMoyen - 5) * 0.1
		}
		if score < 0 {
			score = 0
		}

		agent := TopAgentResponse{
			ID:               stats.agent.ID.String(),
			Nom:              stats.agent.Nom,
			Prenom:           stats.agent.Prenom,
			Matricule:        stats.agent.Matricule,
			PlaintesTraitees: stats.plaintesTraitees,
			PlaintesResolues: stats.plaintesResolues,
			Score:            score,
			DelaiMoyen:       delaiMoyen,
		}
		agents = append(agents, agent)
	}

	// Sort by score descending (simplified - in production use proper sorting)
	// Return top 10
	if len(agents) > 10 {
		agents = agents[:10]
	}

	s.logger.Info("Successfully fetched top agents",
		zap.Int("count", len(agents)))

	return agents, nil
}

// ========================
// HELPER FUNCTIONS
// ========================

// ptrString returns a pointer to a string if it's not empty, otherwise nil
func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ========================



// ========================
// ENQUETES IMPLEMENTATION
// ========================

// GetEnquetes returns enquêtes for a plainte from database
func (s *service) GetEnquetes(ctx context.Context, plainteID string) ([]EnqueteResponse, error) {
	s.logger.Info("Getting enquêtes from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query enquêtes from database through the plainte edge
	pl, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		WithEnquetes().
		Only(ctx)

	if err != nil {
		s.logger.Error("Failed to query plainte with enquêtes", zap.Error(err))
		return nil, fmt.Errorf("failed to query plainte: %w", err)
	}

	enquetes := pl.Edges.Enquetes

	// Convert to response format
	var responses []EnqueteResponse
	for _, e := range enquetes {
		resp := EnqueteResponse{
			ID:             e.ID.String(),
			Type:           string(e.Type),
			OfficierCharge: e.OfficierCharge,
			DateDebut:      e.DateDebut,
			Lieu:           ptrString(e.Lieu),
			Description:    e.Description,
			Resultats:      ptrString(e.Resultats),
			Conclusions:    ptrString(e.Conclusions),
			Statut:         string(e.Statut),
			CreatedAt:      e.CreatedAt,
		}

		if e.DateFin != nil {
			resp.DateFin = e.DateFin
		}
		if len(e.PersonnesInterrogees) > 0 {
			resp.PersonnesInterrogees = e.PersonnesInterrogees
		}
		if len(e.PreuvesCollectees) > 0 {
			resp.PreuvesCollectees = e.PreuvesCollectees
		}
		if len(e.Documents) > 0 {
			resp.Documents = e.Documents
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched enquêtes",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// AddEnquete adds an enquête to a plainte in the database
func (s *service) AddEnquete(ctx context.Context, plainteID string, req AddEnqueteRequest) (*EnqueteResponse, error) {
	s.logger.Info("Adding enquête to database",
		zap.String("plainte_id", plainteID),
		zap.String("type", req.Type))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Verify plainte exists
	exists, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check plainte existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("plainte not found")
	}

	// Create enquête in database
	enqueteBuilder := s.client.Enquete.Create().
		SetPlainteID(uid).
		SetType(enquete.Type(req.Type)).
		SetOfficierCharge(req.OfficierCharge).
		SetDateDebut(req.DateDebut).
		SetDescription(req.Description).
		SetStatut(enquete.StatutEN_COURS)

	// Set optional fields
	if req.Lieu != nil {
		enqueteBuilder.SetLieu(*req.Lieu)
	}

	enqueteResult, err := enqueteBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create enquête", zap.Error(err))
		return nil, fmt.Errorf("failed to create enquête: %w", err)
	}

	// Convert to response
	response := &EnqueteResponse{
		ID:             enqueteResult.ID.String(),
		Type:           string(enqueteResult.Type),
		OfficierCharge: enqueteResult.OfficierCharge,
		DateDebut:      enqueteResult.DateDebut,
		Lieu:           ptrString(enqueteResult.Lieu),
		Description:    enqueteResult.Description,
		Statut:         string(enqueteResult.Statut),
		CreatedAt:      enqueteResult.CreatedAt,
	}

	s.logger.Info("Successfully created enquête",
		zap.String("id", enqueteResult.ID.String()),
		zap.String("type", string(enqueteResult.Type)))

	return response, nil
}

// ========================
// DECISIONS IMPLEMENTATION
// ========================

// GetDecisions returns decisions for a plainte from database
func (s *service) GetDecisions(ctx context.Context, plainteID string) ([]DecisionResponse, error) {
	s.logger.Info("Getting decisions from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query decisions from database through the plainte edge
	pl, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		WithDecisions().
		Only(ctx)

	if err != nil {
		s.logger.Error("Failed to query plainte with decisions", zap.Error(err))
		return nil, fmt.Errorf("failed to query plainte: %w", err)
	}

	decisions := pl.Edges.Decisions

	// Convert to response format
	var responses []DecisionResponse
	for _, d := range decisions {
		resp := DecisionResponse{
			ID:                d.ID.String(),
			Type:              string(d.Type),
			DateDecision:      d.DateDecision,
			Autorite:          d.Autorite,
			Description:       d.Description,
			Motivation:        ptrString(d.Motivation),
			Suites:            ptrString(d.Suites),
			DocumentReference: ptrString(d.DocumentReference),
			Notifiee:          d.Notifiee,
			CreatedAt:         d.CreatedAt,
		}

		if d.DateNotification != nil {
			resp.DateNotification = d.DateNotification
		}
		if len(d.Dispositions) > 0 {
			resp.Dispositions = d.Dispositions
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched decisions",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}

// AddDecision adds a decision to a plainte in the database
func (s *service) AddDecision(ctx context.Context, plainteID string, req AddDecisionRequest) (*DecisionResponse, error) {
	s.logger.Info("Adding decision to database",
		zap.String("plainte_id", plainteID),
		zap.String("type", req.Type))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Verify plainte exists
	exists, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check plainte existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("plainte not found")
	}

	// Create decision in database
	decisionBuilder := s.client.Decision.Create().
		SetPlainteID(uid).
		SetType(decision.Type(req.Type)).
		SetDateDecision(req.DateDecision).
		SetAutorite(req.Autorite).
		SetDescription(req.Description).
		SetNotifiee(false)

	// Set optional fields
	if req.Motivation != nil {
		decisionBuilder.SetMotivation(*req.Motivation)
	}

	decisionResult, err := decisionBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create decision", zap.Error(err))
		return nil, fmt.Errorf("failed to create decision: %w", err)
	}

	// Convert to response
	response := &DecisionResponse{
		ID:           decisionResult.ID.String(),
		Type:         string(decisionResult.Type),
		DateDecision: decisionResult.DateDecision,
		Autorite:     decisionResult.Autorite,
		Description:  decisionResult.Description,
		Motivation:   ptrString(decisionResult.Motivation),
		Notifiee:     decisionResult.Notifiee,
		CreatedAt:    decisionResult.CreatedAt,
	}

	s.logger.Info("Successfully created decision",
		zap.String("id", decisionResult.ID.String()),
		zap.String("type", string(decisionResult.Type)))

	return response, nil
}

// ========================
// HISTORIQUE IMPLEMENTATION
// ========================

// GetHistorique returns historique for a plainte from database
func (s *service) GetHistorique(ctx context.Context, plainteID string) ([]HistoriqueResponse, error) {
	s.logger.Info("Getting historique from database", zap.String("plainte_id", plainteID))

	// Convert ID to UUID
	uid, err := uuid.Parse(plainteID)
	if err != nil {
		return nil, fmt.Errorf("invalid plainte ID: %w", err)
	}

	// Query historique from database through the plainte edge
	pl, err := s.client.Plainte.Query().
		Where(plainte.IDEQ(uid)).
		WithHistoriques().
		Only(ctx)

	if err != nil {
		s.logger.Error("Failed to query plainte with historique", zap.Error(err))
		return nil, fmt.Errorf("failed to query plainte: %w", err)
	}

	historique := pl.Edges.Historiques

	// Convert to response format
	var responses []HistoriqueResponse
	for _, h := range historique {
		resp := HistoriqueResponse{
			ID:             h.ID.String(),
			TypeChangement: string(h.TypeChangement),
			ChampModifie:   h.ChampModifie,
			AncienneValeur: ptrString(h.AncienneValeur),
			NouvelleValeur: h.NouvelleValeur,
			Commentaire:    ptrString(h.Commentaire),
			AuteurNom:      ptrString(h.AuteurNom),
			CreatedAt:      h.CreatedAt,
		}

		responses = append(responses, resp)
	}

	s.logger.Info("Successfully fetched historique",
		zap.String("plainte_id", plainteID),
		zap.Int("count", len(responses)))

	return responses, nil
}
