package objetsperdus

import (
	"context"
	"encoding/json"
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

// Service defines objets perdus service interface
type Service interface {
	Create(ctx context.Context, req *CreateObjetPerduRequest, agentID, commissariatID string) (*ObjetPerduResponse, error)
	GetByID(ctx context.Context, id string) (*ObjetPerduResponse, error)
	List(ctx context.Context, filters *FilterObjetsPerdusRequest, role, userID, commissariatID string) (*ListObjetsPerdusResponse, error)
	Update(ctx context.Context, id string, req *UpdateObjetPerduRequest) (*ObjetPerduResponse, error)
	UpdateStatut(ctx context.Context, id string, req *UpdateStatutRequest, agentID string) (*ObjetPerduResponse, error)
	Delete(ctx context.Context, id string) error
	GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesObjetsPerdusResponse, error)
	GetDashboard(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*DashboardResponse, error)
	CheckMatches(ctx context.Context, req *CheckMatchesRequest) ([]MatchedObjetRetrouve, error)
}

// service implements Service interface
type service struct {
	objetPerduRepo   repository.ObjetPerduRepository
	objetRetrouveRepo repository.ObjetRetrouveRepository
	commissariatRepo repository.CommissariatRepository
	userRepo         repository.UserRepository
	config           *config.Config
	logger           *zap.Logger
}

// NewService creates a new objets perdus service
func NewService(
	objetPerduRepo repository.ObjetPerduRepository,
	objetRetrouveRepo repository.ObjetRetrouveRepository,
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	return &service{
		objetPerduRepo:    objetPerduRepo,
		objetRetrouveRepo: objetRetrouveRepo,
		commissariatRepo:  commissariatRepo,
		userRepo:          userRepo,
		config:            cfg,
		logger:            logger,
	}
}

// generateNumero génère un numéro unique pour l'objet perdu
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

	// Chercher le dernier objet perdu du commissariat pour cette année
	filters := &repository.ObjetPerduFilters{
		CommissariatID: &commissariatID,
		Limit:          1000, // Récupérer beaucoup pour trouver le max
		Offset:         0,
	}

	objets, err := s.objetPerduRepo.List(ctx, filters)

	nextNumber := 1

	// Si des objets existent, extraire le dernier numéro
	if err == nil && len(objets) > 0 {
		// Trouver le numéro max en parcourant tous les objets de l'année
		maxNum := 0
		for _, objet := range objets {
			// Format: OBP-VILLE-COM-YYYY-NNNN
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
		numero := fmt.Sprintf("OBP-%s-COM-%d-%04d", villePrefix, year, nextNumber+retry)

		// Vérifier si le numéro existe déjà
		_, err := s.objetPerduRepo.GetByNumero(ctx, numero)
		if err != nil {
			// Le numéro n'existe pas, on peut l'utiliser
			return numero, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique numero after %d retries", maxRetries)
}

// Create creates a new objet perdu
func (s *service) Create(ctx context.Context, req *CreateObjetPerduRequest, agentID, commissariatID string) (*ObjetPerduResponse, error) {
	// Générer le numéro unique
	numero, err := s.generateNumero(ctx, commissariatID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate numero: %w", err)
	}

	// Parser la date de perte
	datePerte, err := time.Parse("2006-01-02", req.DatePerte)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
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

	// Construire le déclarant
	declarant := map[string]interface{}{
		"nom":       req.Declarant.Nom,
		"prenom":    req.Declarant.Prenom,
		"telephone": req.Declarant.Telephone,
	}
	if req.Declarant.Email != nil {
		declarant["email"] = *req.Declarant.Email
	}
	if req.Declarant.Adresse != nil {
		declarant["adresse"] = *req.Declarant.Adresse
	}
	if req.Declarant.CNI != nil {
		declarant["cni"] = *req.Declarant.CNI
	}

	// Gérer le mode contenant
	isContainer := false
	if req.IsContainer != nil {
		isContainer = *req.IsContainer
	}

	var containerDetails map[string]interface{}
	if isContainer && req.ContainerDetails != nil {
		containerDetails = make(map[string]interface{})
		containerDetails["type"] = req.ContainerDetails.Type

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

		// Sérialiser l'inventaire en JSON
		if req.ContainerDetails.Inventory != nil && len(req.ContainerDetails.Inventory) > 0 {
			containerDetails["inventory"] = req.ContainerDetails.Inventory
		}
	}

	// Créer l'objet perdu
	repoInput := &repository.CreateObjetPerduInput{
		ID:                 uuid.New().String(),
		Numero:             numero,
		TypeObjet:          req.TypeObjet,
		Description:        req.Description,
		ValeurEstimee:      req.ValeurEstimee,
		Couleur:            req.Couleur,
		DetailsSpecifiques: detailsSpecifiques,
		IsContainer:        isContainer,
		ContainerDetails:   containerDetails,
		Declarant:          declarant,
		LieuPerte:          req.LieuPerte,
		AdresseLieu:        req.AdresseLieu,
		DatePerte:          datePerte,
		HeurePerte:         req.HeurePerte,
		Statut:             string(StatutObjetPerduEnRecherche),
		DateDeclaration:    time.Now(),
		Observations:       req.Observations,
		CommissariatID:     commissariatID,
		AgentID:            agentID,
	}

	s.logger.Info("Saving objet perdu to database",
		zap.String("numero", numero),
		zap.String("agent_id", agentID),
		zap.String("commissariat_id", commissariatID),
		zap.Bool("is_container", isContainer),
	)

	objetEnt, err := s.objetPerduRepo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create objet perdu", zap.Error(err))
		return nil, fmt.Errorf("failed to create objet perdu: %w", err)
	}

	s.logger.Info("Objet perdu created successfully",
		zap.String("id", objetEnt.ID.String()),
		zap.String("numero", objetEnt.Numero),
	)

	return s.formatObjetPerdu(objetEnt), nil
}

// GetByID gets an objet perdu by ID
func (s *service) GetByID(ctx context.Context, id string) (*ObjetPerduResponse, error) {
	objet, err := s.objetPerduRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "objet perdu not found" {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to get objet perdu: %w", err)
	}

	return s.formatObjetPerdu(objet), nil
}

// List lists objets perdus with filters
func (s *service) List(ctx context.Context, filters *FilterObjetsPerdusRequest, role, userID, commissariatID string) (*ListObjetsPerdusResponse, error) {
	repoFilters := &repository.ObjetPerduFilters{}

	if filters.Statut != nil {
		repoFilters.Statut = filters.Statut
	}
	if filters.TypeObjet != nil {
		repoFilters.TypeObjet = filters.TypeObjet
	}
	if filters.IsContainer != nil {
		repoFilters.IsContainer = filters.IsContainer
	}
	if filters.CommissariatID != nil {
		repoFilters.CommissariatID = filters.CommissariatID
	} else if role != "ADMIN" && commissariatID != "" {
		// Pour les non-admins, filtrer par commissariat
		repoFilters.CommissariatID = &commissariatID
	}
	if filters.DateDebut != nil {
		repoFilters.DateDebut = filters.DateDebut
	}
	if filters.DateFin != nil {
		repoFilters.DateFin = filters.DateFin
	}
	if filters.Search != nil {
		repoFilters.Search = filters.Search
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

	objets, err := s.objetPerduRepo.List(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list objets perdus: %w", err)
	}

	total, err := s.objetPerduRepo.Count(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to count objets perdus: %w", err)
	}

	responses := make([]ObjetPerduResponse, len(objets))
	for i, objet := range objets {
		responses[i] = *s.formatObjetPerdu(objet)
	}

	return &ListObjetsPerdusResponse{
		Objets: responses,
		Total:  int64(total),
		Page:   page,
		Limit:  limit,
	}, nil
}

// Update updates an objet perdu
func (s *service) Update(ctx context.Context, id string, req *UpdateObjetPerduRequest) (*ObjetPerduResponse, error) {
	repoInput := &repository.UpdateObjetPerduInput{}

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

	// Gérer les nouveaux champs
	if req.IsContainer != nil {
		repoInput.IsContainer = req.IsContainer
	}

	if req.ContainerDetails != nil {
		containerDetails := make(map[string]interface{})
		containerDetails["type"] = req.ContainerDetails.Type

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

		if req.ContainerDetails.Inventory != nil {
			containerDetails["inventory"] = req.ContainerDetails.Inventory
		}

		repoInput.ContainerDetails = containerDetails
	}

	if req.Declarant != nil {
		declarant := map[string]interface{}{
			"nom":       req.Declarant.Nom,
			"prenom":    req.Declarant.Prenom,
			"telephone": req.Declarant.Telephone,
		}
		if req.Declarant.Email != nil {
			declarant["email"] = *req.Declarant.Email
		}
		if req.Declarant.Adresse != nil {
			declarant["adresse"] = *req.Declarant.Adresse
		}
		if req.Declarant.CNI != nil {
			declarant["cni"] = *req.Declarant.CNI
		}
		repoInput.Declarant = declarant
	}
	if req.LieuPerte != nil {
		repoInput.LieuPerte = req.LieuPerte
	}
	if req.AdresseLieu != nil {
		repoInput.AdresseLieu = req.AdresseLieu
	}
	if req.DatePerte != nil {
		datePerte, err := time.Parse("2006-01-02", *req.DatePerte)
		if err == nil {
			repoInput.DatePerte = &datePerte
		}
	}
	if req.HeurePerte != nil {
		repoInput.HeurePerte = req.HeurePerte
	}
	if req.Observations != nil {
		repoInput.Observations = req.Observations
	}

	objet, err := s.objetPerduRepo.Update(ctx, id, repoInput)
	if err != nil {
		if err.Error() == "objet perdu not found" {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to update objet perdu: %w", err)
	}

	return s.formatObjetPerdu(objet), nil
}

// UpdateStatut updates the statut of an objet perdu
func (s *service) UpdateStatut(ctx context.Context, id string, req *UpdateStatutRequest, agentID string) (*ObjetPerduResponse, error) {
	repoInput := &repository.UpdateObjetPerduInput{
		Statut: &req.Statut,
	}

	if req.DateRetrouve != nil {
		repoInput.DateRetrouve = req.DateRetrouve
	}

	objet, err := s.objetPerduRepo.Update(ctx, id, repoInput)
	if err != nil {
		if err.Error() == "objet perdu not found" {
			return nil, fmt.Errorf("objet perdu not found")
		}
		return nil, fmt.Errorf("failed to update statut: %w", err)
	}

	return s.formatObjetPerdu(objet), nil
}

// Delete deletes an objet perdu
func (s *service) Delete(ctx context.Context, id string) error {
	err := s.objetPerduRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "objet perdu not found" {
			return fmt.Errorf("objet perdu not found")
		}
		return fmt.Errorf("failed to delete objet perdu: %w", err)
	}
	return nil
}

// formatObjetPerdu formats an ent.ObjetPerdu to ObjetPerduResponse
func (s *service) formatObjetPerdu(objet *ent.ObjetPerdu) *ObjetPerduResponse {
	response := &ObjetPerduResponse{
		ID:                      objet.ID.String(),
		Numero:                  objet.Numero,
		TypeObjet:               objet.TypeObjet,
		Description:             objet.Description,
		IsContainer:             objet.IsContainer,
		Declarant:               objet.Declarant,
		LieuPerte:               objet.LieuPerte,
		DatePerte:               objet.DatePerte,
		DatePerteFormatee:       objet.DatePerte.Format("02/01/2006"),
		Statut:                  StatutObjetPerdu(objet.Statut),
		DateDeclaration:         objet.DateDeclaration,
		DateDeclarationFormatee: objet.DateDeclaration.Format("02/01/2006 à 15:04"),
		CreatedAt:               objet.CreatedAt,
		UpdatedAt:               objet.UpdatedAt,
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

	// Gérer les détails du contenant
	if objet.ContainerDetails != nil && len(objet.ContainerDetails) > 0 {
		containerDetails := &ContainerDetails{}

		if typeVal, ok := objet.ContainerDetails["type"].(string); ok {
			containerDetails.Type = typeVal
		}
		if couleur, ok := objet.ContainerDetails["couleur"].(string); ok {
			containerDetails.Couleur = &couleur
		}
		if marque, ok := objet.ContainerDetails["marque"].(string); ok {
			containerDetails.Marque = &marque
		}
		if taille, ok := objet.ContainerDetails["taille"].(string); ok {
			containerDetails.Taille = &taille
		}
		if signes, ok := objet.ContainerDetails["signesDistinctifs"].(string); ok {
			containerDetails.SignesDistinctifs = &signes
		}

		// Désérialiser l'inventaire
		if inventoryData, ok := objet.ContainerDetails["inventory"]; ok {
			// L'inventaire peut être soit un tableau d'objets JSON, soit déjà désérialisé
			var inventory []InventoryItem

			// Essayer de le convertir depuis JSON si c'est une string
			if inventoryJSON, ok := inventoryData.(string); ok {
				json.Unmarshal([]byte(inventoryJSON), &inventory)
			} else if inventorySlice, ok := inventoryData.([]interface{}); ok {
				// Si c'est déjà un slice d'interfaces, le convertir
				for _, item := range inventorySlice {
					if itemMap, ok := item.(map[string]interface{}); ok {
						invItem := InventoryItem{}
						if id, ok := itemMap["id"].(float64); ok {
							invItem.ID = int(id)
						}
						if category, ok := itemMap["category"].(string); ok {
							invItem.Category = category
						}
						if icon, ok := itemMap["icon"].(string); ok {
							invItem.Icon = icon
						}
						if name, ok := itemMap["name"].(string); ok {
							invItem.Name = name
						}
						if color, ok := itemMap["color"].(string); ok {
							invItem.Color = color
						}
						if brand, ok := itemMap["brand"].(string); ok {
							invItem.Brand = &brand
						}
						if serial, ok := itemMap["serial"].(string); ok {
							invItem.Serial = &serial
						}
						if description, ok := itemMap["description"].(string); ok {
							invItem.Description = &description
						}
						if identityType, ok := itemMap["identityType"].(string); ok {
							invItem.IdentityType = &identityType
						}
						if identityNumber, ok := itemMap["identityNumber"].(string); ok {
							invItem.IdentityNumber = &identityNumber
						}
						if identityName, ok := itemMap["identityName"].(string); ok {
							invItem.IdentityName = &identityName
						}
						if cardType, ok := itemMap["cardType"].(string); ok {
							invItem.CardType = &cardType
						}
						if cardBank, ok := itemMap["cardBank"].(string); ok {
							invItem.CardBank = &cardBank
						}
						if cardLast4, ok := itemMap["cardLast4"].(string); ok {
							invItem.CardLast4 = &cardLast4
						}
						inventory = append(inventory, invItem)
					}
				}
			}

			if len(inventory) > 0 {
				containerDetails.Inventory = inventory
			}
		}

		response.ContainerDetails = containerDetails
	}

	if objet.AdresseLieu != nil {
		response.AdresseLieu = objet.AdresseLieu
	}
	if objet.HeurePerte != nil {
		response.HeurePerte = objet.HeurePerte
	}
	if objet.DateRetrouve != nil {
		response.DateRetrouve = objet.DateRetrouve
		formatted := objet.DateRetrouve.Format("02/01/2006")
		response.DateRetrouveFormatee = &formatted
	}
	if objet.Observations != nil {
		response.Observations = objet.Observations
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

// GetStatistiques calcule les statistiques des objets perdus
func (s *service) GetStatistiques(ctx context.Context, commissariatID *string, dateDebut, dateFin, periode *string) (*StatistiquesObjetsPerdusResponse, error) {
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
	stats, err := s.objetPerduRepo.GetStatistiques(ctx, commissariatID, debut, fin, periode)
	if err != nil {
		s.logger.Error("Failed to get statistics from repository", zap.Error(err))
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	s.logger.Info("Stats reçues du repository",
		zap.Any("stats", stats),
		zap.Bool("hasEvolutionTotal", stats["evolutionTotal"] != nil),
		zap.Bool("hasEvolutionEnRecherche", stats["evolutionEnRecherche"] != nil),
		zap.Bool("hasEvolutionRetrouves", stats["evolutionRetrouves"] != nil),
		zap.Bool("hasEvolutionTauxRetrouve", stats["evolutionTauxRetrouve"] != nil),
	)

	// Convertir en response avec valeurs par défaut pour les évolutions
	response := &StatistiquesObjetsPerdusResponse{
		Total:                 int64(stats["total"].(int)),
		EnRecherche:           int64(stats["enRecherche"].(int)),
		Retrouves:             int64(stats["retrouves"].(int)),
		Clotures:              int64(stats["clotures"].(int)),
		EvolutionTotal:        "0", // Valeur par défaut
		EvolutionEnRecherche:  "0", // Valeur par défaut
		EvolutionRetrouves:    "0", // Valeur par défaut
		EvolutionClotures:     "0", // Valeur par défaut
		EvolutionTauxRetrouve: "0", // Valeur par défaut
	}

	if tauxRetrouve, ok := stats["tauxRetrouve"].(float64); ok {
		response.TauxRetrouve = tauxRetrouve
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

	if evolutionEnRechercheVal, exists := stats["evolutionEnRecherche"]; exists {
		if evolutionEnRecherche, ok := evolutionEnRechercheVal.(string); ok {
			response.EvolutionEnRecherche = evolutionEnRecherche
		} else {
			s.logger.Warn("evolutionEnRecherche n'est pas une string", zap.Any("value", evolutionEnRechercheVal))
			response.EvolutionEnRecherche = "0"
		}
	} else {
		s.logger.Warn("evolutionEnRecherche n'existe pas dans le map")
		response.EvolutionEnRecherche = "0"
	}

	if evolutionRetrouvesVal, exists := stats["evolutionRetrouves"]; exists {
		if evolutionRetrouves, ok := evolutionRetrouvesVal.(string); ok {
			response.EvolutionRetrouves = evolutionRetrouves
		} else {
			s.logger.Warn("evolutionRetrouves n'est pas une string", zap.Any("value", evolutionRetrouvesVal))
			response.EvolutionRetrouves = "0"
		}
	} else {
		s.logger.Warn("evolutionRetrouves n'existe pas dans le map")
		response.EvolutionRetrouves = "0"
	}

	if evolutionCloturesVal, exists := stats["evolutionClotures"]; exists {
		if evolutionClotures, ok := evolutionCloturesVal.(string); ok {
			response.EvolutionClotures = evolutionClotures
		} else {
			s.logger.Warn("evolutionClotures n'est pas une string", zap.Any("value", evolutionCloturesVal))
			response.EvolutionClotures = "0"
		}
	} else {
		s.logger.Warn("evolutionClotures n'existe pas dans le map")
		response.EvolutionClotures = "0"
	}

	if evolutionTauxRetrouveVal, exists := stats["evolutionTauxRetrouve"]; exists {
		if evolutionTauxRetrouve, ok := evolutionTauxRetrouveVal.(string); ok {
			response.EvolutionTauxRetrouve = evolutionTauxRetrouve
		} else {
			s.logger.Warn("evolutionTauxRetrouve n'est pas une string", zap.Any("value", evolutionTauxRetrouveVal))
			response.EvolutionTauxRetrouve = "0"
		}
	} else {
		s.logger.Warn("evolutionTauxRetrouve n'existe pas dans le map")
		response.EvolutionTauxRetrouve = "0"
	}

	s.logger.Info("Statistiques response finale",
		zap.Int64("total", response.Total),
		zap.String("evolutionTotal", response.EvolutionTotal),
		zap.String("evolutionEnRecherche", response.EvolutionEnRecherche),
		zap.String("evolutionRetrouves", response.EvolutionRetrouves),
		zap.String("evolutionClotures", response.EvolutionClotures),
		zap.String("evolutionTauxRetrouve", response.EvolutionTauxRetrouve),
	)

	return response, nil
}

// GetDashboard gets dashboard data for objets perdus
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
	stats, err := s.objetPerduRepo.GetStatistiques(ctx, commissariatID, debut, fin, periode)
	if err != nil {
		return nil, err
	}

	// Convertir en réponse dashboard
	total := int64(stats["total"].(int))
	enRecherche := int64(stats["enRecherche"].(int))
	retrouves := int64(stats["retrouves"].(int))
	clotures := int64(stats["clotures"].(int))
	tauxRetrouve := float64(stats["tauxRetrouve"].(float64))
	evolutionTotal := stats["evolutionTotal"].(string)
	evolutionEnRecherche := stats["evolutionEnRecherche"].(string)
	evolutionRetrouves := stats["evolutionRetrouves"].(string)
	evolutionClotures := stats["evolutionClotures"].(string)
	evolutionTauxRetrouve := stats["evolutionTauxRetrouve"].(string)

	dashboardStats := DashboardStatsValue{
		Total:                 total,
		EvolutionTotal:        evolutionTotal,
		EnRecherche:           enRecherche,
		Retrouves:             retrouves,
		Clotures:              clotures,
		TauxRetrouve:          tauxRetrouve,
		EvolutionEnRecherche:  evolutionEnRecherche,
		EvolutionRetrouves:    evolutionRetrouves,
		EvolutionClotures:     evolutionClotures,
		EvolutionTauxRetrouve: evolutionTauxRetrouve,
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

	// Récupérer la liste des objets perdus pour calculer les statistiques par type
	filters := &repository.ObjetPerduFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	objetsPerdus, err := s.objetPerduRepo.List(ctx, filters)

	// Calculer les statistiques pour chaque type dans l'ordre défini
	for _, typeInfo := range typesOrdered {
		nombre := 0
		enRechercheType := 0
		retrouvesType := 0
		cloturesType := 0

		// Compter les objets de ce type
		if err == nil && objetsPerdus != nil {
			for _, objetPerdu := range objetsPerdus {
				if objetPerdu.TypeObjet == typeInfo.key {
					nombre++
					if string(objetPerdu.Statut) == string(StatutObjetPerduEnRecherche) {
						enRechercheType++
					} else if string(objetPerdu.Statut) == string(StatutObjetPerduRetrouve) {
						retrouvesType++
					} else if string(objetPerdu.Statut) == string(StatutObjetPerduCloture) {
						cloturesType++
					}
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
	filters := &repository.ObjetPerduFilters{
		CommissariatID: commissariatID,
		DateDebut:      debut,
		DateFin:        fin,
	}
	objetsPerdus, err := s.objetPerduRepo.List(ctx, filters)

	if err != nil || objetsPerdus == nil {
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
			totalObjetsPerdus := 0
			recherche := 0
			retrouves := 0
			clotures := 0

			for _, objetPerdu := range objetsPerdus {
				dateLocale := objetPerdu.DateDeclaration.In(location)
				heure := dateLocale.Hour() // Maintenant c'est la bonne heure !
				if heure >= tranche.heureDebut && heure < tranche.heureFin {
					totalObjetsPerdus++
					if objetPerdu.Statut == "EN_RECHERCHE" {
						recherche++
					} else if objetPerdu.Statut == "RETROUVÉ" {
						retrouves++
					} else if objetPerdu.Statut == "CLÔTURÉ" {
						clotures++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:       tranche.label,
				ObjetsPerdus: totalObjetsPerdus,
				Recherche:    recherche,
				Retrouves:    retrouves,
				Clotures:     clotures,
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
			totalObjetsPerdus := 0
			recherche := 0
			retrouves := 0
			clotures := 0

			for _, objetPerdu := range objetsPerdus {
				if objetPerdu.DateDeclaration.Weekday() == j.jour {
					totalObjetsPerdus++
					if objetPerdu.Statut == "EN_RECHERCHE" {
						recherche++
					} else if objetPerdu.Statut == "RETROUVÉ" {
						retrouves++
					} else if objetPerdu.Statut == "CLÔTURÉ" {
						clotures++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:       j.label,
				ObjetsPerdus: totalObjetsPerdus,
				Recherche:    recherche,
				Retrouves:    retrouves,
				Clotures:     clotures,
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

			totalObjetsPerdus := 0
			recherche := 0
			retrouves := 0
			clotures := 0

			for _, objetPerdu := range objetsPerdus {
				if (objetPerdu.DateDeclaration.After(currentWeekStart) || objetPerdu.DateDeclaration.Equal(currentWeekStart)) &&
					(objetPerdu.DateDeclaration.Before(currentWeekEnd) || objetPerdu.DateDeclaration.Equal(currentWeekEnd)) {
					totalObjetsPerdus++
					if objetPerdu.Statut == "EN_RECHERCHE" {
						recherche++
					} else if objetPerdu.Statut == "RETROUVÉ" {
						retrouves++
					} else if objetPerdu.Statut == "CLÔTURÉ" {
						clotures++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:       fmt.Sprintf("Sem %d", weekNum),
				ObjetsPerdus: totalObjetsPerdus,
				Recherche:    recherche,
				Retrouves:    retrouves,
				Clotures:     clotures,
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
			totalObjetsPerdus := 0
			recherche := 0
			retrouves := 0
			clotures := 0

			for _, objetPerdu := range objetsPerdus {
				if int(objetPerdu.DateDeclaration.Month()) == mois {
					totalObjetsPerdus++
					if objetPerdu.Statut == "EN_RECHERCHE" {
						recherche++
					} else if objetPerdu.Statut == "RETROUVÉ" {
						retrouves++
					} else if objetPerdu.Statut == "CLÔTURÉ" {
						clotures++
					}
				}
			}

			activityData = append(activityData, DashboardActivityData{
				Period:       moisLabels[mois-1],
				ObjetsPerdus: totalObjetsPerdus,
				Recherche:    recherche,
				Retrouves:    retrouves,
				Clotures:     clotures,
			})
		}

	default:
		// Données par année pour "tout"
		// Grouper les objets par année
		anneesMap := make(map[int]struct {
			total     int
			recherche int
			retrouves int
			clotures  int
		})

		for _, objetPerdu := range objetsPerdus {
			annee := objetPerdu.DateDeclaration.Year()
			stats := anneesMap[annee]
			stats.total++
			if objetPerdu.Statut == "EN_RECHERCHE" {
				stats.recherche++
			} else if objetPerdu.Statut == "RETROUVÉ" {
				stats.retrouves++
			} else if objetPerdu.Statut == "CLÔTURÉ" {
				stats.clotures++
			}
			anneesMap[annee] = stats
		}

		// Trier et afficher les 5 dernières années
		currentYear := time.Now().Year()
		for i := currentYear - 4; i <= currentYear; i++ {
			stats := anneesMap[i]
			activityData = append(activityData, DashboardActivityData{
				Period:       fmt.Sprintf("%d", i),
				ObjetsPerdus: stats.total,
				Recherche:    stats.recherche,
				Retrouves:    stats.retrouves,
				Clotures:     stats.clotures,
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
				Period: tranche, ObjetsPerdus: 0, Recherche: 0, Retrouves: 0, Clotures: 0,
			})
		}
	case "semaine":
		jours := []string{"Lun", "Mar", "Mer", "Jeu", "Ven", "Sam", "Dim"}
		for _, jour := range jours {
			activityData = append(activityData, DashboardActivityData{
				Period: jour, ObjetsPerdus: 0, Recherche: 0, Retrouves: 0, Clotures: 0,
			})
		}
	case "mois":
		for i := 1; i <= 4; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period: fmt.Sprintf("Sem %d", i), ObjetsPerdus: 0, Recherche: 0, Retrouves: 0, Clotures: 0,
			})
		}
	case "annee":
		mois := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}
		for _, m := range mois {
			activityData = append(activityData, DashboardActivityData{
				Period: m, ObjetsPerdus: 0, Recherche: 0, Retrouves: 0, Clotures: 0,
			})
		}
	default:
		currentYear := time.Now().Year()
		for i := currentYear - 4; i <= currentYear; i++ {
			activityData = append(activityData, DashboardActivityData{
				Period: fmt.Sprintf("%d", i), ObjetsPerdus: 0, Recherche: 0, Retrouves: 0, Clotures: 0,
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

// CheckMatches vérifie si des objets retrouvés correspondent aux identifiants ultra-uniques fournis
func (s *service) CheckMatches(ctx context.Context, req *CheckMatchesRequest) ([]MatchedObjetRetrouve, error) {
	s.logger.Info("🔍 CheckMatches appelé",
		zap.String("typeObjet", req.TypeObjet),
		zap.Any("identifiers", req.Identifiers),
	)

	var matches []MatchedObjetRetrouve

	// 1. Chercher dans les objets retrouvés directs (non-contenants)
	directMatches, err := s.searchDirectMatches(ctx, req)
	if err != nil {
		s.logger.Error("Erreur lors de la recherche directe", zap.Error(err))
	} else {
		matches = append(matches, directMatches...)
	}

	// 2. Chercher dans les inventaires des contenants
	inventoryMatches, err := s.searchInventoryMatches(ctx, req)
	if err != nil {
		s.logger.Error("Erreur lors de la recherche dans les inventaires", zap.Error(err))
	} else {
		matches = append(matches, inventoryMatches...)
	}

	s.logger.Info("✅ CheckMatches terminé",
		zap.Int("directMatches", len(directMatches)),
		zap.Int("inventoryMatches", len(inventoryMatches)),
		zap.Int("totalMatches", len(matches)),
	)

	return matches, nil
}

// searchDirectMatches cherche dans les objets retrouvés directs
func (s *service) searchDirectMatches(ctx context.Context, req *CheckMatchesRequest) ([]MatchedObjetRetrouve, error) {
	var matches []MatchedObjetRetrouve

	// Récupérer tous les objets retrouvés avec le même type et statut DISPONIBLE
	filters := &repository.ObjetRetrouveFilters{
		TypeObjet: &req.TypeObjet,
	}

	statut := "DISPONIBLE"
	filters.Statut = &statut

	objetsRetrouves, err := s.objetRetrouveRepo.List(ctx, filters)
	if err != nil {
		return matches, err
	}

	s.logger.Info("📋 Objets retrouvés à analyser",
		zap.Int("count", len(objetsRetrouves)),
	)

	// Pour chaque objet retrouvé, vérifier s'il matche avec les identifiants
	for _, objetRetrouve := range objetsRetrouves {
		matchScore := 0
		matchedField := ""

		// Extraire les détails spécifiques
		detailsMap := objetRetrouve.DetailsSpecifiques

		// Vérifier chaque identifiant fourni
		for key, value := range req.Identifiers {
			if value == nil || value == "" {
				continue
			}

			valueStr := fmt.Sprintf("%v", value)

			// Comparer avec les champs de l'objet retrouvé
			switch key {
			case "imei":
				if detailVal, ok := detailsMap["imei"].(string); ok && detailVal == valueStr {
					matchScore = 99
					matchedField = "IMEI"
					break
				}
			case "numeroDocument":
				if detailVal, ok := detailsMap["numeroDocument"].(string); ok && detailVal == valueStr {
					matchScore = 99
					matchedField = "Numéro de document"
					break
				}
			case "numeroSerie":
				if detailVal, ok := detailsMap["numeroSerie"].(string); ok && detailVal == valueStr {
					matchScore = 99
					matchedField = "Numéro de série"
					break
				}
			case "numeroSerieOrdinateur":
				if detailVal, ok := detailsMap["numeroSerieOrdinateur"].(string); ok && detailVal == valueStr {
					matchScore = 99
					matchedField = "Numéro de série ordinateur"
					break
				}
			case "numeroCadre":
				if detailVal, ok := detailsMap["numeroCadre"].(string); ok && detailVal == valueStr {
					matchScore = 99
					matchedField = "Numéro de cadre vélo"
					break
				}
			}

			// Si on a trouvé un match, pas besoin de continuer
			if matchScore > 0 {
				break
			}
		}

		// Si un match a été trouvé, l'ajouter à la liste
		if matchScore > 0 {
			s.logger.Info("✅ Match trouvé (direct)",
				zap.String("objetId", objetRetrouve.ID.String()),
				zap.String("numero", objetRetrouve.Numero),
				zap.Int("score", matchScore),
				zap.String("field", matchedField),
			)

			match := MatchedObjetRetrouve{
				ID:                     objetRetrouve.ID.String(),
				Numero:                 objetRetrouve.Numero,
				TypeObjet:              objetRetrouve.TypeObjet,
				Description:            objetRetrouve.Description,
				ValeurEstimee:          objetRetrouve.ValeurEstimee,
				Couleur:                objetRetrouve.Couleur,
				DetailsSpecifiques:     objetRetrouve.DetailsSpecifiques,
				IsContainer:            objetRetrouve.IsContainer,
				ContainerDetails:       objetRetrouve.ContainerDetails,
				LieuTrouvaille:         objetRetrouve.LieuTrouvaille,
				DateTrouvaille:         objetRetrouve.DateTrouvaille.Format("2006-01-02"),
				DateTrouvailleFormatee: objetRetrouve.DateTrouvaille.Format("02/01/2006"),
				Statut:                 string(objetRetrouve.Statut),
				Deposant:               objetRetrouve.Deposant,
				MatchScore:             matchScore,
				MatchedField:           matchedField,
				MatchedIn:              "direct",
			}

			// Ajouter les infos du commissariat si disponibles
			if objetRetrouve.Edges.Commissariat != nil {
				match.Commissariat = &CommissariatSummary{
					ID:    objetRetrouve.Edges.Commissariat.ID.String(),
					Nom:   objetRetrouve.Edges.Commissariat.Nom,
					Code:  objetRetrouve.Edges.Commissariat.Code,
					Ville: objetRetrouve.Edges.Commissariat.Ville,
				}
			}

			matches = append(matches, match)
		}
	}

	return matches, nil
}

// searchInventoryMatches cherche dans les inventaires des contenants
func (s *service) searchInventoryMatches(ctx context.Context, req *CheckMatchesRequest) ([]MatchedObjetRetrouve, error) {
	var matches []MatchedObjetRetrouve

	// Récupérer tous les contenants (objets avec isContainer = true) en statut DISPONIBLE
	isContainer := true
	statut := "DISPONIBLE"
	filters := &repository.ObjetRetrouveFilters{
		IsContainer: &isContainer,
		Statut:      &statut,
	}

	contenants, err := s.objetRetrouveRepo.List(ctx, filters)
	if err != nil {
		return matches, err
	}

	s.logger.Info("📦 Contenants à analyser",
		zap.Int("count", len(contenants)),
	)

	// Pour chaque contenant, vérifier son inventaire
	for _, contenant := range contenants {
		if contenant.ContainerDetails == nil {
			continue
		}

		// Récupérer l'inventaire
		inventoryData, ok := contenant.ContainerDetails["inventory"]
		if !ok {
			continue
		}

		var inventory []interface{}
		switch v := inventoryData.(type) {
		case []interface{}:
			inventory = v
		case string:
			// Si c'est une string JSON, la parser
			if err := json.Unmarshal([]byte(v), &inventory); err != nil {
				s.logger.Warn("Erreur parsing inventaire JSON", zap.Error(err))
				continue
			}
		default:
			continue
		}

		// Vérifier chaque objet de l'inventaire
		for _, itemData := range inventory {
			itemMap, ok := itemData.(map[string]interface{})
			if !ok {
				continue
			}

			matchScore := 0
			matchedField := ""

			// Vérifier les identifiants ultra-uniques dans l'inventaire
			for key, value := range req.Identifiers {
				if value == nil || value == "" {
					continue
				}

				valueStr := fmt.Sprintf("%v", value)

				switch key {
				case "imei":
					if serialVal, ok := itemMap["serial"].(string); ok && serialVal == valueStr {
						matchScore = 99
						matchedField = "IMEI (dans inventaire)"
						break
					}
				case "numeroDocument", "numeroSerie", "numeroSerieOrdinateur", "numeroCadre":
					if serialVal, ok := itemMap["serial"].(string); ok && serialVal == valueStr {
						matchScore = 99
						matchedField = fmt.Sprintf("%s (dans inventaire)", key)
						break
					}
				}

				// Vérifier identityNumber pour les documents
				if key == "numeroDocument" {
					if identityNum, ok := itemMap["identityNumber"].(string); ok && identityNum == valueStr {
						matchScore = 99
						matchedField = "Numéro de document (dans inventaire)"
						break
					}
				}

				// Vérifier cardLast4 pour les cartes bancaires
				if key == "cardLast4" {
					if cardLast4, ok := itemMap["cardLast4"].(string); ok && cardLast4 == valueStr {
						matchScore = 95
						matchedField = "4 derniers chiffres carte (dans inventaire)"
						break
					}
				}

				if matchScore > 0 {
					break
				}
			}

			// Si un match a été trouvé dans l'inventaire
			if matchScore > 0 {
				s.logger.Info("✅ Match trouvé (inventaire)",
					zap.String("contenantId", contenant.ID.String()),
					zap.String("numero", contenant.Numero),
					zap.Int("score", matchScore),
					zap.String("field", matchedField),
					zap.Any("item", itemMap),
				)

				match := MatchedObjetRetrouve{
					ID:                     contenant.ID.String(),
					Numero:                 contenant.Numero,
					TypeObjet:              contenant.TypeObjet,
					Description:            contenant.Description,
					ValeurEstimee:          contenant.ValeurEstimee,
					Couleur:                contenant.Couleur,
					DetailsSpecifiques:     contenant.DetailsSpecifiques,
					IsContainer:            true,
					ContainerDetails:       contenant.ContainerDetails,
					LieuTrouvaille:         contenant.LieuTrouvaille,
					DateTrouvaille:         contenant.DateTrouvaille.Format("2006-01-02"),
					DateTrouvailleFormatee: contenant.DateTrouvaille.Format("02/01/2006"),
					Statut:                 string(contenant.Statut),
					Deposant:               contenant.Deposant,
					MatchScore:             matchScore,
					MatchedField:           matchedField,
					MatchedIn:              "inventory",
					InventoryItem:          itemMap,
				}

				// Ajouter les infos du commissariat si disponibles
				if contenant.Edges.Commissariat != nil {
					match.Commissariat = &CommissariatSummary{
						ID:    contenant.Edges.Commissariat.ID.String(),
						Nom:   contenant.Edges.Commissariat.Nom,
						Code:  contenant.Edges.Commissariat.Code,
						Ville: contenant.Edges.Commissariat.Ville,
					}
				}

				matches = append(matches, match)
			}
		}
	}

	return matches, nil
}
