package infraction

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines infraction service interface
type Service interface {
	Create(ctx context.Context, input *CreateInfractionRequest) (*InfractionResponse, error)
	GetByID(ctx context.Context, id string) (*InfractionResponse, error)
	GetByNumeroPV(ctx context.Context, numeroPV string) (*InfractionResponse, error)
	List(ctx context.Context, filters *ListInfractionsRequest) (*ListInfractionsResponse, error)
	Update(ctx context.Context, id string, input *UpdateInfractionRequest) (*InfractionResponse, error)
	Delete(ctx context.Context, id string) error
	GetByControle(ctx context.Context, controleID string) (*ListInfractionsResponse, error)
	GetByVehicule(ctx context.Context, vehiculeID string) (*ListInfractionsResponse, error)
	GetByConducteur(ctx context.Context, conducteurID string) (*ListInfractionsResponse, error)
	GetByStatut(ctx context.Context, statut string) (*ListInfractionsResponse, error)
	GetStatistics(ctx context.Context, filters *ListInfractionsRequest) (*InfractionStatisticsResponse, error)
	GeneratePV(ctx context.Context, infractionID string) (*PVGenerationResponse, error)
	ValidateInfraction(ctx context.Context, infractionID string) (*InfractionValidationResponse, error)
	ArchiveInfraction(ctx context.Context, infractionID string) (*InfractionArchiveResponse, error)
	UnarchiveInfraction(ctx context.Context, infractionID string) (*InfractionArchiveResponse, error)
	RecordPayment(ctx context.Context, infractionID string, input *PaymentRequest) (*PaymentResponse, error)
	GroupByType(ctx context.Context, filters *ListInfractionsRequest) ([]*InfractionsByTypeResponse, error)
	GetTypesInfractions(ctx context.Context) ([]*TypeInfractionSummary, error)
	GetCategories(ctx context.Context) ([]*CategorieResponse, error)
	GetDashboard(ctx context.Context, input *DashboardRequest) (*DashboardResponse, error)
}

// service implements Service interface
type service struct {
	infractionRepo     repository.InfractionRepository
	infractionTypeRepo repository.InfractionTypeRepository
	controleRepo       repository.ControleRepository
	vehiculeRepo       repository.VehiculeRepository
	conducteurRepo     repository.ConducteurRepository
	pvRepo             repository.PVRepository
	logger             *zap.Logger
}

// NewService creates a new infraction service
func NewService(
	infractionRepo repository.InfractionRepository,
	infractionTypeRepo repository.InfractionTypeRepository,
	controleRepo repository.ControleRepository,
	vehiculeRepo repository.VehiculeRepository,
	conducteurRepo repository.ConducteurRepository,
	pvRepo repository.PVRepository,
	logger *zap.Logger,
) Service {
	return &service{
		infractionRepo:     infractionRepo,
		infractionTypeRepo: infractionTypeRepo,
		controleRepo:       controleRepo,
		vehiculeRepo:       vehiculeRepo,
		conducteurRepo:     conducteurRepo,
		pvRepo:             pvRepo,
		logger:             logger,
	}
}

// Create creates a new infraction
func (s *service) Create(ctx context.Context, input *CreateInfractionRequest) (*InfractionResponse, error) {
	// Validation métier
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Vérifier que les entités existent et récupérer le type d'infraction pour les montants
	typeInfraction, err := s.validateAndGetTypeInfraction(ctx, input)
	if err != nil {
		return nil, err
	}

	// Valeurs par défaut
	if input.Statut == "" {
		input.Statut = "CONSTATEE"
	}

	// Calculer montant et points basé sur le type d'infraction
	montantAmende := typeInfraction.Amende
	pointsRetires := typeInfraction.Points

	// Ajustements pour excès de vitesse
	if input.VitesseRetenue != nil && input.VitesseLimitee != nil {
		exces := *input.VitesseRetenue - *input.VitesseLimitee
		montantAmende, pointsRetires = s.calculateSpeedingPenalty(exces, typeInfraction.Amende, typeInfraction.Points)
	}

	// Majoration pour flagrant délit
	if input.FlagrantDelit {
		montantAmende = montantAmende * 1.5
	}

	repoInput := &repository.CreateInfractionInput{
		ID:                   uuid.New().String(),
		DateInfraction:       input.DateInfraction,
		LieuInfraction:       input.LieuInfraction,
		Circonstances:        input.Circonstances,
		VitesseRetenue:       input.VitesseRetenue,
		VitesseLimitee:       input.VitesseLimitee,
		AppareilMesure:       input.AppareilMesure,
		MontantAmende:        montantAmende,
		PointsRetires:        pointsRetires,
		Statut:               input.Statut,
		Observations:         input.Observations,
		FlagrantDelit:        input.FlagrantDelit,
		Accident:             input.Accident,
		ControleID:           input.ControleID,
		TypeInfractionID:     input.TypeInfractionID,
		VehiculeID:           input.VehiculeID,
		ConducteurID:         input.ConducteurID,
	}

	infractionEnt, err := s.infractionRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create infraction", zap.Error(err))
		return nil, fmt.Errorf("failed to create infraction: %w", err)
	}

	return s.entityToResponse(infractionEnt), nil
}

// GetByID gets infraction by ID
func (s *service) GetByID(ctx context.Context, id string) (*InfractionResponse, error) {
	infractionEnt, err := s.infractionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(infractionEnt), nil
}

// GetByNumeroPV gets infraction by numero PV
func (s *service) GetByNumeroPV(ctx context.Context, numeroPV string) (*InfractionResponse, error) {
	infractionEnt, err := s.infractionRepo.GetByNumeroPV(ctx, numeroPV)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(infractionEnt), nil
}

// List gets infractions with filters
func (s *service) List(ctx context.Context, input *ListInfractionsRequest) (*ListInfractionsResponse, error) {
	filters := s.buildRepositoryFilters(input)

	infractionsEnt, err := s.infractionRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	infractions := make([]*InfractionResponse, len(infractionsEnt))
	for i, inf := range infractionsEnt {
		infractions[i] = s.entityToResponse(inf)
	}

	return &ListInfractionsResponse{
		Infractions: infractions,
		Total:       len(infractions),
	}, nil
}

// Update updates infraction
func (s *service) Update(ctx context.Context, id string, input *UpdateInfractionRequest) (*InfractionResponse, error) {
	// Validation métier
	if err := s.validateUpdateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Recalculer montant et points si le type change
	var montantAmende *float64
	var pointsRetires *int

	if input.TypeInfractionID != nil {
		// Récupérer le nouveau type d'infraction
		typeInfraction, err := s.infractionTypeRepo.GetByID(ctx, *input.TypeInfractionID)
		if err != nil {
			return nil, fmt.Errorf("invalid infraction type: %w", err)
		}

		if !typeInfraction.Active {
			return nil, fmt.Errorf("infraction type is not active")
		}

		// Recalculer les valeurs basées sur le nouveau type
		newAmende := typeInfraction.Amende
		newPoints := typeInfraction.Points

		// Ajuster pour excès de vitesse si applicable
		if input.VitesseRetenue != nil && input.VitesseLimitee != nil {
			exces := *input.VitesseRetenue - *input.VitesseLimitee
			newAmende, newPoints = s.calculateSpeedingPenalty(exces, typeInfraction.Amende, typeInfraction.Points)
		}

		// Majoration pour flagrant délit
		if input.FlagrantDelit != nil && *input.FlagrantDelit {
			newAmende = newAmende * 1.5
		}

		montantAmende = &newAmende
		pointsRetires = &newPoints
	}

	repoInput := &repository.UpdateInfractionInput{
		NumeroPV:            input.NumeroPV,
		DateInfraction:      input.DateInfraction,
		LieuInfraction:      input.LieuInfraction,
		Circonstances:       input.Circonstances,
		VitesseRetenue:      input.VitesseRetenue,
		VitesseLimitee:      input.VitesseLimitee,
		AppareilMesure:      input.AppareilMesure,
		MontantAmende:       montantAmende,
		PointsRetires:       pointsRetires,
		Statut:              input.Statut,
		Observations:        input.Observations,
		FlagrantDelit:       input.FlagrantDelit,
		Accident:            input.Accident,
		TypeInfractionID:    input.TypeInfractionID,
	}

	infractionEnt, err := s.infractionRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(infractionEnt), nil
}

// Delete deletes infraction
func (s *service) Delete(ctx context.Context, id string) error {
	return s.infractionRepo.Delete(ctx, id)
}

// GetByControle gets infractions by controle
func (s *service) GetByControle(ctx context.Context, controleID string) (*ListInfractionsResponse, error) {
	infractionsEnt, err := s.infractionRepo.GetByControle(ctx, controleID)
	if err != nil {
		return nil, err
	}

	infractions := make([]*InfractionResponse, len(infractionsEnt))
	for i, inf := range infractionsEnt {
		infractions[i] = s.entityToResponse(inf)
	}

	return &ListInfractionsResponse{
		Infractions: infractions,
		Total:       len(infractions),
	}, nil
}

// GetByVehicule gets infractions by vehicule
func (s *service) GetByVehicule(ctx context.Context, vehiculeID string) (*ListInfractionsResponse, error) {
	infractionsEnt, err := s.infractionRepo.GetByVehicule(ctx, vehiculeID)
	if err != nil {
		return nil, err
	}

	infractions := make([]*InfractionResponse, len(infractionsEnt))
	for i, inf := range infractionsEnt {
		infractions[i] = s.entityToResponse(inf)
	}

	return &ListInfractionsResponse{
		Infractions: infractions,
		Total:       len(infractions),
	}, nil
}

// GetByConducteur gets infractions by conducteur
func (s *service) GetByConducteur(ctx context.Context, conducteurID string) (*ListInfractionsResponse, error) {
	infractionsEnt, err := s.infractionRepo.GetByConducteur(ctx, conducteurID)
	if err != nil {
		return nil, err
	}

	infractions := make([]*InfractionResponse, len(infractionsEnt))
	for i, inf := range infractionsEnt {
		infractions[i] = s.entityToResponse(inf)
	}

	return &ListInfractionsResponse{
		Infractions: infractions,
		Total:       len(infractions),
	}, nil
}

// GetByStatut gets infractions by statut
func (s *service) GetByStatut(ctx context.Context, statut string) (*ListInfractionsResponse, error) {
	infractionsEnt, err := s.infractionRepo.GetByStatut(ctx, statut)
	if err != nil {
		return nil, err
	}

	infractions := make([]*InfractionResponse, len(infractionsEnt))
	for i, inf := range infractionsEnt {
		infractions[i] = s.entityToResponse(inf)
	}

	return &ListInfractionsResponse{
		Infractions: infractions,
		Total:       len(infractions),
	}, nil
}

// GetStatistics gets statistics for infractions
func (s *service) GetStatistics(ctx context.Context, input *ListInfractionsRequest) (*InfractionStatisticsResponse, error) {
	repoFilters := &repository.InfractionStatsFilters{
		DateDebut: input.DateDebut,
		DateFin:   input.DateFin,
	}

	stats, err := s.infractionRepo.GetStatistics(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	response := &InfractionStatisticsResponse{
		Total:              stats.Total,
		ParStatut:          stats.ParStatut,
		ParType:            stats.ParType,
		ParMois:            stats.ParMois,
		MontantTotal:       stats.MontantTotal,
		PointsTotal:        stats.PointsTotal,
		FlagrantDelitTotal: stats.FlagrantDelitTotal,
		AccidentTotal:      stats.AccidentTotal,
		TopInfractions:     make([]TypeInfractionStats, len(stats.TopInfractions)),
		PeriodeDebut:       input.DateDebut,
		PeriodeFin:         input.DateFin,
	}

	for i, top := range stats.TopInfractions {
		response.TopInfractions[i] = TypeInfractionStats{
			TypeCode:     top.TypeCode,
			TypeLibelle:  top.TypeLibelle,
			Count:        top.Count,
			MontantTotal: top.MontantTotal,
		}
	}

	return response, nil
}

// GeneratePV generates a PV for an infraction
func (s *service) GeneratePV(ctx context.Context, infractionID string) (*PVGenerationResponse, error) {
	// Récupérer l'infraction
	infraction, err := s.GetByID(ctx, infractionID)
	if err != nil {
		return nil, fmt.Errorf("infraction not found: %w", err)
	}

	// Vérifier qu'un PV n'existe pas déjà
	if infraction.ProcesVerbal != nil {
		return &PVGenerationResponse{
			ProcesVerbalID: infraction.ProcesVerbal.ID,
			NumeroPV:       infraction.ProcesVerbal.NumeroPV,
			Success:        false,
			Message:        "PV already exists for this infraction",
		}, nil
	}

	// Générer numéro PV unique
	numeroPV := s.generateNumeroPV()

	// Calculer la date limite de paiement (45 jours)
	dateLimite := time.Now().AddDate(0, 0, 45)

	// Créer le ProcesVerbal dans la base de données
	pvInput := &repository.CreatePVInput{
		ID:                 uuid.New().String(),
		NumeroPV:           numeroPV,
		DateEmission:       time.Now(),
		MontantTotal:       infraction.MontantAmende,
		DateLimitePaiement: &dateLimite,
		Statut:             "EMIS",
		InfractionIDs:      []string{infractionID},
	}

	pvEnt, err := s.pvRepo.Create(ctx, pvInput)
	if err != nil {
		s.logger.Error("Failed to create PV", zap.Error(err))
		return nil, fmt.Errorf("failed to create PV: %w", err)
	}

	// Mettre à jour l'infraction avec le numéro PV et changer le statut
	updateInput := &UpdateInfractionRequest{
		NumeroPV: &numeroPV,
		Statut:   stringPtr("VALIDEE"),
	}

	_, err = s.Update(ctx, infractionID, updateInput)
	if err != nil {
		s.logger.Error("Failed to update infraction after PV creation", zap.Error(err))
		// Le PV a été créé, on ne retourne pas d'erreur mais on log
	}

	return &PVGenerationResponse{
		ProcesVerbalID: pvEnt.ID.String(),
		NumeroPV:       pvEnt.NumeroPv,
		DateEmission:   pvEnt.DateEmission,
		MontantTotal:   pvEnt.MontantTotal,
		Success:        true,
		Message:        "PV generated successfully",
	}, nil
}

// ValidateInfraction validates an infraction
func (s *service) ValidateInfraction(ctx context.Context, infractionID string) (*InfractionValidationResponse, error) {
	// Récupérer l'infraction
	infraction, err := s.GetByID(ctx, infractionID)
	if err != nil {
		return nil, fmt.Errorf("infraction not found: %w", err)
	}

	if infraction.Statut == "VALIDEE" {
		return &InfractionValidationResponse{
			InfractionID:    infractionID,
			NumeroPV:        infraction.NumeroPV,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DateValidation:  time.Now(),
			Success:         false,
			Message:         "Infraction already validated",
		}, nil
	}

	// Valider l'infraction
	updateInput := &UpdateInfractionRequest{
		Statut: stringPtr("VALIDEE"),
	}

	_, err = s.Update(ctx, infractionID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to validate infraction: %w", err)
	}

	return &InfractionValidationResponse{
		InfractionID:    infractionID,
		NumeroPV:        infraction.NumeroPV,
		StatutPrecedent: infraction.Statut,
		NouveauStatut:   "VALIDEE",
		DateValidation:  time.Now(),
		Success:         true,
		Message:         "Infraction validated successfully",
	}, nil
}

// ArchiveInfraction archives an infraction (only PAYEE or ANNULEE can be archived)
func (s *service) ArchiveInfraction(ctx context.Context, infractionID string) (*InfractionArchiveResponse, error) {
	// Récupérer l'infraction
	infraction, err := s.GetByID(ctx, infractionID)
	if err != nil {
		return nil, fmt.Errorf("infraction not found: %w", err)
	}

	// Vérifier que l'infraction peut être archivée (PAYEE ou ANNULEE uniquement)
	if infraction.Statut != "PAYEE" && infraction.Statut != "ANNULEE" {
		return &InfractionArchiveResponse{
			InfractionID:    infractionID,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DateArchivage:   time.Now(),
			Success:         false,
			Message:         "Only PAYEE or ANNULEE infractions can be archived",
		}, nil
	}

	if infraction.Statut == "ARCHIVEE" {
		return &InfractionArchiveResponse{
			InfractionID:    infractionID,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DateArchivage:   time.Now(),
			Success:         false,
			Message:         "Infraction already archived",
		}, nil
	}

	// Archiver l'infraction
	updateInput := &UpdateInfractionRequest{
		Statut: stringPtr("ARCHIVEE"),
	}

	_, err = s.Update(ctx, infractionID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to archive infraction: %w", err)
	}

	return &InfractionArchiveResponse{
		InfractionID:    infractionID,
		StatutPrecedent: infraction.Statut,
		NouveauStatut:   "ARCHIVEE",
		DateArchivage:   time.Now(),
		Success:         true,
		Message:         "Infraction archived successfully",
	}, nil
}

// UnarchiveInfraction unarchives an infraction (changes status back to PAYEE)
func (s *service) UnarchiveInfraction(ctx context.Context, infractionID string) (*InfractionArchiveResponse, error) {
	// Récupérer l'infraction
	infraction, err := s.GetByID(ctx, infractionID)
	if err != nil {
		return nil, fmt.Errorf("infraction not found: %w", err)
	}

	// Vérifier que l'infraction est archivée
	if infraction.Statut != "ARCHIVEE" {
		return &InfractionArchiveResponse{
			InfractionID:    infractionID,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DateArchivage:   time.Now(),
			Success:         false,
			Message:         "Only ARCHIVEE infractions can be unarchived",
		}, nil
	}

	// Désarchiver l'infraction (revenir à PAYEE par défaut)
	updateInput := &UpdateInfractionRequest{
		Statut: stringPtr("PAYEE"),
	}

	_, err = s.Update(ctx, infractionID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to unarchive infraction: %w", err)
	}

	return &InfractionArchiveResponse{
		InfractionID:    infractionID,
		StatutPrecedent: "ARCHIVEE",
		NouveauStatut:   "PAYEE",
		DateArchivage:   time.Now(),
		Success:         true,
		Message:         "Infraction unarchived successfully",
	}, nil
}

// RecordPayment records a payment for an infraction
func (s *service) RecordPayment(ctx context.Context, infractionID string, input *PaymentRequest) (*PaymentResponse, error) {
	// Récupérer l'infraction
	infraction, err := s.GetByID(ctx, infractionID)
	if err != nil {
		return nil, fmt.Errorf("infraction not found: %w", err)
	}

	// Vérifier que l'infraction peut être payée
	if infraction.Statut == "PAYEE" || infraction.Statut == "ARCHIVEE" {
		return &PaymentResponse{
			InfractionID:    infractionID,
			NumeroPV:        infraction.NumeroPV,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DatePaiement:    time.Now(),
			Success:         false,
			Message:         "Cette infraction est déjà payée ou archivée",
		}, nil
	}

	if infraction.Statut == "ANNULEE" || infraction.Statut == "CONTESTEE" {
		return &PaymentResponse{
			InfractionID:    infractionID,
			NumeroPV:        infraction.NumeroPV,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DatePaiement:    time.Now(),
			Success:         false,
			Message:         "Cette infraction ne peut pas être payée (annulée ou contestée)",
		}, nil
	}

	// Vérifier le montant
	if input.Montant < infraction.MontantAmende {
		return &PaymentResponse{
			InfractionID:    infractionID,
			NumeroPV:        infraction.NumeroPV,
			StatutPrecedent: infraction.Statut,
			NouveauStatut:   infraction.Statut,
			DatePaiement:    time.Now(),
			Success:         false,
			Message:         fmt.Sprintf("Le montant payé (%.0f) est inférieur au montant de l'amende (%.0f)", input.Montant, infraction.MontantAmende),
		}, nil
	}

	// Mettre à jour le statut à PAYEE
	updateInput := &UpdateInfractionRequest{
		Statut: stringPtr("PAYEE"),
	}

	_, err = s.Update(ctx, infractionID, updateInput)
	if err != nil {
		s.logger.Error("Failed to update infraction payment status", zap.Error(err))
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	s.logger.Info("Payment recorded successfully",
		zap.String("infraction_id", infractionID),
		zap.String("mode_paiement", input.ModePaiement),
		zap.Float64("montant", input.Montant),
	)

	return &PaymentResponse{
		InfractionID:    infractionID,
		NumeroPV:        infraction.NumeroPV,
		MontantPaye:     input.Montant,
		ModePaiement:    input.ModePaiement,
		Reference:       input.Reference,
		StatutPrecedent: infraction.Statut,
		NouveauStatut:   "PAYEE",
		DatePaiement:    time.Now(),
		Success:         true,
		Message:         "Paiement enregistré avec succès",
	}, nil
}

// GroupByType groups infractions by type
func (s *service) GroupByType(ctx context.Context, input *ListInfractionsRequest) ([]*InfractionsByTypeResponse, error) {
	// Récupérer toutes les infractions
	infractions, err := s.List(ctx, input)
	if err != nil {
		return nil, err
	}

	// Grouper par type
	typeMap := make(map[string]*InfractionsByTypeResponse)

	for _, inf := range infractions.Infractions {
		if inf.TypeInfraction == nil {
			continue
		}

		typeCode := inf.TypeInfraction.Code
		if _, exists := typeMap[typeCode]; !exists {
			typeMap[typeCode] = &InfractionsByTypeResponse{
				TypeInfraction: inf.TypeInfraction,
				Infractions:    []*InfractionResponse{},
				Count:          0,
				MontantTotal:   0,
				PointsTotal:    0,
			}
		}

		group := typeMap[typeCode]
		group.Infractions = append(group.Infractions, inf)
		group.Count++
		group.MontantTotal += inf.MontantAmende
		group.PointsTotal += inf.PointsRetires
	}

	// Convertir en slice
	result := make([]*InfractionsByTypeResponse, 0, len(typeMap))
	for _, group := range typeMap {
		result = append(result, group)
	}

	return result, nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreateInfractionRequest) error {
	if input.LieuInfraction == "" {
		return fmt.Errorf("lieu_infraction is required")
	}
	if input.ControleID == "" {
		return fmt.Errorf("controle_id is required")
	}
	if input.TypeInfractionID == "" {
		return fmt.Errorf("type_infraction_id is required")
	}
	if input.VehiculeID == "" {
		return fmt.Errorf("vehicule_id is required")
	}
	if input.ConducteurID == "" {
		return fmt.Errorf("conducteur_id is required")
	}
	if input.DateInfraction.IsZero() {
		return fmt.Errorf("date_infraction is required")
	}

	// Validation excès de vitesse
	if input.VitesseRetenue != nil && input.VitesseLimitee != nil {
		if *input.VitesseRetenue <= *input.VitesseLimitee {
			return fmt.Errorf("vitesse_retenue must be greater than vitesse_limitee")
		}
	}

	return nil
}

func (s *service) validateUpdateInput(input *UpdateInfractionRequest) error {
	if input.LieuInfraction != nil && *input.LieuInfraction == "" {
		return fmt.Errorf("lieu_infraction cannot be empty")
	}

	// Validation excès de vitesse
	if input.VitesseRetenue != nil && input.VitesseLimitee != nil {
		if *input.VitesseRetenue <= *input.VitesseLimitee {
			return fmt.Errorf("vitesse_retenue must be greater than vitesse_limitee")
		}
	}

	return nil
}

func (s *service) validateAndGetTypeInfraction(ctx context.Context, input *CreateInfractionRequest) (*ent.InfractionType, error) {
	// Récupérer le type d'infraction depuis la base de données
	typeInfraction, err := s.infractionTypeRepo.GetByID(ctx, input.TypeInfractionID)
	if err != nil {
		s.logger.Error("Failed to get infraction type",
			zap.String("type_id", input.TypeInfractionID),
			zap.Error(err))
		return nil, fmt.Errorf("invalid infraction type: %w", err)
	}

	// Vérifier que le type est actif
	if !typeInfraction.Active {
		return nil, fmt.Errorf("infraction type is not active")
	}

	return typeInfraction, nil
}

func (s *service) calculateSpeedingPenalty(exces, baseAmende float64, basePoints int) (float64, int) {
	// Barème français simplifié
	if exces <= 5 {
		return 68.0, 1
	} else if exces <= 10 {
		return 135.0, 1
	} else if exces <= 20 {
		return 135.0, 2
	} else if exces <= 30 {
		return 135.0, 3
	} else if exces <= 40 {
		return 135.0, 4
	} else if exces <= 50 {
		return 1500.0, 6
	} else {
		return 1500.0, 6 // + suspension possible
	}
}

func (s *service) generateNumeroPV() string {
	// Générer un numéro PV unique (format simplifié)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("PV%d", timestamp)
}

func (s *service) buildRepositoryFilters(input *ListInfractionsRequest) *repository.InfractionFilters {
	if input == nil {
		return nil
	}

	return &repository.InfractionFilters{
		ControleID:       input.ControleID,
		VehiculeID:       input.VehiculeID,
		ConducteurID:     input.ConducteurID,
		TypeInfractionID: input.TypeInfractionID,
		Statut:           input.Statut,
		LieuInfraction:   input.LieuInfraction,
		DateDebut:        input.DateDebut,
		DateFin:          input.DateFin,
		FlagrantDelit:    input.FlagrantDelit,
		Accident:         input.Accident,
		Limit:            input.Limit,
		Offset:           input.Offset,
	}
}

func (s *service) entityToResponse(infractionEnt *ent.Infraction) *InfractionResponse {
	response := &InfractionResponse{
		ID:             infractionEnt.ID.String(),
		NumeroPV:       infractionEnt.NumeroPv,
		DateInfraction: infractionEnt.DateInfraction,
		LieuInfraction: infractionEnt.LieuInfraction,
		Circonstances:  infractionEnt.Circonstances,
		AppareilMesure: infractionEnt.AppareilMesure,
		MontantAmende:  infractionEnt.MontantAmende,
		PointsRetires:  infractionEnt.PointsRetires,
		Statut:         infractionEnt.Statut,
		Observations:   infractionEnt.Observations,
		FlagrantDelit:  infractionEnt.FlagrantDelit,
		Accident:       infractionEnt.Accident,
		CreatedAt:      infractionEnt.CreatedAt,
		UpdatedAt:      infractionEnt.UpdatedAt,
	}

	// Gérer les champs de vitesse optionnels
	if infractionEnt.VitesseRetenue != 0 {
		response.VitesseRetenue = &infractionEnt.VitesseRetenue
	}
	if infractionEnt.VitesseLimitee != 0 {
		response.VitesseLimitee = &infractionEnt.VitesseLimitee
	}

	// Ajouter les relations si chargées
	if infractionEnt.Edges.Controle != nil {
		agentNom := ""
		if infractionEnt.Edges.Controle.Edges.Agent != nil {
			agentNom = fmt.Sprintf("%s %s",
				infractionEnt.Edges.Controle.Edges.Agent.Prenom,
				infractionEnt.Edges.Controle.Edges.Agent.Nom)
		}

		response.Controle = &ControleSummary{
			ID:           infractionEnt.Edges.Controle.ID.String(),
			DateControle: infractionEnt.Edges.Controle.DateControle,
			LieuControle: infractionEnt.Edges.Controle.LieuControle,
			TypeControle: string(infractionEnt.Edges.Controle.TypeControle),
			AgentNom:     agentNom,
		}
	}

	if infractionEnt.Edges.TypeInfraction != nil {
		response.TypeInfraction = &TypeInfractionSummary{
			ID:          infractionEnt.Edges.TypeInfraction.ID.String(),
			Code:        infractionEnt.Edges.TypeInfraction.Code,
			Libelle:     infractionEnt.Edges.TypeInfraction.Libelle,
			Description: infractionEnt.Edges.TypeInfraction.Description,
			Amende:      infractionEnt.Edges.TypeInfraction.Amende,
			Points:      infractionEnt.Edges.TypeInfraction.Points,
			Categorie:   infractionEnt.Edges.TypeInfraction.Categorie,
		}
	}

	if infractionEnt.Edges.Vehicule != nil {
		response.Vehicule = &VehiculeSummary{
			ID:              infractionEnt.Edges.Vehicule.ID.String(),
			Immatriculation: infractionEnt.Edges.Vehicule.Immatriculation,
			Marque:          infractionEnt.Edges.Vehicule.Marque,
			Modele:          infractionEnt.Edges.Vehicule.Modele,
			TypeVehicule:    infractionEnt.Edges.Vehicule.TypeVehicule,
		}
	}

	if infractionEnt.Edges.Conducteur != nil {
		response.Conducteur = &ConducteurSummary{
			ID:           infractionEnt.Edges.Conducteur.ID.String(),
			Nom:          infractionEnt.Edges.Conducteur.Nom,
			Prenom:       infractionEnt.Edges.Conducteur.Prenom,
			NumeroPermis: infractionEnt.Edges.Conducteur.NumeroPermis,
			PointsPermis: infractionEnt.Edges.Conducteur.PointsPermis,
		}
	}

	if infractionEnt.Edges.ProcesVerbal != nil {
		response.ProcesVerbal = &ProcesVerbalSummary{
			ID:           infractionEnt.Edges.ProcesVerbal.ID.String(),
			NumeroPV:     infractionEnt.Edges.ProcesVerbal.NumeroPv,
			DateEmission: infractionEnt.Edges.ProcesVerbal.DateEmission,
			MontantTotal: infractionEnt.Edges.ProcesVerbal.MontantTotal,
			Statut:       infractionEnt.Edges.ProcesVerbal.Statut,
		}
	}

	return response
}

// GetTypesInfractions returns all infraction types from the database
func (s *service) GetTypesInfractions(ctx context.Context) ([]*TypeInfractionSummary, error) {
	// Récupérer les types actifs depuis la base de données
	typesEnt, err := s.infractionTypeRepo.GetActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get infraction types from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get infraction types: %w", err)
	}

	// Convertir en réponses
	types := make([]*TypeInfractionSummary, len(typesEnt))
	for i, t := range typesEnt {
		types[i] = &TypeInfractionSummary{
			ID:          t.ID.String(),
			Code:        t.Code,
			Libelle:     t.Libelle,
			Description: t.Description,
			Amende:      t.Amende,
			Points:      t.Points,
			Categorie:   t.Categorie,
		}
	}

	return types, nil
}

// GetCategories returns all infraction categories with their type counts
func (s *service) GetCategories(ctx context.Context) ([]*CategorieResponse, error) {
	// Récupérer les types actifs pour compter par catégorie
	typesEnt, err := s.infractionTypeRepo.GetActive(ctx)
	if err != nil {
		s.logger.Error("Failed to get infraction types for categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Grouper par catégorie et compter
	categoryMap := make(map[string]*CategorieResponse)
	for _, t := range typesEnt {
		if _, exists := categoryMap[t.Categorie]; !exists {
			categoryMap[t.Categorie] = &CategorieResponse{
				Code:    t.Categorie,
				Libelle: getCategorieLibelle(t.Categorie),
				NbTypes: 0,
			}
		}
		categoryMap[t.Categorie].NbTypes++
	}

	// Convertir en slice
	categories := make([]*CategorieResponse, 0, len(categoryMap))
	for _, cat := range categoryMap {
		categories = append(categories, cat)
	}

	return categories, nil
}

// getCategorieLibelle returns a human-readable label for a category code
func getCategorieLibelle(code string) string {
	labels := map[string]string{
		"DOCUMENTS":     "Documents",
		"VITESSE":       "Excès de vitesse",
		"SECURITE":      "Sécurité routière",
		"STATIONNEMENT": "Stationnement",
		"COMPORTEMENT":  "Comportement au volant",
		"VEHICULE":      "État du véhicule",
		"ALCOOL":        "Alcool et stupéfiants",
		"CIRCULATION":   "Règles de circulation",
	}
	if label, ok := labels[code]; ok {
		return label
	}
	return code
}

// GetDashboard returns dashboard data for the frontend
func (s *service) GetDashboard(ctx context.Context, input *DashboardRequest) (*DashboardResponse, error) {
	// Calculer les dates selon la période
	var dateDebut, dateFin time.Time
	now := time.Now()

	if input.DateDebut != nil && input.DateFin != nil {
		dateDebut = *input.DateDebut
		dateFin = *input.DateFin
	} else {
		switch input.Periode {
		case "jour":
			dateDebut = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			dateFin = now
		case "semaine":
			dateDebut = now.AddDate(0, 0, -7)
			dateFin = now
		case "mois":
			dateDebut = now.AddDate(0, -1, 0)
			dateFin = now
		case "annee":
			dateDebut = now.AddDate(-1, 0, 0)
			dateFin = now
		case "tout":
			dateDebut = time.Date(2020, 1, 1, 0, 0, 0, 0, now.Location())
			dateFin = now
		default:
			dateDebut = now.AddDate(0, -1, 0)
			dateFin = now
		}
	}

	// Récupérer toutes les infractions pour la période
	listRequest := &ListInfractionsRequest{
		DateDebut: &dateDebut,
		DateFin:   &dateFin,
		Limit:     10000,
	}

	infractionsResp, err := s.List(ctx, listRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get infractions: %w", err)
	}

	infractions := infractionsResp.Infractions

	// Requête séparée pour les infractions des dernières 24h (indépendant du filtre)
	yesterday := now.AddDate(0, 0, -1)
	last24hRequest := &ListInfractionsRequest{
		DateDebut: &yesterday,
		DateFin:   &now,
		Limit:     10000,
	}
	last24hResp, err := s.List(ctx, last24hRequest)
	infractions24h := 0
	if err == nil && last24hResp != nil {
		infractions24h = last24hResp.Total
	}

	// Calculer les stats
	response := &DashboardResponse{}

	// Stats de base
	totalMontant := 0.0
	contestees := 0
	payees := 0
	typesMap := make(map[string]bool)

	// Compteurs par catégorie
	categoryCountMap := make(map[string]int)
	typeCountMap := make(map[string]int)
	typeNameMap := make(map[string]string)
	typeCategoryMap := make(map[string]string)

	// Données d'activité par période
	activityMap := make(map[string]*ActivityDataEntry)

	for _, inf := range infractions {
		totalMontant += inf.MontantAmende

		if inf.Statut == "CONTESTEE" {
			contestees++
		}
		if inf.Statut == "PAYEE" || inf.Statut == "ARCHIVEE" {
			payees++
		}

		if inf.TypeInfraction != nil {
			typesMap[inf.TypeInfraction.Code] = true
			typeCountMap[inf.TypeInfraction.Code]++
			typeNameMap[inf.TypeInfraction.Code] = inf.TypeInfraction.Libelle

			category := s.getCategoryForType(inf.TypeInfraction.Categorie)
			categoryCountMap[category]++
			typeCategoryMap[inf.TypeInfraction.Code] = category
		}

		// Grouper par période pour le graphique
		periodKey := s.getPeriodKey(inf.DateInfraction, input.Periode)
		if _, exists := activityMap[periodKey]; !exists {
			activityMap[periodKey] = &ActivityDataEntry{Period: periodKey}
		}
		entry := activityMap[periodKey]
		entry.Total++

		if inf.TypeInfraction != nil {
			category := s.getCategoryForType(inf.TypeInfraction.Categorie)
			switch category {
			case "Documents":
				entry.Documents++
			case "Sécurité":
				entry.Securite++
			case "Comportement":
				entry.Comportement++
			case "État technique":
				entry.Technique++
			}
		}
	}

	totalInf := len(infractions)

	// Formater les montants
	revenus := s.formatMontant(totalMontant)
	montantMoyen := "0"
	if totalInf > 0 {
		montantMoyen = fmt.Sprintf("%.0f", totalMontant/float64(totalInf))
	}

	// Calculer les taux
	tauxContestation := "0%"
	tauxPaiement := "0%"
	if totalInf > 0 {
		tauxContestation = fmt.Sprintf("%.1f%%", float64(contestees)/float64(totalInf)*100)
		tauxPaiement = fmt.Sprintf("%.1f%%", float64(payees)/float64(totalInf)*100)
	}

	response.Stats = DashboardStats{
		TotalInfractions: totalInf,
		Revenus:          revenus,
		TotalTypes:       len(typesMap),
		MontantMoyen:     montantMoyen,
		TauxContestation: tauxContestation,
		TauxPaiement:     tauxPaiement,
		Infractions24h:   infractions24h,
		Evolution:        5.2, // Placeholder pour l'évolution
	}

	// Données du graphique circulaire (par catégorie)
	categoryColors := map[string]string{
		"Documents":      "#3b82f6",
		"Sécurité":       "#ef4444",
		"Comportement":   "#eab308",
		"État technique": "#a855f7",
		"Chargement":     "#f97316",
		"Environnement":  "#22c55e",
		"Autres":         "#6b7280",
	}

	for cat, count := range categoryCountMap {
		color := categoryColors[cat]
		if color == "" {
			color = "#6b7280"
		}
		response.PieData = append(response.PieData, PieDataEntry{
			Name:  cat,
			Value: count,
			Color: color,
		})
	}

	// Données des catégories
	categoryBgColors := map[string]string{
		"Documents":      "bg-blue-50",
		"Sécurité":       "bg-red-50",
		"Comportement":   "bg-yellow-50",
		"État technique": "bg-purple-50",
		"Chargement":     "bg-orange-50",
		"Environnement":  "bg-green-50",
		"Autres":         "bg-gray-50",
	}
	categoryIconColors := map[string]string{
		"Documents":      "text-blue-600",
		"Sécurité":       "text-red-600",
		"Comportement":   "text-yellow-600",
		"État technique": "text-purple-600",
		"Chargement":     "text-orange-600",
		"Environnement":  "text-green-600",
		"Autres":         "text-gray-600",
	}

	for cat, count := range categoryCountMap {
		response.Categories = append(response.Categories, CategoryDataEntry{
			ID:          s.slugify(cat),
			Title:       cat,
			Count:       count,
			BgColor:     categoryBgColors[cat],
			IconColor:   categoryIconColors[cat],
			Infractions: []string{},
			Evolution:   0,
		})
	}

	// Top infractions
	type typeCount struct {
		code  string
		count int
	}
	var topTypes []typeCount
	for code, count := range typeCountMap {
		topTypes = append(topTypes, typeCount{code, count})
	}
	// Trier par count décroissant
	for i := 0; i < len(topTypes); i++ {
		for j := i + 1; j < len(topTypes); j++ {
			if topTypes[j].count > topTypes[i].count {
				topTypes[i], topTypes[j] = topTypes[j], topTypes[i]
			}
		}
	}
	// Prendre les 5 premiers
	for i := 0; i < len(topTypes) && i < 5; i++ {
		tc := topTypes[i]
		percentage := 0.0
		if totalInf > 0 {
			percentage = float64(tc.count) / float64(totalInf) * 100
		}
		response.TopInfractions = append(response.TopInfractions, TopInfractionEntry{
			Name:       typeNameMap[tc.code],
			Count:      tc.count,
			Percentage: percentage,
			Category:   typeCategoryMap[tc.code],
		})
	}

	// Convertir activityMap en slice triée
	for _, entry := range activityMap {
		response.ActivityData = append(response.ActivityData, *entry)
	}
	// Trier par période (ordre chronologique)
	// Simple sort par string pour l'instant
	for i := 0; i < len(response.ActivityData); i++ {
		for j := i + 1; j < len(response.ActivityData); j++ {
			if response.ActivityData[j].Period < response.ActivityData[i].Period {
				response.ActivityData[i], response.ActivityData[j] = response.ActivityData[j], response.ActivityData[i]
			}
		}
	}

	// Limiter à 10 périodes
	if len(response.ActivityData) > 10 {
		response.ActivityData = response.ActivityData[len(response.ActivityData)-10:]
	}

	// Données d'évolution (placeholder)
	for cat := range categoryCountMap {
		response.EvolutionData = append(response.EvolutionData, EvolutionDataEntry{
			Category:  cat,
			Evolution: 0,
		})
	}

	return response, nil
}

// Helper methods pour le dashboard

func (s *service) getCategoryForType(typeCategorie string) string {
	switch typeCategorie {
	case "Documents", "DOCUMENTS":
		return "Documents"
	case "Sécurité", "SECURITE", "Securite":
		return "Sécurité"
	case "Vitesse", "VITESSE":
		return "Comportement"
	case "Signalisation", "SIGNALISATION":
		return "Comportement"
	case "VEHICULE", "Vehicule":
		return "État technique"
	case "Stationnement", "STATIONNEMENT":
		return "Chargement"
	case "CHARGEMENT", "Chargement":
		return "Chargement"
	case "ENVIRONNEMENT", "Environnement":
		return "Environnement"
	default:
		return "Autres"
	}
}

func (s *service) getPeriodKey(date time.Time, periode string) string {
	switch periode {
	case "jour":
		return fmt.Sprintf("%02dh", date.Hour())
	case "semaine":
		return date.Weekday().String()[:3]
	case "mois":
		return fmt.Sprintf("%02d/%02d", date.Day(), date.Month())
	case "annee":
		return date.Month().String()[:3]
	default:
		return fmt.Sprintf("%02d/%02d", date.Day(), date.Month())
	}
}

func (s *service) formatMontant(montant float64) string {
	if montant >= 1000000 {
		return fmt.Sprintf("%.1fM", montant/1000000)
	}
	if montant >= 1000 {
		return fmt.Sprintf("%.0fK", montant/1000)
	}
	return fmt.Sprintf("%.0f", montant)
}

func (s *service) slugify(str string) string {
	result := ""
	for _, c := range str {
		if c == ' ' {
			result += "-"
		} else if c >= 'a' && c <= 'z' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	return result
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}