package verification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/checkoption"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/zap"
)

// Service defines the verification service interface
type Service interface {
	// CheckItems (catalogue)
	ListCheckItems(ctx context.Context, applicableTo string, category string) (*ListCheckItemsResponse, error)
	GetCheckItemByID(ctx context.Context, id string) (*CheckItemResponse, error)

	// CheckOptions (résultats de vérifications)
	GetVerifications(ctx context.Context, sourceType string, sourceID string) (*ListVerificationsResponse, error)
	SaveVerification(ctx context.Context, sourceType string, sourceID string, req *CreateCheckOptionRequest) (*CheckOptionResponse, error)
	SaveBatchVerifications(ctx context.Context, sourceType string, sourceID string, req *BatchCheckOptionsRequest) (*ListVerificationsResponse, error)
	DeleteVerification(ctx context.Context, id string) error
}

type service struct {
	repo   repository.VerificationRepository
	logger *zap.Logger
}

// NewService creates a new verification service
func NewService(repo repository.VerificationRepository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// ListCheckItems lists check items from the catalogue
func (s *service) ListCheckItems(ctx context.Context, applicableTo string, category string) (*ListCheckItemsResponse, error) {
	items, err := s.repo.ListCheckItems(ctx, applicableTo, category, true)
	if err != nil {
		s.logger.Error("Failed to list check items", zap.Error(err))
		return nil, err
	}

	response := &ListCheckItemsResponse{
		Items: make([]*CheckItemResponse, len(items)),
		Total: len(items),
	}

	for i, item := range items {
		response.Items[i] = checkItemToResponse(item)
	}

	return response, nil
}

// GetCheckItemByID gets a check item by ID
func (s *service) GetCheckItemByID(ctx context.Context, id string) (*CheckItemResponse, error) {
	item, err := s.repo.GetCheckItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return checkItemToResponse(item), nil
}

// GetVerifications gets all verifications for a source
func (s *service) GetVerifications(ctx context.Context, sourceType string, sourceID string) (*ListVerificationsResponse, error) {
	options, err := s.repo.GetVerificationsBySource(ctx, sourceType, sourceID)
	if err != nil {
		s.logger.Error("Failed to get verifications",
			zap.String("source_type", sourceType),
			zap.String("source_id", sourceID),
			zap.Error(err))
		return nil, err
	}

	response := &ListVerificationsResponse{
		Verifications: make([]*CheckOptionResponse, len(options)),
		Total:         len(options),
	}

	for i, opt := range options {
		response.Verifications[i] = checkOptionToResponse(opt)

		// Compteurs
		switch opt.ResultStatus {
		case checkoption.ResultStatusPASS:
			response.TotalOk++
		case checkoption.ResultStatusFAIL:
			response.TotalEchec++
		case checkoption.ResultStatusWARNING:
			response.TotalAttention++
		case checkoption.ResultStatusNOT_CHECKED:
			response.TotalNonVerifie++
		}
		response.MontantTotal += opt.FineAmount
	}

	return response, nil
}

// SaveVerification saves a single verification
func (s *service) SaveVerification(ctx context.Context, sourceType string, sourceID string, req *CreateCheckOptionRequest) (*CheckOptionResponse, error) {
	// Get the check item
	checkItem, err := s.repo.GetCheckItemByID(ctx, req.CheckItemID)
	if err != nil {
		return nil, errors.New("check item not found")
	}

	// Determine fine amount
	fineAmount := 0
	if req.MontantAmende != nil {
		fineAmount = *req.MontantAmende
	} else if req.Resultat == "FAIL" {
		fineAmount = checkItem.FineAmount
	}

	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	// Create the CheckOption entity
	opt := &ent.CheckOption{
		SourceType:   checkoption.SourceType(sourceType),
		SourceID:     sourceID,
		ResultStatus: checkoption.ResultStatus(req.Resultat),
		Notes:        notes,
		FineAmount:   fineAmount,
		CheckedAt:    time.Now(),
		Edges: ent.CheckOptionEdges{
			CheckItem: checkItem,
		},
	}

	created, err := s.repo.CreateVerification(ctx, opt)
	if err != nil {
		s.logger.Error("Failed to create verification", zap.Error(err))
		return nil, err
	}

	// Re-fetch with edges
	created.Edges.CheckItem = checkItem

	return checkOptionToResponse(created), nil
}

// SaveBatchVerifications saves multiple verifications at once
func (s *service) SaveBatchVerifications(ctx context.Context, sourceType string, sourceID string, req *BatchCheckOptionsRequest) (*ListVerificationsResponse, error) {
	// Delete existing verifications for this source
	if err := s.repo.DeleteVerificationsBySource(ctx, sourceType, sourceID); err != nil {
		s.logger.Warn("Failed to delete existing verifications", zap.Error(err))
	}

	// Create new verifications
	for _, v := range req.Verifications {
		_, err := s.SaveVerification(ctx, sourceType, sourceID, &v)
		if err != nil {
			s.logger.Error("Failed to save verification",
				zap.String("check_item_id", v.CheckItemID),
				zap.Error(err))
			// Continue with other verifications
		}
	}

	// Return the updated list
	return s.GetVerifications(ctx, sourceType, sourceID)
}

// DeleteVerification deletes a verification
func (s *service) DeleteVerification(ctx context.Context, id string) error {
	return s.repo.DeleteVerification(ctx, id)
}

// Helper functions

func checkItemToResponse(item *ent.CheckItem) *CheckItemResponse {
	return &CheckItemResponse{
		ID:            item.ID.String(),
		Code:          item.ItemCode,
		Nom:           item.ItemName,
		Categorie:     string(item.ItemCategory),
		Description:   item.Description,
		Icon:          item.Icon,
		Obligatoire:   item.IsMandatory,
		Actif:         item.IsActive,
		Ordre:         item.DisplayOrder,
		MontantAmende: item.FineAmount,
		PointsRetrait: item.PointsRetrait,
		ApplicableA:   string(item.ApplicableTo),
	}
}

func checkOptionToResponse(opt *ent.CheckOption) *CheckOptionResponse {
	resp := &CheckOptionResponse{
		ID:               opt.ID.String(),
		Resultat:         string(opt.ResultStatus),
		Notes:            opt.Notes,
		MontantAmende:    opt.FineAmount,
		DateVerification: opt.CheckedAt,
	}

	// Add check item info if available
	if opt.Edges.CheckItem != nil {
		item := opt.Edges.CheckItem
		resp.CheckItemID = item.ID.String()
		resp.CheckItemCode = item.ItemCode
		resp.CheckItemNom = item.ItemName
		resp.Categorie = string(item.ItemCategory)
		resp.Icon = item.Icon
		resp.Obligatoire = item.IsMandatory
		resp.PointsRetrait = item.PointsRetrait
	}

	return resp
}

// NewServiceProvider provides the service for dependency injection
func NewServiceProvider(client *ent.Client, logger *zap.Logger) Service {
	repo := repository.NewVerificationRepository(client, logger)
	return NewService(repo, logger)
}

// GenerateCheckOptionID generates a unique ID for a check option
func GenerateCheckOptionID(sourceType string, sourceID string, itemCode string) string {
	return fmt.Sprintf("chkopt-%s-%s-%s", sourceType[:3], sourceID, itemCode)
}
