package controle

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines controle service interface
type Service interface {
	Create(ctx context.Context, input *CreateControleRequest) (*ControleResponse, error)
	GetByID(ctx context.Context, id string) (*ControleResponse, error)
	List(ctx context.Context, filters *ListControlesRequest) (*ListControlesResponse, error)
	Update(ctx context.Context, id string, input *UpdateControleRequest) (*ControleResponse, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string, filters *ListControlesRequest) (*ListControlesResponse, error)
	GetByVehicule(ctx context.Context, vehiculeID string) (*ListControlesResponse, error)
	GetByConducteur(ctx context.Context, conducteurID string) (*ListControlesResponse, error)
	GetByDateRange(ctx context.Context, start, end time.Time) (*ListControlesResponse, error)
	GetStatistics(ctx context.Context, filters *StatisticsFilters) (*ControleStatisticsResponse, error)
	ChangerStatut(ctx context.Context, controleID string, input *ChangerStatutRequest) (*ControleResponse, error)
	GeneratePV(ctx context.Context, controleID string, input *GeneratePVRequest) (*GeneratePVResponse, error)
	Archive(ctx context.Context, id string) (*ControleResponse, error)
	Unarchive(ctx context.Context, id string) (*ControleResponse, error)
}

// service implements Service interface
type service struct {
	controleRepo     repository.ControleRepository
	infractionRepo   repository.InfractionRepository
	pvRepo           repository.PVRepository
	verificationRepo repository.VerificationRepository
	logger           *zap.Logger
}

// NewService creates a new controle service
func NewService(
	controleRepo repository.ControleRepository,
	infractionRepo repository.InfractionRepository,
	pvRepo repository.PVRepository,
	verificationRepo repository.VerificationRepository,
	logger *zap.Logger,
) Service {
	return &service{
		controleRepo:     controleRepo,
		infractionRepo:   infractionRepo,
		pvRepo:           pvRepo,
		verificationRepo: verificationRepo,
		logger:           logger,
	}
}

// Create creates a new controle
func (s *service) Create(ctx context.Context, input *CreateControleRequest) (*ControleResponse, error) {
	// Validation
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Generate reference
	reference := fmt.Sprintf("CTRL-%d-%s", time.Now().Year(), uuid.New().String()[:8])

	// Default values
	typeControle := input.TypeControle
	if typeControle == "" {
		typeControle = "GENERAL"
	}
	statut := input.Statut
	if statut == "" {
		statut = "EN_COURS"
	}
	vehiculeType := input.VehiculeType
	if vehiculeType == "" {
		vehiculeType = "VOITURE"
	}

	repoInput := &repository.CreateControleInput{
		ID:           uuid.New().String(),
		Reference:    reference,
		DateControle: input.DateControle,
		LieuControle: input.LieuControle,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		TypeControle: typeControle,
		Statut:       statut,
		Observations: input.Observations,
		AgentID:      input.AgentID,
		CommissariatID: input.CommissariatID,
		VehiculeID:     input.VehiculeID,
		ConducteurID:   input.ConducteurID,
		// Données véhicule embarquées
		VehiculeImmatriculation: input.VehiculeImmatriculation,
		VehiculeMarque:          input.VehiculeMarque,
		VehiculeModele:          input.VehiculeModele,
		VehiculeType:            vehiculeType,
		VehiculeAnnee:           input.VehiculeAnnee,
		VehiculeCouleur:         input.VehiculeCouleur,
		VehiculeNumeroChassis:   input.VehiculeNumeroChassis,
		// Données conducteur embarquées
		ConducteurNumeroPermis: input.ConducteurNumeroPermis,
		ConducteurNom:          input.ConducteurNom,
		ConducteurPrenom:       input.ConducteurPrenom,
		ConducteurTelephone:    input.ConducteurTelephone,
		ConducteurAdresse:      input.ConducteurAdresse,
	}

	controleEnt, err := s.controleRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create controle", zap.Error(err))
		return nil, fmt.Errorf("failed to create controle: %w", err)
	}

	// Process initial options (vérifications) if provided - atomic creation
	if len(input.InitialOptions) > 0 {
		s.logger.Info("Creating initial check options for controle",
			zap.String("controle_id", controleEnt.ID.String()),
			zap.Int("nb_options", len(input.InitialOptions)))

		for _, opt := range input.InitialOptions {
			verificationInput := &repository.CreateVerificationInput{
				SourceType:    "CONTROL",
				SourceID:      controleEnt.ID.String(),
				CheckItemID:   opt.CheckItemID,
				ResultStatus:  opt.Resultat,
				Notes:         opt.Notes,
				MontantAmende: opt.MontantAmende,
			}
			_, err := s.verificationRepo.CreateVerificationFromInput(ctx, verificationInput)
			if err != nil {
				s.logger.Warn("Failed to create check option", zap.Error(err), zap.String("check_item_id", opt.CheckItemID))
				// Continue with other options, don't fail the whole creation
			}
		}

		// Update counters on the controle if provided
		if input.TotalVerifications != nil || input.VerificationsOk != nil ||
			input.VerificationsEchec != nil || input.MontantTotalAmendes != nil {
			updateInput := &repository.UpdateControleInput{
				TotalVerifications:  input.TotalVerifications,
				VerificationsOk:     input.VerificationsOk,
				VerificationsEchec:  input.VerificationsEchec,
				MontantTotalAmendes: input.MontantTotalAmendes,
			}
			controleEnt, _ = s.controleRepo.Update(ctx, controleEnt.ID.String(), updateInput)
		}
	}

	// Reload with edges
	controleEnt, err = s.controleRepo.GetByID(ctx, controleEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(controleEnt), nil
}

// GetByID gets controle by ID
func (s *service) GetByID(ctx context.Context, id string) (*ControleResponse, error) {
	controleEnt, err := s.controleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use context-aware method to load CheckOptions from database
	return s.entityToResponseWithContext(ctx, controleEnt), nil
}

// List gets controles with filters
func (s *service) List(ctx context.Context, input *ListControlesRequest) (*ListControlesResponse, error) {
	filters := &repository.ControleFilters{
		AgentID:                 input.AgentID,
		VehiculeID:              input.VehiculeID,
		ConducteurID:            input.ConducteurID,
		CommissariatID:          input.CommissariatID,
		TypeControle:            input.TypeControle,
		Statut:                  input.Statut,
		LieuControle:            input.LieuControle,
		VehiculeImmatriculation: input.VehiculeImmatriculation,
		DateDebut:               input.DateDebut,
		DateFin:                 input.DateFin,
		IsArchived:              input.IsArchived,
		Limit:                   input.Limit,
		Offset:                  input.Offset,
	}

	controlesEnt, err := s.controleRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	total, err := s.controleRepo.Count(ctx, filters)
	if err != nil {
		total = len(controlesEnt)
	}

	controles := make([]*ControleResponse, len(controlesEnt))
	for i, c := range controlesEnt {
		controles[i] = s.entityToResponse(c)
	}

	return &ListControlesResponse{
		Controles: controles,
		Total:     total,
	}, nil
}

// Update updates controle
func (s *service) Update(ctx context.Context, id string, input *UpdateControleRequest) (*ControleResponse, error) {
	repoInput := &repository.UpdateControleInput{
		DateControle:        input.DateControle,
		LieuControle:        input.LieuControle,
		Latitude:            input.Latitude,
		Longitude:           input.Longitude,
		TypeControle:        input.TypeControle,
		Statut:              input.Statut,
		Observations:        input.Observations,
		TotalVerifications:  input.TotalVerifications,
		VerificationsOk:     input.VerificationsOk,
		VerificationsEchec:  input.VerificationsEchec,
		MontantTotalAmendes: input.MontantTotalAmendes,
	}

	controleEnt, err := s.controleRepo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	// Reload with edges
	controleEnt, err = s.controleRepo.GetByID(ctx, controleEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(controleEnt), nil
}

// Delete deletes controle
func (s *service) Delete(ctx context.Context, id string) error {
	return s.controleRepo.Delete(ctx, id)
}

// GetByAgent gets controles by agent
func (s *service) GetByAgent(ctx context.Context, agentID string, input *ListControlesRequest) (*ListControlesResponse, error) {
	filters := s.buildFilters(input)
	controlesEnt, err := s.controleRepo.GetByAgent(ctx, agentID, filters)
	if err != nil {
		return nil, err
	}

	controles := make([]*ControleResponse, len(controlesEnt))
	for i, c := range controlesEnt {
		controles[i] = s.entityToResponse(c)
	}

	return &ListControlesResponse{
		Controles: controles,
		Total:     len(controles),
	}, nil
}

// GetByVehicule gets controles by vehicule
func (s *service) GetByVehicule(ctx context.Context, vehiculeID string) (*ListControlesResponse, error) {
	controlesEnt, err := s.controleRepo.GetByVehicule(ctx, vehiculeID)
	if err != nil {
		return nil, err
	}

	controles := make([]*ControleResponse, len(controlesEnt))
	for i, c := range controlesEnt {
		controles[i] = s.entityToResponse(c)
	}

	return &ListControlesResponse{
		Controles: controles,
		Total:     len(controles),
	}, nil
}

// GetByConducteur gets controles by conducteur
func (s *service) GetByConducteur(ctx context.Context, conducteurID string) (*ListControlesResponse, error) {
	controlesEnt, err := s.controleRepo.GetByConducteur(ctx, conducteurID)
	if err != nil {
		return nil, err
	}

	controles := make([]*ControleResponse, len(controlesEnt))
	for i, c := range controlesEnt {
		controles[i] = s.entityToResponse(c)
	}

	return &ListControlesResponse{
		Controles: controles,
		Total:     len(controles),
	}, nil
}

// GetByDateRange gets controles by date range
func (s *service) GetByDateRange(ctx context.Context, start, end time.Time) (*ListControlesResponse, error) {
	controlesEnt, err := s.controleRepo.GetByDateRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	controles := make([]*ControleResponse, len(controlesEnt))
	for i, c := range controlesEnt {
		controles[i] = s.entityToResponse(c)
	}

	return &ListControlesResponse{
		Controles: controles,
		Total:     len(controles),
	}, nil
}

// GetStatistics gets statistics for controles
func (s *service) GetStatistics(ctx context.Context, filters *StatisticsFilters) (*ControleStatisticsResponse, error) {
	var repoFilters *repository.ControleStatsFilters
	var agentID *string
	var dateDebut, dateFin *time.Time

	if filters != nil {
		repoFilters = &repository.ControleStatsFilters{
			AgentID:   filters.AgentID,
			DateDebut: filters.DateDebut,
			DateFin:   filters.DateFin,
		}
		agentID = filters.AgentID
		dateDebut = filters.DateDebut
		dateFin = filters.DateFin
	}

	stats, err := s.controleRepo.GetStatistics(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	return &ControleStatisticsResponse{
		AgentID:             agentID,
		PeriodeDebut:        dateDebut,
		PeriodeFin:          dateFin,
		Total:               stats.Total,
		EnCours:             stats.EnCours,
		Termine:             stats.Termine,
		Conforme:            stats.Conforme,
		NonConforme:         stats.NonConforme,
		ParType:             stats.ParType,
		ParJour:             stats.ParJour,
		InfractionsAvec:     stats.InfractionsAvec,
		InfractionsSans:     stats.InfractionsSans,
		MontantTotalAmendes: stats.MontantTotalAmendes,
	}, nil
}

// ChangerStatut changes the status of a controle
func (s *service) ChangerStatut(ctx context.Context, controleID string, input *ChangerStatutRequest) (*ControleResponse, error) {
	s.logger.Info("Changing controle status",
		zap.String("controle_id", controleID),
		zap.String("new_status", input.Statut))

	updateInput := &repository.UpdateControleInput{
		Statut:              &input.Statut,
		Observations:        input.Observations,
		TotalVerifications:  input.TotalVerifications,
		VerificationsOk:     input.VerificationsOk,
		VerificationsEchec:  input.VerificationsEchec,
		MontantTotalAmendes: input.MontantTotalAmendes,
	}

	controleEnt, err := s.controleRepo.Update(ctx, controleID, updateInput)
	if err != nil {
		return nil, fmt.Errorf("failed to update controle status: %w", err)
	}

	// Reload with edges
	controleEnt, err = s.controleRepo.GetByID(ctx, controleEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(controleEnt), nil
}

// GeneratePV generates a PV from a controle with specified infractions
func (s *service) GeneratePV(ctx context.Context, controleID string, input *GeneratePVRequest) (*GeneratePVResponse, error) {
	s.logger.Info("Generating PV for controle", zap.String("controle_id", controleID), zap.Int("nb_infractions", len(input.Infractions)))

	// Verify controle exists
	controleEnt, err := s.controleRepo.GetByID(ctx, controleID)
	if err != nil {
		return nil, fmt.Errorf("controle not found: %w", err)
	}

	// Get specified infractions
	var totalMontant float64
	validInfractions := make([]string, 0)

	for _, infractionID := range input.Infractions {
		infraction, err := s.infractionRepo.GetByID(ctx, infractionID)
		if err != nil {
			s.logger.Warn("Infraction not found", zap.String("infraction_id", infractionID))
			continue
		}

		// Verify infraction belongs to this controle
		if infraction.Edges.Controle == nil || infraction.Edges.Controle.ID.String() != controleID {
			continue
		}

		// Verify infraction doesn't already have a PV
		if infraction.Edges.ProcesVerbal != nil {
			continue
		}

		totalMontant += infraction.MontantAmende
		validInfractions = append(validInfractions, infractionID)
	}

	if len(validInfractions) == 0 {
		return nil, fmt.Errorf("no valid infractions to generate PV")
	}

	// Generate unique PV number
	numeroPV := fmt.Sprintf("PV%s%06d", time.Now().Format("20060102"), time.Now().Nanosecond()/1000)

	// Calculate payment deadline (45 days by default)
	dateLimite := time.Now().AddDate(0, 0, 45)

	// Create PV avec toutes les infractions
	pvInput := &repository.CreatePVInput{
		ID:                 uuid.New().String(),
		NumeroPV:           numeroPV,
		DateEmission:       time.Now(),
		MontantTotal:       totalMontant,
		DateLimitePaiement: &dateLimite,
		Statut:             "EMIS",
		InfractionIDs:      validInfractions, // Toutes les infractions liées à ce PV
		ControleID:         &controleID,      // Lier le PV au contrôle
	}

	if controleEnt.Observations != "" {
		pvInput.Observations = &controleEnt.Observations
	}

	pvEnt, err := s.pvRepo.Create(ctx, pvInput)
	if err != nil {
		s.logger.Error("Failed to create PV", zap.Error(err))
		return nil, fmt.Errorf("failed to create PV: %w", err)
	}

	return &GeneratePVResponse{
		ID:                 pvEnt.ID.String(),
		NumeroPV:           pvEnt.NumeroPv,
		DateEmission:       pvEnt.DateEmission,
		MontantTotal:       pvEnt.MontantTotal,
		DateLimitePaiement: pvEnt.DateLimitePaiement,
		Statut:             pvEnt.Statut,
		ControleID:         controleID,
		NbInfractions:      len(validInfractions),
	}, nil
}

// Archive archives a controle
func (s *service) Archive(ctx context.Context, id string) (*ControleResponse, error) {
	s.logger.Info("Archiving controle", zap.String("id", id))

	controleEnt, err := s.controleRepo.Archive(ctx, id)
	if err != nil {
		s.logger.Error("Failed to archive controle", zap.Error(err))
		return nil, fmt.Errorf("failed to archive controle: %w", err)
	}

	// Reload with edges
	controleEnt, err = s.controleRepo.GetByID(ctx, controleEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(controleEnt), nil
}

// Unarchive unarchives a controle
func (s *service) Unarchive(ctx context.Context, id string) (*ControleResponse, error) {
	s.logger.Info("Unarchiving controle", zap.String("id", id))

	controleEnt, err := s.controleRepo.Unarchive(ctx, id)
	if err != nil {
		s.logger.Error("Failed to unarchive controle", zap.Error(err))
		return nil, fmt.Errorf("failed to unarchive controle: %w", err)
	}

	// Reload with edges
	controleEnt, err = s.controleRepo.GetByID(ctx, controleEnt.ID.String())
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(controleEnt), nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreateControleRequest) error {
	if input.LieuControle == "" {
		return fmt.Errorf("lieu_controle is required")
	}
	if input.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if input.VehiculeImmatriculation == "" {
		return fmt.Errorf("vehicule_immatriculation is required")
	}
	if input.VehiculeMarque == "" {
		return fmt.Errorf("vehicule_marque is required")
	}
	if input.VehiculeModele == "" {
		return fmt.Errorf("vehicule_modele is required")
	}
	if input.ConducteurNumeroPermis == "" {
		return fmt.Errorf("conducteur_numero_permis is required")
	}
	if input.ConducteurNom == "" {
		return fmt.Errorf("conducteur_nom is required")
	}
	if input.ConducteurPrenom == "" {
		return fmt.Errorf("conducteur_prenom is required")
	}
	if input.DateControle.IsZero() {
		return fmt.Errorf("date_controle is required")
	}

	return nil
}

func (s *service) buildFilters(input *ListControlesRequest) *repository.ControleFilters {
	if input == nil {
		return nil
	}

	return &repository.ControleFilters{
		TypeControle:            input.TypeControle,
		Statut:                  input.Statut,
		LieuControle:            input.LieuControle,
		VehiculeImmatriculation: input.VehiculeImmatriculation,
		DateDebut:               input.DateDebut,
		DateFin:                 input.DateFin,
		IsArchived:              input.IsArchived,
		Limit:                   input.Limit,
		Offset:                  input.Offset,
	}
}

func (s *service) entityToResponse(controleEnt *ent.Controle) *ControleResponse {
	response := &ControleResponse{
		ID:           controleEnt.ID.String(),
		Reference:    controleEnt.Reference,
		DateControle: controleEnt.DateControle,
		LieuControle: controleEnt.LieuControle,
		Latitude:     controleEnt.Latitude,
		Longitude:    controleEnt.Longitude,
		TypeControle: string(controleEnt.TypeControle),
		Statut:       string(controleEnt.Statut),
		Observations: controleEnt.Observations,
		// Compteurs
		TotalVerifications:  controleEnt.TotalVerifications,
		VerificationsOk:     controleEnt.VerificationsOk,
		VerificationsEchec:  controleEnt.VerificationsEchec,
		MontantTotalAmendes: controleEnt.MontantTotalAmendes,
		// Données véhicule embarquées
		VehiculeImmatriculation: controleEnt.VehiculeImmatriculation,
		VehiculeMarque:          controleEnt.VehiculeMarque,
		VehiculeModele:          controleEnt.VehiculeModele,
		VehiculeAnnee:           controleEnt.VehiculeAnnee,
		VehiculeCouleur:         controleEnt.VehiculeCouleur,
		VehiculeNumeroChassis:   controleEnt.VehiculeNumeroChassis,
		VehiculeType:            string(controleEnt.VehiculeType),
		// Données conducteur embarquées
		ConducteurNumeroPermis: controleEnt.ConducteurNumeroPermis,
		ConducteurNom:          controleEnt.ConducteurNom,
		ConducteurPrenom:       controleEnt.ConducteurPrenom,
		ConducteurTelephone:    controleEnt.ConducteurTelephone,
		ConducteurAdresse:      controleEnt.ConducteurAdresse,
		// Archivage
		IsArchived: controleEnt.IsArchived,
		ArchivedAt: controleEnt.ArchivedAt,
		// Timestamps
		CreatedAt: controleEnt.CreatedAt,
		UpdatedAt: controleEnt.UpdatedAt,
	}

	// Add relations if loaded
	if controleEnt.Edges.Agent != nil {
		agentSummary := &AgentSummary{
			ID:        controleEnt.Edges.Agent.ID.String(),
			Matricule: controleEnt.Edges.Agent.Matricule,
			Nom:       controleEnt.Edges.Agent.Nom,
			Prenom:    controleEnt.Edges.Agent.Prenom,
			Role:      controleEnt.Edges.Agent.Role,
		}
		// Add extended agent fields if available
		if controleEnt.Edges.Agent.Grade != "" {
			agentSummary.Grade = controleEnt.Edges.Agent.Grade
		}
		if controleEnt.Edges.Agent.Telephone != "" {
			agentSummary.Telephone = controleEnt.Edges.Agent.Telephone
		}
		if controleEnt.Edges.Agent.Email != "" {
			agentSummary.Email = controleEnt.Edges.Agent.Email
		}
		// Get commissariat name if agent has commissariat edge loaded
		if controleEnt.Edges.Agent.Edges.Commissariat != nil {
			agentSummary.Commissariat = controleEnt.Edges.Agent.Edges.Commissariat.Nom
		}
		response.Agent = agentSummary
	}

	if controleEnt.Edges.Vehicule != nil {
		response.Vehicule = &VehiculeSummary{
			ID:              controleEnt.Edges.Vehicule.ID.String(),
			Immatriculation: controleEnt.Edges.Vehicule.Immatriculation,
			Marque:          controleEnt.Edges.Vehicule.Marque,
			Modele:          controleEnt.Edges.Vehicule.Modele,
			TypeVehicule:    controleEnt.Edges.Vehicule.TypeVehicule,
			Annee:           controleEnt.Edges.Vehicule.Annee,
			Couleur:         controleEnt.Edges.Vehicule.Couleur,
			NumeroSerie:     controleEnt.Edges.Vehicule.NumeroChassis,
		}
	}

	if controleEnt.Edges.Conducteur != nil {
		permisValide := controleEnt.Edges.Conducteur.PermisValideJusqu.IsZero() ||
			controleEnt.Edges.Conducteur.PermisValideJusqu.After(time.Now())
		conducteurSummary := &ConducteurSummary{
			ID:           controleEnt.Edges.Conducteur.ID.String(),
			Nom:          controleEnt.Edges.Conducteur.Nom,
			Prenom:       controleEnt.Edges.Conducteur.Prenom,
			NumeroPermis: controleEnt.Edges.Conducteur.NumeroPermis,
			PointsPermis: controleEnt.Edges.Conducteur.PointsPermis,
			PermisValide: permisValide,
			Telephone:    controleEnt.Edges.Conducteur.Telephone,
			Email:        controleEnt.Edges.Conducteur.Email,
			Adresse:      controleEnt.Edges.Conducteur.Adresse,
			CNI:          controleEnt.Edges.Conducteur.NumeroCni,
		}
		// Set validite permis if available
		if !controleEnt.Edges.Conducteur.PermisValideJusqu.IsZero() {
			validiteStr := controleEnt.Edges.Conducteur.PermisValideJusqu.Format("02/01/2006")
			conducteurSummary.ValiditePermis = &validiteStr
		}
		response.Conducteur = conducteurSummary
	}

	if controleEnt.Edges.Commissariat != nil {
		response.Commissariat = &CommissariatSummary{
			ID:   controleEnt.Edges.Commissariat.ID.String(),
			Nom:  controleEnt.Edges.Commissariat.Nom,
			Code: controleEnt.Edges.Commissariat.Code,
		}
	}

	if controleEnt.Edges.Infractions != nil {
		response.NombreInfractions = len(controleEnt.Edges.Infractions)
		response.Infractions = make([]*InfractionSummary, len(controleEnt.Edges.Infractions))

		for i, infraction := range controleEnt.Edges.Infractions {
			typeInfraction := "Non spécifié"
			if infraction.Edges.TypeInfraction != nil {
				typeInfraction = infraction.Edges.TypeInfraction.Libelle
			}

			response.Infractions[i] = &InfractionSummary{
				ID:             infraction.ID.String(),
				NumeroPV:       infraction.NumeroPv,
				DateInfraction: infraction.DateInfraction,
				TypeInfraction: typeInfraction,
				MontantAmende:  infraction.MontantAmende,
				PointsRetires:  infraction.PointsRetires,
				Statut:         infraction.Statut,
			}
		}
	}

	// Documents et éléments vérifiés seront chargés via loadCheckOptions si disponible
	// Sinon, générer les données par défaut
	response.DocumentsVerifies = s.generateDocumentsVerifies(controleEnt)
	response.ElementsControles = s.generateElementsControles(controleEnt)

	// Add PV and Amende if exists for this controle (proces_verbal edge)
	if controleEnt.Edges.ProcesVerbal != nil {
		pv := controleEnt.Edges.ProcesVerbal
		infractionDescriptions := make([]string, 0)
		if pv.Edges.Infractions != nil {
			for _, inf := range pv.Edges.Infractions {
				if inf.Edges.TypeInfraction != nil {
					infractionDescriptions = append(infractionDescriptions, inf.Edges.TypeInfraction.Libelle)
				}
			}
		}
		response.PV = &PVSummary{
			ID:           pv.ID.String(),
			Numero:       pv.NumeroPv,
			DateEmission: pv.DateEmission,
			Infractions:  infractionDescriptions,
			Gravite:      "CLASSE_2", // Default, could be computed from infractions
		}

		// Amende is derived from the ProcesVerbal
		amendeStatut := "EN_ATTENTE"
		if pv.Statut == "PAYE" {
			amendeStatut = "PAYE"
		}
		response.Amende = &AmendeSummary{
			ID:      pv.ID.String(),      // Using PV ID since amende is embedded in PV
			Numero:  pv.NumeroPv,         // Same as PV number
			Montant: pv.MontantTotal,
			Statut:  amendeStatut,
		}
	}

	// Generate recommandations based on controle status and infractions
	response.Recommandations = s.generateRecommandations(controleEnt)

	// Calculate duree (time between creation and update)
	if !controleEnt.UpdatedAt.IsZero() && !controleEnt.CreatedAt.IsZero() {
		duree := controleEnt.UpdatedAt.Sub(controleEnt.CreatedAt)
		if duree.Minutes() < 1 {
			response.Duree = "Moins d'1 minute"
		} else if duree.Hours() < 1 {
			response.Duree = fmt.Sprintf("%d minutes", int(duree.Minutes()))
		} else {
			response.Duree = fmt.Sprintf("%d heures %d minutes", int(duree.Hours()), int(duree.Minutes())%60)
		}
	}

	// Photos - charger depuis les documents liés au contrôle
	response.Photos = []*PhotoControle{}
	if controleEnt.Edges.Documents != nil {
		for _, doc := range controleEnt.Edges.Documents {
			// Filtrer uniquement les documents de type PHOTO
			if doc.TypeDocument == "PHOTO" {
				photo := &PhotoControle{
					ID:          doc.ID.String(),
					Filename:    doc.NomFichier,
					URL:         "/api/documents/" + doc.ID.String() + "/download", // URL pour télécharger
					Description: doc.Description,
					CreatedAt:   doc.CreatedAt,
				}
				response.Photos = append(response.Photos, photo)
			}
		}
	}

	return response
}

// generateDocumentsVerifies generates document verification data based on controle
func (s *service) generateDocumentsVerifies(controleEnt *ent.Controle) []*DocumentVerifie {
	docs := []*DocumentVerifie{}

	// Carte grise - always check based on vehicule immatriculation
	carteGriseStatus := "OK"
	carteGriseDetails := "Carte grise valide"
	if controleEnt.VehiculeImmatriculation == "" {
		carteGriseStatus = "NOK"
		carteGriseDetails = "Immatriculation non renseignée"
	}
	docs = append(docs, &DocumentVerifie{
		Type:    "CARTE_GRISE",
		Statut:  carteGriseStatus,
		Details: carteGriseDetails,
	})

	// Assurance - check vehicule edges if available
	assuranceStatus := "OK"
	assuranceDetails := "Assurance en cours de validité"
	var assuranceValidite *string
	if controleEnt.Edges.Vehicule != nil {
		if controleEnt.Edges.Vehicule.AssuranceValidite.Before(time.Now()) && !controleEnt.Edges.Vehicule.AssuranceValidite.IsZero() {
			assuranceStatus = "NOK"
			assuranceDetails = "Assurance expirée"
		}
		if !controleEnt.Edges.Vehicule.AssuranceValidite.IsZero() {
			validite := controleEnt.Edges.Vehicule.AssuranceValidite.Format("02/01/2006")
			assuranceValidite = &validite
		}
	}
	docs = append(docs, &DocumentVerifie{
		Type:     "ASSURANCE",
		Statut:   assuranceStatus,
		Details:  assuranceDetails,
		Validite: assuranceValidite,
	})

	// Contrôle technique - check vehicule edges if available
	ctStatus := "OK"
	ctDetails := "Contrôle technique à jour"
	var ctValidite *string
	if controleEnt.Edges.Vehicule != nil {
		if controleEnt.Edges.Vehicule.ControleTechniqueValidite.Before(time.Now()) && !controleEnt.Edges.Vehicule.ControleTechniqueValidite.IsZero() {
			ctStatus = "NOK"
			ctDetails = "Contrôle technique expiré"
		}
		if !controleEnt.Edges.Vehicule.ControleTechniqueValidite.IsZero() {
			validite := controleEnt.Edges.Vehicule.ControleTechniqueValidite.Format("02/01/2006")
			ctValidite = &validite
		}
	}
	docs = append(docs, &DocumentVerifie{
		Type:     "CONTROLE_TECHNIQUE",
		Statut:   ctStatus,
		Details:  ctDetails,
		Validite: ctValidite,
	})

	// Permis de conduire - check conducteur data
	permisStatus := "OK"
	permisDetails := "Permis de conduire valide"
	var permisValidite *string
	if controleEnt.Edges.Conducteur != nil {
		if !controleEnt.Edges.Conducteur.PermisValideJusqu.IsZero() {
			validite := controleEnt.Edges.Conducteur.PermisValideJusqu.Format("02/01/2006")
			permisValidite = &validite
			if controleEnt.Edges.Conducteur.PermisValideJusqu.Before(time.Now()) {
				permisStatus = "NOK"
				permisDetails = "Permis de conduire expiré"
			}
		}
	} else if controleEnt.ConducteurNumeroPermis == "" {
		permisStatus = "NOK"
		permisDetails = "Numéro de permis non renseigné"
	}
	docs = append(docs, &DocumentVerifie{
		Type:     "PERMIS_CONDUIRE",
		Statut:   permisStatus,
		Details:  permisDetails,
		Validite: permisValidite,
	})

	return docs
}

// generateElementsControles generates element verification data based on controle type
func (s *service) generateElementsControles(controleEnt *ent.Controle) []*ElementControle {
	elements := []*ElementControle{}

	// Default elements that are always checked
	defaultElements := []struct {
		Type    string
		Details string
	}{
		{"ECLAIRAGE", "Système d'éclairage vérifié"},
		{"FREINAGE", "Système de freinage vérifié"},
		{"PNEUMATIQUES", "État des pneumatiques vérifié"},
		{"CEINTURES", "Ceintures de sécurité vérifiées"},
	}

	// Add default elements with computed status based on controle result
	statut := "OK"
	if controleEnt.Statut == "NON_CONFORME" {
		statut = "NOK"
	}

	for _, elem := range defaultElements {
		elements = append(elements, &ElementControle{
			Type:    elem.Type,
			Statut:  statut,
			Details: elem.Details,
		})
	}

	// Add equipment checks for security controles
	if controleEnt.TypeControle == "SECURITE" || controleEnt.TypeControle == "MIXTE" {
		equipmentElements := []struct {
			Type    string
			Details string
		}{
			{"EXTINCTEUR", "Présence d'extincteur vérifiée"},
			{"TRIANGLE", "Triangle de signalisation vérifié"},
			{"GILET", "Gilet de sécurité vérifié"},
		}

		for _, elem := range equipmentElements {
			elements = append(elements, &ElementControle{
				Type:    elem.Type,
				Statut:  "OK", // Default to OK for equipment
				Details: elem.Details,
			})
		}
	}

	return elements
}

// generateRecommandations generates recommendations based on controle data
func (s *service) generateRecommandations(controleEnt *ent.Controle) []string {
	recommandations := []string{}

	// Check vehicule status
	if controleEnt.Edges.Vehicule != nil {
		vehicule := controleEnt.Edges.Vehicule
		if !vehicule.AssuranceValidite.IsZero() && vehicule.AssuranceValidite.Before(time.Now()) {
			recommandations = append(recommandations, "Renouveler l'assurance du véhicule")
		}
		if !vehicule.ControleTechniqueValidite.IsZero() && vehicule.ControleTechniqueValidite.Before(time.Now()) {
			recommandations = append(recommandations, "Effectuer le contrôle technique")
		}
	}

	// Check conducteur status
	if controleEnt.Edges.Conducteur != nil {
		conducteur := controleEnt.Edges.Conducteur
		if !conducteur.PermisValideJusqu.IsZero() && conducteur.PermisValideJusqu.Before(time.Now()) {
			recommandations = append(recommandations, "Renouveler le permis de conduire")
		}
		if conducteur.PointsPermis <= 6 {
			recommandations = append(recommandations, "Stage de récupération de points recommandé")
		}
	}

	// Add infractions-based recommendations
	if controleEnt.Edges.Infractions != nil && len(controleEnt.Edges.Infractions) > 0 {
		recommandations = append(recommandations, "Régulariser les infractions en cours")
	}

	// Non-conforme controle recommendations
	if controleEnt.Statut == "NON_CONFORME" {
		recommandations = append(recommandations, "Présenter le véhicule à un nouveau contrôle après régularisation")
	}

	return recommandations
}

// entityToResponseWithContext converts entity to response and loads CheckOptions from database
func (s *service) entityToResponseWithContext(ctx context.Context, controleEnt *ent.Controle) *ControleResponse {
	// Get base response
	response := s.entityToResponse(controleEnt)

	// Load CheckOptions from database if verificationRepo is available
	if s.verificationRepo != nil {
		checkOptions, err := s.verificationRepo.GetVerificationsBySource(ctx, "CONTROL", controleEnt.ID.String())
		if err == nil && len(checkOptions) > 0 {
			// Clear generated data and use real data from database
			response.DocumentsVerifies = []*DocumentVerifie{}
			response.ElementsControles = []*ElementControle{}
			// Préparer les infractions depuis les CheckOptions FAIL
			infractionsFromCheckOptions := []*InfractionSummary{}

			for _, opt := range checkOptions {
				if opt.Edges.CheckItem == nil {
					continue
				}
				item := opt.Edges.CheckItem

				// Map result status to statut
				statut := "N/A"
				switch string(opt.ResultStatus) {
				case "PASS":
					statut = "OK"
				case "FAIL":
					statut = "NOK"
				case "WARNING":
					statut = "ATTENTION"
				case "NOT_CHECKED":
					statut = "N/A"
				}

				// Récupérer l'URL de la photo preuve si disponible
				var photoURL *string
				if opt.Edges.EvidenceFile != nil {
					url := "/api/documents/" + opt.Edges.EvidenceFile.ID.String() + "/download"
					photoURL = &url
				}

				// Separate by category: DOCUMENT goes to DocumentsVerifies, others to ElementsControles
				if string(item.ItemCategory) == "DOCUMENT" {
					docVerifie := &DocumentVerifie{
						Type:    item.ItemCode,
						Statut:  statut,
						Details: item.Description,
						Photo:   photoURL,
					}
					if opt.Notes != "" {
						docVerifie.Details = opt.Notes
					}
					response.DocumentsVerifies = append(response.DocumentsVerifies, docVerifie)
				} else {
					elemControle := &ElementControle{
						Type:    item.ItemCode,
						Statut:  statut,
						Details: item.Description,
						Photo:   photoURL,
					}
					if opt.Notes != "" {
						elemControle.Details = opt.Notes
					}
					response.ElementsControles = append(response.ElementsControles, elemControle)
				}

				// Si FAIL et infraction liée, ajouter aux infractions
				if string(opt.ResultStatus) == "FAIL" && opt.Edges.Infraction != nil {
					infraction := opt.Edges.Infraction
					typeInfraction := item.ItemName // Utiliser le nom du CheckItem comme type
					if infraction.Edges.TypeInfraction != nil {
						typeInfraction = infraction.Edges.TypeInfraction.Libelle
					}

					infractionsFromCheckOptions = append(infractionsFromCheckOptions, &InfractionSummary{
						ID:             infraction.ID.String(),
						NumeroPV:       infraction.NumeroPv,
						DateInfraction: infraction.DateInfraction,
						TypeInfraction: typeInfraction,
						MontantAmende:  infraction.MontantAmende,
						PointsRetires:  infraction.PointsRetires,
						Statut:         infraction.Statut,
					})
				}
			}

			// Remplacer les infractions par celles issues des CheckOptions si disponibles
			if len(infractionsFromCheckOptions) > 0 {
				response.Infractions = infractionsFromCheckOptions
				response.NombreInfractions = len(infractionsFromCheckOptions)
			}
		}
	}

	return response
}
