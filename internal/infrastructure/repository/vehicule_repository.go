package repository

import (
	"context"
	"fmt"
	"strings"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/vehicule"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VehiculeRepository defines vehicule repository interface
type VehiculeRepository interface {
	Create(ctx context.Context, input *CreateVehiculeInput) (*ent.Vehicule, error)
	GetByID(ctx context.Context, id string) (*ent.Vehicule, error)
	GetByImmatriculation(ctx context.Context, immatriculation string) (*ent.Vehicule, error)
	List(ctx context.Context, filters *VehiculeFilters) ([]*ent.Vehicule, error)
	Update(ctx context.Context, id string, input *UpdateVehiculeInput) (*ent.Vehicule, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]*ent.Vehicule, error)
	GetByProprietaire(ctx context.Context, nom, prenom string) ([]*ent.Vehicule, error)
}

// CreateVehiculeInput represents input for creating vehicule
type CreateVehiculeInput struct {
	ID                             string
	Immatriculation                string
	Marque                         string
	Modele                         string
	Couleur                        *string
	TypeVehicule                   string
	Energie                        *string
	DatePremiereMiseEnCirculation  *string
	NumeroChassis                  *string
	ProprietaireNom                *string
	ProprietairePrenom             *string
	ProprietaireAdresse            *string
	AssuranceCompagnie             *string
	AssuranceNumero                *string
	AssuranceValidite              *string
	ControleTechniqueValidite      *string
}

// UpdateVehiculeInput represents input for updating vehicule
type UpdateVehiculeInput struct {
	Marque                         *string
	Modele                         *string
	Couleur                        *string
	TypeVehicule                   *string
	Energie                        *string
	DatePremiereMiseEnCirculation  *string
	NumeroChassis                  *string
	ProprietaireNom                *string
	ProprietairePrenom             *string
	ProprietaireAdresse            *string
	AssuranceCompagnie             *string
	AssuranceNumero                *string
	AssuranceValidite              *string
	ControleTechniqueValidite      *string
	Active                         *bool
}

// VehiculeFilters represents filters for listing vehicules
type VehiculeFilters struct {
	Marque          *string
	Modele          *string
	TypeVehicule    *string
	Active          *bool
	ProprietaireNom *string
	Limit           int
	Offset          int
}

// vehiculeRepository implements VehiculeRepository
type vehiculeRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewVehiculeRepository creates a new vehicule repository
func NewVehiculeRepository(client *ent.Client, logger *zap.Logger) VehiculeRepository {
	return &vehiculeRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new vehicule
func (r *vehiculeRepository) Create(ctx context.Context, input *CreateVehiculeInput) (*ent.Vehicule, error) {
	r.logger.Info("Creating vehicule", zap.String("immatriculation", input.Immatriculation))

	id, _ := uuid.Parse(input.ID)
	create := r.client.Vehicule.Create().
		SetID(id).
		SetImmatriculation(strings.ToUpper(input.Immatriculation)).
		SetMarque(input.Marque).
		SetModele(input.Modele).
		SetTypeVehicule(input.TypeVehicule)

	if input.Couleur != nil {
		create = create.SetCouleur(*input.Couleur)
	}
	if input.Energie != nil {
		create = create.SetEnergie(*input.Energie)
	}
	if input.NumeroChassis != nil {
		create = create.SetNumeroChassis(*input.NumeroChassis)
	}
	if input.ProprietaireNom != nil {
		create = create.SetProprietaireNom(*input.ProprietaireNom)
	}
	if input.ProprietairePrenom != nil {
		create = create.SetProprietairePrenom(*input.ProprietairePrenom)
	}
	if input.ProprietaireAdresse != nil {
		create = create.SetProprietaireAdresse(*input.ProprietaireAdresse)
	}
	if input.AssuranceCompagnie != nil {
		create = create.SetAssuranceCompagnie(*input.AssuranceCompagnie)
	}
	if input.AssuranceNumero != nil {
		create = create.SetAssuranceNumero(*input.AssuranceNumero)
	}

	vehiculeEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create vehicule", zap.Error(err))
		return nil, fmt.Errorf("failed to create vehicule: %w", err)
	}

	return vehiculeEnt, nil
}

// GetByID gets vehicule by ID
func (r *vehiculeRepository) GetByID(ctx context.Context, id string) (*ent.Vehicule, error) {
	uid, _ := uuid.Parse(id)
	vehiculeEnt, err := r.client.Vehicule.
		Query().
		Where(vehicule.ID(uid)).
		WithControles().
		WithInfractions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("vehicule not found")
		}
		r.logger.Error("Failed to get vehicule by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get vehicule: %w", err)
	}

	return vehiculeEnt, nil
}

// GetByImmatriculation gets vehicule by immatriculation
func (r *vehiculeRepository) GetByImmatriculation(ctx context.Context, immatriculation string) (*ent.Vehicule, error) {
	vehiculeEnt, err := r.client.Vehicule.
		Query().
		Where(vehicule.Immatriculation(strings.ToUpper(immatriculation))).
		WithControles().
		WithInfractions().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("vehicule not found")
		}
		r.logger.Error("Failed to get vehicule by immatriculation", 
			zap.String("immatriculation", immatriculation), zap.Error(err))
		return nil, fmt.Errorf("failed to get vehicule: %w", err)
	}

	return vehiculeEnt, nil
}

// List gets vehicules with filters
func (r *vehiculeRepository) List(ctx context.Context, filters *VehiculeFilters) ([]*ent.Vehicule, error) {
	query := r.client.Vehicule.Query()

	if filters != nil {
		if filters.Marque != nil {
			query = query.Where(vehicule.MarqueContains(*filters.Marque))
		}
		if filters.Modele != nil {
			query = query.Where(vehicule.ModeleContains(*filters.Modele))
		}
		if filters.TypeVehicule != nil {
			query = query.Where(vehicule.TypeVehicule(*filters.TypeVehicule))
		}
		if filters.Active != nil {
			query = query.Where(vehicule.Active(*filters.Active))
		}
		if filters.ProprietaireNom != nil {
			query = query.Where(vehicule.ProprietaireNomContains(*filters.ProprietaireNom))
		}

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	vehicules, err := query.
		WithControles().
		Order(ent.Desc(vehicule.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list vehicules", zap.Error(err))
		return nil, fmt.Errorf("failed to list vehicules: %w", err)
	}

	return vehicules, nil
}

// Update updates vehicule
func (r *vehiculeRepository) Update(ctx context.Context, id string, input *UpdateVehiculeInput) (*ent.Vehicule, error) {
	r.logger.Info("Updating vehicule", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.Vehicule.UpdateOneID(uid)

	if input.Marque != nil {
		update = update.SetMarque(*input.Marque)
	}
	if input.Modele != nil {
		update = update.SetModele(*input.Modele)
	}
	if input.Couleur != nil {
		update = update.SetCouleur(*input.Couleur)
	}
	if input.TypeVehicule != nil {
		update = update.SetTypeVehicule(*input.TypeVehicule)
	}
	if input.Energie != nil {
		update = update.SetEnergie(*input.Energie)
	}
	if input.NumeroChassis != nil {
		update = update.SetNumeroChassis(*input.NumeroChassis)
	}
	if input.ProprietaireNom != nil {
		update = update.SetProprietaireNom(*input.ProprietaireNom)
	}
	if input.ProprietairePrenom != nil {
		update = update.SetProprietairePrenom(*input.ProprietairePrenom)
	}
	if input.ProprietaireAdresse != nil {
		update = update.SetProprietaireAdresse(*input.ProprietaireAdresse)
	}
	if input.AssuranceCompagnie != nil {
		update = update.SetAssuranceCompagnie(*input.AssuranceCompagnie)
	}
	if input.AssuranceNumero != nil {
		update = update.SetAssuranceNumero(*input.AssuranceNumero)
	}
	if input.Active != nil {
		update = update.SetActive(*input.Active)
	}

	vehiculeEnt, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update vehicule", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update vehicule: %w", err)
	}

	return vehiculeEnt, nil
}

// Delete deletes vehicule
func (r *vehiculeRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting vehicule", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.Vehicule.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete vehicule", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete vehicule: %w", err)
	}

	return nil
}

// Search searches vehicules by query
func (r *vehiculeRepository) Search(ctx context.Context, query string) ([]*ent.Vehicule, error) {
	searchQuery := strings.ToLower(strings.TrimSpace(query))
	if searchQuery == "" {
		return []*ent.Vehicule{}, nil
	}

	vehicules, err := r.client.Vehicule.Query().
		Where(
			vehicule.Or(
				vehicule.ImmatriculationContains(strings.ToUpper(searchQuery)),
				vehicule.MarqueContains(searchQuery),
				vehicule.ModeleContains(searchQuery),
				vehicule.ProprietaireNomContains(searchQuery),
				vehicule.ProprietairePrenom(searchQuery),
			),
		).
		WithControles().
		Order(ent.Desc(vehicule.FieldCreatedAt)).
		Limit(50).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to search vehicules", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("failed to search vehicules: %w", err)
	}

	return vehicules, nil
}

// GetByProprietaire gets vehicules by proprietaire
func (r *vehiculeRepository) GetByProprietaire(ctx context.Context, nom, prenom string) ([]*ent.Vehicule, error) {
	vehicules, err := r.client.Vehicule.Query().
		Where(
			vehicule.And(
				vehicule.ProprietaireNomEqualFold(nom),
				vehicule.ProprietairePrenomEqualFold(prenom),
			),
		).
		WithControles().
		Order(ent.Desc(vehicule.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to get vehicules by proprietaire", 
			zap.String("nom", nom), zap.String("prenom", prenom), zap.Error(err))
		return nil, fmt.Errorf("failed to get vehicules by proprietaire: %w", err)
	}

	return vehicules, nil
}