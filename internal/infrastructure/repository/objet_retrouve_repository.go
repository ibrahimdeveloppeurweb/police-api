package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/commissariat"
	"police-trafic-api-frontend-aligned/ent/objetretrouve"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ObjetRetrouveRepository defines objet retrouve repository interface
type ObjetRetrouveRepository interface {
	Create(ctx context.Context, input *CreateObjetRetrouveInput) (*ent.ObjetRetrouve, error)
	GetByID(ctx context.Context, id string) (*ent.ObjetRetrouve, error)
	GetByNumero(ctx context.Context, numero string) (*ent.ObjetRetrouve, error)
	List(ctx context.Context, filters *ObjetRetrouveFilters) ([]*ent.ObjetRetrouve, error)
	Count(ctx context.Context, filters *ObjetRetrouveFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateObjetRetrouveInput) (*ent.ObjetRetrouve, error)
	Delete(ctx context.Context, id string) error
	UpdateStatut(ctx context.Context, id string, statut string, dateRestitution *time.Time, proprietaire map[string]interface{}) (*ent.ObjetRetrouve, error)
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin *time.Time, periode *string) (map[string]interface{}, error)
}

// CreateObjetRetrouveInput represents input for creating objet retrouve
type CreateObjetRetrouveInput struct {
	ID                 string
	Numero             string
	TypeObjet          string
	Description        string
	ValeurEstimee      *string
	Couleur            *string
	DetailsSpecifiques map[string]interface{}
	IsContainer        bool
	ContainerDetails   map[string]interface{}
	Deposant           map[string]interface{}
	LieuTrouvaille     string
	AdresseLieu        *string
	DateTrouvaille     time.Time
	HeureTrouvaille    *string
	Statut             string
	DateDepot          time.Time
	Observations       *string
	CommissariatID     string
	AgentID            string
}

// UpdateObjetRetrouveInput represents input for updating objet retrouve
type UpdateObjetRetrouveInput struct {
	TypeObjet          *string
	Description        *string
	ValeurEstimee      *string
	Couleur            *string
	DetailsSpecifiques map[string]interface{}
	IsContainer        bool
	ContainerDetails   map[string]interface{}
	Deposant           map[string]interface{}
	LieuTrouvaille     *string
	AdresseLieu        *string
	DateTrouvaille     *time.Time
	HeureTrouvaille    *string
	Statut             *string
	DateRestitution    *time.Time
	Proprietaire       map[string]interface{}
	Observations       *string
}

// ObjetRetrouveFilters represents filters for listing objets retrouves
type ObjetRetrouveFilters struct {
	Statut         *string
	TypeObjet      *string
	CommissariatID *string
	AgentID        *string
	IsContainer    *bool
	DateDebut      *time.Time
	DateFin        *time.Time
	Search         *string
	Limit          int
	Offset         int
}

// objetRetrouveRepository implements ObjetRetrouveRepository
type objetRetrouveRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewObjetRetrouveRepository creates a new objet retrouve repository
func NewObjetRetrouveRepository(client *ent.Client, logger *zap.Logger) ObjetRetrouveRepository {
	return &objetRetrouveRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new objet retrouve
func (r *objetRetrouveRepository) Create(ctx context.Context, input *CreateObjetRetrouveInput) (*ent.ObjetRetrouve, error) {
	query := r.client.ObjetRetrouve.Create().
		SetID(uuid.MustParse(input.ID)).
		SetNumero(input.Numero).
		SetTypeObjet(input.TypeObjet).
		SetDescription(input.Description).
		SetIsContainer(input.IsContainer).
		SetDeposant(input.Deposant).
		SetLieuTrouvaille(input.LieuTrouvaille).
		SetDateTrouvaille(input.DateTrouvaille).
		SetDateDepot(input.DateDepot).
		SetCommissariatID(uuid.MustParse(input.CommissariatID)).
		SetAgentID(uuid.MustParse(input.AgentID)).
		SetStatut(objetretrouve.Statut(input.Statut))

	if input.ValeurEstimee != nil {
		query.SetValeurEstimee(*input.ValeurEstimee)
	}
	if input.Couleur != nil {
		query.SetCouleur(*input.Couleur)
	}
	if input.DetailsSpecifiques != nil && len(input.DetailsSpecifiques) > 0 {
		query.SetDetailsSpecifiques(input.DetailsSpecifiques)
	}
	if input.ContainerDetails != nil && len(input.ContainerDetails) > 0 {
		query.SetContainerDetails(input.ContainerDetails)
	}
	if input.AdresseLieu != nil {
		query.SetAdresseLieu(*input.AdresseLieu)
	}
	if input.HeureTrouvaille != nil {
		query.SetHeureTrouvaille(*input.HeureTrouvaille)
	}
	if input.Observations != nil {
		query.SetObservations(*input.Observations)
	}

	objet, err := query.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create objet retrouve", zap.Error(err))
		return nil, fmt.Errorf("failed to create objet retrouve: %w", err)
	}

	return objet, nil
}

// GetByID gets an objet retrouve by ID
func (r *objetRetrouveRepository) GetByID(ctx context.Context, id string) (*ent.ObjetRetrouve, error) {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	objet, err := r.client.ObjetRetrouve.Query().
		WithCommissariat().
		WithAgent().
		Where(objetretrouve.ID(objetID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to get objet retrouve: %w", err)
	}

	return objet, nil
}

// GetByNumero gets an objet retrouve by numero
func (r *objetRetrouveRepository) GetByNumero(ctx context.Context, numero string) (*ent.ObjetRetrouve, error) {
	objet, err := r.client.ObjetRetrouve.Query().
		WithCommissariat().
		WithAgent().
		Where(objetretrouve.Numero(numero)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to get objet retrouve: %w", err)
	}

	return objet, nil
}

// List lists objets retrouves with filters
func (r *objetRetrouveRepository) List(ctx context.Context, filters *ObjetRetrouveFilters) ([]*ent.ObjetRetrouve, error) {
	query := r.client.ObjetRetrouve.Query().
		WithCommissariat().
		WithAgent()

	// ===== DEBUG: Log initial =====
	r.logger.Info("🔍 Repository List - DÉBUT",
		zap.Bool("HasDateDebut", filters.DateDebut != nil),
		zap.Bool("HasDateFin", filters.DateFin != nil),
	)

	if filters.DateDebut != nil {
		r.logger.Info("📅 Repository - DateDebut AVANT filtre",
			zap.Time("dateDebut", *filters.DateDebut),
			zap.String("dateDebutString", filters.DateDebut.String()),
			zap.String("timezone", filters.DateDebut.Location().String()),
		)
	}

	if filters.DateFin != nil {
		r.logger.Info("📅 Repository - DateFin AVANT filtre",
			zap.Time("dateFin", *filters.DateFin),
			zap.String("dateFinString", filters.DateFin.String()),
			zap.String("timezone", filters.DateFin.Location().String()),
		)
	}

	// CORRECTION: S'assurer que tous les filtres sont appliqués dans le bon ordre
	if filters.Statut != nil {
		query = query.Where(objetretrouve.StatutEQ(objetretrouve.Statut(*filters.Statut)))
		r.logger.Info("✅ Applied filter: Statut", zap.String("statut", *filters.Statut))
	}

	if filters.TypeObjet != nil {
		query = query.Where(objetretrouve.TypeObjetContainsFold(*filters.TypeObjet))
		r.logger.Info("✅ Applied filter: TypeObjet", zap.String("typeObjet", *filters.TypeObjet))
	}

	if filters.CommissariatID != nil {
		commissariatID, err := uuid.Parse(*filters.CommissariatID)
		if err == nil {
			query = query.Where(objetretrouve.HasCommissariatWith(commissariat.ID(commissariatID)))
			r.logger.Info("✅ Applied filter: CommissariatID", zap.String("commissariatID", *filters.CommissariatID))
		} else {
			r.logger.Warn("Invalid CommissariatID format", zap.String("commissariatID", *filters.CommissariatID), zap.Error(err))
		}
	}

	if filters.AgentID != nil {
		agentID, err := uuid.Parse(*filters.AgentID)
		if err == nil {
			query = query.Where(objetretrouve.HasAgentWith(user.ID(agentID)))
			r.logger.Info("✅ Applied filter: AgentID", zap.String("agentID", *filters.AgentID))
		} else {
			r.logger.Warn("Invalid AgentID format", zap.String("agentID", *filters.AgentID), zap.Error(err))
		}
	}

	if filters.IsContainer != nil {
		query = query.Where(objetretrouve.IsContainerEQ(*filters.IsContainer))
		r.logger.Info("✅ Applied filter: IsContainer", zap.Bool("isContainer", *filters.IsContainer))
	}

	// ===== FILTRES DE DATES =====
	// CORRECTION: Filtrer sur DateTrouvaille au lieu de DateDepot
	if filters.DateDebut != nil {
		r.logger.Info("🚨 AVANT application filtre DateDebut",
			zap.Time("dateDebut", *filters.DateDebut),
		)
		query = query.Where(objetretrouve.DateTrouvailleGTE(*filters.DateDebut))
		r.logger.Info("✅✅✅ FILTRE DateDebut APPLIQUÉ (sur DateTrouvaille)", zap.Time("dateDebut", *filters.DateDebut))
	}

	if filters.DateFin != nil {
		r.logger.Info("🚨 AVANT application filtre DateFin",
			zap.Time("dateFin", *filters.DateFin),
		)
		query = query.Where(objetretrouve.DateTrouvailleLTE(*filters.DateFin))
		r.logger.Info("✅✅✅ FILTRE DateFin APPLIQUÉ (sur DateTrouvaille)", zap.Time("dateFin", *filters.DateFin))
	}

	if filters.Search != nil && *filters.Search != "" {
		search := *filters.Search
		query = query.Where(
			objetretrouve.Or(
				objetretrouve.NumeroContains(search),
				objetretrouve.TypeObjetContains(search),
				objetretrouve.DescriptionContains(search),
				objetretrouve.LieuTrouvailleContains(search),
			),
		)
		r.logger.Info("Applied filter: Search", zap.String("search", search))
	}

	// Pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
		r.logger.Info("Applied pagination: Limit", zap.Int("limit", filters.Limit))
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
		r.logger.Info("Applied pagination: Offset", zap.Int("offset", filters.Offset))
	}

	// Tri par date de dépôt décroissant
	query = query.Order(ent.Desc(objetretrouve.FieldDateDepot))

	// Exécuter la requête
	objets, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("failed to list objets retrouves: %w", err)
	}

	// ===== DEBUG: Résultats =====
	r.logger.Info("📊 Repository List - RÉSULTATS",
		zap.Int("count", len(objets)),
		zap.Bool("hadDateFilters", filters.DateDebut != nil || filters.DateFin != nil),
	)

	// Si on a des filtres de date, comparer avec le total sans filtre
	if filters.DateDebut != nil || filters.DateFin != nil {
		totalQuery := r.client.ObjetRetrouve.Query()
		totalCount, _ := totalQuery.Count(ctx)
		r.logger.Info("🚨 COMPARAISON AVEC/SANS FILTRE",
			zap.Int("avecFiltre", len(objets)),
			zap.Int("sansFiltre", totalCount),
			zap.Bool("filtreFonctionne", len(objets) < totalCount),
		)
	}

	// Log les 3 premiers objets pour debug
	if len(objets) > 0 {
		for i, obj := range objets {
			if i >= 3 {
				break
			}
			r.logger.Info("📋 Sample objet",
				zap.Int("index", i),
				zap.String("id", obj.ID.String()),
				zap.String("numero", obj.Numero),
				zap.Time("dateDepot", obj.DateDepot),
				zap.String("dateDepotString", obj.DateDepot.String()),
			)
		}
	} else {
		r.logger.Warn("⚠️ Aucun objet trouvé avec ces filtres")
	}

	return objets, nil
}

// Count counts objets retrouves with filters
func (r *objetRetrouveRepository) Count(ctx context.Context, filters *ObjetRetrouveFilters) (int, error) {
	query := r.client.ObjetRetrouve.Query()

	if filters.Statut != nil {
		query = query.Where(objetretrouve.StatutEQ(objetretrouve.Statut(*filters.Statut)))
	}
	if filters.TypeObjet != nil {
		query = query.Where(objetretrouve.TypeObjetContainsFold(*filters.TypeObjet))
	}
	if filters.CommissariatID != nil {
		commissariatID, err := uuid.Parse(*filters.CommissariatID)
		if err == nil {
			query = query.Where(objetretrouve.HasCommissariatWith(commissariat.ID(commissariatID)))
		}
	}
	if filters.AgentID != nil {
		agentID, err := uuid.Parse(*filters.AgentID)
		if err == nil {
			query = query.Where(objetretrouve.HasAgentWith(user.ID(agentID)))
		}
	}
	if filters.IsContainer != nil {
		query = query.Where(objetretrouve.IsContainerEQ(*filters.IsContainer))
	}
	// CORRECTION: Filtrer sur DateTrouvaille au lieu de DateDepot
	if filters.DateDebut != nil {
		query = query.Where(objetretrouve.DateTrouvailleGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(objetretrouve.DateTrouvailleLTE(*filters.DateFin))
	}
	if filters.Search != nil && *filters.Search != "" {
		search := *filters.Search
		query = query.Where(
			objetretrouve.Or(
				objetretrouve.NumeroContains(search),
				objetretrouve.TypeObjetContains(search),
				objetretrouve.DescriptionContains(search),
				objetretrouve.LieuTrouvailleContains(search),
			),
		)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count objets retrouves: %w", err)
	}

	return count, nil
}

// Update updates an objet retrouve
func (r *objetRetrouveRepository) Update(ctx context.Context, id string, input *UpdateObjetRetrouveInput) (*ent.ObjetRetrouve, error) {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	query := r.client.ObjetRetrouve.UpdateOneID(objetID)

	if input.TypeObjet != nil {
		query = query.SetTypeObjet(*input.TypeObjet)
	}
	if input.Description != nil {
		query = query.SetDescription(*input.Description)
	}
	if input.ValeurEstimee != nil {
		query = query.SetValeurEstimee(*input.ValeurEstimee)
	}
	if input.Couleur != nil {
		query = query.SetCouleur(*input.Couleur)
	}
	if input.DetailsSpecifiques != nil {
		query = query.SetDetailsSpecifiques(input.DetailsSpecifiques)
	}
	query = query.SetIsContainer(input.IsContainer)
	if input.ContainerDetails != nil {
		query = query.SetContainerDetails(input.ContainerDetails)
	}
	if input.Deposant != nil {
		query = query.SetDeposant(input.Deposant)
	}
	if input.LieuTrouvaille != nil {
		query = query.SetLieuTrouvaille(*input.LieuTrouvaille)
	}
	if input.AdresseLieu != nil {
		query = query.SetAdresseLieu(*input.AdresseLieu)
	}
	if input.DateTrouvaille != nil {
		query = query.SetDateTrouvaille(*input.DateTrouvaille)
	}
	if input.HeureTrouvaille != nil {
		query = query.SetHeureTrouvaille(*input.HeureTrouvaille)
	}
	if input.Statut != nil {
		query = query.SetStatut(objetretrouve.Statut(*input.Statut))
	}
	if input.DateRestitution != nil {
		query = query.SetDateRestitution(*input.DateRestitution)
	}
	if input.Proprietaire != nil {
		query = query.SetProprietaire(input.Proprietaire)
	}
	if input.Observations != nil {
		query = query.SetObservations(*input.Observations)
	}

	objet, err := query.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to update objet retrouve: %w", err)
	}

	return objet, nil
}

// UpdateStatut updates the statut of an objet retrouve
func (r *objetRetrouveRepository) UpdateStatut(ctx context.Context, id string, statut string, dateRestitution *time.Time, proprietaire map[string]interface{}) (*ent.ObjetRetrouve, error) {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	update := r.client.ObjetRetrouve.UpdateOneID(objetID).
		SetStatut(objetretrouve.Statut(statut))

	if dateRestitution != nil {
		update = update.SetDateRestitution(*dateRestitution)
	}
	if proprietaire != nil {
		update = update.SetProprietaire(proprietaire)
	}

	objet, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to update statut: %w", err)
	}

	return objet, nil
}

// Delete deletes an objet retrouve
func (r *objetRetrouveRepository) Delete(ctx context.Context, id string) error {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	err = r.client.ObjetRetrouve.DeleteOneID(objetID).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("objet retrouve not found")
		}
		return fmt.Errorf("failed to delete objet retrouve: %w", err)
	}

	return nil
}

// GetStatistiques calcule les statistiques des objets retrouvés
func (r *objetRetrouveRepository) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin *time.Time, periode *string) (map[string]interface{}, error) {
	query := r.client.ObjetRetrouve.Query()

	if commissariatID != nil {
		commID, _ := uuid.Parse(*commissariatID)
		query = query.Where(objetretrouve.HasCommissariatWith(commissariat.ID(commID)))
	}
	// CORRECTION: Filtrer sur DateTrouvaille pour cohérence
	if dateDebut != nil {
		query = query.Where(objetretrouve.DateTrouvailleGTE(*dateDebut))
	}
	if dateFin != nil {
		query = query.Where(objetretrouve.DateTrouvailleLTE(*dateFin))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Compter par statut
	disponibles, _ := query.Clone().Where(objetretrouve.StatutEQ(objetretrouve.StatutDISPONIBLE)).Count(ctx)
	restitues, _ := query.Clone().Where(objetretrouve.StatutEQ(objetretrouve.StatutRESTITUÉ)).Count(ctx)
	nonReclames, _ := query.Clone().Where(objetretrouve.StatutEQ(objetretrouve.StatutNON_RÉCLAMÉ)).Count(ctx)

	stats := map[string]interface{}{
		"total":       total,
		"disponibles": disponibles,
		"restitues":   restitues,
		"nonReclames": nonReclames,
	}

	// Calculer le taux de restitution
	if total > 0 {
		tauxRestitution := float64(restitues) / float64(total) * 100
		stats["tauxRestitution"] = math.Round(tauxRestitution*100) / 100 // Arrondir à 2 décimales
	} else {
		stats["tauxRestitution"] = 0.0
	}

	// Calculer l'évolution si période fournie
	if periode != nil && *periode != "" && dateDebut != nil && dateFin != nil {
		r.logger.Info("Calcul des évolutions avec période",
			zap.Stringp("periode", periode),
			zap.Bool("hasDateDebut", dateDebut != nil),
			zap.Bool("hasDateFin", dateFin != nil),
		)
		evolution := r.calculerEvolutionPeriode(ctx, commissariatID, *dateDebut, *dateFin, *periode)
		stats["evolutionTotal"] = evolution["total"]
		stats["evolutionDisponibles"] = evolution["disponibles"]
		stats["evolutionRestitues"] = evolution["restitues"]
		stats["evolutionNonReclames"] = evolution["nonReclames"]
		stats["evolutionTauxRestitution"] = evolution["tauxRestitution"]
		r.logger.Info("Évolutions calculées",
			zap.String("evolutionTotal", evolution["total"]),
			zap.String("evolutionDisponibles", evolution["disponibles"]),
			zap.String("evolutionRestitues", evolution["restitues"]),
			zap.String("evolutionNonReclames", evolution["nonReclames"]),
			zap.String("evolutionTauxRestitution", evolution["tauxRestitution"]),
		)
	} else {
		// Si pas de période, retourner "0" pour toutes les évolutions
		r.logger.Info("Pas de période fournie, évolutions à 0",
			zap.Bool("hasPeriode", periode != nil),
			zap.Bool("hasDateDebut", dateDebut != nil),
			zap.Bool("hasDateFin", dateFin != nil),
		)
		stats["evolutionTotal"] = "0"
		stats["evolutionDisponibles"] = "0"
		stats["evolutionRestitues"] = "0"
		stats["evolutionNonReclames"] = "0"
		stats["evolutionTauxRestitution"] = "0"
	}

	// S'assurer que les évolutions sont TOUJOURS présentes dans le map
	if _, exists := stats["evolutionTotal"]; !exists {
		stats["evolutionTotal"] = "0"
	}
	if _, exists := stats["evolutionDisponibles"]; !exists {
		stats["evolutionDisponibles"] = "0"
	}
	if _, exists := stats["evolutionRestitues"]; !exists {
		stats["evolutionRestitues"] = "0"
	}
	if _, exists := stats["evolutionNonReclames"]; !exists {
		stats["evolutionNonReclames"] = "0"
	}
	if _, exists := stats["evolutionTauxRestitution"]; !exists {
		stats["evolutionTauxRestitution"] = "0"
	}

	r.logger.Info("Stats retournées par repository",
		zap.Any("stats", stats),
		zap.String("evolutionTotal", stats["evolutionTotal"].(string)),
		zap.String("evolutionDisponibles", stats["evolutionDisponibles"].(string)),
		zap.String("evolutionRestitues", stats["evolutionRestitues"].(string)),
		zap.String("evolutionNonReclames", stats["evolutionNonReclames"].(string)),
		zap.String("evolutionTauxRestitution", stats["evolutionTauxRestitution"].(string)),
	)

	return stats, nil
}

// calculerEvolutionPeriode calcule l'évolution par rapport à la période précédente
func (r *objetRetrouveRepository) calculerEvolutionPeriode(ctx context.Context, commissariatID *string, dateDebut, dateFin time.Time, typePeriode string) map[string]string {
	var debutPrecedent, finPrecedent time.Time

	// Calculer la période précédente selon le type
	switch typePeriode {
	case "jour":
		debutPrecedent = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day()-1, 0, 0, 0, 0, dateDebut.Location())
		finPrecedent = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day()-1, 23, 59, 59, 0, dateDebut.Location())

	case "semaine":
		debutPrecedent = dateDebut.Add(-7 * 24 * time.Hour)
		finPrecedent = dateDebut.Add(-1 * time.Second)

	case "mois":
		moisPrecedent := dateDebut.AddDate(0, -1, 0)
		debutPrecedent = time.Date(moisPrecedent.Year(), moisPrecedent.Month(), 1, 0, 0, 0, 0, dateDebut.Location())
		finPrecedent = time.Date(dateDebut.Year(), dateDebut.Month(), 1, 0, 0, 0, 0, dateDebut.Location()).Add(-1 * time.Second)

	case "annee":
		anneePrecedente := dateDebut.Year() - 1
		debutPrecedent = time.Date(anneePrecedente, 1, 1, 0, 0, 0, 0, dateDebut.Location())
		finPrecedent = time.Date(anneePrecedente, 12, 31, 23, 59, 59, 0, dateDebut.Location())

	default:
		return map[string]string{
			"total": "0", "disponibles": "0", "restitues": "0", "nonReclames": "0", "tauxRestitution": "0",
		}
	}

	// Récupérer les stats de la période actuelle (sans évolution pour éviter récursion)
	statsActuelles, _ := r.GetStatistiques(ctx, commissariatID, &dateDebut, &dateFin, nil)
	totalActuel := 0
	disponiblesActuel := 0
	restituesActuel := 0
	nonReclamesActuel := 0
	tauxRestitutionActuel := 0.0
	if statsActuelles != nil {
		if t, ok := statsActuelles["total"].(int); ok {
			totalActuel = t
		}
		if d, ok := statsActuelles["disponibles"].(int); ok {
			disponiblesActuel = d
		}
		if res, ok := statsActuelles["restitues"].(int); ok {
			restituesActuel = res
		}
		if nr, ok := statsActuelles["nonReclames"].(int); ok {
			nonReclamesActuel = nr
		}
		if taux, ok := statsActuelles["tauxRestitution"].(float64); ok {
			tauxRestitutionActuel = taux
		}
	}

	// Récupérer les stats de la période précédente (sans évolution)
	statsPrecedentes, _ := r.GetStatistiques(ctx, commissariatID, &debutPrecedent, &finPrecedent, nil)
	totalPrecedent := 0
	disponiblesPrecedent := 0
	restituesPrecedent := 0
	nonReclamesPrecedent := 0
	tauxRestitutionPrecedent := 0.0

	if statsPrecedentes != nil {
		if t, ok := statsPrecedentes["total"].(int); ok {
			totalPrecedent = t
		}
		if d, ok := statsPrecedentes["disponibles"].(int); ok {
			disponiblesPrecedent = d
		}
		if res, ok := statsPrecedentes["restitues"].(int); ok {
			restituesPrecedent = res
		}
		if nr, ok := statsPrecedentes["nonReclames"].(int); ok {
			nonReclamesPrecedent = nr
		}
		if taux, ok := statsPrecedentes["tauxRestitution"].(float64); ok {
			tauxRestitutionPrecedent = taux
		}
	}

	// Debug logs
	r.logger.Info("Calcul évolution objets retrouvés",
		zap.String("periode", typePeriode),
		zap.Time("debutActuel", dateDebut),
		zap.Time("finActuel", dateFin),
		zap.Time("debutPrecedent", debutPrecedent),
		zap.Time("finPrecedent", finPrecedent),
		zap.Int("totalActuel", totalActuel),
		zap.Int("totalPrecedent", totalPrecedent),
		zap.Int("disponiblesActuel", disponiblesActuel),
		zap.Int("disponiblesPrecedent", disponiblesPrecedent),
		zap.Int("restituesActuel", restituesActuel),
		zap.Int("restituesPrecedent", restituesPrecedent),
		zap.Int("nonReclamesActuel", nonReclamesActuel),
		zap.Int("nonReclamesPrecedent", nonReclamesPrecedent),
	)

	// Calculer les différences
	diffTotal := totalActuel - totalPrecedent
	diffDisponibles := disponiblesActuel - disponiblesPrecedent
	diffRestitues := restituesActuel - restituesPrecedent
	diffNonReclames := nonReclamesActuel - nonReclamesPrecedent
	diffTauxRestitution := tauxRestitutionActuel - tauxRestitutionPrecedent

	// Formater avec signes
	evolutionTotal := formatEvolutionObjetRetrouve(diffTotal)
	evolutionDisponibles := formatEvolutionObjetRetrouve(diffDisponibles)
	evolutionRestitues := formatEvolutionObjetRetrouve(diffRestitues)
	evolutionNonReclames := formatEvolutionObjetRetrouve(diffNonReclames)
	// Pour le taux, on arrondit à 1 décimale
	evolutionTauxRestitution := formatEvolutionFloatObjetRetrouve(diffTauxRestitution)

	return map[string]string{
		"total":           evolutionTotal,
		"disponibles":     evolutionDisponibles,
		"restitues":       evolutionRestitues,
		"nonReclames":     evolutionNonReclames,
		"tauxRestitution": evolutionTauxRestitution,
	}
}

// formatEvolutionObjetRetrouve formate un nombre avec son signe
func formatEvolutionObjetRetrouve(diff int) string {
	if diff > 0 {
		return fmt.Sprintf("+%d", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%d", diff)
	}
	return "0"
}

// formatEvolutionFloatObjetRetrouve formate un nombre décimal avec son signe
func formatEvolutionFloatObjetRetrouve(diff float64) string {
	if diff > 0 {
		return fmt.Sprintf("+%.1f", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%.1f", diff)
	}
	return "0"
}
