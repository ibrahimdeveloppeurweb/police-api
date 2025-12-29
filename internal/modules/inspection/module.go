package inspection

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/modules/verification"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides inspection service dependencies
var Module = fx.Module("inspection",
	fx.Provide(
		NewInspectionService,
		fx.Annotate(
			NewInspectionController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewInspectionService creates a new inspection service for DI
func NewInspectionService(
	client *ent.Client,
	verificationRepo repository.VerificationRepository,
	logger *zap.Logger,
) Service {
	return NewService(client, verificationRepo, logger)
}

// NewInspectionController creates a new inspection controller for DI
func NewInspectionController(service Service, verificationService verification.Service) interfaces.Controller {
	return NewController(service, verificationService)
}
