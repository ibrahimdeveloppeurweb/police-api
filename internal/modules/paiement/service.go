package paiement

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines paiement service interface
type Service interface {
	Create(ctx context.Context, input *CreatePaiementRequest) (*PaiementResponse, error)
	GetByID(ctx context.Context, id string) (*PaiementResponse, error)
	GetByNumeroTransaction(ctx context.Context, numero string) (*PaiementResponse, error)
	List(ctx context.Context, filters *ListPaiementsRequest) (*ListPaiementsResponse, error)
	Update(ctx context.Context, id string, input *UpdatePaiementRequest) (*PaiementResponse, error)
	Delete(ctx context.Context, id string) error
	GetByProcesVerbal(ctx context.Context, pvID string) (*ListPaiementsResponse, error)
	Validate(ctx context.Context, id string, input *ValidatePaiementRequest) (*PaiementResponse, error)
	Refuse(ctx context.Context, id string, input *RefusePaiementRequest) (*PaiementResponse, error)
	Rembourser(ctx context.Context, id string, input *RemboursementRequest) (*PaiementResponse, error)
	GetStatistics(ctx context.Context, filters *ListPaiementsRequest) (*PaiementStatisticsResponse, error)
	// Trésor Public
	GenerateRecuTresor(ctx context.Context, input *RecuTresorRequest) (*RecuTresorResponse, error)
	GetRecuTresor(ctx context.Context, paiementID string) (*RecuTresorResponse, error)
}

// service implements Service interface
type service struct {
	paiementRepo repository.PaiementRepository
	logger       *zap.Logger
}

// NewPaiementService creates a new paiement service
func NewPaiementService(
	paiementRepo repository.PaiementRepository,
	logger *zap.Logger,
) Service {
	return &service{
		paiementRepo: paiementRepo,
		logger:       logger,
	}
}

// Create creates a new paiement
func (s *service) Create(ctx context.Context, input *CreatePaiementRequest) (*PaiementResponse, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Générer un numéro de transaction unique
	numeroTransaction := generateNumeroTransaction()

	repoInput := &repository.CreatePaiementInput{
		ID:                uuid.New().String(),
		NumeroTransaction: numeroTransaction,
		DatePaiement:      time.Now(),
		Montant:           input.Montant,
		MoyenPaiement:     input.MoyenPaiement,
		ReferenceExterne:  input.ReferenceExterne,
		Statut:            "EN_COURS",
		CodeAutorisation:  input.CodeAutorisation,
		DetailsPaiement:   input.DetailsPaiement,
		ProcesVerbalID:    input.ProcesVerbalID,
	}

	paiementEnt, err := s.paiementRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create paiement", zap.Error(err))
		return nil, fmt.Errorf("failed to create paiement: %w", err)
	}

	// Recharger avec les relations
	paiementEnt, err = s.paiementRepo.GetByID(ctx, paiementEnt.ID.String())
	if err != nil {
		s.logger.Error("Failed to reload paiement", zap.Error(err))
		return nil, fmt.Errorf("failed to reload paiement: %w", err)
	}

	return s.entityToResponse(paiementEnt), nil
}

// GetByID gets paiement by ID
func (s *service) GetByID(ctx context.Context, id string) (*PaiementResponse, error) {
	paiementEnt, err := s.paiementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(paiementEnt), nil
}

// GetByNumeroTransaction gets paiement by transaction number
func (s *service) GetByNumeroTransaction(ctx context.Context, numero string) (*PaiementResponse, error) {
	paiementEnt, err := s.paiementRepo.GetByNumeroTransaction(ctx, numero)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(paiementEnt), nil
}

// List gets paiements with filters
func (s *service) List(ctx context.Context, input *ListPaiementsRequest) (*ListPaiementsResponse, error) {
	filters := s.buildFilters(input)

	paiementsEnt, err := s.paiementRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	total, err := s.paiementRepo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	paiements := make([]*PaiementResponse, len(paiementsEnt))
	for i, p := range paiementsEnt {
		paiements[i] = s.entityToResponse(p)
	}

	return &ListPaiementsResponse{
		Paiements: paiements,
		Total:     total,
	}, nil
}

// Update updates paiement
func (s *service) Update(ctx context.Context, id string, input *UpdatePaiementRequest) (*PaiementResponse, error) {
	if err := s.validateUpdateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	repoInput := &repository.UpdatePaiementInput{
		Statut:           input.Statut,
		ReferenceExterne: input.ReferenceExterne,
		CodeAutorisation: input.CodeAutorisation,
		MotifRefus:       input.MotifRefus,
	}

	paiementEnt, err := s.paiementRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	paiementEnt, err = s.paiementRepo.GetByID(ctx, paiementEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(paiementEnt), nil
}

// Delete deletes paiement
func (s *service) Delete(ctx context.Context, id string) error {
	// Vérifier que le paiement existe et n'est pas validé
	paiement, err := s.paiementRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if paiement.Statut == "VALIDE" {
		return fmt.Errorf("cannot delete a validated payment")
	}

	return s.paiementRepo.Delete(ctx, id)
}

// GetByProcesVerbal gets paiements by proces verbal ID
func (s *service) GetByProcesVerbal(ctx context.Context, pvID string) (*ListPaiementsResponse, error) {
	paiementsEnt, err := s.paiementRepo.GetByProcesVerbal(ctx, pvID)
	if err != nil {
		return nil, err
	}

	paiements := make([]*PaiementResponse, len(paiementsEnt))
	for i, p := range paiementsEnt {
		paiements[i] = s.entityToResponse(p)
	}

	return &ListPaiementsResponse{
		Paiements: paiements,
		Total:     len(paiements),
	}, nil
}

// Validate validates a paiement
func (s *service) Validate(ctx context.Context, id string, input *ValidatePaiementRequest) (*PaiementResponse, error) {
	// Vérifier que le paiement existe et est en cours
	paiement, err := s.paiementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if paiement.Statut != "EN_COURS" {
		return nil, fmt.Errorf("payment is not in pending status, current status: %s", paiement.Statut)
	}

	now := time.Now()
	statut := "VALIDE"
	repoInput := &repository.UpdatePaiementInput{
		Statut:           &statut,
		CodeAutorisation: &input.CodeAutorisation,
		DateValidation:   &now,
	}

	paiementEnt, err := s.paiementRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	paiementEnt, err = s.paiementRepo.GetByID(ctx, paiementEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(paiementEnt), nil
}

// Refuse refuses a paiement
func (s *service) Refuse(ctx context.Context, id string, input *RefusePaiementRequest) (*PaiementResponse, error) {
	// Vérifier que le paiement existe et est en cours
	paiement, err := s.paiementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if paiement.Statut != "EN_COURS" {
		return nil, fmt.Errorf("payment is not in pending status, current status: %s", paiement.Statut)
	}

	statut := "REFUSE"
	repoInput := &repository.UpdatePaiementInput{
		Statut:     &statut,
		MotifRefus: &input.MotifRefus,
	}

	paiementEnt, err := s.paiementRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	paiementEnt, err = s.paiementRepo.GetByID(ctx, paiementEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(paiementEnt), nil
}

// Rembourser processes a refund
func (s *service) Rembourser(ctx context.Context, id string, input *RemboursementRequest) (*PaiementResponse, error) {
	// Vérifier que le paiement existe et est validé
	paiement, err := s.paiementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if paiement.Statut != "VALIDE" {
		return nil, fmt.Errorf("only validated payments can be refunded, current status: %s", paiement.Statut)
	}

	// Vérifier le montant du remboursement
	if input.MontantRembourse > 0 && input.MontantRembourse > paiement.Montant {
		return nil, fmt.Errorf("refund amount cannot exceed payment amount")
	}

	statut := "REMBOURSE"
	details := fmt.Sprintf("Remboursement - Motif: %s", input.Motif)
	if input.MontantRembourse > 0 {
		details = fmt.Sprintf("%s - Montant: %.2f", details, input.MontantRembourse)
	}

	repoInput := &repository.UpdatePaiementInput{
		Statut:          &statut,
		DetailsPaiement: &details,
	}

	paiementEnt, err := s.paiementRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Recharger avec les relations
	paiementEnt, err = s.paiementRepo.GetByID(ctx, paiementEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(paiementEnt), nil
}

// GetStatistics gets statistics for paiements
func (s *service) GetStatistics(ctx context.Context, input *ListPaiementsRequest) (*PaiementStatisticsResponse, error) {
	filters := s.buildFilters(input)

	stats, err := s.paiementRepo.GetStatistics(ctx, filters)
	if err != nil {
		return nil, err
	}

	evolutionMensuelle := make([]*MontantMensuel, len(stats.EvolutionMensuelle))
	for i, m := range stats.EvolutionMensuelle {
		evolutionMensuelle[i] = &MontantMensuel{
			Mois:    m.Mois,
			Montant: m.Montant,
			Nombre:  m.Nombre,
		}
	}

	return &PaiementStatisticsResponse{
		TotalPaiements:     stats.Total,
		MontantTotal:       stats.MontantTotal,
		MontantValide:      stats.MontantValide,
		MontantEnCours:     stats.MontantEnCours,
		MontantRembourse:   stats.MontantRembourse,
		ParStatut:          stats.ParStatut,
		ParMoyenPaiement:   stats.ParMoyenPaiement,
		EvolutionMensuelle: evolutionMensuelle,
	}, nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreatePaiementRequest) error {
	if input.ProcesVerbalID == "" {
		return fmt.Errorf("proces_verbal_id is required")
	}
	if input.Montant <= 0 {
		return fmt.Errorf("montant must be positive")
	}
	if input.MoyenPaiement == "" {
		return fmt.Errorf("moyen_paiement is required")
	}

	validMoyens := map[string]bool{
		"CB":            true,
		"CHEQUE":        true,
		"ESPECES":       true,
		"VIREMENT":      true,
		"MOBILE_MONEY":  true,
		"TRESOR_PUBLIC": true,
	}
	if !validMoyens[input.MoyenPaiement] {
		return fmt.Errorf("invalid moyen_paiement: %s", input.MoyenPaiement)
	}

	return nil
}

func (s *service) validateUpdateInput(input *UpdatePaiementRequest) error {
	if input.Statut != nil {
		validStatuts := map[string]bool{
			"EN_COURS":   true,
			"VALIDE":     true,
			"REFUSE":     true,
			"REMBOURSE":  true,
		}
		if !validStatuts[*input.Statut] {
			return fmt.Errorf("invalid statut: %s", *input.Statut)
		}
	}
	return nil
}

func (s *service) buildFilters(input *ListPaiementsRequest) *repository.PaiementFilters {
	if input == nil {
		return nil
	}

	return &repository.PaiementFilters{
		ProcesVerbalID: input.ProcesVerbalID,
		Statut:         input.Statut,
		MoyenPaiement:  input.MoyenPaiement,
		DateDebut:      input.DateDebut,
		DateFin:        input.DateFin,
		MontantMin:     input.MontantMin,
		MontantMax:     input.MontantMax,
		Limit:          input.Limit,
		Offset:         input.Offset,
	}
}

func (s *service) entityToResponse(paiementEnt *ent.Paiement) *PaiementResponse {
	response := &PaiementResponse{
		ID:                paiementEnt.ID.String(),
		NumeroTransaction: paiementEnt.NumeroTransaction,
		DatePaiement:      paiementEnt.DatePaiement,
		Montant:           paiementEnt.Montant,
		MoyenPaiement:     paiementEnt.MoyenPaiement,
		ReferenceExterne:  paiementEnt.ReferenceExterne,
		Statut:            paiementEnt.Statut,
		CodeAutorisation:  paiementEnt.CodeAutorisation,
		DetailsPaiement:   paiementEnt.DetailsPaiement,
		MotifRefus:        paiementEnt.MotifRefus,
		CreatedAt:         paiementEnt.CreatedAt,
		UpdatedAt:         paiementEnt.UpdatedAt,
	}

	// Date de validation (si non nulle)
	if !paiementEnt.DateValidation.IsZero() {
		response.DateValidation = &paiementEnt.DateValidation
	}

	// Ajouter le résumé du PV si chargé
	if paiementEnt.Edges.ProcesVerbal != nil {
		pv := paiementEnt.Edges.ProcesVerbal
		response.ProcesVerbal = &ProcesVerbalSummary{
			ID:           pv.ID.String(),
			NumeroPV:     pv.NumeroPv,
			DateEmission: pv.DateEmission,
			MontantTotal: pv.MontantTotal,
			Statut:       pv.Statut,
		}
	}

	return response
}

// generateNumeroTransaction generates a unique transaction number
func generateNumeroTransaction() string {
	now := time.Now()
	return fmt.Sprintf("TXN%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
}

// GenerateRecuTresor generates a treasury receipt for a TRESOR_PUBLIC payment
func (s *service) GenerateRecuTresor(ctx context.Context, input *RecuTresorRequest) (*RecuTresorResponse, error) {
	// Récupérer le paiement
	paiement, err := s.paiementRepo.GetByID(ctx, input.PaiementID)
	if err != nil {
		return nil, err
	}

	// Vérifier que c'est un paiement Trésor Public
	if paiement.MoyenPaiement != "TRESOR_PUBLIC" {
		return nil, fmt.Errorf("payment must be TRESOR_PUBLIC type")
	}

	// Générer le numéro de reçu
	numeroRecu := generateNumeroRecuTresor()

	// Construire les données QR Code (pour vérification)
	qrData := fmt.Sprintf("TRESOR|%s|%s|%.2f|%s",
		numeroRecu, paiement.NumeroTransaction, paiement.Montant, time.Now().Format("2006-01-02"))

	// Récupérer les infos du PV si disponible
	numeroPV := ""
	datePV := time.Now()
	nomContrevenant := ""
	if paiement.Edges.ProcesVerbal != nil {
		pv := paiement.Edges.ProcesVerbal
		numeroPV = pv.NumeroPv
		datePV = pv.DateEmission
		// Essayer de récupérer le nom du contrevenant depuis les infractions du PV
		if pv.Edges.Infractions != nil && len(pv.Edges.Infractions) > 0 {
			for _, inf := range pv.Edges.Infractions {
				if inf.Edges.Conducteur != nil {
					nomContrevenant = inf.Edges.Conducteur.Nom + " " + inf.Edges.Conducteur.Prenom
					break
				}
			}
		}
	}

	// Mettre à jour le paiement avec les infos trésor et le valider
	now := time.Now()
	statut := "VALIDE"
	details := fmt.Sprintf("Paiement Trésor Public - Reçu: %s - Bureau: %s - Agent: %s",
		numeroRecu, input.BureauTresor, input.AgentTresor)

	updateInput := &repository.UpdatePaiementInput{
		Statut:           &statut,
		DateValidation:   &now,
		DetailsPaiement:  &details,
		CodeAutorisation: &numeroRecu,
	}

	_, err = s.paiementRepo.Update(ctx, input.PaiementID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return &RecuTresorResponse{
		NumeroRecu:       numeroRecu,
		DateEmission:     now,
		Montant:          paiement.Montant,
		MontantEnLettres: montantEnLettres(paiement.Montant),
		NumeroPV:         numeroPV,
		DatePV:           datePV,
		NomContrevenant:  nomContrevenant,
		AgentTresor:      input.AgentTresor,
		BureauTresor:     input.BureauTresor,
		QRCodeData:       qrData,
		PaiementID:       input.PaiementID,
		CreatedAt:        now,
	}, nil
}

// GetRecuTresor retrieves an existing treasury receipt for a payment
func (s *service) GetRecuTresor(ctx context.Context, paiementID string) (*RecuTresorResponse, error) {
	paiement, err := s.paiementRepo.GetByID(ctx, paiementID)
	if err != nil {
		return nil, err
	}

	// Vérifier que c'est un paiement Trésor Public validé
	if paiement.MoyenPaiement != "TRESOR_PUBLIC" {
		return nil, fmt.Errorf("payment is not TRESOR_PUBLIC type")
	}

	if paiement.Statut != "VALIDE" || paiement.CodeAutorisation == "" {
		return nil, fmt.Errorf("recu tresor not found")
	}

	// Récupérer les infos du PV
	numeroPV := ""
	datePV := time.Now()
	nomContrevenant := ""
	if paiement.Edges.ProcesVerbal != nil {
		pv := paiement.Edges.ProcesVerbal
		numeroPV = pv.NumeroPv
		datePV = pv.DateEmission
	}

	// Reconstruire les données QR
	qrData := fmt.Sprintf("TRESOR|%s|%s|%.2f|%s",
		paiement.CodeAutorisation, paiement.NumeroTransaction, paiement.Montant, paiement.DateValidation.Format("2006-01-02"))

	// Extraire agent et bureau depuis les détails
	agentTresor := ""
	bureauTresor := ""
	// Format: "Paiement Trésor Public - Reçu: XXX - Bureau: YYY - Agent: ZZZ"
	// Simple extraction
	if paiement.DetailsPaiement != "" {
		// Parse details to extract agent and bureau (simplified)
		agentTresor = "Agent Trésor"
		bureauTresor = "Bureau Trésor"
	}

	return &RecuTresorResponse{
		NumeroRecu:       paiement.CodeAutorisation,
		DateEmission:     paiement.DateValidation,
		Montant:          paiement.Montant,
		MontantEnLettres: montantEnLettres(paiement.Montant),
		NumeroPV:         numeroPV,
		DatePV:           datePV,
		NomContrevenant:  nomContrevenant,
		AgentTresor:      agentTresor,
		BureauTresor:     bureauTresor,
		QRCodeData:       qrData,
		PaiementID:       paiementID,
		CreatedAt:        paiement.DateValidation,
	}, nil
}

// generateNumeroRecuTresor generates a unique treasury receipt number
func generateNumeroRecuTresor() string {
	now := time.Now()
	return fmt.Sprintf("RCU-TR-%s-%06d", now.Format("20060102"), now.Nanosecond()/1000)
}

// montantEnLettres converts amount to words (simplified French version)
func montantEnLettres(montant float64) string {
	// Simplified version - in production, use a proper library
	unites := []string{"", "un", "deux", "trois", "quatre", "cinq", "six", "sept", "huit", "neuf"}
	dizaines := []string{"", "dix", "vingt", "trente", "quarante", "cinquante", "soixante", "soixante-dix", "quatre-vingt", "quatre-vingt-dix"}

	intPart := int(montant)
	centimes := int((montant - float64(intPart)) * 100)

	if intPart == 0 {
		return "zéro franc CFA"
	}

	result := ""

	// Milliers
	if intPart >= 1000 {
		milliers := intPart / 1000
		if milliers == 1 {
			result += "mille "
		} else if milliers < 10 {
			result += unites[milliers] + " mille "
		} else {
			result += fmt.Sprintf("%d mille ", milliers)
		}
		intPart = intPart % 1000
	}

	// Centaines
	if intPart >= 100 {
		centaines := intPart / 100
		if centaines == 1 {
			result += "cent "
		} else {
			result += unites[centaines] + " cent "
		}
		intPart = intPart % 100
	}

	// Dizaines et unités
	if intPart >= 10 {
		d := intPart / 10
		u := intPart % 10
		if d == 1 {
			// 10-19
			special := []string{"dix", "onze", "douze", "treize", "quatorze", "quinze", "seize", "dix-sept", "dix-huit", "dix-neuf"}
			result += special[u] + " "
		} else {
			result += dizaines[d]
			if u > 0 {
				result += "-" + unites[u]
			}
			result += " "
		}
	} else if intPart > 0 {
		result += unites[intPart] + " "
	}

	result += "francs CFA"

	if centimes > 0 {
		result += fmt.Sprintf(" et %d centimes", centimes)
	}

	return result
}
