package pv

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines PV service interface
type Service interface {
	Create(ctx context.Context, input *CreatePVRequest) (*PVResponse, error)
	GetByID(ctx context.Context, id string) (*PVResponse, error)
	GetByNumeroPV(ctx context.Context, numero string) (*PVResponse, error)
	List(ctx context.Context, filters *ListPVRequest) (*ListPVResponse, error)
	Update(ctx context.Context, id string, input *UpdatePVRequest) (*PVResponse, error)
	Delete(ctx context.Context, id string) error
	GetByInfraction(ctx context.Context, infractionID string) (*PVResponse, error)
	Payer(ctx context.Context, id string, input *PayerPVRequest) (*PVResponse, error)
	Contester(ctx context.Context, id string, input *ContesterPVRequest) (*PVResponse, error)
	DeciderContestation(ctx context.Context, id string, input *DecisionContestationRequest) (*PVResponse, error)
	Majorer(ctx context.Context, id string, input *MajorerPVRequest) (*PVResponse, error)
	Annuler(ctx context.Context, id string, input *AnnulerPVRequest) (*PVResponse, error)
	GetExpired(ctx context.Context) (*ListPVResponse, error)
	GetStatistics(ctx context.Context, filters *ListPVRequest) (*PVStatisticsResponse, error)
	EnvoyerRappel(ctx context.Context, id string) (*RappelResponse, error)
	MarquerEnRetard(ctx context.Context, id string) (*PVResponse, error)
}

// service implements Service interface
type service struct {
	pvRepo repository.PVRepository
	logger *zap.Logger
}

// NewPVService creates a new PV service
func NewPVService(
	pvRepo repository.PVRepository,
	logger *zap.Logger,
) Service {
	return &service{
		pvRepo: pvRepo,
		logger: logger,
	}
}

// Create creates a new PV
func (s *service) Create(ctx context.Context, input *CreatePVRequest) (*PVResponse, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Générer un numéro de PV unique
	numeroPV := generateNumeroPV()

	// Calculer la date limite de paiement (45 jours par défaut)
	var dateLimite *time.Time
	if input.DateLimitePaiement != nil {
		dateLimite = input.DateLimitePaiement
	} else {
		dl := time.Now().AddDate(0, 0, 45)
		dateLimite = &dl
	}

	repoInput := &repository.CreatePVInput{
		ID:                 uuid.New().String(),
		NumeroPV:           numeroPV,
		DateEmission:       time.Now(),
		MontantTotal:       input.MontantTotal,
		DateLimitePaiement: dateLimite,
		Statut:             "EMIS",
		Observations:       input.Observations,
		InfractionIDs:      input.InfractionIDs,
		ControleID:         input.ControleID,
		InspectionID:       input.InspectionID,
	}

	pvEnt, err := s.pvRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create PV", zap.Error(err))
		return nil, fmt.Errorf("failed to create PV: %w", err)
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		s.logger.Error("Failed to reload PV", zap.Error(err))
		return nil, fmt.Errorf("failed to reload PV: %w", err)
	}

	return s.entityToResponse(pvEnt), nil
}

// GetByID gets PV by ID
func (s *service) GetByID(ctx context.Context, id string) (*PVResponse, error) {
	pvEnt, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// GetByNumeroPV gets PV by numero
func (s *service) GetByNumeroPV(ctx context.Context, numero string) (*PVResponse, error) {
	pvEnt, err := s.pvRepo.GetByNumeroPV(ctx, numero)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// List gets PVs with filters
func (s *service) List(ctx context.Context, input *ListPVRequest) (*ListPVResponse, error) {
	filters := s.buildFilters(input)

	pvsEnt, err := s.pvRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	total, err := s.pvRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	pvs := make([]*PVResponse, len(pvsEnt))
	for i, pv := range pvsEnt {
		pvs[i] = s.entityToResponse(pv)
	}

	return &ListPVResponse{
		PVs:   pvs,
		Total: total,
	}, nil
}

// Update updates PV
func (s *service) Update(ctx context.Context, id string, input *UpdatePVRequest) (*PVResponse, error) {
	// Vérifier que le PV existe
	_, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	repoInput := &repository.UpdatePVInput{
		MontantTotal:       input.MontantTotal,
		DateLimitePaiement: input.DateLimitePaiement,
		Observations:       input.Observations,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// Delete deletes PV
func (s *service) Delete(ctx context.Context, id string) error {
	// Vérifier que le PV peut être supprimé
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if pv.Statut == "PAYE" {
		return fmt.Errorf("cannot delete a paid PV")
	}

	return s.pvRepo.Delete(ctx, id)
}

// GetByInfraction gets PV by infraction ID
func (s *service) GetByInfraction(ctx context.Context, infractionID string) (*PVResponse, error) {
	pvEnt, err := s.pvRepo.GetByInfraction(ctx, infractionID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// Payer enregistre un paiement sur le PV
func (s *service) Payer(ctx context.Context, id string, input *PayerPVRequest) (*PVResponse, error) {
	// Vérifier que le PV existe et peut être payé
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pv.Statut == "PAYE" {
		return nil, fmt.Errorf("PV already paid")
	}
	if pv.Statut == "ANNULE" {
		return nil, fmt.Errorf("cannot pay a cancelled PV")
	}

	// Calculer le montant à payer (avec majoration si applicable)
	montantDu := pv.MontantTotal
	if pv.MontantMajore > 0 && !pv.DateMajoration.IsZero() && time.Now().After(pv.DateMajoration) {
		montantDu = pv.MontantMajore
	}

	// Déterminer le nouveau statut
	newStatut := "PAYE"
	if input.MontantPaye < montantDu {
		// Paiement partiel - on garde le statut actuel mais on met à jour le montant payé
		newStatut = pv.Statut
	}

	now := time.Now()
	repoInput := &repository.UpdatePVInput{
		Statut:            &newStatut,
		DatePaiement:      &now,
		MontantPaye:       &input.MontantPaye,
		MoyenPaiement:     &input.MoyenPaiement,
		ReferencePaiement: input.ReferencePaiement,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// Contester enregistre une contestation sur le PV
func (s *service) Contester(ctx context.Context, id string, input *ContesterPVRequest) (*PVResponse, error) {
	// Vérifier que le PV existe et peut être contesté
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pv.Statut == "PAYE" {
		return nil, fmt.Errorf("cannot contest a paid PV")
	}
	if pv.Statut == "ANNULE" {
		return nil, fmt.Errorf("cannot contest a cancelled PV")
	}
	if pv.Statut == "CONTESTE" {
		return nil, fmt.Errorf("PV already contested")
	}

	now := time.Now()
	statut := "CONTESTE"
	repoInput := &repository.UpdatePVInput{
		Statut:            &statut,
		DateContestation:  &now,
		MotifContestation: &input.MotifContestation,
		TribunalCompetent: input.TribunalCompetent,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// DeciderContestation enregistre la décision sur une contestation
func (s *service) DeciderContestation(ctx context.Context, id string, input *DecisionContestationRequest) (*PVResponse, error) {
	// Vérifier que le PV existe et est contesté
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pv.Statut != "CONTESTE" {
		return nil, fmt.Errorf("PV is not contested")
	}

	var newStatut string
	var newMontant *float64

	switch input.Decision {
	case "ACCEPTE":
		newStatut = "ANNULE"
	case "REFUSE_PARTIEL":
		newStatut = "EMIS"
		if input.NouveauMontant != nil {
			newMontant = input.NouveauMontant
		}
	case "REFUSE_TOTAL":
		newStatut = "EMIS"
	default:
		return nil, fmt.Errorf("invalid decision: %s", input.Decision)
	}

	repoInput := &repository.UpdatePVInput{
		Statut:               &newStatut,
		DecisionContestation: &input.Motif,
		MontantTotal:         newMontant,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// Majorer applique une majoration sur le PV
func (s *service) Majorer(ctx context.Context, id string, input *MajorerPVRequest) (*PVResponse, error) {
	// Vérifier que le PV existe et peut être majoré
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pv.Statut == "PAYE" {
		return nil, fmt.Errorf("cannot add penalty to a paid PV")
	}
	if pv.Statut == "ANNULE" {
		return nil, fmt.Errorf("cannot add penalty to a cancelled PV")
	}

	statut := "MAJORE"
	repoInput := &repository.UpdatePVInput{
		Statut:         &statut,
		MontantMajore:  &input.MontantMajore,
		DateMajoration: &input.DateMajoration,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// Annuler annule un PV
func (s *service) Annuler(ctx context.Context, id string, input *AnnulerPVRequest) (*PVResponse, error) {
	// Vérifier que le PV existe et peut être annulé
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pv.Statut == "PAYE" {
		return nil, fmt.Errorf("cannot cancel a paid PV")
	}
	if pv.Statut == "ANNULE" {
		return nil, fmt.Errorf("PV already cancelled")
	}

	statut := "ANNULE"
	repoInput := &repository.UpdatePVInput{
		Statut:       &statut,
		Observations: &input.Motif,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}

// GetExpired gets expired PVs
func (s *service) GetExpired(ctx context.Context) (*ListPVResponse, error) {
	pvsEnt, err := s.pvRepo.GetExpired(ctx)
	if err != nil {
		return nil, err
	}

	pvs := make([]*PVResponse, len(pvsEnt))
	for i, pv := range pvsEnt {
		pvs[i] = s.entityToResponse(pv)
	}

	return &ListPVResponse{
		PVs:   pvs,
		Total: len(pvs),
	}, nil
}

// GetStatistics gets statistics for PVs
func (s *service) GetStatistics(ctx context.Context, input *ListPVRequest) (*PVStatisticsResponse, error) {
	filters := s.buildFilters(input)

	stats, err := s.pvRepo.GetStatistics(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &PVStatisticsResponse{
		TotalPV:          stats.Total,
		MontantTotal:     stats.MontantTotal,
		MontantPaye:      stats.MontantPaye,
		MontantImpaye:    stats.MontantImpaye,
		TauxRecouvrement: stats.TauxRecouvrement,
		PVExpires:        stats.PVExpires,
		ParStatut:        stats.ParStatut,
		ParMois:          stats.ParMois,
	}, nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreatePVRequest) error {
	if len(input.InfractionIDs) == 0 {
		return fmt.Errorf("at least one infraction_id is required")
	}
	if input.MontantTotal <= 0 {
		return fmt.Errorf("montant_total must be positive")
	}
	return nil
}

func (s *service) buildFilters(input *ListPVRequest) *repository.PVFilters {
	if input == nil {
		return nil
	}

	return &repository.PVFilters{
		InfractionID: input.InfractionID,
		Statut:       input.Statut,
		DateDebut:    input.DateDebut,
		DateFin:      input.DateFin,
		MontantMin:   input.MontantMin,
		MontantMax:   input.MontantMax,
		Expired:      input.Expired,
		Limit:        input.Limit,
		Offset:       input.Offset,
	}
}

func (s *service) entityToResponse(pvEnt *ent.ProcesVerbal) *PVResponse {
	response := &PVResponse{
		ID:                   pvEnt.ID.String(),
		NumeroPV:             pvEnt.NumeroPv,
		DateEmission:         pvEnt.DateEmission,
		MontantTotal:         pvEnt.MontantTotal,
		MontantMajore:        pvEnt.MontantMajore,
		Statut:               pvEnt.Statut,
		MontantPaye:          pvEnt.MontantPaye,
		MoyenPaiement:        pvEnt.MoyenPaiement,
		ReferencePaiement:    pvEnt.ReferencePaiement,
		MotifContestation:    pvEnt.MotifContestation,
		DecisionContestation: pvEnt.DecisionContestation,
		TribunalCompetent:    pvEnt.TribunalCompetent,
		Observations:         pvEnt.Observations,
		CreatedAt:            pvEnt.CreatedAt,
		UpdatedAt:            pvEnt.UpdatedAt,
	}

	// Dates optionnelles
	if !pvEnt.DateLimitePaiement.IsZero() {
		response.DateLimitePaiement = &pvEnt.DateLimitePaiement
	}
	if !pvEnt.DateMajoration.IsZero() {
		response.DateMajoration = &pvEnt.DateMajoration
	}
	if !pvEnt.DatePaiement.IsZero() {
		response.DatePaiement = &pvEnt.DatePaiement
	}
	if !pvEnt.DateContestation.IsZero() {
		response.DateContestation = &pvEnt.DateContestation
	}

	// Calculer si le PV est expiré
	if !pvEnt.DateLimitePaiement.IsZero() && pvEnt.DateLimitePaiement.Before(time.Now()) &&
		pvEnt.Statut != "PAYE" && pvEnt.Statut != "ANNULE" {
		response.EstExpire = true
	}

	// Calculer le montant restant
	response.MontantRestant = pvEnt.MontantTotal - pvEnt.MontantPaye
	if pvEnt.MontantMajore > 0 && response.EstExpire {
		response.MontantRestant = pvEnt.MontantMajore - pvEnt.MontantPaye
	}

	// Relations - Multiple infractions
	if len(pvEnt.Edges.Infractions) > 0 {
		response.Infractions = make([]*InfractionSummary, 0, len(pvEnt.Edges.Infractions))
		for _, inf := range pvEnt.Edges.Infractions {
			typeInfr := "Non spécifié"
			if inf.Edges.TypeInfraction != nil {
				typeInfr = inf.Edges.TypeInfraction.Libelle
			}
			response.Infractions = append(response.Infractions, &InfractionSummary{
				ID:             inf.ID.String(),
				NumeroPV:       inf.NumeroPv,
				DateInfraction: inf.DateInfraction,
				TypeInfraction: typeInfr,
				LieuInfraction: inf.LieuInfraction,
				MontantAmende:  inf.MontantAmende,
			})
		}
	}

	if pvEnt.Edges.Paiements != nil {
		response.NombrePaiements = len(pvEnt.Edges.Paiements)
	}

	if pvEnt.Edges.Recours != nil {
		response.NombreRecours = len(pvEnt.Edges.Recours)
	}

	return response
}

// generateNumeroPV generates a unique PV number
func generateNumeroPV() string {
	now := time.Now()
	return fmt.Sprintf("PV%s%06d", now.Format("20060102"), now.Nanosecond()/1000)
}

// EnvoyerRappel envoie un rappel de paiement pour un PV
func (s *service) EnvoyerRappel(ctx context.Context, id string) (*RappelResponse, error) {
	// Vérifier que le PV existe
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Vérifier que le PV peut recevoir un rappel
	if pv.Statut == "PAYE" {
		return &RappelResponse{
			PVID:     id,
			NumeroPV: pv.NumeroPv,
			Success:  false,
			Message:  "Le PV est déjà payé",
		}, nil
	}
	if pv.Statut == "ANNULE" {
		return &RappelResponse{
			PVID:     id,
			NumeroPV: pv.NumeroPv,
			Success:  false,
			Message:  "Le PV est annulé",
		}, nil
	}

	// Calculer le montant dû
	montantDu := pv.MontantTotal - pv.MontantPaye
	if pv.MontantMajore > 0 {
		montantDu = pv.MontantMajore - pv.MontantPaye
	}

	// Log l'envoi du rappel (en production, envoyer SMS/email)
	s.logger.Info("Envoi rappel PV",
		zap.String("pv_id", id),
		zap.String("numero_pv", pv.NumeroPv),
		zap.Float64("montant_du", montantDu))

	return &RappelResponse{
		PVID:         id,
		NumeroPV:     pv.NumeroPv,
		DateRappel:   time.Now(),
		NumeroRappel: 1, // À implémenter: compteur de rappels
		MontantDu:    montantDu,
		DateLimite:   pv.DateLimitePaiement,
		Success:      true,
		Message:      "Rappel envoyé avec succès",
	}, nil
}

// MarquerEnRetard marque un PV comme étant en retard de paiement
func (s *service) MarquerEnRetard(ctx context.Context, id string) (*PVResponse, error) {
	// Vérifier que le PV existe
	pv, err := s.pvRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Vérifier que le PV peut être marqué en retard
	if pv.Statut == "PAYE" {
		return nil, fmt.Errorf("cannot mark paid PV as late")
	}
	if pv.Statut == "ANNULE" {
		return nil, fmt.Errorf("cannot mark cancelled PV as late")
	}
	if pv.Statut == "EN_RETARD" {
		return nil, fmt.Errorf("PV already marked as late")
	}

	// Marquer comme en retard
	statut := "EN_RETARD"
	repoInput := &repository.UpdatePVInput{
		Statut: &statut,
	}

	pvEnt, err := s.pvRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	pvEnt, err = s.pvRepo.GetByID(ctx, pvEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(pvEnt), nil
}
