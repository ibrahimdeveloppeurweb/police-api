package verification

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides verification service dependencies
var Module = fx.Module("verification",
	fx.Provide(
		NewVerificationService,
	),
)

// NewVerificationService creates a new verification service for DI
func NewVerificationService(client *ent.Client, logger *zap.Logger) Service {
	repo := repository.NewVerificationRepository(client, logger)
	return NewService(repo, logger)
}
