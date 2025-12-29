package repository

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/commissariat"
	"police-trafic-api-frontend-aligned/ent/convocation"
	"police-trafic-api-frontend-aligned/ent/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConvocationRepository defines convocation repository interface
type ConvocationRepository interface {
	Create(ctx context.Context, conv *ent.Convocation) (*ent.Convocation, error)
	GetByID(ctx context.Context, id string) (*ent.Convocation, error)
	GetByNumero(ctx context.Context, numero string) (*ent.Convocation, error)
	List(ctx context.Context, filters *ConvocationFilters) ([]*ent.Convocation, error)
	Count(ctx context.Context, filters *ConvocationFilters) (int64, error)
	Update(ctx context.Context, id string, conv *ent.Convocation) (*ent.Convocation, error)
	Delete(ctx context.Context, id string) error
	Client() *ent.Client
}

// ConvocationFilters represents filters for listing convocations
type ConvocationFilters struct {
	Statut          *string
	TypeConvocation *string
	StatutPersonne  *string
	CommissariatID  *string
	AgentID         *string
	DateDebut       *time.Time
	DateFin         *time.Time
	Search          *string
	Page            int
	Limit           int
	Offset          int
}

// convocationRepository implements ConvocationRepository
type convocationRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewConvocationRepository creates a new convocation repository
func NewConvocationRepository(client *ent.Client, logger *zap.Logger) ConvocationRepository {
	return &convocationRepository{
		client: client,
		logger: logger,
	}
}

// Client returns the ent client
func (r *convocationRepository) Client() *ent.Client {
	return r.client
}

// Create creates a new convocation
func (r *convocationRepository) Create(ctx context.Context, conv *ent.Convocation) (*ent.Convocation, error) {
	r.logger.Info("Creating convocation in repository",
		zap.String("numero", conv.Numero),
		zap.String("nom", conv.ConvoqueNom))

	// Note: Cette fonction n'est pas vraiment utilisée car on crée directement
	// depuis le service avec CreateBuilder. On la garde pour la compatibilité.
	created, err := r.client.Convocation.Create().
		SetNumero(conv.Numero).
		SetTypeConvocation(conv.TypeConvocation).
		SetConvoqueNom(conv.ConvoqueNom).
		SetConvoquePrenom(conv.ConvoquePrenom).
		SetConvoqueTelephone(conv.ConvoqueTelephone).
		SetStatutPersonne(conv.StatutPersonne).
		SetDateCreation(conv.DateCreation).
		SetStatut(conv.Statut).
		SetModeEnvoi(conv.ModeEnvoi).
		SetLieuRdv(conv.LieuRdv).
		SetMotif(conv.Motif).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to create convocation", zap.Error(err))
		return nil, fmt.Errorf("failed to create convocation: %w", err)
	}

	return created, nil
}

// GetByID gets convocation by ID
func (r *convocationRepository) GetByID(ctx context.Context, id string) (*ent.Convocation, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	conv, err := r.client.Convocation.
		Query().
		Where(convocation.ID(uid)).
		WithAgent().
		WithCommissariat().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("convocation not found")
		}
		r.logger.Error("Failed to get convocation by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get convocation: %w", err)
	}

	return conv, nil
}

// GetByNumero gets convocation by numero
func (r *convocationRepository) GetByNumero(ctx context.Context, numero string) (*ent.Convocation, error) {
	conv, err := r.client.Convocation.
		Query().
		Where(convocation.Numero(numero)).
		WithAgent().
		WithCommissariat().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("convocation not found")
		}
		r.logger.Error("Failed to get convocation by numero", zap.String("numero", numero), zap.Error(err))
		return nil, fmt.Errorf("failed to get convocation: %w", err)
	}

	return conv, nil
}

// List gets convocations with filters
func (r *convocationRepository) List(ctx context.Context, filters *ConvocationFilters) ([]*ent.Convocation, error) {
	query := r.client.Convocation.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	convs, err := query.
		WithAgent().
		WithCommissariat().
		Order(ent.Desc(convocation.FieldDateCreation)).
		All(ctx)

	if err != nil {
		r.logger.Error("Failed to list convocations", zap.Error(err))
		return nil, fmt.Errorf("failed to list convocations: %w", err)
	}

	return convs, nil
}

// Count counts convocations with filters
func (r *convocationRepository) Count(ctx context.Context, filters *ConvocationFilters) (int64, error) {
	query := r.client.Convocation.Query()

	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("Failed to count convocations", zap.Error(err))
		return 0, fmt.Errorf("failed to count convocations: %w", err)
	}

	return int64(count), nil
}

// applyFilters applies filters to convocation query
func (r *convocationRepository) applyFilters(query *ent.ConvocationQuery, filters *ConvocationFilters) *ent.ConvocationQuery {
	if filters.Statut != nil {
		// Convertir le string en enum convocation.Statut
		query = query.Where(convocation.StatutEQ(convocation.Statut(*filters.Statut)))
	}
	if filters.TypeConvocation != nil {
		query = query.Where(convocation.TypeConvocation(*filters.TypeConvocation))
	}
	if filters.StatutPersonne != nil {
		query = query.Where(convocation.StatutPersonne(*filters.StatutPersonne))
	}
	if filters.CommissariatID != nil {
		commID, _ := uuid.Parse(*filters.CommissariatID)
		query = query.Where(convocation.HasCommissariatWith(commissariat.ID(commID)))
	}
	if filters.AgentID != nil {
		agentID, _ := uuid.Parse(*filters.AgentID)
		query = query.Where(convocation.HasAgentWith(user.ID(agentID)))
	}
	if filters.DateDebut != nil {
		query = query.Where(convocation.DateCreationGTE(*filters.DateDebut))
	}
	if filters.DateFin != nil {
		query = query.Where(convocation.DateCreationLTE(*filters.DateFin))
	}
	if filters.Search != nil && *filters.Search != "" {
		query = query.Where(
			convocation.Or(
				convocation.ConvoqueNomContains(*filters.Search),
				convocation.ConvoquePrenomContains(*filters.Search),
				convocation.NumeroContains(*filters.Search),
			),
		)
	}
	return query
}

// Update updates convocation
func (r *convocationRepository) Update(ctx context.Context, id string, conv *ent.Convocation) (*ent.Convocation, error) {
	r.logger.Info("Updating convocation", zap.String("id", id))

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	updated, err := r.client.Convocation.
		UpdateOneID(uid).
		SetStatut(conv.Statut).
		Save(ctx)

	if err != nil {
		r.logger.Error("Failed to update convocation", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update convocation: %w", err)
	}

	return updated, nil
}

// Delete deletes convocation
func (r *convocationRepository) Delete(ctx context.Context, id string) error {
	r.logger.Info("Deleting convocation", zap.String("id", id))

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	err = r.client.Convocation.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete convocation", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete convocation: %w", err)
	}

	return nil
}
