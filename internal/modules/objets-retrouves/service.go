package objetsretrouves

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines objets retrouves service interface
type Service interface {
	Create(ctx context.Context, req *CreateObjetRetrouveRequest, agentID, commissariatID string) (*ObjetRetrouveResponse, error)
	GetByID(ctx context.Context, id string) (*ObjetRetrouveResponse, error)
	List(ctx context.Context, filters *FilterObjetsRetrouvesRequest, role, userID, commissariatID string) (*ListObjetsRetrouvesResponse, error)
	Update(ctx context.Context, id string, req *UpdateObjetRetrouveRequest) (*ObjetRetrouveResponse, error)
	UpdateStatut(ctx context.Context, id string, req *UpdateStatutRequest, agentID string) (*ObjetRetrouveResponse, error)
	Delete(ctx context.Context, id string) error
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesObjetsRetrouvesResponse, error)
	GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardResponse, error)
}

// service implements Service interface
type service struct {
	objetRetrouveRepo repository.ObjetRetrouveRepository
	commissariatRepo  repository.CommissariatRepository
	userRepo          repository.UserRepository
	config            *config.Config
	logger            *zap.Logger
}

// NewService creates a new objets retrouves service
func NewService(
	objetRetrouveRepo repository.ObjetRetrouveRepository,
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	return &service{
		objetRetrouveRepo: objetRetrouveRepo,
		commissariatRepo:  commissariatRepo,
		userRepo:          userRepo,
		config:            cfg,
		logger:            logger,
	}
}

// generateNumero génère un numéro unique pour l'objet retrouvé
func (s *service) generateNumero(ctx context.Context, commissariatID string) (string, error) {
	// Récupérer le commissariat pour obtenir la ville
	commissariat, err := s.commissariatRepo.GetByID(ctx, commissariatID)
	if err != nil {
		return "", fmt.Errorf("commissariat not found")
	}

	year := time.Now().Year()

	// Extraire les 3 premières lettres de la ville en majuscules
	ville := strings.ToUpper(commissariat.Ville)
	villePrefix := ville
	if len(ville) > 3 {
		villePrefix = ville[:3]
	} else if len(ville) < 3 {
		// Si la ville a moins de 3 lettres, compléter avec des X
		villePrefix = ville + strings.Repeat("X", 3-len(ville))
	}

	// Chercher le dernier objet retrouvé du commissariat pour cette année
	filters := &repository.ObjetRetrouveFilters{
		CommissariatID: &commissariatID,
		Limit:          1000, // Récupérer beaucoup pour trouver le max
		Offset:         0,
	}

	objets, err := s.objetRetrouveRepo.List(ctx, filters)

	nextNumber := 1

	// Si des objets existent, extraire le dernier numéro
	if err == nil && len(objets) > 0 {
		// Trouver le numéro max en parcourant tous les objets de l'année
		maxNum := 0
		for _, objet := range objets {
			// Format: OBR-VILLE-COM-YYYY-NNNN
			parts := strings.Split(objet.Numero, "-")
			if len(parts) == 5 {
				if num, err := strconv.Atoi(parts[4]); err == nil && num > maxNum {
					maxNum = num
				}
			}
		}
		if maxNum > 0 {
			nextNumber = maxNum + 1
		}
	}

	// Générer le numéro avec retry pour éviter les collisions
	maxRetries := 10
	for retry := 0; retry < maxRetries; retry++ {
		numero := fmt.Sprintf("OBR-%s-COM-%d-%04d", villePrefix, year, nextNumber+retry)

		// Vérifier si le numéro existe déjà
		_, err := s.objetRetrouveRepo.GetByNumero(ctx, numero)
		if err != nil {
			// Le numéro n'existe pas, on peut l'utiliser
			return numero, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique numero after %d retries", maxRetries)
}

// Create creates a new objet retrouve
func (s *service) Create(ctx context.Context, req *CreateObjetRetrouveRequest, agentID, commissariatID string) (*ObjetRetrouveResponse, error) {
	// Générer le numéro unique
	numero, err := s.generateNumero(ctx, commissariatID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate numero: %w", err)
	}

	// Parser la date de trouvaille
	dateTrouvaille, err := time.Parse("2006-01-02", req.DateTrouvaille)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Déterminer si c'est un contenant
	isContainer := false
	if req.IsContainer != nil {
		isContainer = *req.IsContainer
	}

	// Construire les détails spécifiques
	detailsSpecifiques := make(map[string]interface{})
	if req.DetailsSpecifiques != nil {
		for key, value := range req.DetailsSpecifiques {
			if value != nil && value != "" {
				detailsSpecifiques[key] = value
			}
		}
	}

	// Ajouter les détails du contenant si applicable
	var containerDetails map[string]interface{}
	if isContainer && req.ContainerDetails != nil {
		containerDetails = map[string]interface{}{
			"type": req.ContainerDetails.Type,
		}
		if req.ContainerDetails.Couleur != nil {
			containerDetails["couleur"] = *req.ContainerDetails.Couleur
		}
		if req.ContainerDetails.Marque != nil {
			containerDetails["marque"] = *req.ContainerDetails.Marque
		}
		if req.ContainerDetails.Taille != nil {
			containerDetails["taille"] = *req.ContainerDetails.Taille
		}
		if req.ContainerDetails.SignesDistinctifs != nil {
			containerDetails["signesDistinctifs"] = *req.ContainerDetails.SignesDistinctifs
		}
		if len(req.ContainerDetails.Inventory) > 0 {
			containerDetails["inventory"] = req.ContainerDetails.Inventory
		}
	}

	// Construire le déposant
	deposant := map[string]interface{}{
		"nom":       req.Deposant.Nom,
		"prenom":    req.Deposant.Prenom,
		"telephone": req.Deposant.Telephone,
	}
	if req.Deposant.Email != nil {
		deposant["email"] = *req.Deposant.Email
	}
	if req.Deposant.Adresse != nil {
		deposant["adresse"] = *req.Deposant.Adresse
	}
	if req.Deposant.CNI != nil {
		deposant["cni"] = *req.Deposant.CNI
	}

	// Créer l'objet retrouvé
	repoInput := &repository.CreateObjetRetrouveInput{
		ID:                 uuid.New().String(),
		Numero:             numero,
		TypeObjet:          req.TypeObjet,
		Description:        req.Description,
		ValeurEstimee:      req.ValeurEstimee,
		Couleur:            req.Couleur,
		DetailsSpecifiques: detailsSpecifiques,
		IsContainer:        isContainer,
		ContainerDetails:   containerDetails,
		Deposant:           deposant,
		LieuTrouvaille:     req.LieuTrouvaille,
		AdresseLieu:        req.AdresseLieu,
		DateTrouvaille:     dateTrouvaille,
		HeureTrouvaille:    req.HeureTrouvaille,
		Statut:             string(StatutObjetRetrouveDisponible),
		DateDepot:          time.Now(),
		Observations:       req.Observations,
		CommissariatID:     commissariatID,
		AgentID:            agentID,
	}

	s.logger.Info("Saving objet retrouve to database",
		zap.String("numero", numero),
		zap.String("agent_id", agentID),
		zap.String("commissariat_id", commissariatID),
	)

	objetEnt, err := s.objetRetrouveRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create objet retrouve", zap.Error(err))
		return nil, fmt.Errorf("failed to create objet retrouve: %w", err)
	}

	s.logger.Info("Objet retrouve created successfully",
		zap.String("id", objetEnt.ID.String()),
		zap.String("numero", objetEnt.Numero),
	)

	return s.formatObjetRetrouve(objetEnt), nil
}

// GetByID gets an objet retrouve by ID
func (s *service) GetByID(ctx context.Context, id string) (*ObjetRetrouveResponse, error) {
	objet, err := s.objetRetrouveRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to get objet retrouve: %w", err)
	}

	return s.formatObjetRetrouve(objet), nil
}

// List lists objets retrouves with filters
func (s *service) List(ctx context.Context, filters *FilterObjetsRetrouvesRequest, role, userID, commissariatID string) (*ListObjetsRetrouvesResponse, error) {
	repoFilters := &repository.ObjetRetrouveFilters{}

	// CORRECTION: Ajouter tous les filtres
	if filters.Statut != nil {
		repoFilters.Statut = filters.Statut
		s.logger.Info("Service - Filter Statut", zap.Stringp("statut", filters.Statut))
	}
	if filters.TypeObjet != nil {
		repoFilters.TypeObjet = filters.TypeObjet
		s.logger.Info("Service - Filter TypeObjet", zap.Stringp("typeObjet", filters.TypeObjet))
	}
	if filters.IsContainer != nil {
		repoFilters.IsContainer = filters.IsContainer
		s.logger.Info("Service - Filter IsContainer", zap.Boolp("isContainer", filters.IsContainer))
	}
	if filters.CommissariatID != nil {
		repoFilters.CommissariatID = filters.CommissariatID
		s.logger.Info("Service - Filter CommissariatID from query", zap.Stringp("commissariatID", filters.CommissariatID))
	} else if role != "ADMIN" && commissariatID != "" {
		// Pour les non-admins, filtrer par commissariat
		repoFilters.CommissariatID = &commissariatID
		s.logger.Info("Service - Filter CommissariatID from context", zap.String("commissariatID", commissariatID))
	}
	if filters.DateDebut != nil {
		repoFilters.DateDebut = filters.DateDebut
		s.logger.Info("Service - Filter DateDebut", zap.Time("dateDebut", *filters.DateDebut))
	}
	if filters.DateFin != nil {
		repoFilters.DateFin = filters.DateFin
		s.logger.Info("Service - Filter DateFin", zap.Time("dateFin", *filters.DateFin))
	}
	if filters.Search != nil {
		repoFilters.Search = filters.Search
		s.logger.Info("Service - Filter Search", zap.Stringp("search", filters.Search))
	}

	// Pagination
	page := 1
	limit := 50
	if filters.Page > 0 {
		page = filters.Page
	}
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	repoFilters.Offset = (page - 1) * limit
	repoFilters.Limit = limit

	s.logger.Info("Service List - Calling repository with filters",
		zap.Any("repoFilters", repoFilters),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	objets, err := s.objetRetrouveRepo.List(ctx, repoFilters)
	if err != nil {
		s.logger.Error("Repository List failed", zap.Error(err))
		return nil, fmt.Errorf("failed to list objets retrouves: %w", err)
	}

	total, err := s.objetRetrouveRepo.Count(ctx, repoFilters)
	if err != nil {
		s.logger.Error("Repository Count failed", zap.Error(err))
		return nil, fmt.Errorf("failed to count objets retrouves: %w", err)
	}

	s.logger.Info("Service List - Results from repository",
		zap.Int("objets_count", len(objets)),
		zap.Int("total", total),
	)

	responses := make([]ObjetRetrouveResponse, len(objets))
	for i, objet := range objets {
		responses[i] = *s.formatObjetRetrouve(objet)
	}

	return &ListObjetsRetrouvesResponse{
		Objets: responses,
		Total:  int64(total),
		Page:   page,
		Limit:  limit,
	}, nil
}

// Update updates an objet retrouve
func (s *service) Update(ctx context.Context, id string, req *UpdateObjetRetrouveRequest) (*ObjetRetrouveResponse, error) {
	repoInput := &repository.UpdateObjetRetrouveInput{}

	if req.TypeObjet != nil {
		repoInput.TypeObjet = req.TypeObjet
	}
	if req.Description != nil {
		repoInput.Description = req.Description
	}
	if req.ValeurEstimee != nil {
		repoInput.ValeurEstimee = req.ValeurEstimee
	}
	if req.Couleur != nil {
		repoInput.Couleur = req.Couleur
	}
	if req.DetailsSpecifiques != nil {
		repoInput.DetailsSpecifiques = req.DetailsSpecifiques
	}
	if req.Deposant != nil {
		deposant := map[string]interface{}{
			"nom":       req.Deposant.Nom,
			"prenom":    req.Deposant.Prenom,
			"telephone": req.Deposant.Telephone,
		}
		if req.Deposant.Email != nil {
			deposant["email"] = *req.Deposant.Email
		}
		if req.Deposant.Adresse != nil {
			deposant["adresse"] = *req.Deposant.Adresse
		}
		if req.Deposant.CNI != nil {
			deposant["cni"] = *req.Deposant.CNI
		}
		repoInput.Deposant = deposant
	}
	if req.LieuTrouvaille != nil {
		repoInput.LieuTrouvaille = req.LieuTrouvaille
	}
	if req.AdresseLieu != nil {
		repoInput.AdresseLieu = req.AdresseLieu
	}
	if req.DateTrouvaille != nil {
		dateTrouvaille, err := time.Parse("2006-01-02", *req.DateTrouvaille)
		if err == nil {
			repoInput.DateTrouvaille = &dateTrouvaille
		}
	}
	if req.HeureTrouvaille != nil {
		repoInput.HeureTrouvaille = req.HeureTrouvaille
	}
	if req.Observations != nil {
		repoInput.Observations = req.Observations
	}
	if req.IsContainer != nil {
		repoInput.IsContainer = *req.IsContainer
	}
	if req.ContainerDetails != nil {
		containerDetails := map[string]interface{}{
			"type": req.ContainerDetails.Type,
		}
		if req.ContainerDetails.Couleur != nil {
			containerDetails["couleur"] = *req.ContainerDetails.Couleur
		}
		if req.ContainerDetails.Marque != nil {
			containerDetails["marque"] = *req.ContainerDetails.Marque
		}
		if req.ContainerDetails.Taille != nil {
			containerDetails["taille"] = *req.ContainerDetails.Taille
		}
		if req.ContainerDetails.SignesDistinctifs != nil {
			containerDetails["signesDistinctifs"] = *req.ContainerDetails.SignesDistinctifs
		}
		if len(req.ContainerDetails.Inventory) > 0 {
			containerDetails["inventory"] = req.ContainerDetails.Inventory
		}
		repoInput.ContainerDetails = containerDetails
	}

	objet, err := s.objetRetrouveRepo.Update(ctx, id, repoInput)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to update objet retrouve: %w", err)
	}

	return s.formatObjetRetrouve(objet), nil
}

// UpdateStatut updates the statut of an objet retrouve
func (s *service) UpdateStatut(ctx context.Context, id string, req *UpdateStatutRequest, agentID string) (*ObjetRetrouveResponse, error) {
	var dateRestitution *time.Time
	var proprietaire map[string]interface{}

	if req.DateRestitution != nil {
		dateRestitution = req.DateRestitution
	}
	if req.Proprietaire != nil {
		proprietaire = map[string]interface{}{
			"nom":       req.Proprietaire.Nom,
			"prenom":    req.Proprietaire.Prenom,
			"telephone": req.Proprietaire.Telephone,
		}
		if req.Proprietaire.Email != nil {
			proprietaire["email"] = *req.Proprietaire.Email
		}
		if req.Proprietaire.Adresse != nil {
			proprietaire["adresse"] = *req.Proprietaire.Adresse
		}
		if req.Proprietaire.CNI != nil {
			proprietaire["cni"] = *req.Proprietaire.CNI
		}
	}

	objet, err := s.objetRetrouveRepo.UpdateStatut(ctx, id, req.Statut, dateRestitution, proprietaire)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return nil, fmt.Errorf("objet retrouve not found")
		}
		return nil, fmt.Errorf("failed to update statut: %w", err)
	}

	return s.formatObjetRetrouve(objet), nil
}

// Delete deletes an objet retrouve
func (s *service) Delete(ctx context.Context, id string) error {
	err := s.objetRetrouveRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "objet retrouve not found" {
			return fmt.Errorf("objet retrouve not found")
		}
		return fmt.Errorf("failed to delete objet retrouve: %w", err)
	}
	return nil
}

// formatObjetRetrouve formats an ent.ObjetRetrouve to ObjetRetrouveResponse
func (s *service) formatObjetRetrouve(objet *ent.ObjetRetrouve) *ObjetRetrouveResponse {
	response := &ObjetRetrouveResponse{
		ID:                     objet.ID.String(),
		Numero:                 objet.Numero,
		TypeObjet:              objet.TypeObjet,
		Description:            objet.Description,
		Deposant:               objet.Deposant,
		LieuTrouvaille:         objet.LieuTrouvaille,
		DateTrouvaille:         objet.DateTrouvaille,
		DateTrouvailleFormatee: objet.DateTrouvaille.Format("02/01/2006"),
		Statut:                 StatutObjetRetrouve(objet.Statut),
		DateDepot:              objet.DateDepot,
		DateDepotFormatee:      objet.DateDepot.Format("02/01/2006 à 15:04"),
		CreatedAt:              objet.CreatedAt,
		UpdatedAt:              objet.UpdatedAt,
	}

	if objet.ValeurEstimee != nil {
		response.ValeurEstimee = objet.ValeurEstimee
	}
	if objet.Couleur != nil {
		response.Couleur = objet.Couleur
	}
	if objet.DetailsSpecifiques != nil {
		response.DetailsSpecifiques = objet.DetailsSpecifiques
	}
	if objet.AdresseLieu != nil {
		response.AdresseLieu = objet.AdresseLieu
	}
	if objet.HeureTrouvaille != nil {
		response.HeureTrouvaille = objet.HeureTrouvaille
	}
	if objet.DateRestitution != nil {
		response.DateRestitution = objet.DateRestitution
		formatted := objet.DateRestitution.Format("02/01/2006")
		response.DateRestitutionFormatee = &formatted
	}
	if objet.Proprietaire != nil {
		response.Proprietaire = objet.Proprietaire
	}
	if objet.Observations != nil {
		response.Observations = objet.Observations
	}

	// Gérer le mode contenant
	response.IsContainer = objet.IsContainer
	if objet.IsContainer && objet.ContainerDetails != nil {
		containerDetails := &ContainerDetails{}

		if typeVal, ok := objet.ContainerDetails["type"].(string); ok {
			containerDetails.Type = typeVal
		}
		if couleurVal, ok := objet.ContainerDetails["couleur"].(string); ok {
			containerDetails.Couleur = &couleurVal
		}
		if marqueVal, ok := objet.ContainerDetails["marque"].(string); ok {
			containerDetails.Marque = &marqueVal
		}
		if tailleVal, ok := objet.ContainerDetails["taille"].(string); ok {
			containerDetails.Taille = &tailleVal
		}
		if signesVal, ok := objet.ContainerDetails["signesDistinctifs"].(string); ok {
			containerDetails.SignesDistinctifs = &signesVal
		}

		// Gérer l'inventaire
		if inventoryVal, ok := objet.ContainerDetails["inventory"].([]interface{}); ok {
			for _, itemVal := range inventoryVal {
				if itemMap, ok := itemVal.(map[string]interface{}); ok {
					item := InventoryItem{}

					if idVal, ok := itemMap["id"].(float64); ok {
						item.ID = int(idVal)
					}
					if catVal, ok := itemMap["category"].(string); ok {
						item.Category = catVal
					}
					if iconVal, ok := itemMap["icon"].(string); ok {
						item.Icon = iconVal
					}
					if nameVal, ok := itemMap["name"].(string); ok {
						item.Name = nameVal
					}
					if colorVal, ok := itemMap["color"].(string); ok {
						item.Color = colorVal
					}
					if brandVal, ok := itemMap["brand"].(string); ok {
						item.Brand = &brandVal
					}
					if serialVal, ok := itemMap["serial"].(string); ok {
						item.Serial = &serialVal
					}
					if descVal, ok := itemMap["description"].(string); ok {
						item.Description = &descVal
					}
					if idTypeVal, ok := itemMap["identityType"].(string); ok {
						item.IdentityType = &idTypeVal
					}
					if idNumVal, ok := itemMap["identityNumber"].(string); ok {
						item.IdentityNumber = &idNumVal
					}
					if idNameVal, ok := itemMap["identityName"].(string); ok {
						item.IdentityName = &idNameVal
					}
					if cardTypeVal, ok := itemMap["cardType"].(string); ok {
						item.CardType = &cardTypeVal
					}
					if cardBankVal, ok := itemMap["cardBank"].(string); ok {
						item.CardBank = &cardBankVal
					}
					if cardLast4Val, ok := itemMap["cardLast4"].(string); ok {
						item.CardLast4 = &cardLast4Val
					}

					containerDetails.Inventory = append(containerDetails.Inventory, item)
				}
			}
		}

		response.ContainerDetails = containerDetails
	}

	// Agent
	if objet.Edges.Agent != nil {
		response.Agent = &AgentSummary{
			ID:        objet.Edges.Agent.ID.String(),
			Nom:       objet.Edges.Agent.Nom,
			Prenom:    objet.Edges.Agent.Prenom,
			Matricule: objet.Edges.Agent.Matricule,
		}
	}

	// Commissariat
	if objet.Edges.Commissariat != nil {
		response.Commissariat = &CommissariatSummary{
			ID:    objet.Edges.Commissariat.ID.String(),
			Nom:   objet.Edges.Commissariat.Nom,
			Code:  objet.Edges.Commissariat.Code,
			Ville: objet.Edges.Commissariat.Ville,
		}
	}

	return response
}

// GetStatistiques calcule les statistiques des objets retrouvés
func (s *service) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesObjetsRetrouvesResponse, error) {
	s.logger.Info("GetStatistiques called",
		zap.Stringp("commissariatID", commissariatID),
		zap.Stringp("dateDebut", dateDebut),
		zap.Stringp("dateFin", dateFin),
		zap.Stringp("periode", periode),
	)

	var debut, fin *time.Time

	if dateDebut != nil {
		t, err := parseDateTime(*dateDebut)
		if err == nil {
			debut = &t
		} else {
			s.logger.Warn("Failed to parse dateDebut", zap.String("dateDebut", *dateDebut), zap.Error(err))
		}
	}
	if dateFin != nil {
		t, err := parseDateTime(*dateFin)
		if err == nil {
			fin = &t
		} else {
			s.logger.Warn("Failed to parse dateFin", zap.String("dateFin", *dateFin), zap.Error(err))
		}
	}

	// Récupérer les stats (le repository calcule l'évolution directement)
	stats, err := s.objetRetrouveRepo.GetStatistiques(ctx, commissariatID, debut, fin, periode)
	if err != nil {
		s.logger.Error("Failed to get statistics from repository", zap.Error(err))
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	s.logger.Info("Stats reçues du repository",
		zap.Any("stats", stats),
		zap.Bool("hasEvolutionTotal", stats["evolutionTotal"] != nil),
		zap.Bool("hasEvolutionDisponibles", stats["evolutionDisponibles"] != nil),
		zap.Bool("hasEvolutionRestitues", stats["evolutionRestitues"] != nil),
		zap.Bool("hasEvolutionNonReclames", stats["evolutionNonReclames"] != nil),
		zap.Bool("hasEvolutionTauxRestitution", stats["evolutionTauxRestitution"] != nil),
	)

	// Convertir en response avec valeurs par défaut pour les évolutions
	response := &StatistiquesObjetsRetrouvesResponse{
		Total:                    int64(stats["total"].(int)),
		Disponibles:              int64(stats["disponibles"].(int)),
		Restitues:                int64(stats["restitues"].(int)),
		NonReclames:              int64(stats["nonReclames"].(int)),
		EvolutionTotal:           "0", // Valeur par défaut
		EvolutionDisponibles:     "0", // Valeur par défaut
		EvolutionRestitues:       "0", // Valeur par défaut
		EvolutionNonReclames:     "0", // Valeur par défaut
		EvolutionTauxRestitution: "0", // Valeur par défaut
	}

	if tauxRestitution, ok := stats["tauxRestitution"].(float64); ok {
		response.TauxRestitution = tauxRestitution
	}

	// Récupérer les évolutions calculées par le repository (toujours présentes)
	if evolutionTotalVal, exists := stats["evolutionTotal"]; exists {
		if evolutionTotal, ok := evolutionTotalVal.(string); ok {
			response.EvolutionTotal = evolutionTotal
		} else {
			s.logger.Warn("evolutionTotal n'est pas une string", zap.Any("value", evolutionTotalVal))
			response.EvolutionTotal = "0"
		}
	} else {
		s.logger.Warn("evolutionTotal n'existe pas dans le map")
		response.EvolutionTotal = "0"
	}

	if evolutionDisponiblesVal, exists := stats["evolutionDisponibles"]; exists {
		if evolutionDisponibles, ok := evolutionDisponiblesVal.(string); ok {
			response.EvolutionDisponibles = evolutionDisponibles
		} else {
			s.logger.Warn("evolutionDisponibles n'est pas une string", zap.Any("value", evolutionDisponiblesVal))
			response.EvolutionDisponibles = "0"
		}
	} else {
		s.logger.Warn("evolutionDisponibles n'existe pas dans le map")
		response.EvolutionDisponibles = "0"
	}

	if evolutionRestituesVal, exists := stats["evolutionRestitues"]; exists {
		if evolutionRestitues, ok := evolutionRestituesVal.(string); ok {
			response.EvolutionRestitues = evolutionRestitues
		} else {
			s.logger.Warn("evolutionRestitues n'est pas une string", zap.Any("value", evolutionRestituesVal))
			response.EvolutionRestitues = "0"
		}
	} else {
		s.logger.Warn("evolutionRestitues n'existe pas dans le map")
		response.EvolutionRestitues = "0"
	}

	if evolutionNonReclamesVal, exists := stats["evolutionNonReclames"]; exists {
		if evolutionNonReclames, ok := evolutionNonReclamesVal.(string); ok {
			response.EvolutionNonReclames = evolutionNonReclames
		} else {
			s.logger.Warn("evolutionNonReclames n'est pas une string", zap.Any("value", evolutionNonReclamesVal))
			response.EvolutionNonReclames = "0"
		}
	} else {
		s.logger.Warn("evolutionNonReclames n'existe pas dans le map")
		response.EvolutionNonReclames = "0"
	}

	if evolutionTauxRestitutionVal, exists := stats["evolutionTauxRestitution"]; exists {
		if evolutionTauxRestitution, ok := evolutionTauxRestitutionVal.(string); ok {
			response.EvolutionTauxRestitution = evolutionTauxRestitution
		} else {
			s.logger.Warn("evolutionTauxRestitution n'est pas une string", zap.Any("value", evolutionTauxRestitutionVal))
			response.EvolutionTauxRestitution = "0"
		}
	} else {
		s.logger.Warn("evolutionTauxRestitution n'existe pas dans le map")
		response.EvolutionTauxRestitution = "0"
	}

	s.logger.Info("Statistiques response finale",
		zap.Int64("total", response.Total),
		zap.String("evolutionTotal", response.EvolutionTotal),
		zap.String("evolutionDisponibles", response.EvolutionDisponibles),
		zap.String("evolutionRestitues", response.EvolutionRestitues),
		zap.String("evolutionNonReclames", response.EvolutionNonReclames),
		zap.String("evolutionTauxRestitution", response.EvolutionTauxRestitution),
	)

	return response, nil
}

// GetDashboard gets dashboard data for objets retrouves
func (s *service) GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardResponse, error) {
	var debut, fin *time.Time

	if dateDebut != nil {
		t, err := parseDateTime(*dateDebut)
		if err == nil {
			debut = &t
		}
	}
	if dateFin != nil {
		t, err := parseDateTime(*dateFin)
		if err == nil {
			fin = &t
		}
	}

	// Récupérer les statistiques de base
	stats, err := s.objetRetrouveRepo.GetStatistiques(ctx, commissariatID, debut, fin, periode)
	if err != nil {
		return nil, err
	}

	// Convertir en réponse dashboard
	total := int64(stats["total"].(int))
	disponibles := int64(stats["disponibles"].(int))
	restitues := int64(stats["restitues"].(int))
	nonReclames := int64(stats["nonReclames"].(int))
	tauxRestitution := stats["tauxRestitution"].(float64)
	evolutionTotal := stats["evolutionTotal"].(string)
	evolutionDisponibles := stats["evolutionDisponibles"].(string)
	evolutionRestitues := stats["evolutionRestitues"].(string)
	evolutionNonReclames := stats["evolutionNonReclames"].(string)

	dashboardStats := DashboardStatsValue{
		Total:                total,
		EvolutionTotal:       evolutionTotal,
		Disponibles:          disponibles,
		Restitues:            restitues,
		NonReclames:          nonReclames,
		TauxRestitution:      tauxRestitution,
		EvolutionDisponibles: evolutionDisponibles,
		EvolutionRestitues:   evolutionRestitues,
		EvolutionNonReclames: evolutionNonReclames,
	}

	// TopTypes de statistiques par type
	statsTable := []TopTypes{}

	// Définir l'ordre et les labels de TOUS les types d'objet
	typesOrdered := []struct {
		key   string
		label string
	}{
		{"Sac / Sacoche", "Sac / Sacoche"},
		{"Valise / Bagage", "Valise / Bagage"},
		{"Mallette professionnelle", "Mallette professionnelle"},
		{"Sac à dos", "Sac à dos"},

		// Documents et papiers
		{"Documents d'identité", "Documents d'identité"},
		{"Permis de conduire", "Permis de conduire"},
		{"Passeport", "Passeport"},
		{"Carte grise", "Carte grise"},
		{"Carte d'assurance", "Carte d'assurance"},
		{"Carte bancaire", "Carte bancaire"},
		{"Carte d'étudiant", "Carte d'étudiant"},
		{"Carte de sécurité sociale", "Carte de sécurité sociale"},
		{"Livres et documents", "Livres et documents"},
		{"Carnets et agendas", "Carnets et agendas"},
		{"Cahiers et blocs-notes", "Cahiers et blocs-notes"},
		{"Carnet de santé", "Carnet de santé"},
		{"Carnet de vaccination", "Carnet de vaccination"},
		{"Diplômes et certificats", "Diplômes et certificats"},
		{"Contrats et factures", "Contrats et factures"},

		// Électronique et technologie
		{"Téléphone portable", "Téléphone portable"},
		{"Tablette", "Tablette"},
		{"Ordinateur portable", "Ordinateur portable"},
		{"Ordinateur de bureau", "Ordinateur de bureau"},
		{"Souris d'ordinateur", "Souris d'ordinateur"},
		{"Clavier d'ordinateur", "Clavier d'ordinateur"},
		{"Casque audio", "Casque audio"},
		{"Écouteurs", "Écouteurs"},
		{"Enceinte Bluetooth", "Enceinte Bluetooth"},
		{"Appareil photo", "Appareil photo"},
		{"Caméra", "Caméra"},
		{"Caméscope", "Caméscope"},
		{"Montre connectée", "Montre connectée"},
		{"Bracelet connecté", "Bracelet connecté"},
		{"Chargeur téléphone", "Chargeur téléphone"},
		{"Chargeur ordinateur", "Chargeur ordinateur"},
		{"Batterie externe", "Batterie externe"},
		{"Câble USB", "Câble USB"},
		{"Adaptateur secteur", "Adaptateur secteur"},
		{"Disque dur externe", "Disque dur externe"},
		{"Clé USB", "Clé USB"},
		{"Carte mémoire", "Carte mémoire"},
		{"Lecteur MP3/MP4", "Lecteur MP3/MP4"},
		{"Console de jeu portable", "Console de jeu portable"},
		{"Manette de jeu", "Manette de jeu"},
		{"Télécommande", "Télécommande"},
		{"Calculatrice", "Calculatrice"},

		// Accessoires personnels
		{"Montre", "Montre"},
		{"Lunettes de vue", "Lunettes de vue"},
		{"Lunettes de soleil", "Lunettes de soleil"},
		{"Portefeuille", "Portefeuille"},
		{"Porte-monnaie", "Porte-monnaie"},
		{"Clés", "Clés"},
		{"Porte-clés", "Porte-clés"},
		{"Bijoux", "Bijoux"},
		{"Bague", "Bague"},
		{"Collier", "Collier"},
		{"Bracelet", "Bracelet"},
		{"Boucles d'oreilles", "Boucles d'oreilles"},
		{"Broche", "Broche"},
		{"Pendentif", "Pendentif"},
		{"Chaîne", "Chaîne"},
		{"Sac à main", "Sac à main"},
		{"Sac à dos", "Sac à dos"},
		{"Sac de voyage", "Sac de voyage"},
		{"Sac de sport", "Sac de sport"},
		{"Porte-documents", "Porte-documents"},
		{"Trousses et étuis", "Trousses et étuis"},
		{"Parapluie", "Parapluie"},
		{"Chapeau", "Chapeau"},
		{"Casquette", "Casquette"},
		{"Bonnet", "Bonnet"},
		{"Écharpe", "Écharpe"},
		{"Gants", "Gants"},
		{"Ceinture", "Ceinture"},
		{"Cravate", "Cravate"},
		{"Foulard", "Foulard"},

		// Vêtements et chaussures
		{"Vêtements", "Vêtements"},
		{"T-shirt", "T-shirt"},
		{"Chemise", "Chemise"},
		{"Pantalon", "Pantalon"},
		{"Jean", "Jean"},
		{"Robe", "Robe"},
		{"Jupe", "Jupe"},
		{"Veste", "Veste"},
		{"Manteau", "Manteau"},
		{"Blouson", "Blouson"},
		{"Pull", "Pull"},
		{"Sweat-shirt", "Sweat-shirt"},
		{"Short", "Short"},
		{"Maillot de bain", "Maillot de bain"},
		{"Sous-vêtements", "Sous-vêtements"},
		{"Chaussures", "Chaussures"},
		{"Baskets", "Baskets"},
		{"Chaussures de ville", "Chaussures de ville"},
		{"Sandales", "Sandales"},
		{"Bottes", "Bottes"},
		{"Chaussures de sport", "Chaussures de sport"},
		{"Tongs", "Tongs"},
		{"Chaussures de sécurité", "Chaussures de sécurité"},

		// Véhicules
		{"Vélo", "Vélo"},
		{"Vélo électrique", "Vélo électrique"},
		{"Scooter", "Scooter"},
		{"Trottinette", "Trottinette"},
		{"Trottinette électrique", "Trottinette électrique"},
		{"Casque moto", "Casque moto"},
		{"Casque vélo", "Casque vélo"},
		{"Antivol", "Antivol"},
		{"Rétroviseur", "Rétroviseur"},
		{"Plaque d'immatriculation", "Plaque d'immatriculation"},
		{"Accessoires véhicule", "Accessoires véhicule"},

		// Animaux
		{"Animal de compagnie", "Animal de compagnie"},
		{"Chien", "Chien"},
		{"Chat", "Chat"},
		{"Oiseau", "Oiseau"},
		{"Cage d'animal", "Cage d'animal"},
		{"Laisse et collier", "Laisse et collier"},

		// Articles de sport
		{"Articles sportifs", "Articles sportifs"},
		{"Ballon", "Ballon"},
		{"Raquette de tennis", "Raquette de tennis"},
		{"Raquette de badminton", "Raquette de badminton"},
		{"Club de golf", "Club de golf"},
		{"Équipement de fitness", "Équipement de fitness"},
		{"Tapis de sport", "Tapis de sport"},
		{"Haltères", "Haltères"},
		{"Corde à sauter", "Corde à sauter"},
		{"Planche de surf", "Planche de surf"},
		{"Planche à voile", "Planche à voile"},
		{"Équipement de plongée", "Équipement de plongée"},
		{"Skateboard", "Skateboard"},
		{"Rollers", "Rollers"},
		{"Patins à glace", "Patins à glace"},

		// Outils et équipements
		{"Outils", "Outils"},
		{"Boîte à outils", "Boîte à outils"},
		{"Tournevis", "Tournevis"},
		{"Marteau", "Marteau"},
		{"Clé", "Clé"},
		{"Perceuse", "Perceuse"},
		{"Multimètre", "Multimètre"},
		{"Équipement de jardinage", "Équipement de jardinage"},

		// Médicaments et santé
		{"Médicaments", "Médicaments"},
		{"Trousse de secours", "Trousse de secours"},
		{"Lunettes médicales", "Lunettes médicales"},
		{"Appareil auditif", "Appareil auditif"},
		{"Dentier", "Dentier"},
		{"Béquilles", "Béquilles"},
		{"Fauteuil roulant", "Fauteuil roulant"},

		// Jouets et jeux
		{"Jouets", "Jouets"},
		{"Poupée", "Poupée"},
		{"Peluche", "Peluche"},
		{"Jeu de société", "Jeu de société"},
		{"Console de jeu", "Console de jeu"},
		{"Jeu vidéo", "Jeu vidéo"},
		{"Puzzle", "Puzzle"},

		// Instruments de musique
		{"Instrument de musique", "Instrument de musique"},
		{"Guitare", "Guitare"},
		{"Violon", "Violon"},
		{"Piano portable", "Piano portable"},
		{"Flûte", "Flûte"},
		{"Trompette", "Trompette"},
		{"Tambour", "Tambour"},
		{"Microphone", "Microphone"},
		{"Amplificateur", "Amplificateur"},

		// Articles de cuisine
		{"Articles de cuisine", "Articles de cuisine"},
		{"Thermos", "Thermos"},
		{"Gourde", "Gourde"},
		{"Bouteille", "Bouteille"},
		{"Tupperware", "Tupperware"},
		{"Lunch box", "Lunch box"},

		// Autres
		{"Cigarettes", "Cigarettes"},
		{"Briquet", "Briquet"},
		{"Allumettes", "Allumettes"},
		{"Stylo", "Stylo"},
		{"Crayon", "Crayon"},
		{"Trousse de stylos", "Trousse de stylos"},
		{"Règle", "Règle"},
		{"Compas", "Compas"},
		{"Équerre", "Équerre"},
		{"Trousse scolaire", "Trousse scolaire"},
		{"Cartable", "Cartable"},
		{"Serviette", "Serviette"},
		{"Peigne", "Peigne"},
		{"Brosse à cheveux", "Brosse à cheveux"},
		{"Rasoir", "Rasoir"},
		{"Tondeuse", "Tondeuse"},
		{"Sèche-cheveux", "Sèche-cheveux"},
		{"Fer à repasser", "Fer à repasser"},
		{"Lampe de poche", "Lampe de poche"},
		{"Boussole", "Boussole"},
		{"Jumelles", "Jumelles"},
		{"Téléscope", "Téléscope"},
		{"Lunettes d'observation", "Lunettes d'observation"},
		{"Coffret à bijoux", "Coffret à bijoux"},
		{"Valise", "Valise"},
		{"Bagage", "Bagage"},
		{"Autre", "Autre"},
	}

	// Récupérer la liste des objets retrouves pour calculer les statistiques par type
	filters := &repository.ObjetRetrouveFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	objetsRetrouves, err := s.objetRetrouveRepo.List(ctx, filters)

	// Calculer les statistiques pour chaque type dans l'ordre défini
	for _, typeInfo := range typesOrdered {
		nombre := 0

		// Compter les objets de ce type
		if err == nil && objetsRetrouves != nil {
			for _, objetRetrouve := range objetsRetrouves {
				if objetRetrouve.TypeObjet == typeInfo.key {
					nombre++
				}
			}
		}

		// N'ajouter au tableau que si ce type a au moins un objet
		if nombre > 0 {
			statsTable = append(statsTable, TopTypes{
				Type:  typeInfo.label,
				Count: nombre,
			})
		}
	}

	// Données d'activité par période
	activityData := []DashboardActivityData{}

	// Selon la période, on génère les données d'activité
	if periode != nil && *periode != "" {
		activityData = s.generateActivityData(ctx, commissariatID, debut, fin, *periode)
	}

	return &DashboardResponse{
		Stats:        dashboardStats,
		TopTypes:     statsTable,
		ActivityData: activityData,
	}, nil
}

// generateActivityData génère les données d'activité selon la période
func (s *service) generateActivityData(ctx context.Context, commissariatID *string, debut, fin *time.Time, typePeriode string) []DashboardActivityData {
	activityData := []DashboardActivityData{}

	// Récupérer tous les objets de la période
	filters := &repository.ObjetRetrouveFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	objetsRetrouves, err := s.objetRetrouveRepo.List(ctx, filters)

	if err != nil || objetsRetrouves == nil {
		// Retourner des données vides en cas d'erreur
		return s.generateEmptyActivityData(typePeriode)
	}

	switch typePeriode {
	case "jour":
		// Données par tranches de 4 heures
		tranches := []struct {
			label      string
			heureDebut int
			heureFin   int
		}{
			{"00h-04h", 0, 4},
			{"04h-08h", 4, 8},
			{"08h-12h", 8, 12},
			{"12h-16h", 12, 16},
			{"16h-20h", 16, 20},
			{"20h-24h", 20, 24},
		}

		location, _ := time.LoadLocation("Africa/Abidjan")
		for _, tranche := range tranches {
			totalObjetsRetrouves := 0
			disponibles := 0
			restitues := 0
			nonReclames := 0

			for _, objetRetrouve := range objetsRetrouves {

				dateLocale := objetRetrouve.DateDepot.In(location)
				heure := dateLocale.Hour() // Maintenant c'est la bonne heure !
				if heure >= tranche.heureDebut && heure < tranche.heureFin {
					totalObjetsRetrouves++
					if objetRetrouve.Statut == "DISPONIBLE" {
						disponibles++
					} else if objetRetrouve.Statut == "RESTITUÉ" {
						restitues++
					} else if objetRetrouve.Statut == "NON_RÉCLAMÉ" {
						nonReclames++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:          tranche.label,
				ObjetsRetrouves: totalObjetsRetrouves,
				Disponibles:     disponibles,
				Restitues:       restitues,
				NonReclames:     nonReclames,
			})
		}

	case "semaine":
		// Données par jour de la semaine
		jours := []struct {
			label string
			jour  time.Weekday
		}{
			{"Lun", time.Monday},
			{"Mar", time.Tuesday},
			{"Mer", time.Wednesday},
			{"Jeu", time.Thursday},
			{"Ven", time.Friday},
			{"Sam", time.Saturday},
			{"Dim", time.Sunday},
		}

		for _, j := range jours {
			totalObjetsRetrouves := 0
			disponibles := 0
			restitues := 0
			nonReclames := 0

			for _, objetRetrouve := range objetsRetrouves {
				if objetRetrouve.DateDepot.Weekday() == j.jour {
					totalObjetsRetrouves++
					if objetRetrouve.Statut == "DISPONIBLE" {
						disponibles++
					} else if objetRetrouve.Statut == "RESTITUÉ" {
						restitues++
					} else if objetRetrouve.Statut == "NON_RÉCLAMÉ" {
						nonReclames++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:          j.label,
				ObjetsRetrouves: totalObjetsRetrouves,
				Disponibles:     disponibles,
				Restitues:       restitues,
				NonReclames:     nonReclames,
			})
		}

	case "mois":
		// Données par semaine du mois
		if debut == nil || fin == nil {
			return s.generateEmptyActivityData(typePeriode)
		}

		// Calculer le nombre de semaines dans le mois
		firstDayOfMonth := time.Date(debut.Year(), debut.Month(), 1, 0, 0, 0, 0, debut.Location())
		lastDayOfMonth := time.Date(debut.Year(), debut.Month()+1, 0, 23, 59, 59, 0, debut.Location())

		weekNum := 1
		currentWeekStart := firstDayOfMonth

		for currentWeekStart.Before(lastDayOfMonth) || currentWeekStart.Equal(lastDayOfMonth) {
			currentWeekEnd := currentWeekStart.AddDate(0, 0, 7).Add(-time.Second)
			if currentWeekEnd.After(lastDayOfMonth) {
				currentWeekEnd = lastDayOfMonth
			}

			totalObjetsRetrouves := 0
			disponibles := 0
			restitues := 0
			nonReclames := 0

			for _, objetRetrouve := range objetsRetrouves {
				if (objetRetrouve.DateDepot.After(currentWeekStart) || objetRetrouve.DateDepot.Equal(currentWeekStart)) &&
					(objetRetrouve.DateDepot.Before(currentWeekEnd) || objetRetrouve.DateDepot.Equal(currentWeekEnd)) {
					totalObjetsRetrouves++
					if objetRetrouve.Statut == "DISPONIBLE" {
						disponibles++
					} else if objetRetrouve.Statut == "RESTITUÉ" {
						restitues++
					} else if objetRetrouve.Statut == "NON_RÉCLAMÉ" {
						nonReclames++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:          fmt.Sprintf("Sem %d", weekNum),
				ObjetsRetrouves: totalObjetsRetrouves,
				Disponibles:     disponibles,
				Restitues:       restitues,
				NonReclames:     nonReclames,
			})

			currentWeekStart = currentWeekEnd.Add(time.Second)
			weekNum++

			if weekNum > 5 {
				break // Maximum 5 semaines par mois
			}
		}

	case "annee":
		// Données par mois de l'année
		moisLabels := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}

		for mois := 1; mois <= 12; mois++ {
			totalObjetsRetrouves := 0
			disponibles := 0
			restitues := 0
			nonReclames := 0

			for _, objetRetrouve := range objetsRetrouves {
				if int(objetRetrouve.DateDepot.Month()) == mois {
					totalObjetsRetrouves++
					if objetRetrouve.Statut == "DISPONIBLE" {
						disponibles++
					} else if objetRetrouve.Statut == "RESTITUÉ" {
						restitues++
					} else if objetRetrouve.Statut == "NON_RÉCLAMÉ" {
						nonReclames++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:          moisLabels[mois-1],
				ObjetsRetrouves: totalObjetsRetrouves,
				Disponibles:     disponibles,
				Restitues:       restitues,
				NonReclames:     nonReclames,
			})
		}

	default:
		// Données par année pour "tout"
		// Grouper les objets par année
		anneesMap := make(map[int]struct {
			total       int
			disponibles int
			restitues   int
			nonReclames int
		})

		for _, objetRetrouve := range objetsRetrouves {
			annee := objetRetrouve.DateDepot.Year()
			stats := anneesMap[annee]
			stats.total++
			if objetRetrouve.Statut == "DISPONIBLE" {
				stats.disponibles++
			} else if objetRetrouve.Statut == "RESTITUÉ" {
				stats.restitues++
			} else if objetRetrouve.Statut == "NON_RÉCLAMÉ" {
				stats.nonReclames++
			}
			anneesMap[annee] = stats
		}

		// Trier et afficher les 5 dernières années
		currentYear := time.Now().Year()
		for i := currentYear - 4; i <= currentYear; i++ {
			stats := anneesMap[i]
			activityData = append(activityData, DashboardActivityData{
				Period:          fmt.Sprintf("%d", i),
				ObjetsRetrouves: stats.total,
				Disponibles:     stats.disponibles,
				Restitues:       stats.restitues,
				NonReclames:     stats.nonReclames,
			})
		}
	}

	return activityData
}

// generateEmptyActivityData génère des données vides selon la période
func (s *service) generateEmptyActivityData(typePeriode string) []DashboardActivityData {
	activityData := []DashboardActivityData{}

	switch typePeriode {
	case "jour":
		tranches := []string{"00h-04h", "04h-08h", "08h-12h", "12h-16h", "16h-20h", "20h-24h"}
		for _, tranche := range tranches {
			activityData = append(activityData, DashboardActivityData{
				Period: tranche, ObjetsRetrouves: 0, Disponibles: 0, Restitues: 0, NonReclames: 0,
			})
		}
	case "semaine":
		jours := []string{"Lun", "Mar", "Mer", "Jeu", "Ven", "Sam", "Dim"}
		for _, jour := range jours {
			activityData = append(activityData, DashboardActivityData{
				Period: jour, ObjetsRetrouves: 0, Disponibles: 0, Restitues: 0, NonReclames: 0,
			})
		}
	case "mois":
		for i := 1; i <= 4; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period: fmt.Sprintf("Sem %d", i), ObjetsRetrouves: 0, Disponibles: 0, Restitues: 0, NonReclames: 0,
			})
		}
	case "annee":
		mois := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}
		for _, m := range mois {
			activityData = append(activityData, DashboardActivityData{
				Period: m, ObjetsRetrouves: 0, Disponibles: 0, Restitues: 0, NonReclames: 0,
			})
		}
	default:
		currentYear := time.Now().Year()
		for i := currentYear - 4; i <= currentYear; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period: fmt.Sprintf("%d", i), ObjetsRetrouves: 0, Disponibles: 0, Restitues: 0, NonReclames: 0,
			})
		}
	}

	return activityData
}

// parseDateTime est une fonction helper pour parser les dates
func parseDateTime(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}
