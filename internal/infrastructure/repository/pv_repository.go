package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/controle"
	"police-trafic-api-frontend-aligned/ent/infraction"
	"police-trafic-api-frontend-aligned/ent/procesverbal"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PVRepository defines proces verbal repository interface
type PVRepository interface {
	Create(ctx context.Context, input *CreatePVInput) (*ent.ProcesVerbal, error)
	GetByID(ctx context.Context, id string) (*ent.ProcesVerbal, error)
	GetByNumeroPV(ctx context.Context, numero string) (*ent.ProcesVerbal, error)
	List(ctx context.Context, filters *PVFilters) ([]*ent.ProcesVerbal, error)
	Count(ctx context.Context, filters *PVFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdatePVInput) (*ent.ProcesVerbal, error)
	Delete(ctx context.Context, id string) error
	GetByInfraction(ctx context.Context, infractionID string) (*ent.ProcesVerbal, error)
	GetByStatut(ctx context.Context, statut string) ([]*ent.ProcesVerbal, error)
	GetExpired(ctx context.Context) ([]*ent.ProcesVerbal, error)
	GetStatistics(ctx context.Context, filters *PVFilters) (*PVStatistics, error)
}

// CreatePVInput represents input for creating PV
type CreatePVInput struct {
	ID                 string
	NumeroPV           string
	DateEmission       time.Time
	MontantTotal       float64
	MontantMajore      *float64
	DateLimitePaiement *time.Time
	DateMajoration     *time.Time
	Statut             string
	Observations       *string
	InfractionIDs      []string // Un PV peut avoir plusieurs infractions
	ControleID         *string  // Optionnel: lié à un contrôle
	InspectionID       *string  // Optionnel: lié à une inspection
}

// UpdatePVInput represents input for updating PV
type UpdatePVInput struct {
	MontantTotal         *float64
	MontantMajore        *float64
	DateLimitePaiement   *time.Time
	DateMajoration       *time.Time
	Statut               *string
	DatePaiement         *time.Time
	MontantPaye          *float64
	MoyenPaiement        *string
	ReferencePaiement    *string
	DateContestation     *time.Time
	MotifContestation    *string
	DecisionContestation *string
	TribunalCompetent    *string
	Observations         *string
}

// PVFilters represents filters for listing PVs
type PVFilters struct {
	InfractionID *string
	AgentID      *string
	Statut       *string
	DateDebut    *time.Time
	DateFin      *time.Time
	MontantMin   *float64
	MontantMax   *float64
	Expired      *bool
	Limit        int
	Offset       int
}

// PVStatistics represents statistics for PVs
type PVStatistics struct {
	Total            int                `json:"total"`
	MontantTotal     float64            `json:"montant_total"`
	MontantPaye      float64            `json:"montant_paye"`
	MontantImpaye    float64            `json:"montant_impaye"`
	ParStatut        map[string]int     `json:"par_statut"`
	ParMois          map[string]float64 `json:"par_mois"`
	TauxRecouvrement float64            `json:"taux_recouvrement"`
	PVExpires        int                `json:"pv_expires"`
}

// pvRepository implements PVRepository
type pvRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewPVRepository creates a new PV repository
func NewPVRepository(client *ent.Client, logger *zap.Logger) PVRepository {
	return &pvRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new PV
func (r *pvRepository) Create(ctx context.Context, input *CreatePVInput) (*ent.ProcesVerbal, error) {
	r.logger.Info("Creating PV",
		zap.String("numero_pv", input.NumeroPV),
		zap.Float64("montant", input.MontantTotal))

	id, _ := uuid.Parse(input.ID)
	create := r.client.ProcesVerbal.Create().
		SetID(id).
		SetNumeroPv(input.NumeroPV).
		SetDateEmission(input.DateEmission).
		SetMontantTotal(input.MontantTotal).
		SetStatut(input.Statut)

	// Ajouter les infractions (1 PV pour N infractions)
	if len(input.InfractionIDs) > 0 {
		infUUIDs := make([]uuid.UUID, len(input.InfractionIDs))
		for i, infID := range input.InfractionIDs {
			infUUIDs[i], _ = uuid.Parse(infID)
		}
		create = create.AddInfractionIDs(infUUIDs...)
	}

	// Lier au contrôle ou à l'inspection
	if input.ControleID != nil {
		ctrlID, _ := uuid.Parse(*input.ControleID)
		create = create.SetControleID(ctrlID)
	}
	if input.InspectionID != nil {
		inspID, _ := uuid.Parse(*input.InspectionID)
		create = create.SetInspectionID(inspID)
	}

	if input.MontantMajore != nil {
		create = create.SetMontantMajore(*input.MontantMajore)
	}
	if input.DateLimitePaiement != nil {
		create = create.SetDateLimitePaiement(*input.DateLimitePaiement)
	}
	if input.DateMajoration != nil {
		create = create.SetDateMajoration(*input.DateMajoration)
	}
	if input.Observations != nil {
		create = create.SetObservations(*input.Observations)
	}

	pvEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create PV", zap.Error(err))
		return nil, fmt.Errorf("failed to create PV: %w", err)
	}

	return pvEnt, nil
}

// GetByID gets PV by ID
func (r *pvRepository) GetByID(ctx context.Context, id string) (*ent.ProcesVerbal, error) {
	uid, _ := uuid.Parse(id)
	pvEnt, err := r.client.ProcesVerbal.
		Query().
		Where(procesverbal.ID(uid)).
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction().WithControle()
		}).
		WithControle().
		WithInspection().
		WithPaiements().
		WithRecours().
		WithDocuments().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("pv not found")
		}
		r.logger.Error("Failed to get PV by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get PV: %w", err)
	}

	return pvEnt, nil
}

// GetByNumeroPV gets PV by numero
func (r *pvRepository) GetByNumeroPV(ctx context.Context, numero string) (*ent.ProcesVerbal, error) {
	pvEnt, err := r.client.ProcesVerbal.
		Query().
		Where(procesverbal.NumeroPv(numero)).
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction()
		}).
		WithControle().
		WithInspection().
		WithPaiements().
		WithRecours().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("pv not found")
		}
		r.logger.Error("Failed to get PV by numero", zap.String("numero", numero), zap.Error(err))
		return nil, fmt.Errorf("failed to get PV: %w", err)
	}

	return pvEnt, nil
}

// List gets PVs with filters
func (r *pvRepository) List(ctx context.Context, filters *PVFilters) ([]*ent.ProcesVerbal, error) {
	query := r.client.ProcesVerbal.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	pvs, err := query.
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction()
		}).
		WithControle().
		WithInspection().
		WithPaiements().
		Order(ent.Desc(procesverbal.FieldDateEmission)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list PVs", zap.Error(err))
		return nil, fmt.Errorf("failed to list PVs: %w", err)
	}

	return pvs, nil
}

// Count counts PVs with filters
func (r *pvRepository) Count(ctx context.Context, filters *PVFilters) (int, error) {
	query := r.client.ProcesVerbal.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count PVs", zap.Error(err))
		return 0, fmt.Errorf("failed to count PVs: %w", err)
	}

	return count, nil
}

// applyFilters applies filters to PV query
func (r *pvRepository) applyFilters(query *ent.ProcesVerbalQuery, filters *PVFilters) *ent.ProcesVerbalQuery {
	if filters.InfractionID != nil {
		infID, _ := uuid.Parse(*filters.InfractionID)
		query = query.Where(procesverbal.HasInfractionsWith(infraction.ID(infID)))
	}
	if filters.Statut != nil {
		query = query.Where(procesverbal.Statut(*filters.Statut))
	}
	if filters.DateDebut != nil {
		query = query.Where(procesverbal.DateEmissionGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(procesverbal.DateEmissionLTE(*filters.DateFin))
	}
	if filters.MontantMin != nil {
		query = query.Where(procesverbal.MontantTotalGTE(*filters.MontantMin))
	}
	if filters.MontantMax != nil {
		query = query.Where(procesverbal.MontantTotalLTE(*filters.MontantMax))
	}
	if filters.AgentID != nil {
		// Filtrer les PVs par agent via la relation: PV -> Controle -> Agent
		agentID, _ := uuid.Parse(*filters.AgentID)
		query = query.Where(procesverbal.HasControleWith(controle.HasAgentWith(user.ID(agentID))))
	}
	if filters.Expired != nil && *filters.Expired {
		query = query.Where(
			procesverbal.And(
				procesverbal.DateLimitePaiementLT(time.Now()),
				procesverbal.StatutNEQ("PAYE"),
				procesverbal.StatutNEQ("ANNULE"),
			),
		)
	}
	return query
}

// Update updates PV
func (r *pvRepository) Update(ctx context.Context, id string, input *UpdatePVInput) (*ent.ProcesVerbal, error) {
	r.logger.Info("Updating PV", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.ProcesVerbal.UpdateOneID(uid)

	if input.MontantTotal != nil {
		update = update.SetMontantTotal(*input.MontantTotal)
	}
	if input.MontantMajore != nil {
		update = update.SetMontantMajore(*input.MontantMajore)
	}
	if input.DateLimitePaiement != nil {
		update = update.SetDateLimitePaiement(*input.DateLimitePaiement)
	}
	if input.DateMajoration != nil {
		update = update.SetDateMajoration(*input.DateMajoration)
	}
	if input.Statut != nil {
		update = update.SetStatut(*input.Statut)
	}
	if input.DatePaiement != nil {
		update = update.SetDatePaiement(*input.DatePaiement)
	}
	if input.MontantPaye != nil {
		update = update.SetMontantPaye(*input.MontantPaye)
	}
	if input.MoyenPaiement != nil {
		update = update.SetMoyenPaiement(*input.MoyenPaiement)
	}
	if input.ReferencePaiement != nil {
		update = update.SetReferencePaiement(*input.ReferencePaiement)
	}
	if input.DateContestation != nil {
		update = update.SetDateContestation(*input.DateContestation)
	}
	if input.MotifContestation != nil {
		update = update.SetMotifContestation(*input.MotifContestation)
	}
	if input.DecisionContestation != nil {
		update = update.SetDecisionContestation(*input.DecisionContestation)
	}
	if input.TribunalCompetent != nil {
		update = update.SetTribunalCompetent(*input.TribunalCompetent)
	}
	if input.Observations != nil {
		update = update.SetObservations(*input.Observations)
	}

	pvEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update PV", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update PV: %w", err)
	}

	return pvEnt, nil
}

// Delete deletes PV
func (r *pvRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting PV", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.ProcesVerbal.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete PV", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete PV: %w", err)
	}

	return nil
}

// GetByInfraction gets PV by infraction ID
func (r *pvRepository) GetByInfraction(ctx context.Context, infractionID string) (*ent.ProcesVerbal, error) {
	infID, _ := uuid.Parse(infractionID)
	pvEnt, err := r.client.ProcesVerbal.Query().
		Where(procesverbal.HasInfractionsWith(infraction.ID(infID))).
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction()
		}).
		WithControle().
		WithInspection().
		WithPaiements().
		WithRecours().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("pv not found")
		}
		r.logger.Error("Failed to get PV by infraction",
			zap.String("infractionID", infractionID), zap.Error(err))
		return nil, fmt.Errorf("failed to get PV by infraction: %w", err)
	}

	return pvEnt, nil
}

// GetByStatut gets PVs by statut
func (r *pvRepository) GetByStatut(ctx context.Context, statut string) ([]*ent.ProcesVerbal, error) {
	pvs, err := r.client.ProcesVerbal.Query().
		Where(procesverbal.Statut(statut)).
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction()
		}).
		WithControle().
		WithInspection().
		WithPaiements().
		Order(ent.Desc(procesverbal.FieldDateEmission)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get PVs by statut",
			zap.String("statut", statut), zap.Error(err))
		return nil, fmt.Errorf("failed to get PVs by statut: %w", err)
	}

	return pvs, nil
}

// GetExpired gets expired PVs
func (r *pvRepository) GetExpired(ctx context.Context) ([]*ent.ProcesVerbal, error) {
	pvs, err := r.client.ProcesVerbal.Query().
		Where(
			procesverbal.And(
				procesverbal.DateLimitePaiementLT(time.Now()),
				procesverbal.StatutNEQ("PAYE"),
				procesverbal.StatutNEQ("ANNULE"),
			),
		).
		WithInfractions(func(q *ent.InfractionQuery) {
			q.WithTypeInfraction()
		}).
		WithControle().
		WithInspection().
		Order(ent.Asc(procesverbal.FieldDateLimitePaiement)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get expired PVs", zap.Error(err))
		return nil, fmt.Errorf("failed to get expired PVs: %w", err)
	}

	return pvs, nil
}

// GetStatistics gets statistics for PVs
func (r *pvRepository) GetStatistics(ctx context.Context, filters *PVFilters) (*PVStatistics, error) {
	query := r.client.ProcesVerbal.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	pvs, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to get PVs for statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get PVs for statistics: %w", err)
	}

	stats := &PVStatistics{
		Total:     len(pvs),
		ParStatut: make(map[string]int),
		ParMois:   make(map[string]float64),
	}

	now := time.Now()
	for _, pv := range pvs {
		stats.MontantTotal += pv.MontantTotal
		stats.MontantPaye += pv.MontantPaye
		stats.ParStatut[pv.Statut]++
		stats.ParMois[pv.DateEmission.Format("2006-01")] += pv.MontantTotal

		// PV expiré
		if !pv.DateLimitePaiement.IsZero() && pv.DateLimitePaiement.Before(now) &&
			pv.Statut != "PAYE" && pv.Statut != "ANNULE" {
			stats.PVExpires++
		}
	}

	stats.MontantImpaye = stats.MontantTotal - stats.MontantPaye
	if stats.MontantTotal > 0 {
		stats.TauxRecouvrement = (stats.MontantPaye / stats.MontantTotal) * 100
	}

	return stats, nil
}
