package paiement

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides paiement service dependencies
var Module = fx.Module("paiement",
	fx.Provide(
		NewPaiementServiceProvider,
		fx.Annotate(
			NewPaiementControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewPaiementServiceProvider creates a new paiement service for DI
func NewPaiementServiceProvider(
	paiementRepo repository.PaiementRepository,
	logger *zap.Logger,
) Service {
	return NewPaiementService(paiementRepo, logger)
}

// NewPaiementControllerProvider creates a new paiement controller for DI
func NewPaiementControllerProvider(service Service) interfaces.Controller {
	return NewPaiementController(service)
}
