package alertes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/alertesecuritaire"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines alertes service interface
type Service interface {
	// CRUD de base
	Create(ctx context.Context, req *CreateAlerteRequest, agentID string) (*AlerteResponse, error)
	GetByID(ctx context.Context, id string) (*AlerteResponse, error)
	GetByNumero(ctx context.Context, numero string) (*AlerteResponse, error)
	List(ctx context.Context, filters *FilterAlertesRequest, role, userID, commissariatID string) (*ListAlertesResponse, error)
	Update(ctx context.Context, id string, req *UpdateAlerteRequest) (*AlerteResponse, error)
	Delete(ctx context.Context, id string) error
	
	// Actions principales
	AddSuivi(ctx context.Context, id string, req *AddSuiviRequest, agentID string) (*AlerteResponse, error)
	Diffuser(ctx context.Context, id string, req *BroadcastAlerteRequest, agentID string) (*AlerteResponse, error)
	DiffusionInterne(ctx context.Context, id string, req *AssignAlerteRequest, commissariatID, agentID string) (*AlerteResponse, error)
	Assigner(ctx context.Context, id string, req *AssignAlerteRequest, commissariatID, agentID string) (*AlerteResponse, error)
	Resoudre(ctx context.Context, id string, agentID string) (*AlerteResponse, error)
	Archiver(ctx context.Context, id string, agentID string) (*AlerteResponse, error)
	Cloturer(ctx context.Context, id string, agentID string) (*AlerteResponse, error)
	
	// Intervention
	DeployIntervention(ctx context.Context, id string, req *DeployInterventionRequest, agentID string) (*AlerteResponse, error)
	UpdateIntervention(ctx context.Context, id string, req *UpdateInterventionRequest, agentID string) (*AlerteResponse, error)
	
	// Évaluation et rapport
	AddEvaluation(ctx context.Context, id string, req *AddEvaluationRequest, agentID string) (*AlerteResponse, error)
	AddRapport(ctx context.Context, id string, req *AddRapportRequest, agentID string) (*AlerteResponse, error)
	
	// Témoins, documents et photos
	AddTemoin(ctx context.Context, id string, req *AddTemoinRequest, agentID string) (*AlerteResponse, error)
	AddDocument(ctx context.Context, id string, req *AddDocumentRequest, agentID string) (*AlerteResponse, error)
	AddPhotos(ctx context.Context, id string, photos []string, agentID string) (*AlerteResponse, error)
	
	// Actions
	UpdateActions(ctx context.Context, id string, req *UpdateActionsRequest, agentID string) (*AlerteResponse, error)
	
	// Utilitaires
	GetActives(ctx context.Context, commissariatID *string) ([]*AlerteResponse, error)
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesAlertesResponse, error)
	GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardResponse, error)
	GenererDescription(ctx context.Context, req *GenerateDescriptionRequest) (*GenerateDescriptionResponse, error)
	GenererRapport(ctx context.Context, alerteID string) (*GenerateRapportResponse, error)
}

// service implements alertes service
type service struct {
	alerteRepo       repository.AlerteRepository
	userRepo         repository.UserRepository
	commissariatRepo repository.CommissariatRepository
	config           *config.Config
	logger           *zap.Logger
}

// NewService creates a new alertes service
func NewService(
	alerteRepo repository.AlerteRepository,
	userRepo repository.UserRepository,
	commissariatRepo repository.CommissariatRepository,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	return &service{
		alerteRepo:       alerteRepo,
		userRepo:         userRepo,
		commissariatRepo: commissariatRepo,
		config:           cfg,
		logger:           logger,
	}
}

// generateNumero génère un numéro unique pour l'alerte avec incrémentation
func (s *service) generateNumero(ctx context.Context, commissariatID string) (string, error) {
	// Récupérer le commissariat pour obtenir la ville
	commissariat, err := s.commissariatRepo.GetByID(ctx, commissariatID)
	if err != nil {
		return "", fmt.Errorf("commissariat not found")
	}

	year := time.Now().Year()
	
	// Extraire les 3 premières lettres de la ville en majuscules
	ville := strings.ToUpper(commissariat.Ville)
	villePrefix := ville
	if len(ville) > 3 {
		villePrefix = ville[:3]
	} else if len(ville) < 3 {
		// Si la ville a moins de 3 lettres, compléter avec des X
		villePrefix = ville + strings.Repeat("X", 3-len(ville))
	}
	
	// Chercher la dernière alerte du commissariat pour cette année
	startYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	
	filters := &repository.AlerteFilters{
		CommissariatID: &commissariatID,
		DateDebut:      &startYear,
		DateFin:        &endYear,
		Limit:          1,
		Offset:         0,
	}
	
	alertes, err := s.alerteRepo.List(ctx, filters)
	
	nextNumber := 1
	
	// Si des alertes existent, extraire le dernier numéro
	if err == nil && len(alertes) > 0 {
		// Trouver le numéro max en parcourant toutes les alertes de l'année
		// (car la liste peut ne pas être triée par numéro)
		maxNum := 0
		for _, alerte := range alertes {
			// Format: ALR-VILLE-COM-YYYY-NNNN
			parts := strings.Split(alerte.Numero, "-")
			if len(parts) == 5 {
				if num, err := strconv.Atoi(parts[4]); err == nil && num > maxNum {
					maxNum = num
				}
			}
		}
		if maxNum > 0 {
			nextNumber = maxNum + 1
		}
	}
	
	// Générer le numéro avec retry pour éviter les collisions
	maxRetries := 10
	for retry := 0; retry < maxRetries; retry++ {
		numero := fmt.Sprintf("ALR-%s-COM-%d-%04d", villePrefix, year, nextNumber+retry)
		
		// Vérifier l'unicité
		_, err := s.alerteRepo.GetByNumero(ctx, numero)
		if err != nil {
			// Numéro n'existe pas, c'est bon !
			s.logger.Info("Generated unique numero", zap.String("numero", numero))
			return numero, nil
		}
		
		s.logger.Warn("Numero collision detected, retrying", 
			zap.String("numero", numero), 
			zap.Int("retry", retry))
	}
	
	// En dernier recours, utiliser un timestamp
	timestamp := time.Now().UnixNano() % 1000000
	numero := fmt.Sprintf("ALR-%s-COM-%d-%06d", villePrefix, year, timestamp)
	s.logger.Warn("Using timestamp-based numero after max retries", zap.String("numero", numero))
	
	return numero, nil
}

// Create creates a new alert
func (s *service) Create(ctx context.Context, req *CreateAlerteRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Creating alerte", zap.String("titre", req.Titre))

	// Vérifier que l'agent existe
	agent, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	// Si CommissariatID n'est pas fourni, utiliser celui de l'agent
	commissariatID := req.CommissariatID
	if commissariatID == "" {
		if agent.Edges.Commissariat == nil {
			return nil, fmt.Errorf("agent must be assigned to a commissariat")
		}
		commissariatID = agent.Edges.Commissariat.ID.String()
		s.logger.Info("Using agent's commissariat", zap.String("commissariatId", commissariatID))
	}

	// Vérifier que le commissariat existe
	_, err = s.commissariatRepo.GetByID(ctx, commissariatID)
	if err != nil {
		return nil, fmt.Errorf("commissariat not found")
	}

	// Générer le numéro unique
	numero, err := s.generateNumero(ctx, commissariatID)
	if err != nil {
		return nil, err
	}

	// Préparer les données JSONB
	var personneConcernee, vehicule, suspect map[string]interface{}
	if req.PersonneConcernee != nil {
		personneConcernee = structToMap(req.PersonneConcernee)
	}
	if req.Vehicule != nil {
		vehicule = structToMap(req.Vehicule)
	}
	if req.Suspect != nil {
		suspect = structToMap(req.Suspect)
	}

	// Date de l'alerte
	dateAlerte := time.Now()
	if req.DateAlerte != nil {
		dateAlerte = *req.DateAlerte
	}

	// Niveau par défaut
	niveau := string(NiveauAlerteMoyen)
	if req.Niveau != nil {
		niveau = string(*req.Niveau)
	}

	input := &repository.CreateAlerteInput{
		ID:                    uuid.New().String(),
		Numero:                numero,
		Titre:                 req.Titre,
		Description:           req.Description,
		Contexte:              req.Contexte,
		Niveau:                niveau,
		TypeAlerte:            string(req.Type),
		Lieu:                  req.Lieu,
		Latitude:              req.Latitude,
		Longitude:             req.Longitude,
		PrecisionLocalisation: req.PrecisionLocalisation,
		Risques:               req.Risques,
		PersonneConcernee:     personneConcernee,
		Vehicule:              vehicule,
		Suspect:               suspect,
		CommissariatID:        commissariatID,
		AgentRecepteurID:      agentID,
		DateAlerte:            &dateAlerte,
		Observations:          req.Observations,
	}

	alerte, err := s.alerteRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create alerte: %w", err)
	}

	// Reload with edges
	alerte, err = s.alerteRepo.GetByID(ctx, alerte.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to reload alerte: %w", err)
	}

	// Ajouter un premier suivi
	suivis := []map[string]interface{}{
		{
			"date":    time.Now().Format("2006-01-02"),
			"heure":   time.Now().Format("15:04"),
			"agent":   fmt.Sprintf("%s %s", agent.Nom, agent.Prenom),
			"agentId": agentID,
			"action":  "Alerte créée",
			"statut":  "ACTIVE",
		},
	}
	updateInput := &repository.UpdateAlerteInput{
		Suivis: suivis,
	}
	alerte, _ = s.alerteRepo.Update(ctx, alerte.ID.String(), updateInput)

	return s.alerteToResponse(alerte), nil
}

// GetByID gets alert by ID
func (s *service) GetByID(ctx context.Context, id string) (*AlerteResponse, error) {
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.alerteToResponse(alerte), nil
}

// GetByNumero gets alert by numero
func (s *service) GetByNumero(ctx context.Context, numero string) (*AlerteResponse, error) {
	alerte, err := s.alerteRepo.GetByNumero(ctx, numero)
	if err != nil {
		return nil, err
	}
	return s.alerteToResponse(alerte), nil
}

// List lists alerts with filters and pagination
func (s *service) List(ctx context.Context, filters *FilterAlertesRequest, role, userID, commissariatID string) (*ListAlertesResponse, error) {
	// Set defaults
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	// Adapter les filtres
	var statut, typeAlerte, niveau *string
	if filters.Statut != nil {
		s := string(*filters.Statut)
		statut = &s
	}
	if filters.Type != nil {
		t := string(*filters.Type)
		typeAlerte = &t
	}
	if filters.Niveau != nil {
		n := string(*filters.Niveau)
		niveau = &n
	}

	repoFilters := &repository.AlerteFilters{
		Niveau:         niveau,
		Statut:         statut,
		TypeAlerte:     typeAlerte,
		CommissariatID: filters.CommissariatID,
		DateDebut:      filters.DateDebut,
		DateFin:        filters.DateFin,
		Search:         filters.Search,
		Limit:          filters.Limit,
		Offset:         (filters.Page - 1) * filters.Limit,
	}

	alertes, err := s.alerteRepo.List(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list alertes: %w", err)
	}

	total, err := s.alerteRepo.Count(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count alertes: %w", err)
	}

	responses := make([]AlerteResponse, len(alertes))
	for i, alerte := range alertes {
		responses[i] = *s.alerteToResponse(alerte)
	}

	return &ListAlertesResponse{
		Alertes: responses,
		Total:   int64(total),
		Page:    filters.Page,
		Limit:   filters.Limit,
	}, nil
}

// Update updates an alert
func (s *service) Update(ctx context.Context, id string, req *UpdateAlerteRequest) (*AlerteResponse, error) {
	s.logger.Info("Updating alerte", zap.String("id", id))

	// Préparer les données JSONB si présentes
	var personneConcernee, vehicule, suspect, intervention, evaluation, rapport, actions map[string]interface{}
	var temoins, documents, suivis []map[string]interface{}
	
	if req.PersonneConcernee != nil {
		personneConcernee = structToMap(req.PersonneConcernee)
	}
	if req.Vehicule != nil {
		vehicule = structToMap(req.Vehicule)
	}
	if req.Suspect != nil {
		suspect = structToMap(req.Suspect)
	}
	if req.Intervention != nil {
		intervention = structToMap(req.Intervention)
	}
	if req.Evaluation != nil {
		evaluation = structToMap(req.Evaluation)
	}
	if req.Rapport != nil {
		rapport = structToMap(req.Rapport)
	}
	if req.Actions != nil {
		actions = structToMap(req.Actions)
	}
	if req.Temoins != nil {
		for _, t := range req.Temoins {
			temoins = append(temoins, structToMap(t))
		}
	}
	if req.Documents != nil {
		for _, d := range req.Documents {
			documents = append(documents, structToMap(d))
		}
	}
	if req.Suivis != nil {
		for _, s := range req.Suivis {
			suivis = append(suivis, structToMap(s))
		}
	}

	// Adapter les champs optionnels
	var niveau, statut, typeAlerte *string
	if req.Niveau != nil {
		n := string(*req.Niveau)
		niveau = &n
	}
	if req.Statut != nil {
		st := string(*req.Statut)
		statut = &st
	}
	if req.Type != nil {
		t := string(*req.Type)
		typeAlerte = &t
	}

	input := &repository.UpdateAlerteInput{
		Titre:                    req.Titre,
		Description:              req.Description,
		Contexte:                 req.Contexte,
		Niveau:                   niveau,
		Statut:                   statut,
		TypeAlerte:               typeAlerte,
		Lieu:                     req.Lieu,
		Latitude:                 req.Latitude,
		Longitude:                req.Longitude,
		PrecisionLocalisation:    req.PrecisionLocalisation,
		Risques:                  req.Risques,
		PersonneConcernee:        personneConcernee,
		Vehicule:                 vehicule,
		Suspect:                  suspect,
		Intervention:             intervention,
		Evaluation:               evaluation,
		Actions:                  actions,
		Rapport:                  rapport,
		Temoins:                  temoins,
		Documents:                documents,
		Photos:                   req.Photos,
		Suivis:                   suivis,
		DateResolution:           req.DateResolution,
		DateCloture:              req.DateCloture,
		Observations:             req.Observations,
		Diffusee:                 req.Diffusee,
		DateDiffusion:            req.DateDiffusion,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update alerte: %w", err)
	}

	// Reload with edges
	alerte, err = s.alerteRepo.GetByID(ctx, alerte.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to reload alerte: %w", err)
	}

	return s.alerteToResponse(alerte), nil
}

// Delete deletes an alert
func (s *service) Delete(ctx context.Context, id string) error {
	s.logger.Info("Deleting alerte", zap.String("id", id))
	return s.alerteRepo.Delete(ctx, id)
}

// AddSuivi ajoute un suivi à une alerte
func (s *service) AddSuivi(ctx context.Context, id string, req *AddSuiviRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Adding suivi to alerte", zap.String("id", id))

	// Récupérer l'alerte
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Récupérer l'agent
	agent, err := s.userRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	// Récupérer les suivis existants ou initialiser
	suivis := alerte.Suivis
	if suivis == nil {
		suivis = []map[string]interface{}{}
	}

	// Ajouter le nouveau suivi
	nouveauSuivi := map[string]interface{}{
		"date":    time.Now().Format("2006-01-02"),
		"heure":   time.Now().Format("15:04"),
		"agent":   fmt.Sprintf("%s %s", agent.Nom, agent.Prenom),
		"agentId": agentID,
		"action":  req.Action,
		"statut":  req.Statut,
	}
	suivis = append(suivis, nouveauSuivi)

	// Mettre à jour l'alerte
	updateInput := &repository.UpdateAlerteInput{
		Suivis: suivis,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// Diffuser diffuse une alerte
func (s *service) Diffuser(ctx context.Context, id string, req *BroadcastAlerteRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Broadcasting alerte", zap.String("id", id))

	// Vérifier que l'alerte existe
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Vérifier si déjà diffusée
	if alerte.Diffusee {
		return nil, fmt.Errorf("alerte already broadcasted")
	}

	now := time.Now()
	diffusee := true

	// Préparer les destinataires
	destinataires := map[string]interface{}{
		"diffusionGenerale": req.DiffusionGenerale != nil && *req.DiffusionGenerale,
		"commissariatsIds":  req.CommissariatsIds,
		"agentsIds":         req.AgentsIds,
	}

	updateInput := &repository.UpdateAlerteInput{
		Diffusee:               &diffusee,
		DateDiffusion:          &now,
		DiffusionDestinataires: destinataires,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Alerte diffusée",
		Statut: string(alerte.Statut),
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// DiffusionInterne diffuse l'alerte aux agents du même commissariat
func (s *service) DiffusionInterne(ctx context.Context, id string, req *AssignAlerteRequest, commissariatID, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Diffusion interne de l'alerte", zap.String("id", id), zap.String("commissariatID", commissariatID))

	// Récupérer l'alerte
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Récupérer ou initialiser les assignations
	assignations := alerte.AssignationDestinataires
	if assignations == nil {
		assignations = make(map[string]interface{})
	}

	// Ajouter la diffusion interne pour ce commissariat
	assignations[commissariatID] = map[string]interface{}{
		"assigneeGenerale":     req.AssigneeGenerale != nil && *req.AssigneeGenerale,
		"agentsIds":            req.AgentsIds,
		"dateAssignation":      time.Now().Format(time.RFC3339),
		"agentAssignateurId":   agentID,
		"typeDiffusion":        "INTERNE", // Marquer comme diffusion interne
	}

	updateInput := &repository.UpdateAlerteInput{
		AssignationDestinataires: assignations,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Diffusion interne aux agents du commissariat",
		Statut: string(alerte.Statut),
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	
	s.logger.Info("Diffusion interne effectuée", 
		zap.String("alerteID", id),
		zap.Bool("generale", req.AssigneeGenerale != nil && *req.AssigneeGenerale),
		zap.Int("nbAgents", len(req.AgentsIds)))
	
	return s.alerteToResponse(alerte), nil
}

// Assigner assigne une alerte à des agents (commissariats destinataires)
func (s *service) Assigner(ctx context.Context, id string, req *AssignAlerteRequest, commissariatID, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Assigning alerte", zap.String("id", id))

	// Récupérer l'alerte
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Vérifier que l'alerte est diffusée
	if !alerte.Diffusee {
		return nil, fmt.Errorf("alerte must be broadcasted before assignment")
	}

	// Récupérer ou initialiser les assignations
	assignations := alerte.AssignationDestinataires
	if assignations == nil {
		assignations = make(map[string]interface{})
	}

	// Ajouter l'assignation pour ce commissariat
	assignations[commissariatID] = map[string]interface{}{
		"assigneeGenerale":     req.AssigneeGenerale != nil && *req.AssigneeGenerale,
		"agentsIds":            req.AgentsIds,
		"dateAssignation":      time.Now().Format(time.RFC3339),
		"agentAssignateurId":   agentID,
		"typeDiffusion":        "EXTERNE", // Marquer comme assignation externe
	}

	updateInput := &repository.UpdateAlerteInput{
		AssignationDestinataires: assignations,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Alerte assignée aux agents",
		Statut: string(alerte.Statut),
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// Resoudre marks an alert as resolved
func (s *service) Resoudre(ctx context.Context, id string, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Resolving alerte", zap.String("id", id))

	now := time.Now()
	statut := string(StatutAlerteResolue)
	
	updateInput := &repository.UpdateAlerteInput{
		Statut:         &statut,
		DateResolution: &now,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Alerte résolue",
		Statut: statut,
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// Archiver archives an alert
func (s *service) Archiver(ctx context.Context, id string, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Archiving alerte", zap.String("id", id))

	statut := string(StatutAlerteArchivee)
	
	updateInput := &repository.UpdateAlerteInput{
		Statut: &statut,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Alerte archivée",
		Statut: statut,
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// Cloturer clôture une alerte (résolu + archivé)
func (s *service) Cloturer(ctx context.Context, id string, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Closing alerte", zap.String("id", id))

	now := time.Now()
	statut := string(StatutAlerteArchivee)
	
	updateInput := &repository.UpdateAlerteInput{
		Statut:         &statut,
		DateResolution: &now,
		DateCloture:    &now,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Alerte clôturée",
		Statut: statut,
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// DeployIntervention déploie une intervention
func (s *service) DeployIntervention(ctx context.Context, id string, req *DeployInterventionRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Deploying intervention", zap.String("id", id))

	intervention := map[string]interface{}{
		"statut":      string(StatutInterventionEnCours),
		"equipe":      req.Equipe,
		"moyens":      req.Moyens,
		"heureDepart": time.Now().Format("15:04"),
	}

	updateInput := &repository.UpdateAlerteInput{
		Intervention: intervention,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Intervention déployée",
		Statut: string(alerte.Statut),
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// UpdateIntervention met à jour une intervention
func (s *service) UpdateIntervention(ctx context.Context, id string, req *UpdateInterventionRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Updating intervention", zap.String("id", id))

	// Récupérer l'alerte pour obtenir l'intervention existante
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Récupérer l'intervention existante
	intervention := alerte.Intervention
	if intervention == nil {
		intervention = make(map[string]interface{})
	}

	// Mettre à jour les champs
	if req.Statut != nil {
		intervention["statut"] = string(*req.Statut)
	}
	if req.HeureDepart != nil {
		intervention["heureDepart"] = *req.HeureDepart
	}
	if req.HeureArrivee != nil {
		intervention["heureArrivee"] = *req.HeureArrivee
	}
	if req.HeureFin != nil {
		intervention["heureFin"] = *req.HeureFin
	}
	if req.Moyens != nil {
		intervention["moyens"] = req.Moyens
	}
	if req.TempsReponse != nil {
		intervention["tempsReponse"] = *req.TempsReponse
	}

	updateInput := &repository.UpdateAlerteInput{
		Intervention: intervention,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// AddEvaluation ajoute une évaluation
func (s *service) AddEvaluation(ctx context.Context, id string, req *AddEvaluationRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Adding evaluation", zap.String("id", id))

	evaluation := structToMap(req)

	updateInput := &repository.UpdateAlerteInput{
		Evaluation: evaluation,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Évaluation ajoutée",
		Statut: string(alerte.Statut),
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// AddRapport ajoute un rapport final
func (s *service) AddRapport(ctx context.Context, id string, req *AddRapportRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Adding rapport", zap.String("id", id))

	rapport := structToMap(req)

	updateInput := &repository.UpdateAlerteInput{
		Rapport: rapport,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Ajouter un suivi
	s.AddSuivi(ctx, id, &AddSuiviRequest{
		Action: "Rapport final ajouté",
		Statut: string(alerte.Statut),
	}, agentID)

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// AddTemoin ajoute un témoin
func (s *service) AddTemoin(ctx context.Context, id string, req *AddTemoinRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Adding temoin", zap.String("id", id))

	// Récupérer l'alerte
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Récupérer les témoins existants
	temoins := alerte.Temoins
	if temoins == nil {
		temoins = []map[string]interface{}{}
	}

	// Ajouter le nouveau témoin
	temoins = append(temoins, structToMap(req))

	updateInput := &repository.UpdateAlerteInput{
		Temoins: temoins,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// AddDocument ajoute un document
func (s *service) AddDocument(ctx context.Context, id string, req *AddDocumentRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Adding document", zap.String("id", id))

	// Récupérer l'alerte
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Récupérer les documents existants
	documents := alerte.Documents
	if documents == nil {
		documents = []map[string]interface{}{}
	}

	// Ajouter le nouveau document
	documents = append(documents, structToMap(req))

	updateInput := &repository.UpdateAlerteInput{
		Documents: documents,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// AddPhotos ajoute des photos
func (s *service) AddPhotos(ctx context.Context, id string, photos []string, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Adding photos", zap.String("id", id))

	// Récupérer l'alerte
	alerte, err := s.alerteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Récupérer les photos existantes
	existingPhotos := alerte.Photos
	if existingPhotos == nil {
		existingPhotos = []string{}
	}

	// Ajouter les nouvelles photos
	existingPhotos = append(existingPhotos, photos...)

	updateInput := &repository.UpdateAlerteInput{
		Photos: existingPhotos,
	}

	alerte, err = s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// UpdateActions met à jour les actions
func (s *service) UpdateActions(ctx context.Context, id string, req *UpdateActionsRequest, agentID string) (*AlerteResponse, error) {
	s.logger.Info("Updating actions", zap.String("id", id))

	// Construire actions en s'assurant que tous les champs sont présents (même vides)
	actions := map[string]interface{}{
		"immediate":  req.Immediate,
		"preventive": req.Preventive,
		"suivi":      req.Suivi,
	}
	
	// S'assurer que nil devient []
	if actions["immediate"] == nil {
		actions["immediate"] = []string{}
	}
	if actions["preventive"] == nil {
		actions["preventive"] = []string{}
	}
	if actions["suivi"] == nil {
		actions["suivi"] = []string{}
	}

	updateInput := &repository.UpdateAlerteInput{
		Actions: actions,
	}

	alerte, err := s.alerteRepo.Update(ctx, id, updateInput)
	if err != nil {
		return nil, err
	}

	// Reload
	alerte, _ = s.alerteRepo.GetByID(ctx, id)
	return s.alerteToResponse(alerte), nil
}

// GetActives gets active alerts
func (s *service) GetActives(ctx context.Context, commissariatID *string) ([]*AlerteResponse, error) {
	var alertes []*ent.AlerteSecuritaire
	var err error

	if commissariatID != nil {
		// Filtrer par commissariat si spécifié
		filters := &repository.AlerteFilters{
			CommissariatID: commissariatID,
		}
		allAlertes, err := s.alerteRepo.List(ctx, filters)
	if err != nil {
			return nil, err
		}
		// Filtrer les actives
		for _, a := range allAlertes {
			if a.Statut == alertesecuritaire.StatutACTIVE {
				alertes = append(alertes, a)
			}
		}
	} else {
		alertes, err = s.alerteRepo.GetActives(ctx)
	if err != nil {
			return nil, err
		}
	}

	responses := make([]*AlerteResponse, len(alertes))
	for i, alerte := range alertes {
		responses[i] = s.alerteToResponse(alerte)
	}

	return responses, nil
}

// GetStatistiques gets alert statistics
func (s *service) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesAlertesResponse, error) {
	var debut, fin *time.Time
	
	if dateDebut != nil {
		t, err := parseDateTime(*dateDebut)
		if err == nil {
			debut = &t
		}
	}
	if dateFin != nil {
		t, err := parseDateTime(*dateFin)
		if err == nil {
			fin = &t
		}
	}

	// Récupérer les stats (le repository calcule l'évolution directement)
	stats, err := s.alerteRepo.GetStatistiques(ctx, commissariatID, debut, fin, periode)
	if err != nil {
		return nil, err
	}

	// Convertir en response
	response := &StatistiquesAlertesResponse{
		Total:      int64(stats["total"].(int)),
		Actives:    int64(stats["actives"].(int)),
		Resolues:   int64(stats["resolues"].(int)),
		Archivees:  int64(stats["archivees"].(int)),
		ParNiveau:  make(map[NiveauAlerte]int64),
	}

	if parNiveau, ok := stats["parNiveau"].(map[string]int); ok {
		for niveau, count := range parNiveau {
			response.ParNiveau[NiveauAlerte(niveau)] = int64(count)
		}
	}

	if tauxResolution, ok := stats["tauxResolution"].(float64); ok {
		response.TauxResolution = &tauxResolution
	}

	// Récupérer l'évolution calculée par le repository
	if evolutionAlertes, ok := stats["evolutionAlertes"].(string); ok {
		response.EvolutionAlertes = &evolutionAlertes
	}
	if evolutionResolution, ok := stats["evolutionResolution"].(string); ok {
		response.EvolutionResolution = &evolutionResolution
	}

	return response, nil
}

// GenererDescription génère une description avec IA OpenAI
func (s *service) GenererDescription(ctx context.Context, req *GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
	s.logger.Info("Generating description with AI", zap.String("type", string(req.Type)))

	// Vérifier la clé OpenAI
	var openaiKey string
	if s.config.OpenAI != nil {
		openaiKey = s.config.OpenAI.APIKey
	}
	
	if openaiKey == "" {
		s.logger.Error("OpenAI key not configured")
		return nil, fmt.Errorf("la clé OpenAI n'est pas configurée. Veuillez ajouter OPENAI_KEY dans le fichier .env ou config.yaml")
	}

	// Construire le prompt
	prompt := s.construirePrompt(req)

	// Préparer la requête OpenAI
	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un assistant spécialisé en rédaction de rapports d'incidents pour la Police Nationale de Côte d'Ivoire. Tu dois générer des descriptions professionnelles, claires et détaillées.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature":     0.7,
		"max_tokens":      1000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		s.logger.Error("Failed to marshal OpenAI request", zap.Error(err))
		return nil, fmt.Errorf("erreur lors de la préparation de la requête OpenAI: %v", err)
	}

	// Appeler l'API OpenAI
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error("Failed to create OpenAI request", zap.Error(err))
		return nil, fmt.Errorf("erreur lors de la création de la requête OpenAI: %v", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		s.logger.Error("Failed to call OpenAI API", zap.Error(err))
		return nil, fmt.Errorf("erreur lors de l'appel à l'API OpenAI: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read OpenAI response", zap.Error(err))
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse OpenAI: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("OpenAI API error", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		
		// Parser l'erreur pour un message plus clair
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorMap, ok := errorResp["error"].(map[string]interface{}); ok {
				code := errorMap["code"]
				message := errorMap["message"]
				
				if code == "insufficient_quota" {
					return nil, fmt.Errorf("quota OpenAI dépassé. Veuillez recharger votre compte sur https://platform.openai.com/account/billing")
				}
				if code == "invalid_api_key" {
					return nil, fmt.Errorf("clé API OpenAI invalide. Veuillez vérifier votre clé sur https://platform.openai.com/api-keys")
				}
				return nil, fmt.Errorf("erreur OpenAI: %v", message)
			}
		}
		
		return nil, fmt.Errorf("erreur OpenAI: code %d", resp.StatusCode)
	}

	// Parser la réponse
	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openaiResp); err != nil {
		s.logger.Error("Failed to parse OpenAI response", zap.Error(err))
		return nil, fmt.Errorf("erreur lors du parsing de la réponse OpenAI: %v", err)
	}

	if len(openaiResp.Choices) == 0 {
		s.logger.Error("No choices in OpenAI response")
		return nil, fmt.Errorf("aucune réponse générée par OpenAI")
	}

	// Parser le contenu JSON
	content := openaiResp.Choices[0].Message.Content
	var resultat struct {
		Description string `json:"description"`
		Contexte    string `json:"contexte"`
	}

	if err := json.Unmarshal([]byte(content), &resultat); err != nil {
		s.logger.Error("Failed to parse AI generated content", zap.Error(err), zap.String("content", content))
		return nil, fmt.Errorf("erreur lors du parsing du contenu généré: %v", err)
	}

	s.logger.Info("AI description generated successfully")
	
	return &GenerateDescriptionResponse{
		Success: true,
		Data: GenerateDescriptionData{
			Description: resultat.Description,
			Contexte:    resultat.Contexte,
		},
		Mode:    "openai",
		Message: "Contenu généré avec succès par l'intelligence artificielle OpenAI",
	}, nil
}

// GenererRapport génère un rapport complet avec IA OpenAI
func (s *service) GenererRapport(ctx context.Context, alerteID string) (*GenerateRapportResponse, error) {
	s.logger.Info("Generating rapport with AI", zap.String("alerteId", alerteID))

	// Vérifier la clé OpenAI
	var openaiKey string
	if s.config.OpenAI != nil {
		openaiKey = s.config.OpenAI.APIKey
	}
	
	if openaiKey == "" {
		s.logger.Error("OpenAI key not configured")
		return nil, fmt.Errorf("la clé OpenAI n'est pas configurée")
	}

	// Récupérer l'alerte pour construire le prompt
	alerte, err := s.alerteRepo.GetByID(ctx, alerteID)
	if err != nil {
		return nil, fmt.Errorf("alerte not found")
	}

	// Construire le prompt pour le rapport
	prompt := s.construirePromptRapport(alerte)

	// Préparer la requête OpenAI
	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un assistant spécialisé en rédaction de rapports d'intervention de police. Tu dois générer des rapports professionnels, détaillés et objectifs pour la Police Nationale de Côte d'Ivoire.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature":     0.7,
		"max_tokens":      2000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la préparation de la requête: %v", err)
	}

	// Appeler l'API OpenAI
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'appel à l'API OpenAI: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("OpenAI API error", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("erreur OpenAI: code %d", resp.StatusCode)
	}

	// Parser la réponse
	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing de la réponse: %v", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("aucune réponse générée")
	}

	// Parser le contenu JSON
	content := openaiResp.Choices[0].Message.Content
	var resultat struct {
		Resume          string   `json:"resume"`
		Conclusions     []string `json:"conclusions"`
		Recommandations []string `json:"recommandations"`
	}

	if err := json.Unmarshal([]byte(content), &resultat); err != nil {
		s.logger.Error("Failed to parse AI generated rapport", zap.Error(err))
		return nil, fmt.Errorf("erreur lors du parsing du rapport généré: %v", err)
	}

	s.logger.Info("AI rapport generated successfully")
	
	return &GenerateRapportResponse{
		Success:         true,
		Resume:          resultat.Resume,
		Conclusions:     resultat.Conclusions,
		Recommandations: resultat.Recommandations,
		Mode:            "openai",
		Message:         "Rapport généré avec succès par l'intelligence artificielle OpenAI",
	}, nil
}

// construirePrompt construit le prompt pour l'IA
func (s *service) construirePrompt(req *GenerateDescriptionRequest) string {
	typeLabels := map[TypeAlerte]string{
		TypeAlerteVehiculeVole:          "Véhicule volé",
		TypeAlerteSuspectRecherche:      "Suspect recherché",
		TypeAlerteUrgenceSecurite:       "Urgence sécurité",
		TypeAlerteAlerteGenerale:        "Alerte générale",
		TypeAlerteMaintenanceSysteme:    "Maintenance système",
		TypeAlerteAccident:              "Accident de circulation",
		TypeAlerteIncendie:              "Incendie",
		TypeAlerteAggression:            "Aggression",
		TypeAlerteAmber:                 "Alerte Amber",
		TypeAlerteAutre:                 "Autre",
	}

	typeLabel := typeLabels[req.Type]
	if typeLabel == "" {
		typeLabel = string(req.Type)
	}

	prompt := fmt.Sprintf(`
Génère une description détaillée et un contexte pour cette alerte de la Police Nationale :

Type d'alerte: %s
Titre: %s`, typeLabel, req.Titre)

	if req.Lieu != nil {
		prompt += fmt.Sprintf("\nLieu: %s", *req.Lieu)
	}

	if len(req.Risques) > 0 {
		prompt += fmt.Sprintf("\nRisques identifiés: %s", strings.Join(req.Risques, ", "))
	}

	if req.InformationsComplementaires != nil && len(req.InformationsComplementaires) > 0 {
		infos, _ := json.Marshal(req.InformationsComplementaires)
		prompt += fmt.Sprintf("\nInformations complémentaires: %s", string(infos))
	}

	prompt += `

Retourne UNIQUEMENT un objet JSON avec cette structure exacte (sans texte avant ou après) :
{
  "description": "Une description détaillée professionnelle de 3-4 phrases expliquant la situation, les faits, les circonstances et l'urgence",
  "contexte": "Le contexte de 2-3 phrases expliquant les antécédents, l'environnement, les facteurs aggravants ou les éléments importants à connaître"
}`

	return prompt
}

// construirePromptRapport construit le prompt pour générer un rapport complet
func (s *service) construirePromptRapport(alerte *ent.AlerteSecuritaire) string {
	typeLabels := map[string]string{
		"VEHICULE_VOLE":       "Véhicule volé",
		"SUSPECT_RECHERCHE":   "Suspect recherché",
		"URGENCE_SECURITE":    "Urgence sécurité",
		"ALERTE_GENERALE":     "Alerte générale",
		"MAINTENANCE_SYSTEME": "Maintenance système",
		"ACCIDENT":            "Accident de circulation",
		"INCENDIE":            "Incendie",
		"AGGRESSION":          "Aggression",
		"AUTRE":               "Autre",
	}

	typeLabel := typeLabels[alerte.TypeAlerte]
	if typeLabel == "" {
		typeLabel = alerte.TypeAlerte
	}

	// === INFORMATIONS DE BASE ===
	prompt := fmt.Sprintf(`
Génère un rapport d'intervention complet et professionnel pour cette alerte de police :

═══════════════════════════════════════════════════
INFORMATIONS DE L'ALERTE :
═══════════════════════════════════════════════════
- Numéro: %s
- Type: %s
- Titre: %s
- Description: %s
- Niveau: %s
- Date alerte: %s
`, alerte.Numero, typeLabel, alerte.Titre, alerte.Description, alerte.Niveau, alerte.DateAlerte.Format("02/01/2006 15:04"))

	if alerte.Contexte != nil && *alerte.Contexte != "" {
		prompt += fmt.Sprintf("- Contexte: %s\n", *alerte.Contexte)
	}
	
	if alerte.Lieu != nil {
		prompt += fmt.Sprintf("- Lieu: %s\n", *alerte.Lieu)
	}

	if len(alerte.Risques) > 0 {
		prompt += fmt.Sprintf("- Risques identifiés: %s\n", strings.Join(alerte.Risques, ", "))
	}

	// === PERSONNE CONCERNÉE ===
	if len(alerte.PersonneConcernee) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "PERSONNE CONCERNÉE :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		if nom, ok := alerte.PersonneConcernee["nom"].(string); ok {
			prompt += fmt.Sprintf("- Nom: %s\n", nom)
		}
		if tel, ok := alerte.PersonneConcernee["telephone"].(string); ok {
			prompt += fmt.Sprintf("- Téléphone: %s\n", tel)
		}
		if desc, ok := alerte.PersonneConcernee["description"].(string); ok {
			prompt += fmt.Sprintf("- Description: %s\n", desc)
		}
	}

	// === VÉHICULE ===
	if len(alerte.Vehicule) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "VÉHICULE CONCERNÉ :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		if immat, ok := alerte.Vehicule["immatriculation"].(string); ok {
			prompt += fmt.Sprintf("- Immatriculation: %s\n", immat)
		}
		if marque, ok := alerte.Vehicule["marque"].(string); ok {
			prompt += fmt.Sprintf("- Marque: %s\n", marque)
		}
		if modele, ok := alerte.Vehicule["modele"].(string); ok {
			prompt += fmt.Sprintf("- Modèle: %s\n", modele)
		}
		if couleur, ok := alerte.Vehicule["couleur"].(string); ok {
			prompt += fmt.Sprintf("- Couleur: %s\n", couleur)
		}
	}

	// === SUSPECT ===
	if len(alerte.Suspect) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "SUSPECT RECHERCHÉ :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		if nom, ok := alerte.Suspect["nom"].(string); ok {
			prompt += fmt.Sprintf("- Nom: %s\n", nom)
		}
		if desc, ok := alerte.Suspect["description"].(string); ok {
			prompt += fmt.Sprintf("- Description: %s\n", desc)
		}
		if motif, ok := alerte.Suspect["motif"].(string); ok {
			prompt += fmt.Sprintf("- Motif: %s\n", motif)
		}
	}

	// === INTERVENTION ===
	if len(alerte.Intervention) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "INTERVENTION EFFECTUÉE :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		if statut, ok := alerte.Intervention["statut"].(string); ok {
			prompt += fmt.Sprintf("- Statut: %s\n", statut)
		}
		if equipe, ok := alerte.Intervention["equipe"].([]interface{}); ok && len(equipe) > 0 {
			prompt += fmt.Sprintf("- Nombre d'agents déployés: %d\n", len(equipe))
			for i, membre := range equipe {
				if membreMap, ok := membre.(map[string]interface{}); ok {
					if nom, ok := membreMap["nom"].(string); ok {
						prompt += fmt.Sprintf("  %d. %s", i+1, nom)
						if matricule, ok := membreMap["matricule"].(string); ok {
							prompt += fmt.Sprintf(" (%s)", matricule)
						}
						prompt += "\n"
					}
				}
			}
		}
		if moyens, ok := alerte.Intervention["moyens"].([]interface{}); ok {
			moyensStr := make([]string, 0, len(moyens))
			for _, m := range moyens {
				if str, ok := m.(string); ok {
					moyensStr = append(moyensStr, str)
				}
			}
			if len(moyensStr) > 0 {
				prompt += fmt.Sprintf("- Moyens déployés: %s\n", strings.Join(moyensStr, ", "))
			}
		}
		if heureDepart, ok := alerte.Intervention["heureDepart"].(string); ok {
			prompt += fmt.Sprintf("- Heure de départ: %s\n", heureDepart)
		}
		if heureArrivee, ok := alerte.Intervention["heureArrivee"].(string); ok {
			prompt += fmt.Sprintf("- Heure d'arrivée: %s\n", heureArrivee)
		}
		if heureFin, ok := alerte.Intervention["heureFin"].(string); ok {
			prompt += fmt.Sprintf("- Heure de fin: %s\n", heureFin)
		}
		if tempsReponse, ok := alerte.Intervention["tempsReponse"].(string); ok {
			prompt += fmt.Sprintf("- Temps de réponse: %s\n", tempsReponse)
		}
	}

	// === ÉVALUATION ===
	if len(alerte.Evaluation) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "ÉVALUATION SUR PLACE :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		if situation, ok := alerte.Evaluation["situationReelle"].(string); ok {
			prompt += fmt.Sprintf("- Situation réelle: %s\n", situation)
		}
		if victimes, ok := alerte.Evaluation["victimes"].(float64); ok {
			prompt += fmt.Sprintf("- Victimes: %.0f\n", victimes)
		}
		if degats, ok := alerte.Evaluation["degats"].(string); ok {
			prompt += fmt.Sprintf("- Dégâts: %s\n", degats)
		}
		if mesures, ok := alerte.Evaluation["mesuresPrises"].([]interface{}); ok && len(mesures) > 0 {
			prompt += "- Mesures prises sur place:\n"
			for i, m := range mesures {
				if str, ok := m.(string); ok {
					prompt += fmt.Sprintf("  %d. %s\n", i+1, str)
				}
			}
		}
		if renforts, ok := alerte.Evaluation["renforts"].(bool); ok && renforts {
			if details, ok := alerte.Evaluation["renfortsDetails"].(string); ok {
				prompt += fmt.Sprintf("- Renforts demandés: %s\n", details)
			}
		}
	}

	// === ACTIONS ===
	if len(alerte.Actions) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "ACTIONS MENÉES :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		
		if immediate, ok := alerte.Actions["immediate"].([]interface{}); ok && len(immediate) > 0 {
			prompt += "Actions immédiates:\n"
			for i, a := range immediate {
				if str, ok := a.(string); ok {
					prompt += fmt.Sprintf("  %d. %s\n", i+1, str)
				}
			}
		}
		
		if preventive, ok := alerte.Actions["preventive"].([]interface{}); ok && len(preventive) > 0 {
			prompt += "Actions préventives:\n"
			for i, a := range preventive {
				if str, ok := a.(string); ok {
					prompt += fmt.Sprintf("  %d. %s\n", i+1, str)
				}
			}
		}
		
		if suivi, ok := alerte.Actions["suivi"].([]interface{}); ok && len(suivi) > 0 {
			prompt += "Actions de suivi:\n"
			for i, a := range suivi {
				if str, ok := a.(string); ok {
					prompt += fmt.Sprintf("  %d. %s\n", i+1, str)
				}
			}
		}
	}

	// === TÉMOINS ===
	if len(alerte.Temoins) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += fmt.Sprintf("TÉMOIGNAGES (%d) :\n", len(alerte.Temoins))
		prompt += "═══════════════════════════════════════════════════\n"
		for i, temoinMap := range alerte.Temoins {
			prompt += fmt.Sprintf("Témoin %d:\n", i+1)
			if nom, ok := temoinMap["nom"].(string); ok {
				prompt += fmt.Sprintf("  - Nom: %s\n", nom)
			}
			if declaration, ok := temoinMap["declaration"].(string); ok {
				prompt += fmt.Sprintf("  - Déclaration: %s\n", declaration)
			}
		}
	}

	// === DOCUMENTS ===
	if len(alerte.Documents) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += fmt.Sprintf("DOCUMENTS ATTACHÉS (%d) :\n", len(alerte.Documents))
		prompt += "═══════════════════════════════════════════════════\n"
		for i, docMap := range alerte.Documents {
			if typeDoc, ok := docMap["type"].(string); ok {
				prompt += fmt.Sprintf("%d. %s", i+1, typeDoc)
				if numero, ok := docMap["numero"].(string); ok {
					prompt += fmt.Sprintf(" (Nº %s)", numero)
				}
				if desc, ok := docMap["description"].(string); ok && desc != "" {
					prompt += fmt.Sprintf(" - %s", desc)
				}
				prompt += "\n"
			}
		}
	}

	// === SUIVIS ===
	if len(alerte.Suivis) > 0 {
		prompt += "\n═══════════════════════════════════════════════════\n"
		prompt += "HISTORIQUE DES SUIVIS :\n"
		prompt += "═══════════════════════════════════════════════════\n"
		for i, suiviMap := range alerte.Suivis {
			date := ""
			heure := ""
			action := ""
			agent := ""
			
			if d, ok := suiviMap["date"].(string); ok {
				date = d
			}
			if h, ok := suiviMap["heure"].(string); ok {
				heure = h
			}
			if a, ok := suiviMap["action"].(string); ok {
				action = a
			}
			if ag, ok := suiviMap["agent"].(string); ok {
				agent = ag
			}
			
			prompt += fmt.Sprintf("%d. [%s %s] %s - Par: %s\n", i+1, date, heure, action, agent)
		}
	}

	prompt += `

Génère un rapport professionnel au format JSON avec cette structure EXACTE :
{
  "resume": "Un résumé complet de 4-6 phrases décrivant chronologiquement l'alerte, l'intervention, les actions menées et le résultat",
  "conclusions": [
    "Première conclusion basée sur les faits observés",
    "Deuxième conclusion sur l'efficacité de l'intervention",
    "Troisième conclusion sur les leçons apprises"
  ],
  "recommandations": [
    "Première recommandation pour prévenir des incidents similaires",
    "Deuxième recommandation pour améliorer les procédures",
    "Troisième recommandation pour le suivi"
  ]
}

Sois professionnel, objectif et factuel. Utilise un ton officiel adapté à un rapport de police.`

	return prompt
}

// Helper functions

func (s *service) alerteToResponse(alerte *ent.AlerteSecuritaire) *AlerteResponse {
	resp := &AlerteResponse{
		ID:          alerte.ID.String(),
		Numero:      alerte.Numero,
		Titre:       alerte.Titre,
		Description: alerte.Description,
		Niveau:      NiveauAlerte(alerte.Niveau),
		Statut:      StatutAlerte(alerte.Statut),
		Type:        TypeAlerte(alerte.TypeAlerte),
		DateAlerte:  alerte.DateAlerte,
		Diffusee:    alerte.Diffusee,
		CreatedAt:   alerte.CreatedAt,
		UpdatedAt:   alerte.UpdatedAt,
	}

	// Champs optionnels simples
	if alerte.Contexte != nil {
		resp.Contexte = alerte.Contexte
	}
	if alerte.Lieu != nil {
		resp.Lieu = alerte.Lieu
	}
	if alerte.Latitude != 0 {
		resp.Latitude = &alerte.Latitude
	}
	if alerte.Longitude != 0 {
		resp.Longitude = &alerte.Longitude
	}
	if alerte.PrecisionLocalisation != nil {
		resp.PrecisionLocalisation = alerte.PrecisionLocalisation
	}
	if alerte.Observations != nil {
		resp.Observations = alerte.Observations
	}
	if alerte.DateResolution != nil {
		resp.DateResolution = alerte.DateResolution
	}
	if alerte.DateCloture != nil {
		resp.DateCloture = alerte.DateCloture
	}
	if alerte.DateDiffusion != nil {
		resp.DateDiffusion = alerte.DateDiffusion
	}

	// Les champs JSONB sont déjà désérialisés par Ent
	if alerte.Risques != nil {
		resp.Risques = alerte.Risques
	}

	// Pour les autres champs JSONB, on doit convertir map[string]interface{} vers nos types
	if alerte.PersonneConcernee != nil {
		var pc PersonneConcernee
		if data, err := json.Marshal(alerte.PersonneConcernee); err == nil {
			json.Unmarshal(data, &pc)
			resp.PersonneConcernee = &pc
		}
	}

	if alerte.Vehicule != nil {
		var v VehiculeAlerte
		if data, err := json.Marshal(alerte.Vehicule); err == nil {
			json.Unmarshal(data, &v)
			resp.Vehicule = &v
		}
	}

	if alerte.Suspect != nil {
		var susp Suspect
		if data, err := json.Marshal(alerte.Suspect); err == nil {
			json.Unmarshal(data, &susp)
			resp.Suspect = &susp
		}
	}

	if alerte.Intervention != nil {
		var interv Intervention
		if data, err := json.Marshal(alerte.Intervention); err == nil {
			json.Unmarshal(data, &interv)
			resp.Intervention = &interv
		}
	}

	if alerte.Evaluation != nil {
		var eval Evaluation
		if data, err := json.Marshal(alerte.Evaluation); err == nil {
			json.Unmarshal(data, &eval)
			resp.Evaluation = &eval
		}
	}

	// Initialiser actions avec valeurs par défaut
	resp.Actions = Actions{
		Immediate:  []string{},
		Preventive: []string{},
		Suivi:      []string{},
	}
	
	if alerte.Actions != nil {
		var acts Actions
		if data, err := json.Marshal(alerte.Actions); err == nil {
			json.Unmarshal(data, &acts)
			
			// Garder les valeurs si elles existent, sinon utiliser []
			if acts.Immediate != nil {
				resp.Actions.Immediate = acts.Immediate
			}
			if acts.Preventive != nil {
				resp.Actions.Preventive = acts.Preventive
			}
			if acts.Suivi != nil {
				resp.Actions.Suivi = acts.Suivi
			}
		}
	}

	if alerte.Rapport != nil {
		var rapp Rapport
		if data, err := json.Marshal(alerte.Rapport); err == nil {
			json.Unmarshal(data, &rapp)
			resp.Rapport = &rapp
		}
	}

	if alerte.Temoins != nil {
		var temoins []Temoin
		if data, err := json.Marshal(alerte.Temoins); err == nil {
			json.Unmarshal(data, &temoins)
			resp.Temoins = temoins
		}
	}

	if alerte.Documents != nil {
		var docs []Document
		if data, err := json.Marshal(alerte.Documents); err == nil {
			json.Unmarshal(data, &docs)
			resp.Documents = docs
		}
	}

	if alerte.Photos != nil {
		resp.Photos = alerte.Photos
	}

	if alerte.Suivis != nil {
		var suivis []Suivi
		if data, err := json.Marshal(alerte.Suivis); err == nil {
			json.Unmarshal(data, &suivis)
			resp.Suivis = suivis
		}
	}

	if alerte.DiffusionDestinataires != nil {
		var diff DiffusionDestinataires
		if data, err := json.Marshal(alerte.DiffusionDestinataires); err == nil {
			json.Unmarshal(data, &diff)
			resp.DiffusionDestinataires = &diff
		}
	}

	if alerte.AssignationDestinataires != nil {
		var assign map[string]*AssignationCommissariat
		if data, err := json.Marshal(alerte.AssignationDestinataires); err == nil {
			json.Unmarshal(data, &assign)
			resp.AssignationDestinataires = assign
		}
	}

	// Load commissariat if available
	if comm := alerte.Edges.Commissariat; comm != nil {
		resp.CommissariatID = comm.ID.String()
		resp.Commissariat = &CommissariatSummary{
			ID:    comm.ID.String(),
			Nom:   comm.Nom,
			Code:  comm.Code,
			Ville: comm.Ville,
		}
	}

	// Load agent if available
	if agent := alerte.Edges.Agent; agent != nil {
		resp.AgentRecepteurID = agent.ID.String()
	}

	return resp
}

// GetDashboard gets dashboard data for alerts
func (s *service) GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardResponse, error) {
	var debut, fin *time.Time
	
	if dateDebut != nil {
		t, err := parseDateTime(*dateDebut)
		if err == nil {
			debut = &t
		}
	}
	if dateFin != nil {
		t, err := parseDateTime(*dateFin)
		if err == nil {
			fin = &t
		}
	}

	// Récupérer les statistiques de base
	stats, err := s.alerteRepo.GetStatistiques(ctx, commissariatID, debut, fin, periode)
	if err != nil {
		return nil, err
	}

	// Convertir en réponse dashboard
	total := int(stats["total"].(int))
	actives := int(stats["actives"].(int))
	resolues := int(stats["resolues"].(int))
	evolutionAlertes := stats["evolutionAlertes"].(string)
	evolutionResolution := stats["evolutionResolution"].(string)

	// Calculer les en cours (actives - résolues si logique différente dans votre contexte)
	enCours := actives

	// Temps de réponse moyen
	tempsReponseMoyen := "0 min"
	evolutionTempsReponse := "+0 min"
	if tempsRepMoy, ok := stats["tempsReponseMoyen"].(float64); ok && tempsRepMoy > 0 {
		tempsReponseMoyen = fmt.Sprintf("%.0f min", tempsRepMoy)
		// On peut calculer l'évolution du temps de réponse ici si on a les données
		evolutionTempsReponse = "-5 min" // Par défaut, on peut ajuster
	}

	dashboardStats := DashboardStats{
		TotalAlertes: DashboardStatsValue{
			Total:     total,
			Evolution: evolutionAlertes,
		},
		Resolues: DashboardStatsValue{
			Total:     resolues,
			Evolution: evolutionResolution,
		},
		EnCours: DashboardStatsValue{
			Total:     enCours,
			Evolution: evolutionAlertes, // On utilise la même évolution pour enCours
		},
		TempsReponse: DashboardTempsReponse{
			Moyen:     tempsReponseMoyen,
			Evolution: evolutionTempsReponse,
		},
	}

	// Tableau de statistiques par type
	statsTable := []DashboardStatsTableItem{}
	
	// Définir l'ordre et les labels de TOUS les types d'alertes
	typesOrdered := []struct {
		key   string
		label string
	}{
		{string(TypeAlerteUrgenceSecurite), "Urgence sécurité"},
		{string(TypeAlerteAccident), "Accident"},
		{string(TypeAlerteAggression), "Aggression"},
		{string(TypeAlerteIncendie), "Incendie"},
		{string(TypeAlerteVehiculeVole), "Véhicules volés"},
		{string(TypeAlerteSuspectRecherche), "Avis de recherche"},
		{string(TypeAlerteAlerteGenerale), "Alerte générale"},
		{string(TypeAlerteAmber), "AMBER Alert"},
		{string(TypeAlerteMaintenanceSysteme), "Maintenance système"},
		{string(TypeAlerteAutre), "Autre"},
	}

	// Récupérer la liste des alertes pour calculer les statistiques par type
	filters := &repository.AlerteFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	alertes, err := s.alerteRepo.List(ctx, filters)
	
	// Calculer les statistiques pour chaque type dans l'ordre défini
	for _, typeInfo := range typesOrdered {
		nombre := 0
		resoluesType := 0

		// Compter les alertes de ce type
		if err == nil && alertes != nil {
			for _, alerte := range alertes {
				if alerte.TypeAlerte == typeInfo.key {
					nombre++
					if alerte.Statut == "RESOLUE" {
						resoluesType++
					}
				}
			}
		}

		// N'ajouter au tableau que si ce type a au moins une alerte
		if nombre > 0 {
			// Calculer le taux de résolution
			taux := int((float64(resoluesType) / float64(nombre)) * 100)

			statsTable = append(statsTable, DashboardStatsTableItem{
				Type:     typeInfo.label,
				Nombre:   nombre,
				Resolues: resoluesType,
				Taux:     taux,
			})
		}
	}

	// Données d'activité par période
	activityData := []DashboardActivityData{}
	
	// Selon la période, on génère les données d'activité
	if periode != nil && *periode != "" {
		activityData = s.generateActivityData(ctx, commissariatID, debut, fin, *periode)
	}

	// Récupérer les dernières alertes pour le tableau
	alertsItems := []DashboardAlertItem{}
	filtersAlerts := &repository.AlerteFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	alertesForTable, err := s.alerteRepo.List(ctx, filtersAlerts)
	if err == nil && alertesForTable != nil {
		// Limiter à 10 alertes maximum pour le tableau
		maxAlertes := 10
		if len(alertesForTable) > maxAlertes {
			alertesForTable = alertesForTable[:maxAlertes]
		}
		
		for _, alerte := range alertesForTable {
			typeDiffusion := "Locale"
			if alerte.Diffusee {
				// Analyser les destinataires pour déterminer le type
				if len(alerte.DiffusionDestinataires) > 0 {
					destData := alerte.DiffusionDestinataires
					if diffGen, ok := destData["diffusionGenerale"].(bool); ok && diffGen {
						typeDiffusion = "Nationale"
					} else if comms, ok := destData["commissariatsIds"].([]interface{}); ok && len(comms) > 1 {
						typeDiffusion = "Régionale"
					}
				}
			}

			dateDiffusion := "Non diffusée"
			if alerte.Diffusee && alerte.DateDiffusion != nil {
				dateDiffusion = s.formatDateRelative(*alerte.DateDiffusion)
			}

			status := "ACTIVE"
			switch alerte.Statut {
			case "RESOLUE":
				status = "RÉSOLU"
			case "ARCHIVEE":
				status = "ARCHIVÉE"
			case "ACTIVE":
				status = "EN COURS"
			}

			priorite := "NORMALE"
			switch alerte.Niveau {
			case "CRITIQUE":
				priorite = "CRITIQUE"
			case "ELEVE":
				priorite = "HAUTE"
			}

			villeDiffusion := "N/A"
			if alerte.Edges.Commissariat != nil {
				villeDiffusion = alerte.Edges.Commissariat.Ville
			}

			alertsItems = append(alertsItems, DashboardAlertItem{
				ID:             alerte.ID.String(),
				Code:           alerte.Numero,
				TypeAlerte:     s.getTypeAlerteLabel(alerte.TypeAlerte),
				Libelle:        alerte.Titre,
				TypeDiffusion:  typeDiffusion,
				DateDiffusion:  dateDiffusion,
				Status:         status,
				VilleDiffusion: villeDiffusion,
				Priorite:       priorite,
			})
		}
	}

	return &DashboardResponse{
		Stats:        dashboardStats,
		StatsTable:   statsTable,
		ActivityData: activityData,
		Alerts:       alertsItems,
	}, nil
}

// generateActivityData génère les données d'activité selon la période
func (s *service) generateActivityData(ctx context.Context, commissariatID *string, debut, fin *time.Time, typePeriode string) []DashboardActivityData {
	activityData := []DashboardActivityData{}
	
	// Récupérer toutes les alertes de la période
	filters := &repository.AlerteFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	alertes, err := s.alerteRepo.List(ctx, filters)
	if err != nil || alertes == nil {
		// Retourner des données vides en cas d'erreur
		return s.generateEmptyActivityData(typePeriode)
	}

	switch typePeriode {
	case "jour":
		// Données par tranches de 4 heures
		tranches := []struct {
			label      string
			heureDebut int
			heureFin   int
		}{
			{"00h-04h", 0, 4},
			{"04h-08h", 4, 8},
			{"08h-12h", 8, 12},
			{"12h-16h", 12, 16},
			{"16h-20h", 16, 20},
			{"20h-24h", 20, 24},
		}
		
		location, _ := time.LoadLocation("Africa/Abidjan")
		for _, tranche := range tranches {
			totalAlertes := 0
			enCours := 0
			resolues := 0
			
			for _, alerte := range alertes {
				dateLocale := alerte.DateAlerte.In(location)
				heure := dateLocale.Hour() // Maintenant c'est la bonne heure !
				if heure >= tranche.heureDebut && heure < tranche.heureFin {
					totalAlertes++
					if alerte.Statut == "ACTIVE" {
						enCours++
					} else if alerte.Statut == "RESOLUE" {
						resolues++
					}
				}
			}
			
			activityData = append(activityData, DashboardActivityData{
				Period:   tranche.label,
				Alertes:  totalAlertes,
				EnCours:  enCours,
				Resolues: resolues,
			})
		}
		
	case "semaine":
		// Données par jour de la semaine
		jours := []struct {
			label string
			jour  time.Weekday
		}{
			{"Lun", time.Monday},
			{"Mar", time.Tuesday},
			{"Mer", time.Wednesday},
			{"Jeu", time.Thursday},
			{"Ven", time.Friday},
			{"Sam", time.Saturday},
			{"Dim", time.Sunday},
		}
		
		for _, j := range jours {
			totalAlertes := 0
			enCours := 0
			resolues := 0
			
			for _, alerte := range alertes {
				if alerte.DateAlerte.Weekday() == j.jour {
					totalAlertes++
					if alerte.Statut == "ACTIVE" {
						enCours++
					} else if alerte.Statut == "RESOLUE" {
						resolues++
					}
				}
			}
			
			activityData = append(activityData, DashboardActivityData{
				Period:   j.label,
				Alertes:  totalAlertes,
				EnCours:  enCours,
				Resolues: resolues,
			})
		}
		
	case "mois":
		// Données par semaine du mois
		if debut == nil || fin == nil {
			return s.generateEmptyActivityData(typePeriode)
		}
		
		// Calculer le nombre de semaines dans le mois
		firstDayOfMonth := time.Date(debut.Year(), debut.Month(), 1, 0, 0, 0, 0, debut.Location())
		lastDayOfMonth := time.Date(debut.Year(), debut.Month()+1, 0, 23, 59, 59, 0, debut.Location())
		
		weekNum := 1
		currentWeekStart := firstDayOfMonth
		
		for currentWeekStart.Before(lastDayOfMonth) || currentWeekStart.Equal(lastDayOfMonth) {
			currentWeekEnd := currentWeekStart.AddDate(0, 0, 7).Add(-time.Second)
			if currentWeekEnd.After(lastDayOfMonth) {
				currentWeekEnd = lastDayOfMonth
			}
			
			totalAlertes := 0
			enCours := 0
			resolues := 0
			
			for _, alerte := range alertes {
				if (alerte.DateAlerte.After(currentWeekStart) || alerte.DateAlerte.Equal(currentWeekStart)) &&
					(alerte.DateAlerte.Before(currentWeekEnd) || alerte.DateAlerte.Equal(currentWeekEnd)) {
					totalAlertes++
					if alerte.Statut == "ACTIVE" {
						enCours++
					} else if alerte.Statut == "RESOLUE" {
						resolues++
					}
				}
			}
			
			activityData = append(activityData, DashboardActivityData{
				Period:   fmt.Sprintf("Sem %d", weekNum),
				Alertes:  totalAlertes,
				EnCours:  enCours,
				Resolues: resolues,
			})
			
			currentWeekStart = currentWeekEnd.Add(time.Second)
			weekNum++
			
			if weekNum > 5 {
				break // Maximum 5 semaines par mois
			}
		}
		
	case "annee":
		// Données par mois de l'année
		moisLabels := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}
		
		for mois := 1; mois <= 12; mois++ {
			totalAlertes := 0
			enCours := 0
			resolues := 0
			
			for _, alerte := range alertes {
				if int(alerte.DateAlerte.Month()) == mois {
					totalAlertes++
					if alerte.Statut == "ACTIVE" {
						enCours++
					} else if alerte.Statut == "RESOLUE" {
						resolues++
					}
				}
			}
			
			activityData = append(activityData, DashboardActivityData{
				Period:   moisLabels[mois-1],
				Alertes:  totalAlertes,
				EnCours:  enCours,
				Resolues: resolues,
			})
		}
		
	default:
		// Données par année pour "tout"
		// Grouper les alertes par année
		anneesMap := make(map[int]struct {
			total    int
			enCours  int
			resolues int
		})
		
		for _, alerte := range alertes {
			annee := alerte.DateAlerte.Year()
			stats := anneesMap[annee]
			stats.total++
			if alerte.Statut == "ACTIVE" {
				stats.enCours++
			} else if alerte.Statut == "RESOLUE" {
				stats.resolues++
			}
			anneesMap[annee] = stats
		}
		
		// Trier et afficher les 5 dernières années
		currentYear := time.Now().Year()
		for i := currentYear - 4; i <= currentYear; i++ {
			stats := anneesMap[i]
			activityData = append(activityData, DashboardActivityData{
				Period:   fmt.Sprintf("%d", i),
				Alertes:  stats.total,
				EnCours:  stats.enCours,
				Resolues: stats.resolues,
			})
		}
	}

	return activityData
}

// generateEmptyActivityData génère des données vides selon la période
func (s *service) generateEmptyActivityData(typePeriode string) []DashboardActivityData {
	activityData := []DashboardActivityData{}
	
	switch typePeriode {
	case "jour":
		tranches := []string{"00h-04h", "04h-08h", "08h-12h", "12h-16h", "16h-20h", "20h-24h"}
		for _, tranche := range tranches {
			activityData = append(activityData, DashboardActivityData{
				Period: tranche, Alertes: 0, EnCours: 0, Resolues: 0,
			})
		}
	case "semaine":
		jours := []string{"Lun", "Mar", "Mer", "Jeu", "Ven", "Sam", "Dim"}
		for _, jour := range jours {
			activityData = append(activityData, DashboardActivityData{
				Period: jour, Alertes: 0, EnCours: 0, Resolues: 0,
			})
		}
	case "mois":
		for i := 1; i <= 4; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period: fmt.Sprintf("Sem %d", i), Alertes: 0, EnCours: 0, Resolues: 0,
			})
		}
	case "annee":
		mois := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}
		for _, m := range mois {
			activityData = append(activityData, DashboardActivityData{
				Period: m, Alertes: 0, EnCours: 0, Resolues: 0,
			})
		}
	default:
		currentYear := time.Now().Year()
		for i := currentYear - 4; i <= currentYear; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period: fmt.Sprintf("%d", i), Alertes: 0, EnCours: 0, Resolues: 0,
			})
		}
	}
	
	return activityData
}

// getTypeAlerteLabel retourne le label d'un type d'alerte
func (s *service) getTypeAlerteLabel(typeAlerte string) string {
	labels := map[string]string{
		string(TypeAlerteVehiculeVole):     "Véhicule volé",
		string(TypeAlerteSuspectRecherche): "Personne recherchée",
		string(TypeAlerteUrgenceSecurite):  "Urgence sécurité",
		string(TypeAlerteAlerteGenerale):   "Alerte générale",
		string(TypeAlerteAccident):         "Accident",
		string(TypeAlerteIncendie):         "Incendie",
		string(TypeAlerteAggression):       "Aggression",
		string(TypeAlerteAmber):            "AMBER Alert",
		string(TypeAlerteMaintenanceSysteme): "Maintenance système",
		string(TypeAlerteAutre):            "Autre",
	}
	if label, ok := labels[typeAlerte]; ok {
		return label
	}
	return typeAlerte
}

// formatDateRelative formate une date de manière relative
func (s *service) formatDateRelative(date time.Time) string {
	now := time.Now()
	diff := now.Sub(date)

	if diff.Hours() < 1 {
		mins := int(diff.Minutes())
		return fmt.Sprintf("Il y a %d min", mins)
	} else if diff.Hours() < 24 {
		hours := int(diff.Hours())
		return fmt.Sprintf("Il y a %d h", hours)
	} else if diff.Hours() < 48 {
		return "Hier"
	} else if diff.Hours() < 72 {
		return "Il y a 2 jours"
	} else if diff.Hours() < 168 { // 7 jours
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("Il y a %d jours", days)
	} else if diff.Hours() < 720 { // 30 jours
		weeks := int(diff.Hours() / 168)
		return fmt.Sprintf("Il y a %d semaine(s)", weeks)
	} else {
		months := int(diff.Hours() / 720)
		return fmt.Sprintf("Il y a %d mois", months)
	}
}

// structToMap convertit une struct en map pour JSONB
func structToMap(v interface{}) map[string]interface{} {
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
