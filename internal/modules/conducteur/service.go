package conducteur

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines conducteur service interface
type Service interface {
	Create(ctx context.Context, input *CreateConducteurRequest) (*ConducteurResponse, error)
	GetByID(ctx context.Context, id string) (*ConducteurResponse, error)
	GetByNumeroPermis(ctx context.Context, numeroPermis string) (*ConducteurResponse, error)
	GetByEmail(ctx context.Context, email string) (*ConducteurResponse, error)
	List(ctx context.Context, filters *ListConducteursRequest) (*ListConducteursResponse, error)
	Update(ctx context.Context, id string, input *UpdateConducteurRequest) (*ConducteurResponse, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) (*SearchConducteursResponse, error)
	GetByNomPrenom(ctx context.Context, nom, prenom string) (*ListConducteursResponse, error)
	GetStatistics(ctx context.Context, conducteurID string) (*ConducteurStatisticsResponse, error)
}

// service implements Service interface
type service struct {
	repo   repository.ConducteurRepository
	logger *zap.Logger
}

// NewService creates a new conducteur service
func NewService(repo repository.ConducteurRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new conducteur
func (s *service) Create(ctx context.Context, input *CreateConducteurRequest) (*ConducteurResponse, error) {
	// Validation métier
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Normaliser les données
	s.normalizeInput(input)

	// Vérifier unicité email et permis
	if err := s.checkUniqueness(ctx, input); err != nil {
		return nil, err
	}

	// Points par défaut si non spécifiés
	if input.PointsPermis == 0 {
		input.PointsPermis = 12
	}

	// Nationalité par défaut
	if input.Nationalite == "" {
		input.Nationalite = "FR"
	}

	repoInput := &repository.CreateConducteurInput{
		ID:                  uuid.New().String(),
		Nom:                 input.Nom,
		Prenom:              input.Prenom,
		DateNaissance:       input.DateNaissance,
		LieuNaissance:       input.LieuNaissance,
		Adresse:             input.Adresse,
		CodePostal:          input.CodePostal,
		Ville:               input.Ville,
		Telephone:           input.Telephone,
		Email:               input.Email,
		NumeroPermis:        input.NumeroPermis,
		PermisDelivreLe:     input.PermisDelivreLe,
		PermisValideJusqu:   input.PermisValideJusqu,
		CategoriesPermis:    input.CategoriesPermis,
		PointsPermis:        input.PointsPermis,
		Nationalite:         input.Nationalite,
	}

	conducteurEnt, err := s.repo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create conducteur", zap.Error(err))
		return nil, fmt.Errorf("failed to create conducteur: %w", err)
	}

	return s.entityToResponse(conducteurEnt), nil
}

// GetByID gets conducteur by ID
func (s *service) GetByID(ctx context.Context, id string) (*ConducteurResponse, error) {
	conducteurEnt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(conducteurEnt), nil
}

// GetByNumeroPermis gets conducteur by numero permis
func (s *service) GetByNumeroPermis(ctx context.Context, numeroPermis string) (*ConducteurResponse, error) {
	conducteurEnt, err := s.repo.GetByNumeroPermis(ctx, numeroPermis)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(conducteurEnt), nil
}

// GetByEmail gets conducteur by email
func (s *service) GetByEmail(ctx context.Context, email string) (*ConducteurResponse, error) {
	normalizedEmail := s.normalizeEmail(email)
	conducteurEnt, err := s.repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(conducteurEnt), nil
}

// List gets conducteurs with filters
func (s *service) List(ctx context.Context, input *ListConducteursRequest) (*ListConducteursResponse, error) {
	filters := &repository.ConducteurFilters{
		Nom:         input.Nom,
		Prenom:      input.Prenom,
		Ville:       input.Ville,
		Nationalite: input.Nationalite,
		Active:      input.Active,
		Limit:       input.Limit,
		Offset:      input.Offset,
	}

	conducteursEnt, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	conducteurs := make([]*ConducteurResponse, len(conducteursEnt))
	for i, c := range conducteursEnt {
		conducteurs[i] = s.entityToResponse(c)
	}

	return &ListConducteursResponse{
		Conducteurs: conducteurs,
		Total:       len(conducteurs),
	}, nil
}

// Update updates conducteur
func (s *service) Update(ctx context.Context, id string, input *UpdateConducteurRequest) (*ConducteurResponse, error) {
	// Validation métier
	if err := s.validateUpdateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Normaliser email si présent
	if input.Email != nil {
		normalizedEmail := s.normalizeEmail(*input.Email)
		input.Email = &normalizedEmail
	}

	repoInput := &repository.UpdateConducteurInput{
		Nom:                 input.Nom,
		Prenom:              input.Prenom,
		DateNaissance:       input.DateNaissance,
		LieuNaissance:       input.LieuNaissance,
		Adresse:             input.Adresse,
		CodePostal:          input.CodePostal,
		Ville:               input.Ville,
		Telephone:           input.Telephone,
		Email:               input.Email,
		NumeroPermis:        input.NumeroPermis,
		PermisDelivreLe:     input.PermisDelivreLe,
		PermisValideJusqu:   input.PermisValideJusqu,
		CategoriesPermis:    input.CategoriesPermis,
		PointsPermis:        input.PointsPermis,
		Nationalite:         input.Nationalite,
		Active:              input.Active,
	}

	conducteurEnt, err := s.repo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(conducteurEnt), nil
}

// Delete deletes conducteur
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Search searches conducteurs
func (s *service) Search(ctx context.Context, query string) (*SearchConducteursResponse, error) {
	conducteursEnt, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	conducteurs := make([]*ConducteurResponse, len(conducteursEnt))
	for i, c := range conducteursEnt {
		conducteurs[i] = s.entityToResponse(c)
	}

	return &SearchConducteursResponse{
		Query:   query,
		Results: conducteurs,
		Total:   len(conducteurs),
	}, nil
}

// GetByNomPrenom gets conducteurs by nom and prenom
func (s *service) GetByNomPrenom(ctx context.Context, nom, prenom string) (*ListConducteursResponse, error) {
	conducteursEnt, err := s.repo.GetByNomPrenom(ctx, nom, prenom)
	if err != nil {
		return nil, err
	}

	conducteurs := make([]*ConducteurResponse, len(conducteursEnt))
	for i, c := range conducteursEnt {
		conducteurs[i] = s.entityToResponse(c)
	}

	return &ListConducteursResponse{
		Conducteurs: conducteurs,
		Total:       len(conducteurs),
	}, nil
}

// GetStatistics gets statistics for a conducteur
func (s *service) GetStatistics(ctx context.Context, conducteurID string) (*ConducteurStatisticsResponse, error) {
	// Récupérer le conducteur avec ses relations
	conducteurEnt, err := s.repo.GetByID(ctx, conducteurID)
	if err != nil {
		return nil, err
	}

	stats := &ConducteurStatisticsResponse{
		ConducteurID:       conducteurID,
		PointsRestants:     conducteurEnt.PointsPermis,
		InfractionsParType: make(map[string]int),
	}

	// Calculer les statistiques
	if conducteurEnt.Edges.Controles != nil {
		stats.NombreControles = len(conducteurEnt.Edges.Controles)
		
		// Trouver premier et dernier contrôle
		for _, controle := range conducteurEnt.Edges.Controles {
			if stats.PremierControle == nil || controle.DateControle.Before(*stats.PremierControle) {
				stats.PremierControle = &controle.DateControle
			}
			if stats.DernierControle == nil || controle.DateControle.After(*stats.DernierControle) {
				stats.DernierControle = &controle.DateControle
			}
		}
	}

	if conducteurEnt.Edges.Infractions != nil {
		stats.NombreInfractions = len(conducteurEnt.Edges.Infractions)

		for _, infraction := range conducteurEnt.Edges.Infractions {
			// Calculer points et montants
			stats.PointsRetires += infraction.PointsRetires
			stats.MontantAmendes += infraction.MontantAmende

			// Trouver dernière infraction
			if stats.DerniereInfraction == nil || infraction.DateInfraction.After(*stats.DerniereInfraction) {
				stats.DerniereInfraction = &infraction.DateInfraction
			}

			// Compter par type si chargé
			if infraction.Edges.TypeInfraction != nil {
				typeCode := infraction.Edges.TypeInfraction.Code
				stats.InfractionsParType[typeCode]++
			}
		}

		// Calculer points restants
		stats.PointsRestants = conducteurEnt.PointsPermis - stats.PointsRetires
		if stats.PointsRestants < 0 {
			stats.PointsRestants = 0
		}
	}

	return stats, nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreateConducteurRequest) error {
	if input.Nom == "" {
		return fmt.Errorf("nom is required")
	}
	if input.Prenom == "" {
		return fmt.Errorf("prenom is required")
	}
	if input.DateNaissance.IsZero() {
		return fmt.Errorf("date_naissance is required")
	}

	// Vérifier que la personne a au moins 14 ans (âge minimum pour le permis en France)
	if time.Since(input.DateNaissance).Hours() < 24*365*14 {
		return fmt.Errorf("conducteur must be at least 14 years old")
	}

	// Validation email si fourni
	if input.Email != nil && !s.isValidEmail(*input.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validation numéro de permis si fourni
	if input.NumeroPermis != nil && !s.isValidNumeroPermis(*input.NumeroPermis) {
		return fmt.Errorf("invalid numero permis format")
	}

	// Validation points permis
	if input.PointsPermis < 0 || input.PointsPermis > 12 {
		return fmt.Errorf("points permis must be between 0 and 12")
	}

	return nil
}

func (s *service) validateUpdateInput(input *UpdateConducteurRequest) error {
	if input.Nom != nil && *input.Nom == "" {
		return fmt.Errorf("nom cannot be empty")
	}
	if input.Prenom != nil && *input.Prenom == "" {
		return fmt.Errorf("prenom cannot be empty")
	}
	if input.Email != nil && !s.isValidEmail(*input.Email) {
		return fmt.Errorf("invalid email format")
	}
	if input.NumeroPermis != nil && !s.isValidNumeroPermis(*input.NumeroPermis) {
		return fmt.Errorf("invalid numero permis format")
	}
	if input.PointsPermis != nil && (*input.PointsPermis < 0 || *input.PointsPermis > 12) {
		return fmt.Errorf("points permis must be between 0 and 12")
	}

	return nil
}

func (s *service) normalizeInput(input *CreateConducteurRequest) {
	// Normaliser nom/prénom (première lettre en majuscule)
	input.Nom = s.capitalizeFirst(input.Nom)
	input.Prenom = s.capitalizeFirst(input.Prenom)

	// Normaliser email
	if input.Email != nil {
		normalizedEmail := s.normalizeEmail(*input.Email)
		input.Email = &normalizedEmail
	}

	// Normaliser code postal
	if input.CodePostal != nil {
		*input.CodePostal = strings.TrimSpace(*input.CodePostal)
	}
}

func (s *service) capitalizeFirst(str string) string {
	if str == "" {
		return str
	}
	return strings.ToUpper(string(str[0])) + strings.ToLower(str[1:])
}

func (s *service) normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *service) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (s *service) isValidNumeroPermis(numeroPermis string) bool {
	// Format français: 12 chiffres (simple validation)
	if len(numeroPermis) != 12 {
		return false
	}
	for _, char := range numeroPermis {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func (s *service) checkUniqueness(ctx context.Context, input *CreateConducteurRequest) error {
	// Vérifier email unique si fourni
	if input.Email != nil {
		_, err := s.repo.GetByEmail(ctx, *input.Email)
		if err == nil {
			return fmt.Errorf("email %s already exists", *input.Email)
		}
		if !strings.Contains(err.Error(), "not found") {
			return err // Erreur technique
		}
	}

	// Vérifier numéro de permis unique si fourni
	if input.NumeroPermis != nil {
		_, err := s.repo.GetByNumeroPermis(ctx, *input.NumeroPermis)
		if err == nil {
			return fmt.Errorf("numero permis %s already exists", *input.NumeroPermis)
		}
		if !strings.Contains(err.Error(), "not found") {
			return err // Erreur technique
		}
	}

	return nil
}

func (s *service) entityToResponse(conducteurEnt *ent.Conducteur) *ConducteurResponse {
	response := &ConducteurResponse{
		ID:                  conducteurEnt.ID.String(),
		Nom:                 conducteurEnt.Nom,
		Prenom:              conducteurEnt.Prenom,
		DateNaissance:       conducteurEnt.DateNaissance,
		LieuNaissance:       conducteurEnt.LieuNaissance,
		Adresse:             conducteurEnt.Adresse,
		CodePostal:          conducteurEnt.CodePostal,
		Ville:               conducteurEnt.Ville,
		Telephone:           conducteurEnt.Telephone,
		Email:               conducteurEnt.Email,
		NumeroPermis:        conducteurEnt.NumeroPermis,
		CategoriesPermis:    conducteurEnt.CategoriesPermis,
		PointsPermis:        conducteurEnt.PointsPermis,
		Nationalite:         conducteurEnt.Nationalite,
		Active:              conducteurEnt.Active,
		CreatedAt:           conducteurEnt.CreatedAt,
		UpdatedAt:           conducteurEnt.UpdatedAt,
	}

	// Gérer les champs de date optionnels
	if !conducteurEnt.PermisDelivreLe.IsZero() {
		response.PermisDelivreLe = &conducteurEnt.PermisDelivreLe
	}
	if !conducteurEnt.PermisValideJusqu.IsZero() {
		response.PermisValideJusqu = &conducteurEnt.PermisValideJusqu
		response.PermisValide = conducteurEnt.PermisValideJusqu.After(time.Now())
	} else {
		response.PermisValide = true // Pas d'échéance = valide
	}

	// Ajouter les relations si chargées
	if conducteurEnt.Edges.Controles != nil {
		response.NombreControles = len(conducteurEnt.Edges.Controles)
	}

	if conducteurEnt.Edges.Infractions != nil {
		response.NombreInfractions = len(conducteurEnt.Edges.Infractions)
	}

	return response
}