package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/paiement"
	"police-trafic-api-frontend-aligned/ent/procesverbal"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PaiementRepository defines paiement repository interface
type PaiementRepository interface {
	Create(ctx context.Context, input *CreatePaiementInput) (*ent.Paiement, error)
	GetByID(ctx context.Context, id string) (*ent.Paiement, error)
	GetByNumeroTransaction(ctx context.Context, numero string) (*ent.Paiement, error)
	List(ctx context.Context, filters *PaiementFilters) ([]*ent.Paiement, error)
	Count(ctx context.Context, filters *PaiementFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdatePaiementInput) (*ent.Paiement, error)
	Delete(ctx context.Context, id string) error
	GetByProcesVerbal(ctx context.Context, pvID string) ([]*ent.Paiement, error)
	GetByStatut(ctx context.Context, statut string) ([]*ent.Paiement, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.Paiement, error)
	GetStatistics(ctx context.Context, filters *PaiementFilters) (*PaiementStatistics, error)
}

// CreatePaiementInput represents input for creating paiement
type CreatePaiementInput struct {
	ID                string
	NumeroTransaction string
	DatePaiement      time.Time
	Montant           float64
	MoyenPaiement     string
	ReferenceExterne  *string
	Statut            string
	CodeAutorisation  *string
	DetailsPaiement   *string
	ProcesVerbalID    string
}

// UpdatePaiementInput represents input for updating paiement
type UpdatePaiementInput struct {
	Statut           *string
	ReferenceExterne *string
	CodeAutorisation *string
	DateValidation   *time.Time
	MotifRefus       *string
	DetailsPaiement  *string
}

// PaiementFilters represents filters for listing paiements
type PaiementFilters struct {
	ProcesVerbalID *string
	Statut         *string
	MoyenPaiement  *string
	DateDebut      *time.Time
	DateFin        *time.Time
	MontantMin     *float64
	MontantMax     *float64
	Limit          int
	Offset         int
}

// PaiementStatistics represents statistics for paiements
type PaiementStatistics struct {
	Total              int                `json:"total"`
	MontantTotal       float64            `json:"montant_total"`
	MontantValide      float64            `json:"montant_valide"`
	MontantEnCours     float64            `json:"montant_en_cours"`
	MontantRembourse   float64            `json:"montant_rembourse"`
	ParStatut          map[string]int     `json:"par_statut"`
	ParMoyenPaiement   map[string]float64 `json:"par_moyen_paiement"`
	EvolutionMensuelle []MontantMensuel   `json:"evolution_mensuelle"`
}

// MontantMensuel represents monthly amount
type MontantMensuel struct {
	Mois    string  `json:"mois"`
	Montant float64 `json:"montant"`
	Nombre  int     `json:"nombre"`
}

// paiementRepository implements PaiementRepository
type paiementRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewPaiementRepository creates a new paiement repository
func NewPaiementRepository(client *ent.Client, logger *zap.Logger) PaiementRepository {
	return &paiementRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new paiement
func (r *paiementRepository) Create(ctx context.Context, input *CreatePaiementInput) (*ent.Paiement, error) {
	r.logger.Info("Creating paiement",
		zap.String("numero_transaction", input.NumeroTransaction),
		zap.Float64("montant", input.Montant))

	id, _ := uuid.Parse(input.ID)
	pvID, _ := uuid.Parse(input.ProcesVerbalID)
	create := r.client.Paiement.Create().
		SetID(id).
		SetNumeroTransaction(input.NumeroTransaction).
		SetDatePaiement(input.DatePaiement).
		SetMontant(input.Montant).
		SetMoyenPaiement(input.MoyenPaiement).
		SetStatut(input.Statut).
		SetProcesVerbalID(pvID)

	if input.ReferenceExterne != nil {
		create = create.SetReferenceExterne(*input.ReferenceExterne)
	}
	if input.CodeAutorisation != nil {
		create = create.SetCodeAutorisation(*input.CodeAutorisation)
	}
	if input.DetailsPaiement != nil {
		create = create.SetDetailsPaiement(*input.DetailsPaiement)
	}

	paiementEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create paiement", zap.Error(err))
		return nil, fmt.Errorf("failed to create paiement: %w", err)
	}

	return paiementEnt, nil
}

// GetByID gets paiement by ID
func (r *paiementRepository) GetByID(ctx context.Context, id string) (*ent.Paiement, error) {
	uid, _ := uuid.Parse(id)
	paiementEnt, err := r.client.Paiement.
		Query().
		Where(paiement.ID(uid)).
		WithProcesVerbal().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("paiement not found")
		}
		r.logger.Error("Failed to get paiement by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get paiement: %w", err)
	}

	return paiementEnt, nil
}

// GetByNumeroTransaction gets paiement by transaction number
func (r *paiementRepository) GetByNumeroTransaction(ctx context.Context, numero string) (*ent.Paiement, error) {
	paiementEnt, err := r.client.Paiement.
		Query().
		Where(paiement.NumeroTransaction(numero)).
		WithProcesVerbal().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("paiement not found")
		}
		r.logger.Error("Failed to get paiement by numero", zap.String("numero", numero), zap.Error(err))
		return nil, fmt.Errorf("failed to get paiement: %w", err)
	}

	return paiementEnt, nil
}

// List gets paiements with filters
func (r *paiementRepository) List(ctx context.Context, filters *PaiementFilters) ([]*ent.Paiement, error) {
	query := r.client.Paiement.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	paiements, err := query.
		WithProcesVerbal().
		Order(ent.Desc(paiement.FieldDatePaiement)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list paiements", zap.Error(err))
		return nil, fmt.Errorf("failed to list paiements: %w", err)
	}

	return paiements, nil
}

// Count counts paiements with filters
func (r *paiementRepository) Count(ctx context.Context, filters *PaiementFilters) (int, error) {
	query := r.client.Paiement.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count paiements", zap.Error(err))
		return 0, fmt.Errorf("failed to count paiements: %w", err)
	}

	return count, nil
}

// applyFilters applies filters to paiement query
func (r *paiementRepository) applyFilters(query *ent.PaiementQuery, filters *PaiementFilters) *ent.PaiementQuery {
	if filters.ProcesVerbalID != nil {
		pvID, _ := uuid.Parse(*filters.ProcesVerbalID)
		query = query.Where(paiement.HasProcesVerbalWith(procesverbal.ID(pvID)))
	}
	if filters.Statut != nil {
		query = query.Where(paiement.Statut(*filters.Statut))
	}
	if filters.MoyenPaiement != nil {
		query = query.Where(paiement.MoyenPaiement(*filters.MoyenPaiement))
	}
	if filters.DateDebut != nil {
		query = query.Where(paiement.DatePaiementGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(paiement.DatePaiementLTE(*filters.DateFin))
	}
	if filters.MontantMin != nil {
		query = query.Where(paiement.MontantGTE(*filters.MontantMin))
	}
	if filters.MontantMax != nil {
		query = query.Where(paiement.MontantLTE(*filters.MontantMax))
	}
	return query
}

// Update updates paiement
func (r *paiementRepository) Update(ctx context.Context, id string, input *UpdatePaiementInput) (*ent.Paiement, error) {
	r.logger.Info("Updating paiement", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Paiement.UpdateOneID(uid)

	if input.Statut != nil {
		update = update.SetStatut(*input.Statut)
	}
	if input.ReferenceExterne != nil {
		update = update.SetReferenceExterne(*input.ReferenceExterne)
	}
	if input.CodeAutorisation != nil {
		update = update.SetCodeAutorisation(*input.CodeAutorisation)
	}
	if input.DateValidation != nil {
		update = update.SetDateValidation(*input.DateValidation)
	}
	if input.MotifRefus != nil {
		update = update.SetMotifRefus(*input.MotifRefus)
	}
	if input.DetailsPaiement != nil {
		update = update.SetDetailsPaiement(*input.DetailsPaiement)
	}

	paiementEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update paiement", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update paiement: %w", err)
	}

	return paiementEnt, nil
}

// Delete deletes paiement
func (r *paiementRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting paiement", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Paiement.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete paiement", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete paiement: %w", err)
	}

	return nil
}

// GetByProcesVerbal gets paiements by proces verbal ID
func (r *paiementRepository) GetByProcesVerbal(ctx context.Context, pvID string) ([]*ent.Paiement, error) {
	uid, _ := uuid.Parse(pvID)
	paiements, err := r.client.Paiement.Query().
		Where(paiement.HasProcesVerbalWith(procesverbal.ID(uid))).
		WithProcesVerbal().
		Order(ent.Desc(paiement.FieldDatePaiement)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get paiements by PV",
			zap.String("pvID", pvID), zap.Error(err))
		return nil, fmt.Errorf("failed to get paiements by PV: %w", err)
	}

	return paiements, nil
}

// GetByStatut gets paiements by statut
func (r *paiementRepository) GetByStatut(ctx context.Context, statut string) ([]*ent.Paiement, error) {
	paiements, err := r.client.Paiement.Query().
		Where(paiement.Statut(statut)).
		WithProcesVerbal().
		Order(ent.Desc(paiement.FieldDatePaiement)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get paiements by statut",
			zap.String("statut", statut), zap.Error(err))
		return nil, fmt.Errorf("failed to get paiements by statut: %w", err)
	}

	return paiements, nil
}

// GetByDateRange gets paiements by date range
func (r *paiementRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.Paiement, error) {
	paiements, err := r.client.Paiement.Query().
		Where(
			paiement.And(
				paiement.DatePaiementGTE(start),
				paiement.DatePaiementLTE(end),
			),
		).
		WithProcesVerbal().
		Order(ent.Desc(paiement.FieldDatePaiement)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get paiements by date range",
			zap.Time("start", start), zap.Time("end", end), zap.Error(err))
		return nil, fmt.Errorf("failed to get paiements by date range: %w", err)
	}

	return paiements, nil
}

// GetStatistics gets statistics for paiements
func (r *paiementRepository) GetStatistics(ctx context.Context, filters *PaiementFilters) (*PaiementStatistics, error) {
	query := r.client.Paiement.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	paiements, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to get paiements for statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get paiements for statistics: %w", err)
	}

	stats := &PaiementStatistics{
		Total:            len(paiements),
		ParStatut:        make(map[string]int),
		ParMoyenPaiement: make(map[string]float64),
	}

	// Map pour calculer l'évolution mensuelle
	mensuelMap := make(map[string]*MontantMensuel)

	for _, p := range paiements {
		stats.MontantTotal += p.Montant

		// Par statut
		stats.ParStatut[p.Statut]++

		// Par moyen de paiement
		stats.ParMoyenPaiement[p.MoyenPaiement] += p.Montant

		// Montants par statut
		switch p.Statut {
		case "VALIDE":
			stats.MontantValide += p.Montant
		case "EN_COURS":
			stats.MontantEnCours += p.Montant
		case "REMBOURSE":
			stats.MontantRembourse += p.Montant
		}

		// Evolution mensuelle
		mois := p.DatePaiement.Format("2006-01")
		if _, ok := mensuelMap[mois]; !ok {
			mensuelMap[mois] = &MontantMensuel{Mois: mois}
		}
		mensuelMap[mois].Montant += p.Montant
		mensuelMap[mois].Nombre++
	}

	// Convertir map en slice pour l'évolution mensuelle
	for _, m := range mensuelMap {
		stats.EvolutionMensuelle = append(stats.EvolutionMensuelle, *m)
	}

	return stats, nil
}
