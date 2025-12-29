package vehicule

import (
	"context"
	"fmt"
	"strings"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines vehicule service interface
type Service interface {
	Create(ctx context.Context, input *CreateVehiculeRequest) (*VehiculeResponse, error)
	GetByID(ctx context.Context, id string) (*VehiculeResponse, error)
	GetByImmatriculation(ctx context.Context, immatriculation string) (*VehiculeResponse, error)
	List(ctx context.Context, filters *ListVehiculesRequest) (*ListVehiculesResponse, error)
	Update(ctx context.Context, id string, input *UpdateVehiculeRequest) (*VehiculeResponse, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) (*SearchVehiculesResponse, error)
	GetByProprietaire(ctx context.Context, nom, prenom string) (*ListVehiculesResponse, error)
	GetByMarque(ctx context.Context, marque string) (*ListVehiculesResponse, error)
	GetByType(ctx context.Context, typeVehicule string) (*ListVehiculesResponse, error)
}

// service implements Service interface
type service struct {
	repo   repository.VehiculeRepository
	logger *zap.Logger
}

// NewService creates a new vehicule service
func NewService(repo repository.VehiculeRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new vehicule
func (s *service) Create(ctx context.Context, input *CreateVehiculeRequest) (*VehiculeResponse, error) {
	// Validation métier
	if err := s.validateCreateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Normaliser immatriculation
	input.Immatriculation = s.normalizeImmatriculation(input.Immatriculation)

	// Vérifier unicité
	if err := s.checkImmatriculationUnique(ctx, input.Immatriculation); err != nil {
		return nil, err
	}

	repoInput := &repository.CreateVehiculeInput{
		ID:                             uuid.New().String(),
		Immatriculation:                input.Immatriculation,
		Marque:                         input.Marque,
		Modele:                         input.Modele,
		Couleur:                        input.Couleur,
		TypeVehicule:                   input.TypeVehicule,
		Energie:                        input.Energie,
		NumeroChassis:                  input.NumeroChassis,
		ProprietaireNom:                input.ProprietaireNom,
		ProprietairePrenom:             input.ProprietairePrenom,
		ProprietaireAdresse:            input.ProprietaireAdresse,
		AssuranceCompagnie:             input.AssuranceCompagnie,
		AssuranceNumero:                input.AssuranceNumero,
	}

	vehiculeEnt, err := s.repo.Create(ctx, repoInput)
	if err != nil {
		s.logger.Error("Failed to create vehicule", zap.Error(err))
		return nil, fmt.Errorf("failed to create vehicule: %w", err)
	}

	return s.entityToResponse(vehiculeEnt), nil
}

// GetByID gets vehicule by ID
func (s *service) GetByID(ctx context.Context, id string) (*VehiculeResponse, error) {
	vehiculeEnt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(vehiculeEnt), nil
}

// GetByImmatriculation gets vehicule by immatriculation
func (s *service) GetByImmatriculation(ctx context.Context, immatriculation string) (*VehiculeResponse, error) {
	normalizedImmat := s.normalizeImmatriculation(immatriculation)
	vehiculeEnt, err := s.repo.GetByImmatriculation(ctx, normalizedImmat)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(vehiculeEnt), nil
}

// List gets vehicules with filters
func (s *service) List(ctx context.Context, input *ListVehiculesRequest) (*ListVehiculesResponse, error) {
	filters := &repository.VehiculeFilters{
		Marque:          input.Marque,
		Modele:          input.Modele,
		TypeVehicule:    input.TypeVehicule,
		Active:          input.Active,
		ProprietaireNom: input.ProprietaireNom,
		Limit:           input.Limit,
		Offset:          input.Offset,
	}

	vehiculesEnt, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	vehicules := make([]*VehiculeResponse, len(vehiculesEnt))
	for i, v := range vehiculesEnt {
		vehicules[i] = s.entityToResponse(v)
	}

	return &ListVehiculesResponse{
		Vehicules: vehicules,
		Total:     len(vehicules),
	}, nil
}

// Update updates vehicule
func (s *service) Update(ctx context.Context, id string, input *UpdateVehiculeRequest) (*VehiculeResponse, error) {
	// Validation métier
	if err := s.validateUpdateInput(input); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	repoInput := &repository.UpdateVehiculeInput{
		Marque:                         input.Marque,
		Modele:                         input.Modele,
		Couleur:                        input.Couleur,
		TypeVehicule:                   input.TypeVehicule,
		Energie:                        input.Energie,
		NumeroChassis:                  input.NumeroChassis,
		ProprietaireNom:                input.ProprietaireNom,
		ProprietairePrenom:             input.ProprietairePrenom,
		ProprietaireAdresse:            input.ProprietaireAdresse,
		AssuranceCompagnie:             input.AssuranceCompagnie,
		AssuranceNumero:                input.AssuranceNumero,
		Active:                         input.Active,
	}

	vehiculeEnt, err := s.repo.Update(ctx, id, repoInput)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(vehiculeEnt), nil
}

// Delete deletes vehicule
func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Search searches vehicules
func (s *service) Search(ctx context.Context, query string) (*SearchVehiculesResponse, error) {
	vehiculesEnt, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	vehicules := make([]*VehiculeResponse, len(vehiculesEnt))
	for i, v := range vehiculesEnt {
		vehicules[i] = s.entityToResponse(v)
	}

	return &SearchVehiculesResponse{
		Query:     query,
		Results:   vehicules,
		Total:     len(vehicules),
	}, nil
}

// GetByProprietaire gets vehicules by proprietaire
func (s *service) GetByProprietaire(ctx context.Context, nom, prenom string) (*ListVehiculesResponse, error) {
	vehiculesEnt, err := s.repo.GetByProprietaire(ctx, nom, prenom)
	if err != nil {
		return nil, err
	}

	vehicules := make([]*VehiculeResponse, len(vehiculesEnt))
	for i, v := range vehiculesEnt {
		vehicules[i] = s.entityToResponse(v)
	}

	return &ListVehiculesResponse{
		Vehicules: vehicules,
		Total:     len(vehicules),
	}, nil
}

// GetByMarque gets vehicules by marque
func (s *service) GetByMarque(ctx context.Context, marque string) (*ListVehiculesResponse, error) {
	filters := &repository.VehiculeFilters{
		Marque: &marque,
	}

	vehiculesEnt, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	vehicules := make([]*VehiculeResponse, len(vehiculesEnt))
	for i, v := range vehiculesEnt {
		vehicules[i] = s.entityToResponse(v)
	}

	return &ListVehiculesResponse{
		Vehicules: vehicules,
		Total:     len(vehicules),
	}, nil
}

// GetByType gets vehicules by type
func (s *service) GetByType(ctx context.Context, typeVehicule string) (*ListVehiculesResponse, error) {
	filters := &repository.VehiculeFilters{
		TypeVehicule: &typeVehicule,
	}

	vehiculesEnt, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	vehicules := make([]*VehiculeResponse, len(vehiculesEnt))
	for i, v := range vehiculesEnt {
		vehicules[i] = s.entityToResponse(v)
	}

	return &ListVehiculesResponse{
		Vehicules: vehicules,
		Total:     len(vehicules),
	}, nil
}

// Private helper methods

func (s *service) validateCreateInput(input *CreateVehiculeRequest) error {
	if input.Immatriculation == "" {
		return fmt.Errorf("immatriculation is required")
	}
	if input.Marque == "" {
		return fmt.Errorf("marque is required")
	}
	if input.Modele == "" {
		return fmt.Errorf("modele is required")
	}
	if input.TypeVehicule == "" {
		input.TypeVehicule = "VP" // Default
	}

	// Validation format immatriculation
	if !s.isValidImmatriculation(input.Immatriculation) {
		return fmt.Errorf("invalid immatriculation format")
	}

	return nil
}

func (s *service) validateUpdateInput(input *UpdateVehiculeRequest) error {
	if input.Marque != nil && *input.Marque == "" {
		return fmt.Errorf("marque cannot be empty")
	}
	if input.Modele != nil && *input.Modele == "" {
		return fmt.Errorf("modele cannot be empty")
	}
	return nil
}

func (s *service) normalizeImmatriculation(immat string) string {
	// Normaliser: supprimer espaces, tirets, mettre en majuscules
	normalized := strings.ReplaceAll(immat, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return strings.ToUpper(normalized)
}

func (s *service) isValidImmatriculation(immat string) bool {
	normalized := s.normalizeImmatriculation(immat)
	
	// Format français standard: AA123BB ou 1234AB34
	if len(normalized) < 6 || len(normalized) > 9 {
		return false
	}
	
	// Validation basique (à améliorer selon les besoins spécifiques)
	return true
}

func (s *service) checkImmatriculationUnique(ctx context.Context, immatriculation string) error {
	_, err := s.repo.GetByImmatriculation(ctx, immatriculation)
	if err == nil {
		return fmt.Errorf("vehicule with immatriculation %s already exists", immatriculation)
	}
	if strings.Contains(err.Error(), "not found") {
		return nil // OK, n'existe pas
	}
	return err // Erreur technique
}

func (s *service) entityToResponse(vehiculeEnt *ent.Vehicule) *VehiculeResponse {
	response := &VehiculeResponse{
		ID:                             vehiculeEnt.ID.String(),
		Immatriculation:                vehiculeEnt.Immatriculation,
		Marque:                         vehiculeEnt.Marque,
		Modele:                         vehiculeEnt.Modele,
		Couleur:                        vehiculeEnt.Couleur,
		TypeVehicule:                   vehiculeEnt.TypeVehicule,
		Energie:                        vehiculeEnt.Energie,
		NumeroChassis:                  vehiculeEnt.NumeroChassis,
		ProprietaireNom:                vehiculeEnt.ProprietaireNom,
		ProprietairePrenom:             vehiculeEnt.ProprietairePrenom,
		ProprietaireAdresse:            vehiculeEnt.ProprietaireAdresse,
		AssuranceCompagnie:             vehiculeEnt.AssuranceCompagnie,
		AssuranceNumero:                vehiculeEnt.AssuranceNumero,
		Active:                         vehiculeEnt.Active,
		CreatedAt:                      vehiculeEnt.CreatedAt,
		UpdatedAt:                      vehiculeEnt.UpdatedAt,
	}

	// Ajouter les contrôles si chargés
	if vehiculeEnt.Edges.Controles != nil {
		response.NombreControles = len(vehiculeEnt.Edges.Controles)
	}

	// Ajouter les infractions si chargées
	if vehiculeEnt.Edges.Infractions != nil {
		response.NombreInfractions = len(vehiculeEnt.Edges.Infractions)
	}

	return response
}