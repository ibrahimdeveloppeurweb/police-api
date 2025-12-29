package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/procesverbal"
	"police-trafic-api-frontend-aligned/ent/recours"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RecoursRepository defines recours repository interface
type RecoursRepository interface {
	Create(ctx context.Context, input *CreateRecoursInput) (*ent.Recours, error)
	GetByID(ctx context.Context, id string) (*ent.Recours, error)
	GetByNumeroRecours(ctx context.Context, numero string) (*ent.Recours, error)
	List(ctx context.Context, filters *RecoursFilters) ([]*ent.Recours, error)
	Count(ctx context.Context, filters *RecoursFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateRecoursInput) (*ent.Recours, error)
	Delete(ctx context.Context, id string) error
	GetByProcesVerbal(ctx context.Context, pvID string) ([]*ent.Recours, error)
	GetByStatut(ctx context.Context, statut string) ([]*ent.Recours, error)
	GetByTraitePar(ctx context.Context, userID string) ([]*ent.Recours, error)
	GetStatistics(ctx context.Context, filters *RecoursFilters) (*RecoursStatistics, error)
}

// CreateRecoursInput represents input for creating recours
type CreateRecoursInput struct {
	ID                 string
	NumeroRecours      string
	DateRecours        time.Time
	TypeRecours        string
	Motif              string
	Argumentaire       string
	Statut             string
	AutoriteCompetente *string
	DateLimiteRecours  *time.Time
	Observations       *string
	ProcesVerbalID     string
}

// UpdateRecoursInput represents input for updating recours
type UpdateRecoursInput struct {
	TypeRecours        *string
	Motif              *string
	Argumentaire       *string
	Statut             *string
	DateTraitement     *time.Time
	Decision           *string
	MotifDecision      *string
	AutoriteCompetente *string
	ReferenceDecision  *string
	NouveauMontant     *float64
	RecoursPossible    *bool
	Observations       *string
	TraiteParID        *string
}

// RecoursFilters represents filters for listing recours
type RecoursFilters struct {
	ProcesVerbalID *string
	TypeRecours    *string
	Statut         *string
	TraiteParID    *string
	DateDebut      *time.Time
	DateFin        *time.Time
	Limit          int
	Offset         int
}

// RecoursStatistics represents statistics for recours
type RecoursStatistics struct {
	Total           int            `json:"total"`
	ParStatut       map[string]int `json:"par_statut"`
	ParType         map[string]int `json:"par_type"`
	ParDecision     map[string]int `json:"par_decision"`
	TauxAcceptation float64        `json:"taux_acceptation"`
	DelaiMoyenJours float64        `json:"delai_moyen_jours"`
}

// recoursRepository implements RecoursRepository
type recoursRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewRecoursRepository creates a new recours repository
func NewRecoursRepository(client *ent.Client, logger *zap.Logger) RecoursRepository {
	return &recoursRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new recours
func (r *recoursRepository) Create(ctx context.Context, input *CreateRecoursInput) (*ent.Recours, error) {
	r.logger.Info("Creating recours",
		zap.String("numero_recours", input.NumeroRecours),
		zap.String("type", input.TypeRecours))

	id, _ := uuid.Parse(input.ID)
	pvID, _ := uuid.Parse(input.ProcesVerbalID)
	create := r.client.Recours.Create().
		SetID(id).
		SetNumeroRecours(input.NumeroRecours).
		SetDateRecours(input.DateRecours).
		SetTypeRecours(input.TypeRecours).
		SetMotif(input.Motif).
		SetArgumentaire(input.Argumentaire).
		SetStatut(input.Statut).
		SetProcesVerbalID(pvID)

	if input.AutoriteCompetente != nil {
		create = create.SetAutoriteCompetente(*input.AutoriteCompetente)
	}
	if input.DateLimiteRecours != nil {
		create = create.SetDateLimiteRecours(*input.DateLimiteRecours)
	}
	if input.Observations != nil {
		create = create.SetObservations(*input.Observations)
	}

	recoursEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create recours", zap.Error(err))
		return nil, fmt.Errorf("failed to create recours: %w", err)
	}

	return recoursEnt, nil
}

// GetByID gets recours by ID
func (r *recoursRepository) GetByID(ctx context.Context, id string) (*ent.Recours, error) {
	uid, _ := uuid.Parse(id)
	recoursEnt, err := r.client.Recours.
		Query().
		Where(recours.ID(uid)).
		WithProcesVerbal(func(q *ent.ProcesVerbalQuery) {
			q.WithInfractions()
		}).
		WithDocuments().
		WithTraitePar().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("recours not found")
		}
		r.logger.Error("Failed to get recours by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get recours: %w", err)
	}

	return recoursEnt, nil
}

// GetByNumeroRecours gets recours by numero
func (r *recoursRepository) GetByNumeroRecours(ctx context.Context, numero string) (*ent.Recours, error) {
	recoursEnt, err := r.client.Recours.
		Query().
		Where(recours.NumeroRecours(numero)).
		WithProcesVerbal(func(q *ent.ProcesVerbalQuery) {
			q.WithInfractions()
		}).
		WithDocuments().
		WithTraitePar().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("recours not found")
		}
		r.logger.Error("Failed to get recours by numero", zap.String("numero", numero), zap.Error(err))
		return nil, fmt.Errorf("failed to get recours: %w", err)
	}

	return recoursEnt, nil
}

// List gets recours with filters
func (r *recoursRepository) List(ctx context.Context, filters *RecoursFilters) ([]*ent.Recours, error) {
	query := r.client.Recours.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	recoursList, err := query.
		WithProcesVerbal().
		WithTraitePar().
		Order(ent.Desc(recours.FieldDateRecours)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list recours", zap.Error(err))
		return nil, fmt.Errorf("failed to list recours: %w", err)
	}

	return recoursList, nil
}

// Count counts recours with filters
func (r *recoursRepository) Count(ctx context.Context, filters *RecoursFilters) (int, error) {
	query := r.client.Recours.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count recours", zap.Error(err))
		return 0, fmt.Errorf("failed to count recours: %w", err)
	}

	return count, nil
}

// applyFilters applies filters to recours query
func (r *recoursRepository) applyFilters(query *ent.RecoursQuery, filters *RecoursFilters) *ent.RecoursQuery {
	if filters.ProcesVerbalID != nil {
		pvID, _ := uuid.Parse(*filters.ProcesVerbalID)
		query = query.Where(recours.HasProcesVerbalWith(procesverbal.ID(pvID)))
	}
	if filters.TypeRecours != nil {
		query = query.Where(recours.TypeRecours(*filters.TypeRecours))
	}
	if filters.Statut != nil {
		query = query.Where(recours.Statut(*filters.Statut))
	}
	if filters.TraiteParID != nil {
		trID, _ := uuid.Parse(*filters.TraiteParID)
		query = query.Where(recours.HasTraiteParWith(user.ID(trID)))
	}
	if filters.DateDebut != nil {
		query = query.Where(recours.DateRecoursGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(recours.DateRecoursLTE(*filters.DateFin))
	}
	return query
}

// Update updates recours
func (r *recoursRepository) Update(ctx context.Context, id string, input *UpdateRecoursInput) (*ent.Recours, error) {
	r.logger.Info("Updating recours", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Recours.UpdateOneID(uid)

	if input.TypeRecours != nil {
		update = update.SetTypeRecours(*input.TypeRecours)
	}
	if input.Motif != nil {
		update = update.SetMotif(*input.Motif)
	}
	if input.Argumentaire != nil {
		update = update.SetArgumentaire(*input.Argumentaire)
	}
	if input.Statut != nil {
		update = update.SetStatut(*input.Statut)
	}
	if input.DateTraitement != nil {
		update = update.SetDateTraitement(*input.DateTraitement)
	}
	if input.Decision != nil {
		update = update.SetDecision(*input.Decision)
	}
	if input.MotifDecision != nil {
		update = update.SetMotifDecision(*input.MotifDecision)
	}
	if input.AutoriteCompetente != nil {
		update = update.SetAutoriteCompetente(*input.AutoriteCompetente)
	}
	if input.ReferenceDecision != nil {
		update = update.SetReferenceDecision(*input.ReferenceDecision)
	}
	if input.NouveauMontant != nil {
		update = update.SetNouveauMontant(*input.NouveauMontant)
	}
	if input.RecoursPossible != nil {
		update = update.SetRecoursPossible(*input.RecoursPossible)
	}
	if input.Observations != nil {
		update = update.SetObservations(*input.Observations)
	}
	if input.TraiteParID != nil {
		trID, _ := uuid.Parse(*input.TraiteParID)
		update = update.SetTraiteParID(trID)
	}

	recoursEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update recours", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update recours: %w", err)
	}

	return recoursEnt, nil
}

// Delete deletes recours
func (r *recoursRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting recours", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Recours.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete recours", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete recours: %w", err)
	}

	return nil
}

// GetByProcesVerbal gets recours by PV ID
func (r *recoursRepository) GetByProcesVerbal(ctx context.Context, pvID string) ([]*ent.Recours, error) {
	uid, _ := uuid.Parse(pvID)
	recoursList, err := r.client.Recours.Query().
		Where(recours.HasProcesVerbalWith(procesverbal.ID(uid))).
		WithProcesVerbal().
		WithTraitePar().
		Order(ent.Desc(recours.FieldDateRecours)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get recours by PV",
			zap.String("pvID", pvID), zap.Error(err))
		return nil, fmt.Errorf("failed to get recours by PV: %w", err)
	}

	return recoursList, nil
}

// GetByStatut gets recours by statut
func (r *recoursRepository) GetByStatut(ctx context.Context, statut string) ([]*ent.Recours, error) {
	recoursList, err := r.client.Recours.Query().
		Where(recours.Statut(statut)).
		WithProcesVerbal().
		WithTraitePar().
		Order(ent.Desc(recours.FieldDateRecours)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get recours by statut",
			zap.String("statut", statut), zap.Error(err))
		return nil, fmt.Errorf("failed to get recours by statut: %w", err)
	}

	return recoursList, nil
}

// GetByTraitePar gets recours by user who processed them
func (r *recoursRepository) GetByTraitePar(ctx context.Context, userID string) ([]*ent.Recours, error) {
	uid, _ := uuid.Parse(userID)
	recoursList, err := r.client.Recours.Query().
		Where(recours.HasTraiteParWith(user.ID(uid))).
		WithProcesVerbal().
		WithTraitePar().
		Order(ent.Desc(recours.FieldDateRecours)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get recours by traite_par",
			zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get recours by traite_par: %w", err)
	}

	return recoursList, nil
}

// GetStatistics gets statistics for recours
func (r *recoursRepository) GetStatistics(ctx context.Context, filters *RecoursFilters) (*RecoursStatistics, error) {
	query := r.client.Recours.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	recoursList, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to get recours for statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get recours for statistics: %w", err)
	}

	stats := &RecoursStatistics{
		Total:       len(recoursList),
		ParStatut:   make(map[string]int),
		ParType:     make(map[string]int),
		ParDecision: make(map[string]int),
	}

	var totalDelai float64
	var countWithDelai int
	var acceptes int

	for _, rec := range recoursList {
		stats.ParStatut[rec.Statut]++
		stats.ParType[rec.TypeRecours]++

		if rec.Decision != "" {
			stats.ParDecision[rec.Decision]++
			if rec.Decision == "ACCEPTE" {
				acceptes++
			}
		}

		// Calculer le délai de traitement
		if !rec.DateTraitement.IsZero() {
			delai := rec.DateTraitement.Sub(rec.DateRecours).Hours() / 24
			totalDelai += delai
			countWithDelai++
		}
	}

	// Taux d'acceptation
	if stats.Total > 0 {
		stats.TauxAcceptation = (float64(acceptes) / float64(stats.Total)) * 100
	}

	// Délai moyen
	if countWithDelai > 0 {
		stats.DelaiMoyenJours = totalDelai / float64(countWithDelai)
	}

	return stats, nil
}
