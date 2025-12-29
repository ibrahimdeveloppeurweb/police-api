package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/commissariat"
	"police-trafic-api-frontend-aligned/ent/objetperdu"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ObjetPerduRepository defines objet perdu repository interface
type ObjetPerduRepository interface {
	Create(ctx context.Context, input *CreateObjetPerduInput) (*ent.ObjetPerdu, error)
	GetByID(ctx context.Context, id string) (*ent.ObjetPerdu, error)
	GetByNumero(ctx context.Context, numero string) (*ent.ObjetPerdu, error)
	List(ctx context.Context, filters *ObjetPerduFilters) ([]*ent.ObjetPerdu, error)
	Count(ctx context.Context, filters *ObjetPerduFilters) (int, error)
	Update(ctx context.Context, id string, input *UpdateObjetPerduInput) (*ent.ObjetPerdu, error)
	Delete(ctx context.Context, id string) error
	UpdateStatut(ctx context.Context, id string, statut string) (*ent.ObjetPerdu, error)
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin *time.Time, periode *string) (map[string]interface{}, error)
}

// CreateObjetPerduInput represents input for creating objet perdu
type CreateObjetPerduInput struct {
	ID                 string
	Numero             string
	TypeObjet          string
	Description        string
	ValeurEstimee      *string
	Couleur            *string
	DetailsSpecifiques map[string]interface{}
	
	// Nouveaux champs pour le mode contenant
	IsContainer        bool
	ContainerDetails   map[string]interface{}
	
	Declarant          map[string]interface{}
	LieuPerte          string
	AdresseLieu        *string
	DatePerte          time.Time
	HeurePerte         *string
	Statut             string
	DateDeclaration    time.Time
	Observations       *string
	CommissariatID     string
	AgentID            string
}

// UpdateObjetPerduInput represents input for updating objet perdu
type UpdateObjetPerduInput struct {
	TypeObjet          *string
	Description        *string
	ValeurEstimee      *string
	Couleur            *string
	DetailsSpecifiques map[string]interface{}
	
	// Nouveaux champs pour le mode contenant
	IsContainer        *bool
	ContainerDetails   map[string]interface{}
	
	Declarant          map[string]interface{}
	LieuPerte          *string
	AdresseLieu        *string
	DatePerte          *time.Time
	HeurePerte         *string
	Statut             *string
	DateRetrouve       *time.Time
	Observations       *string
}

// ObjetPerduFilters represents filters for listing objets perdus
type ObjetPerduFilters struct {
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

// objetPerduRepository implements ObjetPerduRepository
type objetPerduRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewObjetPerduRepository creates a new objet perdu repository
func NewObjetPerduRepository(client *ent.Client, logger *zap.Logger) ObjetPerduRepository {
	return &objetPerduRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new objet perdu
func (r *objetPerduRepository) Create(ctx context.Context, input *CreateObjetPerduInput) (*ent.ObjetPerdu, error) {
	query := r.client.ObjetPerdu.Create().
		SetID(uuid.MustParse(input.ID)).
		SetNumero(input.Numero).
		SetTypeObjet(input.TypeObjet).
		SetDescription(input.Description).
		SetDeclarant(input.Declarant).
		SetLieuPerte(input.LieuPerte).
		SetDatePerte(input.DatePerte).
		SetDateDeclaration(input.DateDeclaration).
		SetCommissariatID(uuid.MustParse(input.CommissariatID)).
		SetAgentID(uuid.MustParse(input.AgentID)).
		SetStatut(objetperdu.Statut(input.Statut)).
		SetIsContainer(input.IsContainer)

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
	if input.HeurePerte != nil {
		query.SetHeurePerte(*input.HeurePerte)
	}
	if input.Statut != "" {
		query.SetStatut(objetperdu.Statut(input.Statut))
	}
	if input.Observations != nil {
		query.SetObservations(*input.Observations)
	}

	objet, err := query.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create objet perdu", zap.Error(err))
		return nil, fmt.Errorf("failed to create objet perdu: %w", err)
	}

	return objet, nil
}

// GetByID gets an objet perdu by ID
func (r *objetPerduRepository) GetByID(ctx context.Context, id string) (*ent.ObjetPerdu, error) {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	objet, err := r.client.ObjetPerdu.Query().
		WithCommissariat().
		WithAgent().
		Where(objetperdu.ID(objetID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to get objet perdu: %w", err)
	}

	return objet, nil
}

// GetByNumero gets an objet perdu by numero
func (r *objetPerduRepository) GetByNumero(ctx context.Context, numero string) (*ent.ObjetPerdu, error) {
	objet, err := r.client.ObjetPerdu.Query().
		WithCommissariat().
		WithAgent().
		Where(objetperdu.Numero(numero)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to get objet perdu: %w", err)
	}

	return objet, nil
}

// List lists objets perdus with filters
func (r *objetPerduRepository) List(ctx context.Context, filters *ObjetPerduFilters) ([]*ent.ObjetPerdu, error) {
	query := r.client.ObjetPerdu.Query().
		WithCommissariat().
		WithAgent()

	if filters.Statut != nil {
		query = query.Where(objetperdu.StatutEQ(objetperdu.Statut(*filters.Statut)))
	}
	if filters.TypeObjet != nil {
		query = query.Where(objetperdu.TypeObjet(*filters.TypeObjet))
	}
	if filters.IsContainer != nil {
		query = query.Where(objetperdu.IsContainer(*filters.IsContainer))
	}
	if filters.CommissariatID != nil {
		commissariatID, err := uuid.Parse(*filters.CommissariatID)
		if err == nil {
			query = query.Where(objetperdu.HasCommissariatWith(commissariat.ID(commissariatID)))
		}
	}
	if filters.AgentID != nil {
		agentID, err := uuid.Parse(*filters.AgentID)
		if err == nil {
			query = query.Where(objetperdu.HasAgentWith(user.ID(agentID)))
		}
	}
	if filters.DateDebut != nil {
		query = query.Where(objetperdu.DateDeclarationGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(objetperdu.DateDeclarationLTE(*filters.DateFin))
	}
	if filters.Search != nil && *filters.Search != "" {
		search := *filters.Search
		query = query.Where(
			objetperdu.Or(
				objetperdu.NumeroContains(search),
				objetperdu.TypeObjetContains(search),
				objetperdu.DescriptionContains(search),
			),
		)
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	query = query.Order(ent.Desc(objetperdu.FieldDateDeclaration))

	objets, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list objets perdus: %w", err)
	}

	return objets, nil
}

// Count counts objets perdus with filters
func (r *objetPerduRepository) Count(ctx context.Context, filters *ObjetPerduFilters) (int, error) {
	query := r.client.ObjetPerdu.Query()

	if filters.Statut != nil {
		query = query.Where(objetperdu.StatutEQ(objetperdu.Statut(*filters.Statut)))
	}
	if filters.TypeObjet != nil {
		query = query.Where(objetperdu.TypeObjet(*filters.TypeObjet))
	}
	if filters.IsContainer != nil {
		query = query.Where(objetperdu.IsContainer(*filters.IsContainer))
	}
	if filters.CommissariatID != nil {
		commissariatID, err := uuid.Parse(*filters.CommissariatID)
		if err == nil {
			query = query.Where(objetperdu.HasCommissariatWith(commissariat.ID(commissariatID)))
		}
	}
	if filters.AgentID != nil {
		agentID, err := uuid.Parse(*filters.AgentID)
		if err == nil {
			query = query.Where(objetperdu.HasAgentWith(user.ID(agentID)))
		}
	}
	if filters.DateDebut != nil {
		query = query.Where(objetperdu.DateDeclarationGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(objetperdu.DateDeclarationLTE(*filters.DateFin))
	}
	if filters.Search != nil && *filters.Search != "" {
		search := *filters.Search
		query = query.Where(
			objetperdu.Or(
				objetperdu.NumeroContains(search),
				objetperdu.TypeObjetContains(search),
				objetperdu.DescriptionContains(search),
			),
		)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count objets perdus: %w", err)
	}

	return count, nil
}

// Update updates an objet perdu
func (r *objetPerduRepository) Update(ctx context.Context, id string, input *UpdateObjetPerduInput) (*ent.ObjetPerdu, error) {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	query := r.client.ObjetPerdu.UpdateOneID(objetID)

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
	if input.IsContainer != nil {
		query = query.SetIsContainer(*input.IsContainer)
	}
	if input.ContainerDetails != nil {
		query = query.SetContainerDetails(input.ContainerDetails)
	}
	if input.Declarant != nil {
		query = query.SetDeclarant(input.Declarant)
	}
	if input.LieuPerte != nil {
		query = query.SetLieuPerte(*input.LieuPerte)
	}
	if input.AdresseLieu != nil {
		query = query.SetAdresseLieu(*input.AdresseLieu)
	}
	if input.DatePerte != nil {
		query = query.SetDatePerte(*input.DatePerte)
	}
	if input.HeurePerte != nil {
		query = query.SetHeurePerte(*input.HeurePerte)
	}
	if input.Statut != nil {
		query = query.SetStatut(objetperdu.Statut(*input.Statut))
	}
	if input.DateRetrouve != nil {
		query = query.SetDateRetrouve(*input.DateRetrouve)
	}
	if input.Observations != nil {
		query = query.SetObservations(*input.Observations)
	}

	objet, err := query.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to update objet perdu: %w", err)
	}

	return objet, nil
}

// UpdateStatut updates the statut of an objet perdu
func (r *objetPerduRepository) UpdateStatut(ctx context.Context, id string, statut string) (*ent.ObjetPerdu, error) {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	objet, err := r.client.ObjetPerdu.UpdateOneID(objetID).
		SetStatut(objetperdu.Statut(statut)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to update statut: %w", err)
	}

	return objet, nil
}

// Delete deletes an objet perdu
func (r *objetPerduRepository) Delete(ctx context.Context, id string) error {
	objetID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	err = r.client.ObjetPerdu.DeleteOneID(objetID).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("objet perdu not found")
		}
		return fmt.Errorf("failed to delete objet perdu: %w", err)
	}

	return nil
}

// GetStatistiques calcule les statistiques des objets perdus
func (r *objetPerduRepository) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin *time.Time, periode *string) (map[string]interface{}, error) {
	query := r.client.ObjetPerdu.Query()

	if commissariatID != nil {
		commID, _ := uuid.Parse(*commissariatID)
		query = query.Where(objetperdu.HasCommissariatWith(commissariat.ID(commID)))
	}
	if dateDebut != nil {
		query = query.Where(objetperdu.DateDeclarationGTE(*dateDebut))
	}
	if dateFin != nil {
		query = query.Where(objetperdu.DateDeclarationLTE(*dateFin))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Compter par statut
	enRecherche, _ := query.Clone().Where(objetperdu.StatutEQ(objetperdu.StatutEN_RECHERCHE)).Count(ctx)
	retrouves, _ := query.Clone().Where(objetperdu.StatutEQ(objetperdu.StatutRETROUVÉ)).Count(ctx)
	clotures, _ := query.Clone().Where(objetperdu.StatutEQ(objetperdu.StatutCLÔTURÉ)).Count(ctx)

	stats := map[string]interface{}{
		"total":       total,
		"enRecherche": enRecherche,
		"retrouves":   retrouves,
		"clotures":    clotures,
	}

	// Calculer le taux de retrouvaille
	if total > 0 {
		tauxRetrouve := float64(retrouves) / float64(total) * 100
		stats["tauxRetrouve"] = tauxRetrouve
	} else {
		stats["tauxRetrouve"] = 0.0
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
		stats["evolutionEnRecherche"] = evolution["enRecherche"]
		stats["evolutionRetrouves"] = evolution["retrouves"]
		stats["evolutionClotures"] = evolution["clotures"]
		stats["evolutionTauxRetrouve"] = evolution["tauxRetrouve"]
		r.logger.Info("Évolutions calculées",
			zap.String("evolutionTotal", evolution["total"]),
			zap.String("evolutionEnRecherche", evolution["enRecherche"]),
			zap.String("evolutionRetrouves", evolution["retrouves"]),
			zap.String("evolutionClotures", evolution["clotures"]),
			zap.String("evolutionTauxRetrouve", evolution["tauxRetrouve"]),
		)
	} else {
		// Si pas de période, retourner "0" pour toutes les évolutions
		r.logger.Info("Pas de période fournie, évolutions à 0",
			zap.Bool("hasPeriode", periode != nil),
			zap.Bool("hasDateDebut", dateDebut != nil),
			zap.Bool("hasDateFin", dateFin != nil),
		)
		stats["evolutionTotal"] = "0"
		stats["evolutionEnRecherche"] = "0"
		stats["evolutionRetrouves"] = "0"
		stats["evolutionClotures"] = "0"
		stats["evolutionTauxRetrouve"] = "0"
	}

	// S'assurer que les évolutions sont TOUJOURS présentes dans le map
	if _, exists := stats["evolutionTotal"]; !exists {
		stats["evolutionTotal"] = "0"
	}
	if _, exists := stats["evolutionEnRecherche"]; !exists {
		stats["evolutionEnRecherche"] = "0"
	}
	if _, exists := stats["evolutionRetrouves"]; !exists {
		stats["evolutionRetrouves"] = "0"
	}
	if _, exists := stats["evolutionClotures"]; !exists {
		stats["evolutionClotures"] = "0"
	}
	if _, exists := stats["evolutionTauxRetrouve"]; !exists {
		stats["evolutionTauxRetrouve"] = "0"
	}

	r.logger.Info("Stats retournées par repository",
		zap.Any("stats", stats),
		zap.String("evolutionTotal", stats["evolutionTotal"].(string)),
		zap.String("evolutionEnRecherche", stats["evolutionEnRecherche"].(string)),
		zap.String("evolutionRetrouves", stats["evolutionRetrouves"].(string)),
		zap.String("evolutionClotures", stats["evolutionClotures"].(string)),
		zap.String("evolutionTauxRetrouve", stats["evolutionTauxRetrouve"].(string)),
	)

	return stats, nil
}

// calculerEvolutionPeriode calcule l'évolution par rapport à la période précédente
func (r *objetPerduRepository) calculerEvolutionPeriode(ctx context.Context, commissariatID *string, dateDebut, dateFin time.Time, typePeriode string) map[string]string {
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
	enRechercheActuel := 0
	retrouvesActuel := 0
	cloturesActuel := 0
	tauxRetrouveActuel := 0.0
	if statsActuelles != nil {
		if t, ok := statsActuelles["total"].(int); ok {
			totalActuel = t
		}
		if er, ok := statsActuelles["enRecherche"].(int); ok {
			enRechercheActuel = er
		}
		if ret, ok := statsActuelles["retrouves"].(int); ok {
			retrouvesActuel = ret
		}
		if cl, ok := statsActuelles["clotures"].(int); ok {
			cloturesActuel = cl
		}
		if taux, ok := statsActuelles["tauxRetrouve"].(float64); ok {
			tauxRetrouveActuel = taux
		}
	}
	
	// Récupérer les stats de la période précédente (sans évolution)
	statsPrecedentes, _ := r.GetStatistiques(ctx, commissariatID, &debutPrecedent, &finPrecedent, nil)
	totalPrecedent := 0
	enRecherchePrecedent := 0
	retrouvesPrecedent := 0
	cloturesPrecedent := 0
	tauxRetrouvePrecedent := 0.0
	
	if statsPrecedentes != nil {
		if t, ok := statsPrecedentes["total"].(int); ok {
			totalPrecedent = t
		}
		if er, ok := statsPrecedentes["enRecherche"].(int); ok {
			enRecherchePrecedent = er
		}
		if ret, ok := statsPrecedentes["retrouves"].(int); ok {
			retrouvesPrecedent = ret
		}
		if cl, ok := statsPrecedentes["clotures"].(int); ok {
			cloturesPrecedent = cl
		}
		if taux, ok := statsPrecedentes["tauxRetrouve"].(float64); ok {
			tauxRetrouvePrecedent = taux
		}
	}
	
	// Debug logs
	r.logger.Info("Calcul évolution objets perdus",
		zap.String("periode", typePeriode),
		zap.Time("debutActuel", dateDebut),
		zap.Time("finActuel", dateFin),
		zap.Time("debutPrecedent", debutPrecedent),
		zap.Time("finPrecedent", finPrecedent),
		zap.Int("totalActuel", totalActuel),
		zap.Int("totalPrecedent", totalPrecedent),
		zap.Int("enRechercheActuel", enRechercheActuel),
		zap.Int("enRecherchePrecedent", enRecherchePrecedent),
		zap.Int("retrouvesActuel", retrouvesActuel),
		zap.Int("retrouvesPrecedent", retrouvesPrecedent),
		zap.Int("cloturesActuel", cloturesActuel),
		zap.Int("cloturesPrecedent", cloturesPrecedent),
	)
	
	// Calculer les différences
	diffTotal := totalActuel - totalPrecedent
	diffEnRecherche := enRechercheActuel - enRecherchePrecedent
	diffRetrouves := retrouvesActuel - retrouvesPrecedent
	diffClotures := cloturesActuel - cloturesPrecedent
	diffTauxRetrouve := tauxRetrouveActuel - tauxRetrouvePrecedent
	
	// Formater avec signes
	evolutionTotal := formatEvolutionObjetPerdu(diffTotal)
	evolutionEnRecherche := formatEvolutionObjetPerdu(diffEnRecherche)
	evolutionRetrouves := formatEvolutionObjetPerdu(diffRetrouves)
	evolutionClotures := formatEvolutionObjetPerdu(diffClotures)
	// Pour le taux, on arrondit à 1 décimale
	evolutionTauxRetrouve := formatEvolutionFloatObjetPerdu(diffTauxRetrouve)
	
	return map[string]string{
		"total":         evolutionTotal,
		"enRecherche":   evolutionEnRecherche,
		"retrouves":     evolutionRetrouves,
		"clotures":      evolutionClotures,
		"tauxRetrouve":  evolutionTauxRetrouve,
	}
}

// formatEvolutionObjetPerdu formate un nombre avec son signe
func formatEvolutionObjetPerdu(diff int) string {
	if diff > 0 {
		return fmt.Sprintf("+%d", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%d", diff)
	}
	return "0"
}

// formatEvolutionFloatObjetPerdu formate un nombre décimal avec son signe
func formatEvolutionFloatObjetPerdu(diff float64) string {
	if diff > 0 {
		return fmt.Sprintf("+%.1f", diff)
	} else if diff < 0 {
		return fmt.Sprintf("%.1f", diff)
	}
	return "0"
}
