package recours

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines recours service interface
type Service interface {
	Create(ctx context.Context, input *CreateRecoursRequest) (*RecoursResponse, error)
	GetByID(ctx context.Context, id string) (*RecoursResponse, error)
	GetByNumeroRecours(ctx context.Context, numero string) (*RecoursResponse, error)
	List(ctx context.Context, filters *ListRecoursRequest) (*ListRecoursResponse, error)
	Update(ctx context.Context, id string, input *UpdateRecoursRequest) (*RecoursResponse, error)
	Delete(ctx context.Context, id string) error
	GetByProcesVerbal(ctx context.Context, pvID string) (*ListRecoursResponse, error)
	Traiter(ctx context.Context, id string, input *TraiterRecoursRequest, userID string) (*RecoursResponse, error)
	Assigner(ctx context.Context, id string, input *AssignerRecoursRequest) (*RecoursResponse, error)
	Abandonner(ctx context.Context, id string, input *AbandonnerRecoursRequest) (*RecoursResponse, error)
	GetEnCours(ctx context.Context) (*ListRecoursResponse, error)
	GetStatistics(ctx context.Context, filters *ListRecoursRequest) (*RecoursStatisticsResponse, error)
	GetEtapes(ctx context.Context, id string) ([]*EtapeRecoursResponse, error)
}

// service implements Service interface
type service struct {
	recoursRepo repository.RecoursRepository
	logger      *zap.Logger
}

// NewRecoursService creates a new recours service
func NewRecoursService(
	recoursRepo repository.RecoursRepository,
	logger *zap.Logger,
) Service {
	return &service{
		recoursRepo: recoursRepo,
		logger:      logger,
	}
}

// Create creates a new recours
func (s *service) Create(ctx context.Context, input *CreateRecoursRequest) (*RecoursResponse, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Générer un numéro de recours unique
	numeroRecours := generateNumeroRecours()

	repoInput := &repository.CreateRecoursInput{
		ID:                 uuid.New().String(),
		NumeroRecours:      numeroRecours,
		DateRecours:        time.Now(),
		TypeRecours:        input.TypeRecours,
		Motif:              input.Motif,
		Argumentaire:       input.Argumentaire,
		Statut:             "DEPOSE",
		AutoriteCompetente: input.AutoriteCompetente,
		DateLimiteRecours:  input.DateLimiteRecours,
		Observations:       input.Observations,
		ProcesVerbalID:     input.ProcesVerbalID,
	}

	recoursEnt, err := s.recoursRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create recours", zap.Error(err))
		return nil, fmt.Errorf("failed to create recours: %w", err)
	}

	// Recharger avec les relations
	recoursEnt, err = s.recoursRepo.GetByID(ctx, recoursEnt.ID.String())
	if err != nil {
		s.logger.Error("Failed to reload recours", zap.Error(err))
		return nil, fmt.Errorf("failed to reload recours: %w", err)
	}

	return s.entityToResponse(recoursEnt), nil
}

// GetByID gets recours by ID
func (s *service) GetByID(ctx context.Context, id string) (*RecoursResponse, error) {
	recoursEnt, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(recoursEnt), nil
}

// GetByNumeroRecours gets recours by numero
func (s *service) GetByNumeroRecours(ctx context.Context, numero string) (*RecoursResponse, error) {
	recoursEnt, err := s.recoursRepo.GetByNumeroRecours(ctx, numero)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(recoursEnt), nil
}

// List gets recours with filters
func (s *service) List(ctx context.Context, input *ListRecoursRequest) (*ListRecoursResponse, error) {
	filters := s.buildFilters(input)

	recoursEnt, err := s.recoursRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	total, err := s.recoursRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	recoursList := make([]*RecoursResponse, len(recoursEnt))
	for i, r := range recoursEnt {
		recoursList[i] = s.entityToResponse(r)
	}

	return &ListRecoursResponse{
		Recours: recoursList,
		Total:   total,
	}, nil
}

// Update updates recours
func (s *service) Update(ctx context.Context, id string, input *UpdateRecoursRequest) (*RecoursResponse, error) {
	// Vérifier que le recours existe et peut être modifié
	rec, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Statut != "DEPOSE" && rec.Statut != "EN_COURS" {
		return nil, fmt.Errorf("cannot update a processed recours")
	}

	repoInput := &repository.UpdateRecoursInput{
		TypeRecours:        input.TypeRecours,
		Motif:              input.Motif,
		Argumentaire:       input.Argumentaire,
		AutoriteCompetente: input.AutoriteCompetente,
		Observations:       input.Observations,
	}

	recoursEnt, err := s.recoursRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	recoursEnt, err = s.recoursRepo.GetByID(ctx, recoursEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(recoursEnt), nil
}

// Delete deletes recours
func (s *service) Delete(ctx context.Context, id string) error {
	// Vérifier que le recours peut être supprimé
	rec, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rec.Statut != "DEPOSE" {
		return fmt.Errorf("can only delete recours with DEPOSE status")
	}

	return s.recoursRepo.Delete(ctx, id)
}

// GetByProcesVerbal gets recours by PV ID
func (s *service) GetByProcesVerbal(ctx context.Context, pvID string) (*ListRecoursResponse, error) {
	recoursEnt, err := s.recoursRepo.GetByProcesVerbal(ctx, pvID)
	if err != nil {
		return nil, err
	}

	recoursList := make([]*RecoursResponse, len(recoursEnt))
	for i, r := range recoursEnt {
		recoursList[i] = s.entityToResponse(r)
	}

	return &ListRecoursResponse{
		Recours: recoursList,
		Total:   len(recoursList),
	}, nil
}

// Traiter processes a recours
func (s *service) Traiter(ctx context.Context, id string, input *TraiterRecoursRequest, userID string) (*RecoursResponse, error) {
	// Vérifier que le recours existe et peut être traité
	rec, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Statut != "DEPOSE" && rec.Statut != "EN_COURS" {
		return nil, fmt.Errorf("recours already processed, current status: %s", rec.Statut)
	}

	// Déterminer le nouveau statut
	var newStatut string
	switch input.Decision {
	case "ACCEPTE":
		newStatut = "ACCEPTE"
	case "REFUSE_PARTIEL", "REFUSE_TOTAL":
		newStatut = "REFUSE"
	default:
		return nil, fmt.Errorf("invalid decision: %s", input.Decision)
	}

	now := time.Now()
	recoursPossible := true
	if input.RecoursPossible != nil {
		recoursPossible = *input.RecoursPossible
	}

	repoInput := &repository.UpdateRecoursInput{
		Statut:            &newStatut,
		DateTraitement:    &now,
		Decision:          &input.Decision,
		MotifDecision:     &input.MotifDecision,
		ReferenceDecision: input.ReferenceDecision,
		NouveauMontant:    input.NouveauMontant,
		RecoursPossible:   &recoursPossible,
		TraiteParID:       &userID,
	}

	recoursEnt, err := s.recoursRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	recoursEnt, err = s.recoursRepo.GetByID(ctx, recoursEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(recoursEnt), nil
}

// Assigner assigns a recours to an agent
func (s *service) Assigner(ctx context.Context, id string, input *AssignerRecoursRequest) (*RecoursResponse, error) {
	// Vérifier que le recours existe
	rec, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Statut != "DEPOSE" && rec.Statut != "EN_COURS" {
		return nil, fmt.Errorf("cannot assign a processed recours")
	}

	statut := "EN_COURS"
	repoInput := &repository.UpdateRecoursInput{
		Statut:      &statut,
		TraiteParID: &input.TraiteParID,
	}

	recoursEnt, err := s.recoursRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	recoursEnt, err = s.recoursRepo.GetByID(ctx, recoursEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(recoursEnt), nil
}

// Abandonner abandons a recours
func (s *service) Abandonner(ctx context.Context, id string, input *AbandonnerRecoursRequest) (*RecoursResponse, error) {
	// Vérifier que le recours existe
	rec, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Statut != "DEPOSE" && rec.Statut != "EN_COURS" {
		return nil, fmt.Errorf("cannot abandon a processed recours")
	}

	statut := "ABANDONNE"
	now := time.Now()
	repoInput := &repository.UpdateRecoursInput{
		Statut:         &statut,
		DateTraitement: &now,
		Observations:   &input.Motif,
	}

	recoursEnt, err := s.recoursRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	recoursEnt, err = s.recoursRepo.GetByID(ctx, recoursEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(recoursEnt), nil
}

// GetEnCours gets recours in progress
func (s *service) GetEnCours(ctx context.Context) (*ListRecoursResponse, error) {
	statut := "EN_COURS"
	filters := &repository.RecoursFilters{
		Statut: &statut,
	}

	recoursEnt, err := s.recoursRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Ajouter aussi les recours déposés
	statutDepose := "DEPOSE"
	filtersDepose := &repository.RecoursFilters{
		Statut: &statutDepose,
	}

	recoursDeposes, err := s.recoursRepo.List(ctx, filtersDepose)
	if err != nil {
		return nil, err
	}

	allRecours := append(recoursEnt, recoursDeposes...)

	recoursList := make([]*RecoursResponse, len(allRecours))
	for i, r := range allRecours {
		recoursList[i] = s.entityToResponse(r)
	}

	return &ListRecoursResponse{
		Recours: recoursList,
		Total:   len(recoursList),
	}, nil
}

// GetStatistics gets statistics for recours
func (s *service) GetStatistics(ctx context.Context, input *ListRecoursRequest) (*RecoursStatisticsResponse, error) {
	filters := s.buildFilters(input)

	stats, err := s.recoursRepo.GetStatistics(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &RecoursStatisticsResponse{
		TotalRecours:    stats.Total,
		ParStatut:       stats.ParStatut,
		ParType:         stats.ParType,
		ParDecision:     stats.ParDecision,
		TauxAcceptation: stats.TauxAcceptation,
		DelaiMoyenJours: stats.DelaiMoyenJours,
	}, nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreateRecoursRequest) error {
	if input.ProcesVerbalID == "" {
		return fmt.Errorf("proces_verbal_id is required")
	}
	if input.TypeRecours == "" {
		return fmt.Errorf("type_recours is required")
	}
	if input.Motif == "" {
		return fmt.Errorf("motif is required")
	}
	if input.Argumentaire == "" {
		return fmt.Errorf("argumentaire is required")
	}

	validTypes := map[string]bool{
		"GRACIEUX":     true,
		"CONTENTIEUX":  true,
		"HIERARCHIQUE": true,
	}
	if !validTypes[input.TypeRecours] {
		return fmt.Errorf("invalid type_recours: %s", input.TypeRecours)
	}

	return nil
}

func (s *service) buildFilters(input *ListRecoursRequest) *repository.RecoursFilters {
	if input == nil {
		return nil
	}

	return &repository.RecoursFilters{
		ProcesVerbalID: input.ProcesVerbalID,
		TypeRecours:    input.TypeRecours,
		Statut:         input.Statut,
		TraiteParID:    input.TraiteParID,
		DateDebut:      input.DateDebut,
		DateFin:        input.DateFin,
		Limit:          input.Limit,
		Offset:         input.Offset,
	}
}

func (s *service) entityToResponse(recoursEnt *ent.Recours) *RecoursResponse {
	response := &RecoursResponse{
		ID:                 recoursEnt.ID.String(),
		NumeroRecours:      recoursEnt.NumeroRecours,
		DateRecours:        recoursEnt.DateRecours,
		TypeRecours:        recoursEnt.TypeRecours,
		Motif:              recoursEnt.Motif,
		Argumentaire:       recoursEnt.Argumentaire,
		Statut:             recoursEnt.Statut,
		Decision:           recoursEnt.Decision,
		MotifDecision:      recoursEnt.MotifDecision,
		AutoriteCompetente: recoursEnt.AutoriteCompetente,
		ReferenceDecision:  recoursEnt.ReferenceDecision,
		NouveauMontant:     recoursEnt.NouveauMontant,
		RecoursPossible:    recoursEnt.RecoursPossible,
		Observations:       recoursEnt.Observations,
		CreatedAt:          recoursEnt.CreatedAt,
		UpdatedAt:          recoursEnt.UpdatedAt,
	}

	// Dates optionnelles
	if !recoursEnt.DateTraitement.IsZero() {
		response.DateTraitement = &recoursEnt.DateTraitement
		// Calculer le délai de traitement
		delai := int(recoursEnt.DateTraitement.Sub(recoursEnt.DateRecours).Hours() / 24)
		response.DelaiTraitement = &delai
	}
	if !recoursEnt.DateLimiteRecours.IsZero() {
		response.DateLimiteRecours = &recoursEnt.DateLimiteRecours
	}

	// Relations
	if recoursEnt.Edges.ProcesVerbal != nil {
		pv := recoursEnt.Edges.ProcesVerbal
		response.ProcesVerbal = &ProcesVerbalSummary{
			ID:           pv.ID.String(),
			NumeroPV:     pv.NumeroPv,
			DateEmission: pv.DateEmission,
			MontantTotal: pv.MontantTotal,
			Statut:       pv.Statut,
		}
	}

	if recoursEnt.Edges.TraitePar != nil {
		u := recoursEnt.Edges.TraitePar
		response.TraitePar = &UserSummary{
			ID:        u.ID.String(),
			Matricule: u.Matricule,
			Nom:       u.Nom,
			Prenom:    u.Prenom,
		}
	}

	if recoursEnt.Edges.Documents != nil {
		response.NombreDocuments = len(recoursEnt.Edges.Documents)
	}

	return response
}

// GetEtapes returns the workflow steps for a recours
func (s *service) GetEtapes(ctx context.Context, id string) ([]*EtapeRecoursResponse, error) {
	// Vérifier que le recours existe
	rec, err := s.recoursRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Définir les étapes du workflow selon le type de recours
	etapes := s.buildEtapesForRecours(rec)

	return etapes, nil
}

// buildEtapesForRecours builds the workflow steps based on recours state
func (s *service) buildEtapesForRecours(rec *ent.Recours) []*EtapeRecoursResponse {
	etapes := []*EtapeRecoursResponse{}

	// Étape 1: Dépôt du recours
	etapeDepot := &EtapeRecoursResponse{
		Code:        "DEPOT",
		Libelle:     "Dépôt du recours",
		Description: "Le recours a été déposé et enregistré dans le système",
		Ordre:       1,
	}
	dateRecours := rec.DateRecours
	etapeDepot.DateDebut = &dateRecours
	etapeDepot.DateFin = &dateRecours

	if rec.Statut == "DEPOSE" {
		etapeDepot.Statut = "TERMINEE"
	} else {
		etapeDepot.Statut = "TERMINEE"
	}
	etapes = append(etapes, etapeDepot)

	// Étape 2: Prise en charge
	etapePriseEnCharge := &EtapeRecoursResponse{
		Code:        "PRISE_EN_CHARGE",
		Libelle:     "Prise en charge",
		Description: "Le recours est assigné à un agent pour traitement",
		Ordre:       2,
	}
	if rec.Edges.TraitePar != nil {
		traitePar := rec.Edges.TraitePar.Nom + " " + rec.Edges.TraitePar.Prenom
		etapePriseEnCharge.Responsable = &traitePar
	}

	switch rec.Statut {
	case "DEPOSE":
		etapePriseEnCharge.Statut = "A_VENIR"
	case "EN_COURS":
		etapePriseEnCharge.Statut = "EN_COURS"
		etapePriseEnCharge.DateDebut = &rec.UpdatedAt
	default:
		etapePriseEnCharge.Statut = "TERMINEE"
		etapePriseEnCharge.DateDebut = &rec.UpdatedAt
		if !rec.DateTraitement.IsZero() {
			etapePriseEnCharge.DateFin = &rec.DateTraitement
		}
	}
	etapes = append(etapes, etapePriseEnCharge)

	// Étape 3: Instruction
	etapeInstruction := &EtapeRecoursResponse{
		Code:        "INSTRUCTION",
		Libelle:     "Instruction du dossier",
		Description: "Examen des pièces et arguments du recours",
		Ordre:       3,
	}
	if rec.Edges.TraitePar != nil {
		traitePar := rec.Edges.TraitePar.Nom + " " + rec.Edges.TraitePar.Prenom
		etapeInstruction.Responsable = &traitePar
	}

	switch rec.Statut {
	case "DEPOSE", "EN_COURS":
		if rec.Statut == "EN_COURS" {
			etapeInstruction.Statut = "EN_COURS"
		} else {
			etapeInstruction.Statut = "A_VENIR"
		}
	default:
		etapeInstruction.Statut = "TERMINEE"
		if !rec.DateTraitement.IsZero() {
			etapeInstruction.DateFin = &rec.DateTraitement
		}
	}
	etapes = append(etapes, etapeInstruction)

	// Étape 4: Décision
	etapeDecision := &EtapeRecoursResponse{
		Code:        "DECISION",
		Libelle:     "Décision",
		Description: "Notification de la décision finale",
		Ordre:       4,
	}
	if rec.AutoriteCompetente != "" {
		etapeDecision.Responsable = &rec.AutoriteCompetente
	}

	switch rec.Statut {
	case "DEPOSE", "EN_COURS":
		etapeDecision.Statut = "A_VENIR"
	case "ACCEPTE", "REFUSE", "ABANDONNE":
		etapeDecision.Statut = "TERMINEE"
		if !rec.DateTraitement.IsZero() {
			etapeDecision.DateDebut = &rec.DateTraitement
			etapeDecision.DateFin = &rec.DateTraitement
		}
	default:
		etapeDecision.Statut = "A_VENIR"
	}
	etapes = append(etapes, etapeDecision)

	// Étape 5: Clôture (optionnelle, si recours est possible après)
	if rec.RecoursPossible && (rec.Statut == "REFUSE" || rec.Statut == "ACCEPTE") {
		etapeCloture := &EtapeRecoursResponse{
			Code:        "CLOTURE",
			Libelle:     "Clôture ou nouveau recours",
			Description: "Possibilité de déposer un nouveau recours ou clôture définitive",
			Ordre:       5,
			Statut:      "A_VENIR",
		}
		etapes = append(etapes, etapeCloture)
	}

	return etapes
}

// generateNumeroRecours generates a unique recours number
func generateNumeroRecours() string {
	now := time.Now()
	return fmt.Sprintf("REC%s%06d", now.Format("20060102"), now.Nanosecond()/1000)
}
