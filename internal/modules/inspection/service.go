package inspection

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/inspection"
	"police-trafic-api-frontend-aligned/ent/user"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service interface defines inspection service methods
type Service interface {
	Create(ctx context.Context, req CreateInspectionRequest) (*InspectionResponse, error)
	GetByID(ctx context.Context, id string) (*InspectionResponse, error)
	GetByNumero(ctx context.Context, numero string) (*InspectionResponse, error)
	List(ctx context.Context, req ListInspectionsRequest) (*ListInspectionsResponse, error)
	Update(ctx context.Context, id string, req UpdateInspectionRequest) (*InspectionResponse, error)
	Delete(ctx context.Context, id string) error
	ChangerStatut(ctx context.Context, id string, req ChangerStatutRequest) (*InspectionResponse, error)
	GetStatistics(ctx context.Context) (*InspectionStatisticsResponse, error)
	GetStatisticsWithFilters(ctx context.Context, dateDebut, dateFin *time.Time) (*InspectionStatisticsResponse, error)
	GetByVehicule(ctx context.Context, vehiculeID string) ([]*InspectionResponse, error)
}

type service struct {
	client           *ent.Client
	verificationRepo repository.VerificationRepository
	logger           *zap.Logger
}

// NewService creates a new inspection service
func NewService(client *ent.Client, verificationRepo repository.VerificationRepository, logger *zap.Logger) Service {
	return &service{
		client:           client,
		verificationRepo: verificationRepo,
		logger:           logger,
	}
}

func (s *service) Create(ctx context.Context, req CreateInspectionRequest) (*InspectionResponse, error) {
	s.logger.Info("Creating new inspection", zap.String("vehicule_immatriculation", req.VehiculeImmatriculation))

	// Generate unique ID and numero
	id := uuid.New()
	numero := fmt.Sprintf("INS-%d-%s", time.Now().Year(), uuid.New().String()[:8])

	// Parse inspecteur ID
	inspecteurID, _ := uuid.Parse(req.InspecteurID)

	// Build create query
	create := s.client.Inspection.Create().
		SetID(id).
		SetNumero(numero).
		SetDateInspection(req.DateInspection).
		SetInspecteurID(inspecteurID).
		// Embedded vehicle data
		SetVehiculeImmatriculation(req.VehiculeImmatriculation).
		SetVehiculeMarque(req.VehiculeMarque).
		SetVehiculeModele(req.VehiculeModele).
		SetVehiculeType(inspection.VehiculeType(req.VehiculeType)).
		// Embedded driver data
		SetConducteurNumeroPermis(req.ConducteurNumeroPermis).
		SetConducteurPrenom(req.ConducteurPrenom).
		SetConducteurNom(req.ConducteurNom)

	// Optional vehicle fields
	if req.VehiculeID != nil {
		vehiculeID, _ := uuid.Parse(*req.VehiculeID)
		create.SetVehiculeID(vehiculeID)
	}
	if req.VehiculeAnnee != nil {
		create.SetVehiculeAnnee(*req.VehiculeAnnee)
	}
	if req.VehiculeCouleur != nil {
		create.SetVehiculeCouleur(*req.VehiculeCouleur)
	}
	if req.VehiculeNumeroChassis != nil {
		create.SetVehiculeNumeroChassis(*req.VehiculeNumeroChassis)
	}

	// Optional driver fields
	if req.ConducteurTelephone != nil {
		create.SetConducteurTelephone(*req.ConducteurTelephone)
	}
	if req.ConducteurAdresse != nil {
		create.SetConducteurAdresse(*req.ConducteurAdresse)
	}
	if req.ConducteurTypePiece != nil {
		create.SetConducteurTypePiece(inspection.ConducteurTypePiece(*req.ConducteurTypePiece))
	}
	if req.ConducteurNumeroPiece != nil {
		create.SetConducteurNumeroPiece(*req.ConducteurNumeroPiece)
	}

	// Insurance fields
	if req.AssuranceCompagnie != nil {
		create.SetAssuranceCompagnie(*req.AssuranceCompagnie)
	}
	if req.AssuranceNumeroPolice != nil {
		create.SetAssuranceNumeroPolice(*req.AssuranceNumeroPolice)
	}
	if req.AssuranceDateExpiration != nil {
		create.SetAssuranceDateExpiration(*req.AssuranceDateExpiration)
	}
	if req.AssuranceStatut != nil {
		create.SetAssuranceStatut(inspection.AssuranceStatut(*req.AssuranceStatut))
	}

	// Location fields
	if req.LieuInspection != nil {
		create.SetLieuInspection(*req.LieuInspection)
	}
	if req.Latitude != nil {
		create.SetLatitude(*req.Latitude)
	}
	if req.Longitude != nil {
		create.SetLongitude(*req.Longitude)
	}

	// Other optional fields
	if req.CommissariatID != nil {
		commID, _ := uuid.Parse(*req.CommissariatID)
		create.SetCommissariatID(commID)
	}
	if req.Observations != nil {
		create.SetObservations(*req.Observations)
	}

	ins, err := create.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create inspection", zap.Error(err))
		return nil, fmt.Errorf("failed to create inspection: %w", err)
	}

	// Process initial options (vérifications) if provided - atomic creation
	if len(req.InitialOptions) > 0 && s.verificationRepo != nil {
		s.logger.Info("Creating initial check options for inspection",
			zap.String("inspection_id", ins.ID.String()),
			zap.Int("nb_options", len(req.InitialOptions)))

		for _, opt := range req.InitialOptions {
			verificationInput := &repository.CreateVerificationInput{
				SourceType:    "INSPECTION",
				SourceID:      ins.ID.String(),
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

		// Update counters on the inspection if provided
		if req.TotalVerifications != nil || req.VerificationsOk != nil ||
			req.VerificationsAttention != nil || req.VerificationsEchec != nil || req.MontantTotalAmendes != nil {
			update := s.client.Inspection.UpdateOneID(ins.ID)
			if req.TotalVerifications != nil {
				update.SetTotalVerifications(*req.TotalVerifications)
			}
			if req.VerificationsOk != nil {
				update.SetVerificationsOk(*req.VerificationsOk)
			}
			if req.VerificationsAttention != nil {
				update.SetVerificationsAttention(*req.VerificationsAttention)
			}
			if req.VerificationsEchec != nil {
				update.SetVerificationsEchec(*req.VerificationsEchec)
			}
			if req.MontantTotalAmendes != nil {
				update.SetMontantTotalAmendes(*req.MontantTotalAmendes)
			}
			ins, _ = update.Save(ctx)
		}
	}

	return s.toResponse(ctx, ins)
}

func (s *service) GetByID(ctx context.Context, id string) (*InspectionResponse, error) {
	uid, _ := uuid.Parse(id)
	ins, err := s.client.Inspection.Query().
		Where(inspection.ID(uid)).
		WithVehicule().
		WithInspecteur().
		WithCommissariat().
		WithProcesVerbal().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("inspection not found")
		}
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}

	return s.toResponse(ctx, ins)
}

func (s *service) GetByNumero(ctx context.Context, numero string) (*InspectionResponse, error) {
	ins, err := s.client.Inspection.Query().
		Where(inspection.Numero(numero)).
		WithVehicule().
		WithInspecteur().
		WithCommissariat().
		WithProcesVerbal().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("inspection not found")
		}
		return nil, fmt.Errorf("failed to get inspection: %w", err)
	}

	return s.toResponse(ctx, ins)
}

func (s *service) List(ctx context.Context, req ListInspectionsRequest) (*ListInspectionsResponse, error) {
	query := s.client.Inspection.Query()

	// Apply filters
	if req.VehiculeID != nil {
		query = query.Where(inspection.HasVehicule())
	}
	if req.InspecteurID != nil {
		inspecteurUID, _ := uuid.Parse(*req.InspecteurID)
		query = query.Where(inspection.HasInspecteurWith(user.ID(inspecteurUID)))
	}
	if req.Statut != nil {
		query = query.Where(inspection.StatutEQ(inspection.Statut(*req.Statut)))
	}
	if req.AssuranceStatut != nil {
		query = query.Where(inspection.AssuranceStatutEQ(inspection.AssuranceStatut(*req.AssuranceStatut)))
	}
	if req.DateDebut != nil {
		query = query.Where(inspection.DateInspectionGTE(*req.DateDebut))
	}
	if req.DateFin != nil {
		query = query.Where(inspection.DateInspectionLTE(*req.DateFin))
	}
	if req.VehiculeImmatriculation != nil && *req.VehiculeImmatriculation != "" {
		query = query.Where(inspection.VehiculeImmatriculationContains(*req.VehiculeImmatriculation))
	}
	if req.Search != nil && *req.Search != "" {
		query = query.Where(
			inspection.Or(
				inspection.NumeroContains(*req.Search),
				inspection.VehiculeImmatriculationContains(*req.Search),
				inspection.ConducteurNumeroPermisContains(*req.Search),
			),
		)
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count inspections: %w", err)
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

	// Order by date inspection descending
	query = query.Order(ent.Desc(inspection.FieldDateInspection))

	// Load with edges
	query = query.WithVehicule().WithInspecteur().WithCommissariat().WithProcesVerbal()

	inspections, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list inspections: %w", err)
	}

	responses := make([]*InspectionResponse, len(inspections))
	for i, ins := range inspections {
		resp, err := s.toResponse(ctx, ins)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}

	return &ListInspectionsResponse{
		Inspections: responses,
		Total:       total,
	}, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateInspectionRequest) (*InspectionResponse, error) {
	uid, _ := uuid.Parse(id)
	update := s.client.Inspection.UpdateOneID(uid)

	if req.DateInspection != nil {
		update.SetDateInspection(*req.DateInspection)
	}
	if req.Statut != nil {
		update.SetStatut(inspection.Statut(*req.Statut))
	}
	if req.Observations != nil {
		update.SetObservations(*req.Observations)
	}
	if req.LieuInspection != nil {
		update.SetLieuInspection(*req.LieuInspection)
	}
	if req.Latitude != nil {
		update.SetLatitude(*req.Latitude)
	}
	if req.Longitude != nil {
		update.SetLongitude(*req.Longitude)
	}
	// Insurance updates
	if req.AssuranceCompagnie != nil {
		update.SetAssuranceCompagnie(*req.AssuranceCompagnie)
	}
	if req.AssuranceNumeroPolice != nil {
		update.SetAssuranceNumeroPolice(*req.AssuranceNumeroPolice)
	}
	if req.AssuranceDateExpiration != nil {
		update.SetAssuranceDateExpiration(*req.AssuranceDateExpiration)
	}
	if req.AssuranceStatut != nil {
		update.SetAssuranceStatut(inspection.AssuranceStatut(*req.AssuranceStatut))
	}
	// Counter updates
	if req.TotalVerifications != nil {
		update.SetTotalVerifications(*req.TotalVerifications)
	}
	if req.VerificationsOk != nil {
		update.SetVerificationsOk(*req.VerificationsOk)
	}
	if req.VerificationsAttention != nil {
		update.SetVerificationsAttention(*req.VerificationsAttention)
	}
	if req.VerificationsEchec != nil {
		update.SetVerificationsEchec(*req.VerificationsEchec)
	}
	if req.MontantTotalAmendes != nil {
		update.SetMontantTotalAmendes(*req.MontantTotalAmendes)
	}

	ins, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("inspection not found")
		}
		return nil, fmt.Errorf("failed to update inspection: %w", err)
	}

	return s.toResponse(ctx, ins)
}

func (s *service) Delete(ctx context.Context, id string) error {
	uid, _ := uuid.Parse(id)
	err := s.client.Inspection.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("inspection not found")
		}
		return fmt.Errorf("failed to delete inspection: %w", err)
	}
	return nil
}

func (s *service) ChangerStatut(ctx context.Context, id string, req ChangerStatutRequest) (*InspectionResponse, error) {
	uid, _ := uuid.Parse(id)
	update := s.client.Inspection.UpdateOneID(uid).
		SetStatut(inspection.Statut(req.Statut))

	if req.Observations != nil {
		update.SetObservations(*req.Observations)
	}
	if req.TotalVerifications != nil {
		update.SetTotalVerifications(*req.TotalVerifications)
	}
	if req.VerificationsOk != nil {
		update.SetVerificationsOk(*req.VerificationsOk)
	}
	if req.VerificationsAttention != nil {
		update.SetVerificationsAttention(*req.VerificationsAttention)
	}
	if req.VerificationsEchec != nil {
		update.SetVerificationsEchec(*req.VerificationsEchec)
	}
	if req.MontantTotalAmendes != nil {
		update.SetMontantTotalAmendes(*req.MontantTotalAmendes)
	}

	ins, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("inspection not found")
		}
		return nil, fmt.Errorf("failed to change statut: %w", err)
	}

	return s.toResponse(ctx, ins)
}

func (s *service) GetStatistics(ctx context.Context) (*InspectionStatisticsResponse, error) {
	// Get all inspections for statistics
	inspections, err := s.client.Inspection.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspections for statistics: %w", err)
	}

	stats := &InspectionStatisticsResponse{
		Total:     len(inspections),
		ParStatut: make(map[string]int),
	}

	var completedTotal int

	for _, ins := range inspections {
		// Count by status
		switch ins.Statut {
		case inspection.StatutEN_ATTENTE:
			stats.EnAttente++
		case inspection.StatutEN_COURS:
			stats.EnCours++
		case inspection.StatutTERMINE:
			stats.Termine++
			completedTotal++
		case inspection.StatutCONFORME:
			stats.Conforme++
			completedTotal++
		case inspection.StatutNON_CONFORME:
			stats.NonConforme++
			completedTotal++
		}

		stats.ParStatut[string(ins.Statut)]++

		// Count invalid insurance
		if ins.AssuranceStatut == inspection.AssuranceStatutEXPIREE ||
			ins.AssuranceStatut == inspection.AssuranceStatutSUSPENDUE ||
			ins.AssuranceStatut == inspection.AssuranceStatutANNULEE {
			stats.AssuranceInvalide++
		}

		// Sum fine amounts
		stats.MontantTotalAmendes += ins.MontantTotalAmendes
	}

	// Calculate pass rate (taux de conformité)
	if completedTotal > 0 {
		stats.TauxConformite = float64(stats.Conforme) / float64(completedTotal) * 100
	}

	return stats, nil
}

func (s *service) GetStatisticsWithFilters(ctx context.Context, dateDebut, dateFin *time.Time) (*InspectionStatisticsResponse, error) {
	// Build query with date filters
	query := s.client.Inspection.Query()

	if dateDebut != nil {
		query = query.Where(inspection.DateInspectionGTE(*dateDebut))
	}
	if dateFin != nil {
		query = query.Where(inspection.DateInspectionLTE(*dateFin))
	}

	inspections, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspections for statistics: %w", err)
	}

	stats := &InspectionStatisticsResponse{
		Total:     len(inspections),
		ParStatut: make(map[string]int),
	}

	var completedTotal int

	for _, ins := range inspections {
		// Count by status
		switch ins.Statut {
		case inspection.StatutEN_ATTENTE:
			stats.EnAttente++
		case inspection.StatutEN_COURS:
			stats.EnCours++
		case inspection.StatutTERMINE:
			stats.Termine++
			completedTotal++
		case inspection.StatutCONFORME:
			stats.Conforme++
			completedTotal++
		case inspection.StatutNON_CONFORME:
			stats.NonConforme++
			completedTotal++
		}

		stats.ParStatut[string(ins.Statut)]++

		// Count invalid insurance
		if ins.AssuranceStatut == inspection.AssuranceStatutEXPIREE ||
			ins.AssuranceStatut == inspection.AssuranceStatutSUSPENDUE ||
			ins.AssuranceStatut == inspection.AssuranceStatutANNULEE {
			stats.AssuranceInvalide++
		}

		// Sum fine amounts
		stats.MontantTotalAmendes += ins.MontantTotalAmendes
	}

	// Calculate pass rate (taux de conformité)
	if completedTotal > 0 {
		stats.TauxConformite = float64(stats.Conforme) / float64(completedTotal) * 100
	}

	return stats, nil
}

func (s *service) GetByVehicule(ctx context.Context, vehiculeID string) ([]*InspectionResponse, error) {
	inspections, err := s.client.Inspection.Query().
		Where(inspection.HasVehiculeWith()).
		WithVehicule().
		WithInspecteur().
		WithCommissariat().
		WithProcesVerbal().
		Order(ent.Desc(inspection.FieldDateInspection)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspections by vehicule: %w", err)
	}

	responses := make([]*InspectionResponse, len(inspections))
	for i, ins := range inspections {
		resp, err := s.toResponse(ctx, ins)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}

	return responses, nil
}

func (s *service) toResponse(ctx context.Context, ins *ent.Inspection) (*InspectionResponse, error) {
	resp := &InspectionResponse{
		ID:             ins.ID.String(),
		Numero:         ins.Numero,
		DateInspection: ins.DateInspection,
		Statut:         string(ins.Statut),
		Observations:   ins.Observations,
		// Counters
		TotalVerifications:     ins.TotalVerifications,
		VerificationsOk:        ins.VerificationsOk,
		VerificationsAttention: ins.VerificationsAttention,
		VerificationsEchec:     ins.VerificationsEchec,
		MontantTotalAmendes:    ins.MontantTotalAmendes,
		// Embedded vehicle data
		VehiculeImmatriculation: ins.VehiculeImmatriculation,
		VehiculeMarque:          ins.VehiculeMarque,
		VehiculeModele:          ins.VehiculeModele,
		VehiculeAnnee:           ins.VehiculeAnnee,
		VehiculeCouleur:         ins.VehiculeCouleur,
		VehiculeNumeroChassis:   ins.VehiculeNumeroChassis,
		VehiculeType:            string(ins.VehiculeType),
		// Embedded driver data
		ConducteurNumeroPermis: ins.ConducteurNumeroPermis,
		ConducteurPrenom:       ins.ConducteurPrenom,
		ConducteurNom:          ins.ConducteurNom,
		ConducteurTelephone:    ins.ConducteurTelephone,
		ConducteurAdresse:      ins.ConducteurAdresse,
		// Insurance data
		AssuranceCompagnie:      ins.AssuranceCompagnie,
		AssuranceNumeroPolice:   ins.AssuranceNumeroPolice,
		AssuranceDateExpiration: ins.AssuranceDateExpiration,
		AssuranceStatut:         string(ins.AssuranceStatut),
		// Location
		LieuInspection: ins.LieuInspection,
		Latitude:       ins.Latitude,
		Longitude:      ins.Longitude,
		// Timestamps
		CreatedAt: ins.CreatedAt,
		UpdatedAt: ins.UpdatedAt,
	}

	// Add driver ID type if set
	if ins.ConducteurTypePiece != "" {
		resp.ConducteurTypePiece = string(ins.ConducteurTypePiece)
	}
	resp.ConducteurNumeroPiece = ins.ConducteurNumeroPiece

	// Load edges if present
	if ins.Edges.Vehicule != nil {
		resp.Vehicule = &VehiculeSummary{
			ID:              ins.Edges.Vehicule.ID.String(),
			Immatriculation: ins.Edges.Vehicule.Immatriculation,
			Marque:          ins.Edges.Vehicule.Marque,
			Modele:          ins.Edges.Vehicule.Modele,
			ProprietaireNom: ins.Edges.Vehicule.ProprietaireNom,
		}
	}

	if ins.Edges.Inspecteur != nil {
		resp.Inspecteur = &AgentSummary{
			ID:        ins.Edges.Inspecteur.ID.String(),
			Matricule: ins.Edges.Inspecteur.Matricule,
			Nom:       ins.Edges.Inspecteur.Nom,
			Prenom:    ins.Edges.Inspecteur.Prenom,
		}
	}

	if ins.Edges.Commissariat != nil {
		resp.Commissariat = &CommissariatSummary{
			ID:   ins.Edges.Commissariat.ID.String(),
			Nom:  ins.Edges.Commissariat.Nom,
			Code: ins.Edges.Commissariat.Code,
		}
	}

	if ins.Edges.ProcesVerbal != nil {
		resp.ProcesVerbal = &PVSummary{
			ID:           ins.Edges.ProcesVerbal.ID.String(),
			NumeroPV:     ins.Edges.ProcesVerbal.NumeroPv,
			DateEmission: ins.Edges.ProcesVerbal.DateEmission,
			MontantTotal: ins.Edges.ProcesVerbal.MontantTotal,
			Statut:       ins.Edges.ProcesVerbal.Statut,
		}
	}

	return resp, nil
}
