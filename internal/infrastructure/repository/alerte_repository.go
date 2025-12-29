package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/alertesecuritaire"
	"police-trafic-api-frontend-aligned/ent/commissariat"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AlerteRepository defines alerte repository interface
type AlerteRepository interface {
	Create(ctx context.Context, input *CreateAlerteInput) (*ent.AlerteSecuritaire, error)
	GetByID(ctx context.Context, id string) (*ent.AlerteSecuritaire, error)
	GetByNumero(ctx context.Context, numero string) (*ent.AlerteSecuritaire, error)
	List(ctx context.Context, filters *AlerteFilters) ([]*ent.AlerteSecuritaire, error)
	Count(ctx context.Context, filters *AlerteFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateAlerteInput) (*ent.AlerteSecuritaire, error)
	Delete(ctx context.Context, id string) error
	GetByCommissariat(ctx context.Context, commissariatID string) ([]*ent.AlerteSecuritaire, error)
	GetByStatut(ctx context.Context, statut string) ([]*ent.AlerteSecuritaire, error)
	GetActives(ctx context.Context) ([]*ent.AlerteSecuritaire, error)
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin *time.Time, periode *string) (map[string]interface{}, error)
}

// CreateAlerteInput represents input for creating alerte
type CreateAlerteInput struct {
	ID                    string
	Numero                string
	Titre                 string
	Description           string
	Contexte              *string
	Niveau                string
	TypeAlerte            string
	Lieu                  *string
	Latitude              *float64
	Longitude             *float64
	PrecisionLocalisation *string
	Risques               []string
	PersonneConcernee     map[string]interface{}
	Vehicule              map[string]interface{}
	Suspect               map[string]interface{}
	CommissariatID        string
	AgentRecepteurID      string
	DateAlerte            *time.Time
	Observations          *string
}

// UpdateAlerteInput represents input for updating alerte
type UpdateAlerteInput struct {
	Titre                    *string
	Description              *string
	Contexte                 *string
	Niveau                   *string
	Statut                   *string
	TypeAlerte               *string
	Lieu                     *string
	Latitude                 *float64
	Longitude                *float64
	PrecisionLocalisation    *string
	Risques                  []string
	PersonneConcernee        map[string]interface{}
	Vehicule                 map[string]interface{}
	Suspect                  map[string]interface{}
	Intervention             map[string]interface{}
	Evaluation               map[string]interface{}
	Actions                  map[string]interface{}
	Rapport                  map[string]interface{}
	Temoins                  []map[string]interface{}
	Documents                []map[string]interface{}
	Photos                   []string
	Suivis                   []map[string]interface{}
	Diffusee                 *bool
	DateDiffusion            *time.Time
	DiffusionDestinataires   map[string]interface{}
	AssignationDestinataires map[string]interface{}
	DateResolution           *time.Time
	DateCloture              *time.Time
	Observations             *string
}

// AlerteFilters represents filters for listing alertes
type AlerteFilters struct {
	Niveau         *string
	Statut         *string
	TypeAlerte     *string
	CommissariatID *string
	DateDebut      *time.Time
	DateFin        *time.Time
	Search         *string
	Limit          int
	Offset         int
}

// alerteRepository implements AlerteRepository
type alerteRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewAlerteRepository creates a new alerte repository
func NewAlerteRepository(client *ent.Client, logger *zap.Logger) AlerteRepository {
	return &alerteRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new alerte
func (r *alerteRepository) Create(ctx context.Context, input *CreateAlerteInput) (*ent.AlerteSecuritaire, error) {
	r.logger.Info("Creating alerte", zap.String("titre", input.Titre))

	id, _ := uuid.Parse(input.ID)
	commID, _ := uuid.Parse(input.CommissariatID)
	agentID, _ := uuid.Parse(input.AgentRecepteurID)

	create := r.client.AlerteSecuritaire.Create().
		SetID(id).
		SetNumero(input.Numero).
		SetTitre(input.Titre).
		SetDescription(input.Description).
		SetTypeAlerte(input.TypeAlerte).
		SetNiveau(alertesecuritaire.Niveau(input.Niveau)).
		SetCommissariatID(commID).
		SetAgentID(agentID)

	if input.Contexte != nil {
		create = create.SetContexte(*input.Contexte)
	}
	if input.Lieu != nil {
		create = create.SetLieu(*input.Lieu)
	}
	if input.Latitude != nil {
		create = create.SetLatitude(*input.Latitude)
	}
	if input.Longitude != nil {
		create = create.SetLongitude(*input.Longitude)
	}
	if input.PrecisionLocalisation != nil {
		create = create.SetPrecisionLocalisation(*input.PrecisionLocalisation)
	}
	if input.DateAlerte != nil {
		create = create.SetDateAlerte(*input.DateAlerte)
	}
	if input.Observations != nil {
		create = create.SetObservations(*input.Observations)
	}

	// Données JSONB
	if input.Risques != nil {
		create = create.SetRisques(input.Risques)
	}
	if input.PersonneConcernee != nil {
		create = create.SetPersonneConcernee(input.PersonneConcernee)
	}
	if input.Vehicule != nil {
		create = create.SetVehicule(input.Vehicule)
	}
	if input.Suspect != nil {
		create = create.SetSuspect(input.Suspect)
	}

	// Initialiser actions avec la structure par défaut (toujours)
	create = create.SetActions(map[string]interface{}{
		"immediate":  []string{},
		"preventive": []string{},
		"suivi":      []string{},
	})

	alerte, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create alerte", zap.Error(err))
		return nil, fmt.Errorf("failed to create alerte: %w", err)
	}

	return alerte, nil
}

// GetByID gets alerte by ID
func (r *alerteRepository) GetByID(ctx context.Context, id string) (*ent.AlerteSecuritaire, error) {
	uid, _ := uuid.Parse(id)
	alerte, err := r.client.AlerteSecuritaire.
		Query().
		Where(alertesecuritaire.ID(uid)).
		WithCommissariat().
		WithAgent().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("alerte not found")
		}
		r.logger.Error("Failed to get alerte by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get alerte: %w", err)
	}

	return alerte, nil
}

// List gets alertes with filters
func (r *alerteRepository) List(ctx context.Context, filters *AlerteFilters) ([]*ent.AlerteSecuritaire, error) {
	query := r.client.AlerteSecuritaire.Query()

	if filters != nil {
		r.logger.Info("Applying filters to alertes query",
			zap.Any("commissariatId", filters.CommissariatID),
			zap.Any("dateDebut", filters.DateDebut),
			zap.Any("dateFin", filters.DateFin),
			zap.Any("statut", filters.Statut),
			zap.Any("niveau", filters.Niveau),
			zap.Any("search", filters.Search),
			zap.Int("limit", filters.Limit),
			zap.Int("offset", filters.Offset),
		)

		query = r.applyFilters(query, filters)

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	alerteList, err := query.
		WithCommissariat().
		WithAgent().
		Order(ent.Desc(alertesecuritaire.FieldDateAlerte)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list alertes", zap.Error(err))
		return nil, fmt.Errorf("failed to list alertes: %w", err)
	}

	r.logger.Info("Alertes list result", zap.Int("count", len(alerteList)))

	return alerteList, nil
}

// Count counts alertes with filters
func (r *alerteRepository) Count(ctx context.Context, filters *AlerteFilters) (int, error) {
	query := r.client.AlerteSecuritaire.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count alertes", zap.Error(err))
		return 0, fmt.Errorf("failed to count alertes: %w", err)
	}

	return count, nil
}

// applyFilters applies filters to alerte query
func (r *alerteRepository) applyFilters(query *ent.AlerteSecuritaireQuery, filters *AlerteFilters) *ent.AlerteSecuritaireQuery {
	if filters.Niveau != nil {
		query = query.Where(alertesecuritaire.NiveauEQ(alertesecuritaire.Niveau(*filters.Niveau)))
	}
	if filters.Statut != nil {
		query = query.Where(alertesecuritaire.StatutEQ(alertesecuritaire.Statut(*filters.Statut)))
	}
	if filters.TypeAlerte != nil {
		query = query.Where(alertesecuritaire.TypeAlerte(*filters.TypeAlerte))
	}
	if filters.CommissariatID != nil {
		commID, _ := uuid.Parse(*filters.CommissariatID)
		query = query.Where(alertesecuritaire.HasCommissariatWith(commissariat.ID(commID)))
	}
	if filters.DateDebut != nil {
		query = query.Where(alertesecuritaire.DateAlerteGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(alertesecuritaire.DateAlerteLTE(*filters.DateFin))
	}
	if filters.Search != nil && *filters.Search != "" {
		query = query.Where(
			alertesecuritaire.Or(
				alertesecuritaire.NumeroContains(*filters.Search),      // Recherche par numéro
				alertesecuritaire.TypeAlerteContains(*filters.Search),  // Recherche par type
				alertesecuritaire.TitreContains(*filters.Search),       // Recherche par titre
				alertesecuritaire.DescriptionContains(*filters.Search), // Recherche par description
			),
		)
	}
	return query
}

// Update updates alerte
func (r *alerteRepository) Update(ctx context.Context, id string, input *UpdateAlerteInput) (*ent.AlerteSecuritaire, error) {
	r.logger.Info("Updating alerte", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.AlerteSecuritaire.UpdateOneID(uid)

	// Champs de base
	if input.Titre != nil {
		update = update.SetTitre(*input.Titre)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Contexte != nil {
		update = update.SetContexte(*input.Contexte)
	}
	if input.Niveau != nil {
		update = update.SetNiveau(alertesecuritaire.Niveau(*input.Niveau))
	}
	if input.Statut != nil {
		update = update.SetStatut(alertesecuritaire.Statut(*input.Statut))
	}
	if input.TypeAlerte != nil {
		update = update.SetTypeAlerte(*input.TypeAlerte)
	}

	// Localisation
	if input.Lieu != nil {
		update = update.SetLieu(*input.Lieu)
	}
	if input.Latitude != nil {
		update = update.SetLatitude(*input.Latitude)
	}
	if input.Longitude != nil {
		update = update.SetLongitude(*input.Longitude)
	}
	if input.PrecisionLocalisation != nil {
		update = update.SetPrecisionLocalisation(*input.PrecisionLocalisation)
	}

	// Données JSONB
	if input.Risques != nil {
		update = update.SetRisques(input.Risques)
	}
	if input.PersonneConcernee != nil {
		update = update.SetPersonneConcernee(input.PersonneConcernee)
	}
	if input.Vehicule != nil {
		update = update.SetVehicule(input.Vehicule)
	}
	if input.Suspect != nil {
		update = update.SetSuspect(input.Suspect)
	}
	if input.Intervention != nil {
		update = update.SetIntervention(input.Intervention)
	}
	if input.Evaluation != nil {
		update = update.SetEvaluation(input.Evaluation)
	}
	if input.Actions != nil {
		update = update.SetActions(input.Actions)
	}
	if input.Rapport != nil {
		update = update.SetRapport(input.Rapport)
	}
	if input.Temoins != nil {
		update = update.SetTemoins(input.Temoins)
	}
	if input.Documents != nil {
		update = update.SetDocuments(input.Documents)
	}
	if input.Photos != nil {
		update = update.SetPhotos(input.Photos)
	}
	if input.Suivis != nil {
		update = update.SetSuivis(input.Suivis)
	}

	// Diffusion
	if input.Diffusee != nil {
		update = update.SetDiffusee(*input.Diffusee)
	}
	if input.DateDiffusion != nil {
		update = update.SetDateDiffusion(*input.DateDiffusion)
	}
	if input.DiffusionDestinataires != nil {
		update = update.SetDiffusionDestinataires(input.DiffusionDestinataires)
	}
	if input.AssignationDestinataires != nil {
		update = update.SetAssignationDestinataires(input.AssignationDestinataires)
	}

	// Résolution et clôture
	if input.DateResolution != nil {
		update = update.SetDateResolution(*input.DateResolution)
	}
	if input.DateCloture != nil {
		update = update.SetDateCloture(*input.DateCloture)
	}
	if input.Observations != nil {
		update = update.SetObservations(*input.Observations)
	}

	alerte, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update alerte", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update alerte: %w", err)
	}

	return alerte, nil
}

// Delete deletes alerte
func (r *alerteRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting alerte", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.AlerteSecuritaire.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete alerte", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete alerte: %w", err)
	}

	return nil
}

// GetByCommissariat gets alertes by commissariat ID
func (r *alerteRepository) GetByCommissariat(ctx context.Context, commissariatID string) ([]*ent.AlerteSecuritaire, error) {
	commID, _ := uuid.Parse(commissariatID)
	alerteList, err := r.client.AlerteSecuritaire.
		Query().
		Where(alertesecuritaire.HasCommissariatWith(commissariat.ID(commID))).
		WithCommissariat().
		WithAgent().
		Order(ent.Desc(alertesecuritaire.FieldDateAlerte)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get alertes by commissariat", zap.String("commissariatID", commissariatID), zap.Error(err))
		return nil, fmt.Errorf("failed to get alertes: %w", err)
	}

	return alerteList, nil
}

// GetByStatut gets alertes by statut
func (r *alerteRepository) GetByStatut(ctx context.Context, statut string) ([]*ent.AlerteSecuritaire, error) {
	alerteList, err := r.client.AlerteSecuritaire.
		Query().
		Where(alertesecuritaire.StatutEQ(alertesecuritaire.Statut(statut))).
		WithCommissariat().
		WithAgent().
		Order(ent.Desc(alertesecuritaire.FieldDateAlerte)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get alertes by statut", zap.String("statut", statut), zap.Error(err))
		return nil, fmt.Errorf("failed to get alertes: %w", err)
	}

	return alerteList, nil
}

// GetActives gets active alertes
func (r *alerteRepository) GetActives(ctx context.Context) ([]*ent.AlerteSecuritaire, error) {
	alerteList, err := r.client.AlerteSecuritaire.
		Query().
		Where(alertesecuritaire.StatutEQ(alertesecuritaire.StatutACTIVE)).
		WithCommissariat().
		WithAgent().
		Order(ent.Desc(alertesecuritaire.FieldDateAlerte)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get active alertes", zap.Error(err))
		return nil, fmt.Errorf("failed to get alertes: %w", err)
	}

	return alerteList, nil
}

// GetByNumero gets alerte by numero
func (r *alerteRepository) GetByNumero(ctx context.Context, numero string) (*ent.AlerteSecuritaire, error) {
	alerte, err := r.client.AlerteSecuritaire.
		Query().
		Where(alertesecuritaire.NumeroEQ(numero)).
		WithCommissariat().
		WithAgent().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("alerte not found")
		}
		r.logger.Error("Failed to get alerte by numero", zap.String("numero", numero), zap.Error(err))
		return nil, fmt.Errorf("failed to get alerte: %w", err)
	}

	return alerte, nil
}

// GetStatistiques calcule les statistiques des alertes
func (r *alerteRepository) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin *time.Time, periode *string) (map[string]interface{}, error) {
	query := r.client.AlerteSecuritaire.Query()

	if commissariatID != nil {
		commID, _ := uuid.Parse(*commissariatID)
		query = query.Where(alertesecuritaire.HasCommissariatWith(commissariat.ID(commID)))
	}
	if dateDebut != nil {
		query = query.Where(alertesecuritaire.DateAlerteGTE(*dateDebut))
	}
	if dateFin != nil {
		query = query.Where(alertesecuritaire.DateAlerteLTE(*dateFin))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Compter par statut
	actives, _ := query.Clone().Where(alertesecuritaire.StatutEQ(alertesecuritaire.StatutACTIVE)).Count(ctx)
	resolues, _ := query.Clone().Where(alertesecuritaire.StatutEQ(alertesecuritaire.StatutRESOLUE)).Count(ctx)
	archivees, _ := query.Clone().Where(alertesecuritaire.StatutEQ(alertesecuritaire.StatutARCHIVEE)).Count(ctx)

	// Compter par niveau
	faible, _ := query.Clone().Where(alertesecuritaire.NiveauEQ(alertesecuritaire.NiveauFAIBLE)).Count(ctx)
	moyen, _ := query.Clone().Where(alertesecuritaire.NiveauEQ(alertesecuritaire.NiveauMOYEN)).Count(ctx)
	eleve, _ := query.Clone().Where(alertesecuritaire.NiveauEQ(alertesecuritaire.NiveauELEVE)).Count(ctx)
	critique, _ := query.Clone().Where(alertesecuritaire.NiveauEQ(alertesecuritaire.NiveauCRITIQUE)).Count(ctx)

	stats := map[string]interface{}{
		"total":     total,
		"actives":   actives,
		"resolues":  resolues,
		"archivees": archivees,
		"parNiveau": map[string]int{
			"FAIBLE":   faible,
			"MOYEN":    moyen,
			"ELEVE":    eleve,
			"CRITIQUE": critique,
		},
	}

	// Calculer le taux de résolution
	if total > 0 {
		tauxResolution := float64(resolues) / float64(total) * 100
		stats["tauxResolution"] = tauxResolution
	}

	// Calculer l'évolution si période fournie
	if periode != nil && dateDebut != nil && dateFin != nil {
		evolution := r.calculerEvolutionPeriode(ctx, commissariatID, *dateDebut, *dateFin, *periode)
		stats["evolutionAlertes"] = evolution["alertes"]
		stats["evolutionResolution"] = evolution["resolutions"]
	}

	return stats, nil
}

// calculerEvolutionPeriode calcule l'évolution par rapport à la période précédente
func (r *alerteRepository) calculerEvolutionPeriode(ctx context.Context, commissariatID *string, dateDebut, dateFin time.Time, typePeriode string) map[string]string {
	var debutPrecedent, finPrecedent time.Time

	// Calculer la période précédente selon le type
	switch typePeriode {
	case "jour":
		// Période actuelle: aujourd'hui (00:00:00 à 23:59:59)
		// Période précédente: hier (00:00:00 à 23:59:59)
		debutPrecedent = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day()-1, 0, 0, 0, 0, dateDebut.Location())
		finPrecedent = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day()-1, 23, 59, 59, 0, dateDebut.Location())

	case "semaine":
		// Période actuelle: du lundi de cette semaine à aujourd'hui
		// Période précédente: du lundi de la semaine dernière au dimanche de la semaine dernière
		debutPrecedent = dateDebut.Add(-7 * 24 * time.Hour)
		finPrecedent = dateDebut.Add(-1 * time.Second) // Juste avant le début de la semaine actuelle

	case "mois":
		// Période actuelle: du 1er du mois en cours à aujourd'hui
		// Période précédente: tout le mois précédent (du 1er au dernier jour)
		moisPrecedent := dateDebut.AddDate(0, -1, 0)
		debutPrecedent = time.Date(moisPrecedent.Year(), moisPrecedent.Month(), 1, 0, 0, 0, 0, dateDebut.Location())
		// Dernier jour du mois précédent
		finPrecedent = time.Date(dateDebut.Year(), dateDebut.Month(), 1, 0, 0, 0, 0, dateDebut.Location()).Add(-1 * time.Second)

	case "annee":
		// Période actuelle: du 1er janvier de l'année en cours à aujourd'hui
		// Période précédente: toute l'année précédente (1er janvier au 31 décembre)
		anneePrecedente := dateDebut.Year() - 1
		debutPrecedent = time.Date(anneePrecedente, 1, 1, 0, 0, 0, 0, dateDebut.Location())
		finPrecedent = time.Date(anneePrecedente, 12, 31, 23, 59, 59, 0, dateDebut.Location())

	default:
		// Par défaut: utiliser les dates fournies et les décaler de -1 an
		// Période actuelle: dateDebut à dateFin (fournis par l'API)
		// Période précédente: (dateDebut - 1 an) à (dateFin - 1 an)
		debutPrecedent = dateDebut.AddDate(-1, 0, 0)
		finPrecedent = dateFin.AddDate(-1, 0, 0)
	}

	// Récupérer les stats de la période actuelle (sans évolution pour éviter récursion)
	statsActuelles, _ := r.GetStatistiques(ctx, commissariatID, &dateDebut, &dateFin, nil)
	totalActuel := 0
	resoluesActuel := 0
	if statsActuelles != nil {
		if t, ok := statsActuelles["total"].(int); ok {
			totalActuel = t
		}
		if res, ok := statsActuelles["resolues"].(int); ok {
			resoluesActuel = res
		}
	}

	// Récupérer les stats de la période précédente (sans évolution)
	statsPrecedentes, _ := r.GetStatistiques(ctx, commissariatID, &debutPrecedent, &finPrecedent, nil)
	totalPrecedent := 0
	resoluesPrecedent := 0

	if statsPrecedentes != nil {
		if t, ok := statsPrecedentes["total"].(int); ok {
			totalPrecedent = t
		}
		if res, ok := statsPrecedentes["resolues"].(int); ok {
			resoluesPrecedent = res
		}
	}

	// Debug logs
	r.logger.Info("Calcul évolution",
		zap.String("periode", typePeriode),
		zap.Time("debutActuel", dateDebut),
		zap.Time("finActuel", dateFin),
		zap.Time("debutPrecedent", debutPrecedent),
		zap.Time("finPrecedent", finPrecedent),
		zap.Int("totalActuel", totalActuel),
		zap.Int("totalPrecedent", totalPrecedent),
		zap.Int("resoluesActuel", resoluesActuel),
		zap.Int("resoluesPrecedent", resoluesPrecedent),
	)

	// Calculer les différences
	diffAlertes := totalActuel - totalPrecedent
	diffResolutions := resoluesActuel - resoluesPrecedent

	// Formater avec signes
	evolutionAlertes := formatEvolution(diffAlertes)
	evolutionResolution := formatEvolution(diffResolutions)

	return map[string]string{
		"alertes":     evolutionAlertes,
		"resolutions": evolutionResolution,
	}
}

// formatEvolution formate un nombre avec son signe
func formatEvolution(diff int) string {
	if diff > 0 {
		return fmt.Sprintf("+%d", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%d", diff)
	}
	return "0"
}
