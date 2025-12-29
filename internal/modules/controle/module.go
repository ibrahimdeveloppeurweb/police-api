package controle

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/modules/verification"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides controle service dependencies
var Module = fx.Module("controle",
	fx.Provide(
		NewControleService,
		fx.Annotate(
			NewControleController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewControleService creates a new controle service for DI
func NewControleService(
	controleRepo repository.ControleRepository,
	infractionRepo repository.InfractionRepository,
	pvRepo repository.PVRepository,
	verificationRepo repository.VerificationRepository,
	logger *zap.Logger,
) Service {
	return NewService(controleRepo, infractionRepo, pvRepo, verificationRepo, logger)
}

// NewControleController creates a new controle controller for DI
func NewControleController(service Service, verificationService verification.Service) interfaces.Controller {
	return NewController(service, verificationService)
}