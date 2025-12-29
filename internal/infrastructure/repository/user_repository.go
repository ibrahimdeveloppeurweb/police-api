package repository

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/commissariat"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserRepository defines user repository interface
type UserRepository interface {
	Create(ctx context.Context, userInput *CreateUserInput) (*ent.User, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	GetByMatricule(ctx context.Context, matricule string) (*ent.User, error)
	GetByEmail(ctx context.Context, email string) (*ent.User, error)
	List(ctx context.Context) ([]*ent.User, error)
	ListWithFilters(ctx context.Context, filters *UserFilters) ([]*ent.User, error)
	Update(ctx context.Context, id string, userInput *UpdateUserInput) (*ent.User, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filters *UserFilters) (int, error)
}

// UserFilters represents filters for listing users
type UserFilters struct {
	CommissariatID *string
	Role           *string
	StatutService  *string
	Active         *bool
	Search         *string
}

// CreateUserInput represents input for creating user
type CreateUserInput struct {
	ID             string
	Matricule      string
	Nom            string
	Prenom         string
	Email          string
	Password       string
	Role           string
	Grade          *string
	Telephone      *string
	CommissariatID *string
}

// UpdateUserInput represents input for updating user
type UpdateUserInput struct {
	Nom            *string
	Prenom         *string
	Email          *string
	Password       *string
	Role           *string
	Grade          *string
	Telephone      *string
	StatutService  *string
	Localisation   *string
	Activite       *string
	Active         *bool
	CommissariatID *string
}

// userRepository implements UserRepository
type userRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(client *ent.Client, logger *zap.Logger) UserRepository {
	return &userRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, input *CreateUserInput) (*ent.User, error) {
	r.logger.Info("Creating user", zap.String("matricule", input.Matricule))

	id, _ := uuid.Parse(input.ID)
	create := r.client.User.
		Create().
		SetID(id).
		SetMatricule(input.Matricule).
		SetNom(input.Nom).
		SetPrenom(input.Prenom).
		SetEmail(input.Email).
		SetPassword(input.Password).
		SetRole(input.Role).
		SetStatutService("HORS_SERVICE")

	if input.Grade != nil {
		create = create.SetGrade(*input.Grade)
	}
	if input.Telephone != nil {
		create = create.SetTelephone(*input.Telephone)
	}
	if input.CommissariatID != nil {
		commID, _ := uuid.Parse(*input.CommissariatID)
		create = create.SetCommissariatID(commID)
	}

	userEnt, err := create.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Reload with commissariat edge
	return r.GetByID(ctx, userEnt.ID.String())
}

// GetByID gets user by ID with all relations
func (r *userRepository) GetByID(ctx context.Context, id string) (*ent.User, error) {
	uid, _ := uuid.Parse(id)
	userEnt, err := r.client.User.
		Query().
		Where(user.ID(uid)).
		WithCommissariat().
		WithEquipe().
		WithSuperieur().
		WithMissions(func(q *ent.MissionQuery) {
			q.Order(ent.Desc("date_debut")).Limit(10)
		}).
		WithObjectifs(func(q *ent.ObjectifQuery) {
			q.Order(ent.Desc("created_at"))
		}).
		WithObservations(func(q *ent.ObservationQuery) {
			q.Order(ent.Desc("created_at")).Limit(10)
		}).
		WithCompetences().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return userEnt, nil
}

// GetByMatricule gets user by matricule with commissariat
func (r *userRepository) GetByMatricule(ctx context.Context, matricule string) (*ent.User, error) {
	user, err := r.client.User.
		Query().
		Where(user.Matricule(matricule)).
		WithCommissariat().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by matricule", zap.String("matricule", matricule), zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail gets user by email with commissariat
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	user, err := r.client.User.
		Query().
		Where(user.Email(email)).
		WithCommissariat().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// List gets all active users with commissariat
func (r *userRepository) List(ctx context.Context) ([]*ent.User, error) {
	users, err := r.client.User.
		Query().
		Where(user.Active(true)).
		WithCommissariat().
		Order(ent.Desc(user.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// ListWithFilters gets users with filters and commissariat
func (r *userRepository) ListWithFilters(ctx context.Context, filters *UserFilters) ([]*ent.User, error) {
	query := r.client.User.Query().WithCommissariat()

	if filters != nil {
		if filters.Active != nil {
			query = query.Where(user.Active(*filters.Active))
		}
		if filters.Role != nil && *filters.Role != "" {
			query = query.Where(user.Role(*filters.Role))
		}
		if filters.StatutService != nil && *filters.StatutService != "" {
			query = query.Where(user.StatutService(*filters.StatutService))
		}
		if filters.CommissariatID != nil && *filters.CommissariatID != "" {
			commID, _ := uuid.Parse(*filters.CommissariatID)
			query = query.Where(user.HasCommissariatWith(commissariat.ID(commID)))
		}
		if filters.Search != nil && *filters.Search != "" {
			search := "%" + *filters.Search + "%"
			query = query.Where(
				user.Or(
					user.NomContainsFold(*filters.Search),
					user.PrenomContainsFold(*filters.Search),
					user.MatriculeContainsFold(*filters.Search),
					user.EmailContains(search),
				),
			)
		}
	}

	users, err := query.Order(ent.Desc(user.FieldCreatedAt)).All(ctx)
	if err != nil {
		r.logger.Error("Failed to list users with filters", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Count counts users with optional filters
func (r *userRepository) Count(ctx context.Context, filters *UserFilters) (int, error) {
	query := r.client.User.Query()

	if filters != nil {
		if filters.Active != nil {
			query = query.Where(user.Active(*filters.Active))
		}
		if filters.Role != nil && *filters.Role != "" {
			query = query.Where(user.Role(*filters.Role))
		}
		if filters.StatutService != nil && *filters.StatutService != "" {
			query = query.Where(user.StatutService(*filters.StatutService))
		}
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count users", zap.Error(err))
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Update updates user
func (r *userRepository) Update(ctx context.Context, id string, input *UpdateUserInput) (*ent.User, error) {
	r.logger.Info("Updating user", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	update := r.client.User.UpdateOneID(uid)

	if input.Nom != nil {
		update = update.SetNom(*input.Nom)
	}
	if input.Prenom != nil {
		update = update.SetPrenom(*input.Prenom)
	}
	if input.Email != nil {
		update = update.SetEmail(*input.Email)
	}
	if input.Password != nil {
		update = update.SetPassword(*input.Password)
	}
	if input.Role != nil {
		update = update.SetRole(*input.Role)
	}
	if input.Grade != nil {
		update = update.SetGrade(*input.Grade)
	}
	if input.Telephone != nil {
		update = update.SetTelephone(*input.Telephone)
	}
	if input.StatutService != nil {
		update = update.SetStatutService(*input.StatutService)
	}
	if input.Localisation != nil {
		update = update.SetLocalisation(*input.Localisation)
	}
	if input.Activite != nil {
		update = update.SetActivite(*input.Activite)
	}
	if input.Active != nil {
		update = update.SetActive(*input.Active)
	}
	if input.CommissariatID != nil {
		commID, _ := uuid.Parse(*input.CommissariatID)
		update = update.SetCommissariatID(commID)
	}

	_, err := update.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update user", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Reload with commissariat edge
	return r.GetByID(ctx, id)
}

// Delete deletes user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting user", zap.String("id", id))

	uid, _ := uuid.Parse(id)
	err := r.client.User.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}