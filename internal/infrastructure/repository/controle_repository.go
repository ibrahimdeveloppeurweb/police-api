package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/commissariat"
	"police-trafic-api-frontend-aligned/ent/conducteur"
	"police-trafic-api-frontend-aligned/ent/controle"
	"police-trafic-api-frontend-aligned/ent/user"
	"police-trafic-api-frontend-aligned/ent/vehicule"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ControleRepository defines controle repository interface
type ControleRepository interface {
	Create(ctx context.Context, input *CreateControleInput) (*ent.Controle, error)
	GetByID(ctx context.Context, id string) (*ent.Controle, error)
	List(ctx context.Context, filters *ControleFilters) ([]*ent.Controle, error)
	Count(ctx context.Context, filters *ControleFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateControleInput) (*ent.Controle, error)
	Delete(ctx context.Context, id string) error
	GetByAgent(ctx context.Context, agentID string, filters *ControleFilters) ([]*ent.Controle, error)
	GetByVehicule(ctx context.Context, vehiculeID string) ([]*ent.Controle, error)
	GetByConducteur(ctx context.Context, conducteurID string) ([]*ent.Controle, error)
	GetByCommissariat(ctx context.Context, commissariatID string, filters *ControleFilters) ([]*ent.Controle, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.Controle, error)
	GetStatistics(ctx context.Context, filters *ControleStatsFilters) (*ControleStatistics, error)
	Archive(ctx context.Context, id string) (*ent.Controle, error)
	Unarchive(ctx context.Context, id string) (*ent.Controle, error)
}

// CreateControleInput represents input for creating controle
type CreateControleInput struct {
	ID        string
	Reference string
	// Date et localisation
	DateControle time.Time
	LieuControle string
	Latitude     *float64
	Longitude    *float64
	// Info contrôle
	TypeControle string // DOCUMENT, SECURITE, GENERAL, MIXTE
	Statut       string // EN_COURS, TERMINE, CONFORME, NON_CONFORME
	Observations *string
	// Relations
	AgentID        string
	VehiculeID     *string
	ConducteurID   *string
	CommissariatID *string
	// Données véhicule embarquées (dénormalisées)
	VehiculeImmatriculation string
	VehiculeMarque          string
	VehiculeModele          string
	VehiculeAnnee           *int
	VehiculeCouleur         *string
	VehiculeNumeroChassis   *string
	VehiculeType            string // VOITURE, SUV, CAMION, CAMIONNETTE, MOTO, BUS, AUTRE
	// Données conducteur embarquées (dénormalisées)
	ConducteurNumeroPermis string
	ConducteurNom          string
	ConducteurPrenom       string
	ConducteurTelephone    *string
	ConducteurAdresse      *string
}

// UpdateControleInput represents input for updating controle
type UpdateControleInput struct {
	DateControle *time.Time
	LieuControle *string
	Latitude     *float64
	Longitude    *float64
	TypeControle *string
	Statut       *string
	Observations *string
	// Compteurs
	TotalVerifications  *int
	VerificationsOk     *int
	VerificationsEchec  *int
	MontantTotalAmendes *int
}

// ControleFilters represents filters for listing controles
type ControleFilters struct {
	AgentID                 *string
	VehiculeID              *string
	ConducteurID            *string
	CommissariatID          *string
	TypeControle            *string
	Statut                  *string
	LieuControle            *string
	VehiculeImmatriculation *string
	DateDebut               *time.Time
	DateFin                 *time.Time
	IsArchived              *bool // Filter by archive status
	Limit                   int
	Offset                  int
}

// ControleStatsFilters represents filters for statistics
type ControleStatsFilters struct {
	AgentID   *string
	DateDebut *time.Time
	DateFin   *time.Time
}

type ControleStatistics struct {
	Total               int            `json:"total"`
	EnCours             int            `json:"en_cours"`
	Termine             int            `json:"termine"`
	Conforme            int            `json:"conforme"`
	NonConforme         int            `json:"non_conforme"`
	ParType             map[string]int `json:"par_type"`
	ParJour             map[string]int `json:"par_jour"`
	InfractionsAvec     int            `json:"infractions_avec"`
	InfractionsSans     int            `json:"infractions_sans"`
	MontantTotalAmendes int            `json:"montant_total_amendes"`
}

// controleRepository implements ControleRepository
type controleRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewControleRepository creates a new controle repository
func NewControleRepository(client *ent.Client, logger *zap.Logger) ControleRepository {
	return &controleRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new controle
func (r *controleRepository) Create(ctx context.Context, input *CreateControleInput) (*ent.Controle, error) {
	r.logger.Info("Creating controle",
		zap.String("lieu", input.LieuControle), zap.String("agent_id", input.AgentID))

	id, _ := uuid.Parse(input.ID)
	agentID, _ := uuid.Parse(input.AgentID)
	create := r.client.Controle.Create().
		SetID(id).
		SetDateControle(input.DateControle).
		SetLieuControle(input.LieuControle).
		SetTypeControle(controle.TypeControle(input.TypeControle)).
		SetStatut(controle.Statut(input.Statut)).
		SetAgentID(agentID).
		// Données véhicule embarquées
		SetVehiculeImmatriculation(input.VehiculeImmatriculation).
		SetVehiculeMarque(input.VehiculeMarque).
		SetVehiculeModele(input.VehiculeModele).
		SetVehiculeType(controle.VehiculeType(input.VehiculeType)).
		// Données conducteur embarquées
		SetConducteurNumeroPermis(input.ConducteurNumeroPermis).
		SetConducteurNom(input.ConducteurNom).
		SetConducteurPrenom(input.ConducteurPrenom)

	// Champs optionnels
	if input.Reference != "" {
		create = create.SetReference(input.Reference)
	}
	if input.Observations != nil {
		create = create.SetObservations(*input.Observations)
	}
	if input.Latitude != nil {
		create = create.SetLatitude(*input.Latitude)
	}
	if input.Longitude != nil {
		create = create.SetLongitude(*input.Longitude)
	}
	if input.VehiculeID != nil {
		vehID, _ := uuid.Parse(*input.VehiculeID)
		create = create.SetVehiculeID(vehID)
	}
	if input.ConducteurID != nil {
		condID, _ := uuid.Parse(*input.ConducteurID)
		create = create.SetConducteurID(condID)
	}
	if input.CommissariatID != nil {
		commID, _ := uuid.Parse(*input.CommissariatID)
		create = create.SetCommissariatID(commID)
	}
	if input.VehiculeAnnee != nil {
		create = create.SetVehiculeAnnee(*input.VehiculeAnnee)
	}
	if input.VehiculeCouleur != nil {
		create = create.SetVehiculeCouleur(*input.VehiculeCouleur)
	}
	if input.VehiculeNumeroChassis != nil {
		create = create.SetVehiculeNumeroChassis(*input.VehiculeNumeroChassis)
	}
	if input.ConducteurTelephone != nil {
		create = create.SetConducteurTelephone(*input.ConducteurTelephone)
	}
	if input.ConducteurAdresse != nil {
		create = create.SetConducteurAdresse(*input.ConducteurAdresse)
	}

	controleEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create controle", zap.Error(err))
		return nil, fmt.Errorf("failed to create controle: %w", err)
	}

	return controleEnt, nil
}

// GetByID gets controle by ID
func (r *controleRepository) GetByID(ctx context.Context, id string) (*ent.Controle, error) {
	uid, _ := uuid.Parse(id)
	controleEnt, err := r.client.Controle.
		Query().
		Where(controle.ID(uid)).
		WithAgent(func(q *ent.UserQuery) {
			q.WithCommissariat()
		}).
		WithVehicule().
		WithConducteur().
		WithCommissariat().
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction().WithProcesVerbal()
		}).
		WithDocuments().
		WithProcesVerbal(func(q *ent.ProcesVerbalQuery) {
			q.WithInfractions(func(iq *ent.InfractionQuery) {
				iq.WithTypeInfraction()
			})
		}).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("controle not found")
		}
		r.logger.Error("Failed to get controle by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get controle: %w", err)
	}

	return controleEnt, nil
}

// List gets controles with filters
func (r *controleRepository) List(ctx context.Context, filters *ControleFilters) ([]*ent.Controle, error) {
	query := r.client.Controle.Query()

	if filters != nil {
		if filters.AgentID != nil {
			agentUID, _ := uuid.Parse(*filters.AgentID)
			query = query.Where(controle.HasAgentWith(user.ID(agentUID)))
		}
		if filters.VehiculeID != nil {
			vehUID, _ := uuid.Parse(*filters.VehiculeID)
			query = query.Where(controle.HasVehiculeWith(vehicule.ID(vehUID)))
		}
		if filters.ConducteurID != nil {
			condUID, _ := uuid.Parse(*filters.ConducteurID)
			query = query.Where(controle.HasConducteurWith(conducteur.ID(condUID)))
		}
		if filters.CommissariatID != nil {
			commUID, _ := uuid.Parse(*filters.CommissariatID)
			query = query.Where(controle.HasCommissariatWith(commissariat.ID(commUID)))
		}
		if filters.TypeControle != nil {
			query = query.Where(controle.TypeControleEQ(controle.TypeControle(*filters.TypeControle)))
		}
		if filters.Statut != nil {
			query = query.Where(controle.StatutEQ(controle.Statut(*filters.Statut)))
		}
		if filters.LieuControle != nil {
			query = query.Where(controle.LieuControleContains(*filters.LieuControle))
		}
		if filters.VehiculeImmatriculation != nil {
			query = query.Where(controle.VehiculeImmatriculationContains(*filters.VehiculeImmatriculation))
		}
		if filters.DateDebut != nil {
			query = query.Where(controle.DateControleGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(controle.DateControleLTE(*filters.DateFin))
		}
		if filters.IsArchived != nil {
			query = query.Where(controle.IsArchivedEQ(*filters.IsArchived))
		}

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	controles, err := query.
		WithAgent().
		WithVehicule().
		WithConducteur().
		WithCommissariat().
		WithInfractions().
		Order(ent.Desc(controle.FieldDateControle)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list controles", zap.Error(err))
		return nil, fmt.Errorf("failed to list controles: %w", err)
	}

	return controles, nil
}

// Count counts controles with filters
func (r *controleRepository) Count(ctx context.Context, filters *ControleFilters) (int, error) {
	query := r.client.Controle.Query()

	if filters != nil {
		if filters.AgentID != nil {
			agentUID, _ := uuid.Parse(*filters.AgentID)
			query = query.Where(controle.HasAgentWith(user.ID(agentUID)))
		}
		if filters.VehiculeID != nil {
			vehUID, _ := uuid.Parse(*filters.VehiculeID)
			query = query.Where(controle.HasVehiculeWith(vehicule.ID(vehUID)))
		}
		if filters.ConducteurID != nil {
			condUID, _ := uuid.Parse(*filters.ConducteurID)
			query = query.Where(controle.HasConducteurWith(conducteur.ID(condUID)))
		}
		if filters.CommissariatID != nil {
			commUID, _ := uuid.Parse(*filters.CommissariatID)
			query = query.Where(controle.HasCommissariatWith(commissariat.ID(commUID)))
		}
		if filters.TypeControle != nil {
			query = query.Where(controle.TypeControleEQ(controle.TypeControle(*filters.TypeControle)))
		}
		if filters.Statut != nil {
			query = query.Where(controle.StatutEQ(controle.Statut(*filters.Statut)))
		}
		if filters.DateDebut != nil {
			query = query.Where(controle.DateControleGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(controle.DateControleLTE(*filters.DateFin))
		}
		if filters.IsArchived != nil {
			query = query.Where(controle.IsArchivedEQ(*filters.IsArchived))
		}
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count controles", zap.Error(err))
		return 0, fmt.Errorf("failed to count controles: %w", err)
	}

	return count, nil
}

// Update updates controle
func (r *controleRepository) Update(ctx context.Context, id string, input *UpdateControleInput) (*ent.Controle, error) {
	r.logger.Info("Updating controle", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Controle.UpdateOneID(uid)

	if input.DateControle != nil {
		update = update.SetDateControle(*input.DateControle)
	}
	if input.LieuControle != nil {
		update = update.SetLieuControle(*input.LieuControle)
	}
	if input.Latitude != nil {
		update = update.SetLatitude(*input.Latitude)
	}
	if input.Longitude != nil {
		update = update.SetLongitude(*input.Longitude)
	}
	if input.Observations != nil {
		update = update.SetObservations(*input.Observations)
	}
	if input.TypeControle != nil {
		update = update.SetTypeControle(controle.TypeControle(*input.TypeControle))
	}
	if input.Statut != nil {
		update = update.SetStatut(controle.Statut(*input.Statut))
	}
	if input.TotalVerifications != nil {
		update = update.SetTotalVerifications(*input.TotalVerifications)
	}
	if input.VerificationsOk != nil {
		update = update.SetVerificationsOk(*input.VerificationsOk)
	}
	if input.VerificationsEchec != nil {
		update = update.SetVerificationsEchec(*input.VerificationsEchec)
	}
	if input.MontantTotalAmendes != nil {
		update = update.SetMontantTotalAmendes(*input.MontantTotalAmendes)
	}

	controleEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update controle", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update controle: %w", err)
	}

	return controleEnt, nil
}

// Delete deletes controle
func (r *controleRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting controle", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Controle.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete controle", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete controle: %w", err)
	}

	return nil
}

// GetByAgent gets controles by agent
func (r *controleRepository) GetByAgent(ctx context.Context, agentID string, filters *ControleFilters) ([]*ent.Controle, error) {
	agentUID, _ := uuid.Parse(agentID)
	query := r.client.Controle.Query().
		Where(controle.HasAgentWith(user.ID(agentUID)))

	if filters != nil {
		if filters.TypeControle != nil {
			query = query.Where(controle.TypeControleEQ(controle.TypeControle(*filters.TypeControle)))
		}
		if filters.Statut != nil {
			query = query.Where(controle.StatutEQ(controle.Statut(*filters.Statut)))
		}
		if filters.DateDebut != nil {
			query = query.Where(controle.DateControleGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(controle.DateControleLTE(*filters.DateFin))
		}

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	controles, err := query.
		WithAgent().
		WithVehicule().
		WithConducteur().
		WithInfractions().
		Order(ent.Desc(controle.FieldDateControle)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get controles by agent",
			zap.String("agentID", agentID), zap.Error(err))
		return nil, fmt.Errorf("failed to get controles by agent: %w", err)
	}

	return controles, nil
}

// GetByVehicule gets controles by vehicule
func (r *controleRepository) GetByVehicule(ctx context.Context, vehiculeID string) ([]*ent.Controle, error) {
	vehUID, _ := uuid.Parse(vehiculeID)
	controles, err := r.client.Controle.Query().
		Where(controle.HasVehiculeWith(vehicule.ID(vehUID))).
		WithAgent().
		WithVehicule().
		WithConducteur().
		WithInfractions().
		Order(ent.Desc(controle.FieldDateControle)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get controles by vehicule",
			zap.String("vehiculeID", vehiculeID), zap.Error(err))
		return nil, fmt.Errorf("failed to get controles by vehicule: %w", err)
	}

	return controles, nil
}

// GetByConducteur gets controles by conducteur
func (r *controleRepository) GetByConducteur(ctx context.Context, conducteurID string) ([]*ent.Controle, error) {
	condUID, _ := uuid.Parse(conducteurID)
	controles, err := r.client.Controle.Query().
		Where(controle.HasConducteurWith(conducteur.ID(condUID))).
		WithAgent().
		WithVehicule().
		WithConducteur().
		WithInfractions().
		Order(ent.Desc(controle.FieldDateControle)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get controles by conducteur",
			zap.String("conducteurID", conducteurID), zap.Error(err))
		return nil, fmt.Errorf("failed to get controles by conducteur: %w", err)
	}

	return controles, nil
}

// GetByCommissariat gets controles by commissariat
func (r *controleRepository) GetByCommissariat(ctx context.Context, commissariatID string, filters *ControleFilters) ([]*ent.Controle, error) {
	commUID, _ := uuid.Parse(commissariatID)
	query := r.client.Controle.Query().
		Where(controle.HasCommissariatWith(commissariat.ID(commUID)))

	if filters != nil {
		if filters.TypeControle != nil {
			query = query.Where(controle.TypeControleEQ(controle.TypeControle(*filters.TypeControle)))
		}
		if filters.Statut != nil {
			query = query.Where(controle.StatutEQ(controle.Statut(*filters.Statut)))
		}
		if filters.DateDebut != nil {
			query = query.Where(controle.DateControleGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(controle.DateControleLTE(*filters.DateFin))
		}
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	controles, err := query.
		WithAgent().
		WithVehicule().
		WithConducteur().
		WithInfractions().
		Order(ent.Desc(controle.FieldDateControle)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get controles by commissariat",
			zap.String("commissariatID", commissariatID), zap.Error(err))
		return nil, fmt.Errorf("failed to get controles by commissariat: %w", err)
	}

	return controles, nil
}

// GetByDateRange gets controles by date range
func (r *controleRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.Controle, error) {
	controles, err := r.client.Controle.Query().
		Where(
			controle.And(
				controle.DateControleGTE(start),
				controle.DateControleLTE(end),
			),
		).
		WithAgent().
		WithVehicule().
		WithConducteur().
		WithInfractions().
		Order(ent.Desc(controle.FieldDateControle)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get controles by date range",
			zap.Time("start", start), zap.Time("end", end), zap.Error(err))
		return nil, fmt.Errorf("failed to get controles by date range: %w", err)
	}

	return controles, nil
}

// GetStatistics gets statistics for controles
func (r *controleRepository) GetStatistics(ctx context.Context, filters *ControleStatsFilters) (*ControleStatistics, error) {
	query := r.client.Controle.Query()

	if filters != nil {
		if filters.AgentID != nil {
			agentUID, _ := uuid.Parse(*filters.AgentID)
			query = query.Where(controle.HasAgentWith(user.ID(agentUID)))
		}
		if filters.DateDebut != nil {
			query = query.Where(controle.DateControleGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			// Add one day to include the entire end date
			endDate := filters.DateFin.Add(24 * time.Hour)
			query = query.Where(controle.DateControleLT(endDate))
		}
	}

	controles, err := query.WithInfractions().All(ctx)
	if err != nil {
		r.logger.Error("Failed to get controles for statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get controles for statistics: %w", err)
	}

	stats := &ControleStatistics{
		Total:   len(controles),
		ParType: make(map[string]int),
		ParJour: make(map[string]int),
	}

	for _, ctrl := range controles {
		// Statuts
		switch ctrl.Statut {
		case controle.StatutEN_COURS:
			stats.EnCours++
		case controle.StatutTERMINE:
			stats.Termine++
		case controle.StatutCONFORME:
			stats.Conforme++
		case controle.StatutNON_CONFORME:
			stats.NonConforme++
		}

		// Types
		stats.ParType[string(ctrl.TypeControle)]++

		// Par jour
		jour := ctrl.DateControle.Format("2006-01-02")
		stats.ParJour[jour]++

		// Infractions et calcul du montant total
		if len(ctrl.Edges.Infractions) > 0 {
			stats.InfractionsAvec++
			// Calculer le montant total des amendes à partir des infractions
			for _, inf := range ctrl.Edges.Infractions {
				stats.MontantTotalAmendes += int(inf.MontantAmende)
			}
		} else {
			stats.InfractionsSans++
		}
	}

	return stats, nil
}

// Archive archives a controle
func (r *controleRepository) Archive(ctx context.Context, id string) (*ent.Controle, error) {
	r.logger.Info("Archiving controle", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	controleEnt, err := r.client.Controle.UpdateOneID(uid).
		SetIsArchived(true).
		SetArchivedAt(time.Now()).
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("controle not found")
		}
		r.logger.Error("Failed to archive controle", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to archive controle: %w", err)
	}

	return controleEnt, nil
}

// Unarchive removes controle from archives
func (r *controleRepository) Unarchive(ctx context.Context, id string) (*ent.Controle, error) {
	r.logger.Info("Unarchiving controle", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	controleEnt, err := r.client.Controle.UpdateOneID(uid).
		SetIsArchived(false).
		ClearArchivedAt().
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("controle not found")
		}
		r.logger.Error("Failed to unarchive controle", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to unarchive controle: %w", err)
	}

	return controleEnt, nil
}
