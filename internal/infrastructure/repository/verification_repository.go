package repository

import (
	"context"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/checkitem"
	"police-trafic-api-frontend-aligned/ent/checkoption"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateVerificationInput represents input for creating a verification
type CreateVerificationInput struct {
	SourceType    string
	SourceID      string
	CheckItemID   string
	ResultStatus  string
	Notes         *string
	MontantAmende *int
}

// VerificationRepository handles verification (CheckOption) database operations
type VerificationRepository interface {
	// CheckItems (catalogue)
	ListCheckItems(ctx context.Context, applicableTo string, category string, activeOnly bool) ([]*ent.CheckItem, error)
	GetCheckItemByID(ctx context.Context, id string) (*ent.CheckItem, error)
	GetCheckItemByCode(ctx context.Context, code string) (*ent.CheckItem, error)

	// CheckOptions (résultats de vérifications)
	GetVerificationsBySource(ctx context.Context, sourceType string, sourceID string) ([]*ent.CheckOption, error)
	CreateVerification(ctx context.Context, opt *ent.CheckOption) (*ent.CheckOption, error)
	CreateVerificationFromInput(ctx context.Context, input *CreateVerificationInput) (*ent.CheckOption, error)
	UpdateVerification(ctx context.Context, id string, opt *ent.CheckOption) (*ent.CheckOption, error)
	DeleteVerification(ctx context.Context, id string) error
	DeleteVerificationsBySource(ctx context.Context, sourceType string, sourceID string) error
}

type verificationRepository struct {
	client *ent.Client
	logger *zap.Logger
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(client *ent.Client, logger *zap.Logger) VerificationRepository {
	return &verificationRepository{
		client: client,
		logger: logger,
	}
}

// ListCheckItems lists check items from the catalogue
func (r *verificationRepository) ListCheckItems(ctx context.Context, applicableTo string, category string, activeOnly bool) ([]*ent.CheckItem, error) {
	query := r.client.CheckItem.Query()

	if applicableTo != "" {
		// Filter by applicable_to (INSPECTION, CONTROL, BOTH)
		query = query.Where(
			checkitem.Or(
				checkitem.ApplicableToEQ(checkitem.ApplicableTo(applicableTo)),
				checkitem.ApplicableToEQ(checkitem.ApplicableToBOTH),
			),
		)
	}

	if category != "" {
		query = query.Where(checkitem.ItemCategoryEQ(checkitem.ItemCategory(category)))
	}

	if activeOnly {
		query = query.Where(checkitem.IsActiveEQ(true))
	}

	return query.
		Order(ent.Asc(checkitem.FieldDisplayOrder)).
		All(ctx)
}

// GetCheckItemByID gets a check item by ID
func (r *verificationRepository) GetCheckItemByID(ctx context.Context, id string) (*ent.CheckItem, error) {
	uid, _ := uuid.Parse(id)
	return r.client.CheckItem.Get(ctx, uid)
}

// GetCheckItemByCode gets a check item by code
func (r *verificationRepository) GetCheckItemByCode(ctx context.Context, code string) (*ent.CheckItem, error) {
	return r.client.CheckItem.Query().
		Where(checkitem.ItemCodeEQ(code)).
		Only(ctx)
}

// GetVerificationsBySource gets all verifications for a source (controle or inspection)
func (r *verificationRepository) GetVerificationsBySource(ctx context.Context, sourceType string, sourceID string) ([]*ent.CheckOption, error) {
	return r.client.CheckOption.Query().
		Where(
			checkoption.SourceTypeEQ(checkoption.SourceType(sourceType)),
			checkoption.SourceIDEQ(sourceID),
		).
		WithCheckItem().
		WithEvidenceFile().                                    // Charger le document preuve (photo)
		WithInfraction(func(q *ent.InfractionQuery) { // Charger l'infraction liée (si FAIL)
			q.WithTypeInfraction()
		}).
		All(ctx)
}

// CreateVerification creates a new verification result
func (r *verificationRepository) CreateVerification(ctx context.Context, opt *ent.CheckOption) (*ent.CheckOption, error) {
	// Generate ID if not provided
	id := opt.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	return r.client.CheckOption.Create().
		SetID(id).
		SetSourceType(opt.SourceType).
		SetSourceID(opt.SourceID).
		SetResultStatus(opt.ResultStatus).
		SetNotes(opt.Notes).
		SetFineAmount(opt.FineAmount).
		SetCheckedAt(opt.CheckedAt).
		SetCheckItemID(opt.Edges.CheckItem.ID).
		Save(ctx)
}

// CreateVerificationFromInput creates a new verification from simple input struct
func (r *verificationRepository) CreateVerificationFromInput(ctx context.Context, input *CreateVerificationInput) (*ent.CheckOption, error) {
	checkItemID, err := uuid.Parse(input.CheckItemID)
	if err != nil {
		return nil, err
	}

	create := r.client.CheckOption.Create().
		SetID(uuid.New()).
		SetSourceType(checkoption.SourceType(input.SourceType)).
		SetSourceID(input.SourceID).
		SetResultStatus(checkoption.ResultStatus(input.ResultStatus)).
		SetCheckItemID(checkItemID)

	if input.Notes != nil {
		create.SetNotes(*input.Notes)
	}

	if input.MontantAmende != nil {
		create.SetFineAmount(*input.MontantAmende)
	}

	return create.Save(ctx)
}

// UpdateVerification updates a verification result
func (r *verificationRepository) UpdateVerification(ctx context.Context, id string, opt *ent.CheckOption) (*ent.CheckOption, error) {
	uid, _ := uuid.Parse(id)
	return r.client.CheckOption.UpdateOneID(uid).
		SetResultStatus(opt.ResultStatus).
		SetNotes(opt.Notes).
		SetFineAmount(opt.FineAmount).
		Save(ctx)
}

// DeleteVerification deletes a verification
func (r *verificationRepository) DeleteVerification(ctx context.Context, id string) error {
	uid, _ := uuid.Parse(id)
	return r.client.CheckOption.DeleteOneID(uid).Exec(ctx)
}

// DeleteVerificationsBySource deletes all verifications for a source
func (r *verificationRepository) DeleteVerificationsBySource(ctx context.Context, sourceType string, sourceID string) error {
	_, err := r.client.CheckOption.Delete().
		Where(
			checkoption.SourceTypeEQ(checkoption.SourceType(sourceType)),
			checkoption.SourceIDEQ(sourceID),
		).
		Exec(ctx)
	return err
}
