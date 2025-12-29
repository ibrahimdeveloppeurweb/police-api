package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/conducteur"
	"police-trafic-api-frontend-aligned/ent/controle"
	"police-trafic-api-frontend-aligned/ent/infraction"
	"police-trafic-api-frontend-aligned/ent/infractiontype"
	"police-trafic-api-frontend-aligned/ent/user"
	"police-trafic-api-frontend-aligned/ent/vehicule"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InfractionRepository defines infraction repository interface
type InfractionRepository interface {
	Create(ctx context.Context, input *CreateInfractionInput) (*ent.Infraction, error)
	GetByID(ctx context.Context, id string) (*ent.Infraction, error)
	GetByNumeroPV(ctx context.Context, numeroPV string) (*ent.Infraction, error)
	List(ctx context.Context, filters *InfractionFilters) ([]*ent.Infraction, error)
	Update(ctx context.Context, id string, input *UpdateInfractionInput) (*ent.Infraction, error)
	Delete(ctx context.Context, id string) error
	GetByControle(ctx context.Context, controleID string) ([]*ent.Infraction, error)
	GetByVehicule(ctx context.Context, vehiculeID string) ([]*ent.Infraction, error)
	GetByConducteur(ctx context.Context, conducteurID string) ([]*ent.Infraction, error)
	GetByStatut(ctx context.Context, statut string) ([]*ent.Infraction, error)
	GetStatistics(ctx context.Context, filters *InfractionStatsFilters) (*InfractionStatistics, error)
}

// CreateInfractionInput represents input for creating infraction
type CreateInfractionInput struct {
	ID                   string
	NumeroPV             *string
	DateInfraction       time.Time
	LieuInfraction       string
	Circonstances        *string
	VitesseRetenue       *float64
	VitesseLimitee       *float64
	AppareilMesure       *string
	MontantAmende        float64
	PointsRetires        int
	Statut               string
	Observations         *string
	FlagrantDelit        bool
	Accident             bool
	ControleID           string
	TypeInfractionID     string
	VehiculeID           string
	ConducteurID         string
}

// UpdateInfractionInput represents input for updating infraction
type UpdateInfractionInput struct {
	NumeroPV             *string
	DateInfraction       *time.Time
	LieuInfraction       *string
	Circonstances        *string
	VitesseRetenue       *float64
	VitesseLimitee       *float64
	AppareilMesure       *string
	MontantAmende        *float64
	PointsRetires        *int
	Statut               *string
	Observations         *string
	FlagrantDelit        *bool
	Accident             *bool
	TypeInfractionID     *string
}

// InfractionFilters represents filters for listing infractions
type InfractionFilters struct {
	ControleID       *string
	VehiculeID       *string
	ConducteurID     *string
	AgentID          *string
	TypeInfractionID *string
	Statut           *string
	LieuInfraction   *string
	DateDebut        *time.Time
	DateFin          *time.Time
	FlagrantDelit    *bool
	Accident         *bool
	Limit            int
	Offset           int
}

// InfractionStatsFilters represents filters for statistics
type InfractionStatsFilters struct {
	DateDebut    *time.Time
	DateFin      *time.Time
	AgentID      *string
	TypeControle *string
}

// InfractionStatistics represents statistics for infractions
type InfractionStatistics struct {
	Total              int                     `json:"total"`
	ParStatut          map[string]int          `json:"par_statut"`
	ParType            map[string]int          `json:"par_type"`
	ParMois            map[string]int          `json:"par_mois"`
	MontantTotal       float64                 `json:"montant_total"`
	PointsTotal        int                     `json:"points_total"`
	FlagrantDelitTotal int                     `json:"flagrant_delit_total"`
	AccidentTotal      int                     `json:"accident_total"`
	TopInfractions     []InfractionTypeStats   `json:"top_infractions"`
}

// InfractionTypeStats represents statistics by infraction type
type InfractionTypeStats struct {
	TypeCode     string  `json:"type_code"`
	TypeLibelle  string  `json:"type_libelle"`
	Count        int     `json:"count"`
	MontantTotal float64 `json:"montant_total"`
}

// infractionRepository implements InfractionRepository
type infractionRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewInfractionRepository creates a new infraction repository
func NewInfractionRepository(client *ent.Client, logger *zap.Logger) InfractionRepository {
	return &infractionRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new infraction
func (r *infractionRepository) Create(ctx context.Context, input *CreateInfractionInput) (*ent.Infraction, error) {
	r.logger.Info("Creating infraction",
		zap.String("lieu", input.LieuInfraction), zap.String("controle_id", input.ControleID))

	id, _ := uuid.Parse(input.ID)
	controleID, _ := uuid.Parse(input.ControleID)
	typeInfractionID, _ := uuid.Parse(input.TypeInfractionID)
	vehiculeID, _ := uuid.Parse(input.VehiculeID)
	conducteurID, _ := uuid.Parse(input.ConducteurID)

	create := r.client.Infraction.Create().
		SetID(id).
		SetDateInfraction(input.DateInfraction).
		SetLieuInfraction(input.LieuInfraction).
		SetMontantAmende(input.MontantAmende).
		SetPointsRetires(input.PointsRetires).
		SetStatut(input.Statut).
		SetFlagrantDelit(input.FlagrantDelit).
		SetAccident(input.Accident).
		SetControleID(controleID).
		SetTypeInfractionID(typeInfractionID).
		SetVehiculeID(vehiculeID).
		SetConducteurID(conducteurID)

	if input.NumeroPV != nil {
		create = create.SetNumeroPv(*input.NumeroPV)
	}
	if input.Circonstances != nil {
		create = create.SetCirconstances(*input.Circonstances)
	}
	if input.VitesseRetenue != nil {
		create = create.SetVitesseRetenue(*input.VitesseRetenue)
	}
	if input.VitesseLimitee != nil {
		create = create.SetVitesseLimitee(*input.VitesseLimitee)
	}
	if input.AppareilMesure != nil {
		create = create.SetAppareilMesure(*input.AppareilMesure)
	}
	if input.Observations != nil {
		create = create.SetObservations(*input.Observations)
	}

	infractionEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create infraction", zap.Error(err))
		return nil, fmt.Errorf("failed to create infraction: %w", err)
	}

	return infractionEnt, nil
}

// GetByID gets infraction by ID
func (r *infractionRepository) GetByID(ctx context.Context, id string) (*ent.Infraction, error) {
	uid, _ := uuid.Parse(id)
	infractionEnt, err := r.client.Infraction.
		Query().
		Where(infraction.ID(uid)).
		WithControle(func(q *ent.ControleQuery) {
			q.WithAgent()
		}).
		WithTypeInfraction().
		WithVehicule().
		WithConducteur().
		WithProcesVerbal(func(q *ent.ProcesVerbalQuery) {
			q.WithPaiements().WithRecours()
		}).
		WithDocuments().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("infraction not found")
		}
		r.logger.Error("Failed to get infraction by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get infraction: %w", err)
	}

	return infractionEnt, nil
}

// GetByNumeroPV gets infraction by numero PV
func (r *infractionRepository) GetByNumeroPV(ctx context.Context, numeroPV string) (*ent.Infraction, error) {
	infractionEnt, err := r.client.Infraction.
		Query().
		Where(infraction.NumeroPv(numeroPV)).
		WithControle(func(q *ent.ControleQuery) {
			q.WithAgent()
		}).
		WithTypeInfraction().
		WithVehicule().
		WithConducteur().
		WithProcesVerbal().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("infraction not found")
		}
		r.logger.Error("Failed to get infraction by numero PV", 
			zap.String("numeroPV", numeroPV), zap.Error(err))
		return nil, fmt.Errorf("failed to get infraction: %w", err)
	}

	return infractionEnt, nil
}

// List gets infractions with filters
func (r *infractionRepository) List(ctx context.Context, filters *InfractionFilters) ([]*ent.Infraction, error) {
	query := r.client.Infraction.Query()

	if filters != nil {
		if filters.ControleID != nil {
			ctrlID, _ := uuid.Parse(*filters.ControleID)
			query = query.Where(infraction.HasControleWith(controle.ID(ctrlID)))
		}
		if filters.VehiculeID != nil {
			vehID, _ := uuid.Parse(*filters.VehiculeID)
			query = query.Where(infraction.HasVehiculeWith(vehicule.ID(vehID)))
		}
		if filters.ConducteurID != nil {
			condID, _ := uuid.Parse(*filters.ConducteurID)
			query = query.Where(infraction.HasConducteurWith(conducteur.ID(condID)))
		}
		if filters.TypeInfractionID != nil {
			typeID, _ := uuid.Parse(*filters.TypeInfractionID)
			query = query.Where(infraction.HasTypeInfractionWith(infractiontype.ID(typeID)))
		}
		if filters.Statut != nil {
			query = query.Where(infraction.Statut(*filters.Statut))
		}
		if filters.LieuInfraction != nil {
			query = query.Where(infraction.LieuInfractionContains(*filters.LieuInfraction))
		}
		if filters.DateDebut != nil {
			query = query.Where(infraction.DateInfractionGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(infraction.DateInfractionLTE(*filters.DateFin))
		}
		if filters.FlagrantDelit != nil {
			query = query.Where(infraction.FlagrantDelit(*filters.FlagrantDelit))
		}
		if filters.Accident != nil {
			query = query.Where(infraction.Accident(*filters.Accident))
		}

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	infractions, err := query.
		WithControle(func(q *ent.ControleQuery) {
			q.WithAgent()
		}).
		WithTypeInfraction().
		WithVehicule().
		WithConducteur().
		WithProcesVerbal().
		Order(ent.Desc(infraction.FieldDateInfraction)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list infractions", zap.Error(err))
		return nil, fmt.Errorf("failed to list infractions: %w", err)
	}

	return infractions, nil
}

// Update updates infraction
func (r *infractionRepository) Update(ctx context.Context, id string, input *UpdateInfractionInput) (*ent.Infraction, error) {
	r.logger.Info("Updating infraction", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Infraction.UpdateOneID(uid)

	if input.NumeroPV != nil {
		update = update.SetNumeroPv(*input.NumeroPV)
	}
	if input.DateInfraction != nil {
		update = update.SetDateInfraction(*input.DateInfraction)
	}
	if input.LieuInfraction != nil {
		update = update.SetLieuInfraction(*input.LieuInfraction)
	}
	if input.Circonstances != nil {
		update = update.SetCirconstances(*input.Circonstances)
	}
	if input.VitesseRetenue != nil {
		update = update.SetVitesseRetenue(*input.VitesseRetenue)
	}
	if input.VitesseLimitee != nil {
		update = update.SetVitesseLimitee(*input.VitesseLimitee)
	}
	if input.AppareilMesure != nil {
		update = update.SetAppareilMesure(*input.AppareilMesure)
	}
	if input.MontantAmende != nil {
		update = update.SetMontantAmende(*input.MontantAmende)
	}
	if input.PointsRetires != nil {
		update = update.SetPointsRetires(*input.PointsRetires)
	}
	if input.Statut != nil {
		update = update.SetStatut(*input.Statut)
	}
	if input.Observations != nil {
		update = update.SetObservations(*input.Observations)
	}
	if input.FlagrantDelit != nil {
		update = update.SetFlagrantDelit(*input.FlagrantDelit)
	}
	if input.Accident != nil {
		update = update.SetAccident(*input.Accident)
	}
	if input.TypeInfractionID != nil {
		typeID, _ := uuid.Parse(*input.TypeInfractionID)
		update = update.SetTypeInfractionID(typeID)
	}

	infractionEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update infraction", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update infraction: %w", err)
	}

	return infractionEnt, nil
}

// Delete deletes infraction
func (r *infractionRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting infraction", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Infraction.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete infraction", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete infraction: %w", err)
	}

	return nil
}

// GetByControle gets infractions by controle
func (r *infractionRepository) GetByControle(ctx context.Context, controleID string) ([]*ent.Infraction, error) {
	ctrlID, _ := uuid.Parse(controleID)
	infractions, err := r.client.Infraction.Query().
		Where(infraction.HasControleWith(controle.ID(ctrlID))).
		WithTypeInfraction().
		WithVehicule().
		WithConducteur().
		WithProcesVerbal().
		Order(ent.Desc(infraction.FieldDateInfraction)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get infractions by controle", 
			zap.String("controleID", controleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get infractions by controle: %w", err)
	}

	return infractions, nil
}

// GetByVehicule gets infractions by vehicule
func (r *infractionRepository) GetByVehicule(ctx context.Context, vehiculeID string) ([]*ent.Infraction, error) {
	vehID, _ := uuid.Parse(vehiculeID)
	infractions, err := r.client.Infraction.Query().
		Where(infraction.HasVehiculeWith(vehicule.ID(vehID))).
		WithControle().
		WithTypeInfraction().
		WithConducteur().
		WithProcesVerbal().
		Order(ent.Desc(infraction.FieldDateInfraction)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get infractions by vehicule", 
			zap.String("vehiculeID", vehiculeID), zap.Error(err))
		return nil, fmt.Errorf("failed to get infractions by vehicule: %w", err)
	}

	return infractions, nil
}

// GetByConducteur gets infractions by conducteur
func (r *infractionRepository) GetByConducteur(ctx context.Context, conducteurID string) ([]*ent.Infraction, error) {
	condID, _ := uuid.Parse(conducteurID)
	infractions, err := r.client.Infraction.Query().
		Where(infraction.HasConducteurWith(conducteur.ID(condID))).
		WithControle().
		WithTypeInfraction().
		WithVehicule().
		WithProcesVerbal().
		Order(ent.Desc(infraction.FieldDateInfraction)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get infractions by conducteur", 
			zap.String("conducteurID", conducteurID), zap.Error(err))
		return nil, fmt.Errorf("failed to get infractions by conducteur: %w", err)
	}

	return infractions, nil
}

// GetByStatut gets infractions by statut
func (r *infractionRepository) GetByStatut(ctx context.Context, statut string) ([]*ent.Infraction, error) {
	infractions, err := r.client.Infraction.Query().
		Where(infraction.Statut(statut)).
		WithControle().
		WithTypeInfraction().
		WithVehicule().
		WithConducteur().
		WithProcesVerbal().
		Order(ent.Desc(infraction.FieldDateInfraction)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get infractions by statut", 
			zap.String("statut", statut), zap.Error(err))
		return nil, fmt.Errorf("failed to get infractions by statut: %w", err)
	}

	return infractions, nil
}

// GetStatistics gets statistics for infractions
func (r *infractionRepository) GetStatistics(ctx context.Context, filters *InfractionStatsFilters) (*InfractionStatistics, error) {
	query := r.client.Infraction.Query()

	if filters != nil {
		if filters.DateDebut != nil {
			query = query.Where(infraction.DateInfractionGTE(*filters.DateDebut))
		}
		if filters.DateFin != nil {
			query = query.Where(infraction.DateInfractionLTE(*filters.DateFin))
		}
		if filters.AgentID != nil {
			agentID, _ := uuid.Parse(*filters.AgentID)
			query = query.Where(infraction.HasControleWith(controle.HasAgentWith(user.ID(agentID))))
		}
	}

	infractions, err := query.WithTypeInfraction().All(ctx)
	if err != nil {
		r.logger.Error("Failed to get infractions for statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to get infractions for statistics: %w", err)
	}

	stats := &InfractionStatistics{
		Total:          len(infractions),
		ParStatut:      make(map[string]int),
		ParType:        make(map[string]int),
		ParMois:        make(map[string]int),
		TopInfractions: make([]InfractionTypeStats, 0),
	}

	typeStats := make(map[string]*InfractionTypeStats)

	for _, inf := range infractions {
		// Statuts
		stats.ParStatut[inf.Statut]++

		// Types
		if inf.Edges.TypeInfraction != nil {
			typeCode := inf.Edges.TypeInfraction.Code
			stats.ParType[typeCode]++

			if _, exists := typeStats[typeCode]; !exists {
				typeStats[typeCode] = &InfractionTypeStats{
					TypeCode:    typeCode,
					TypeLibelle: inf.Edges.TypeInfraction.Libelle,
					Count:       0,
				}
			}
			typeStats[typeCode].Count++
			typeStats[typeCode].MontantTotal += inf.MontantAmende
		}

		// Par mois
		mois := inf.DateInfraction.Format("2006-01")
		stats.ParMois[mois]++

		// Totaux
		stats.MontantTotal += inf.MontantAmende
		stats.PointsTotal += inf.PointsRetires

		if inf.FlagrantDelit {
			stats.FlagrantDelitTotal++
		}
		if inf.Accident {
			stats.AccidentTotal++
		}
	}

	// Convertir typeStats en slice pour TopInfractions
	for _, ts := range typeStats {
		stats.TopInfractions = append(stats.TopInfractions, *ts)
	}

	return stats, nil
}