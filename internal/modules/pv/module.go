package pv

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides PV service dependencies
var Module = fx.Module("pv",
	fx.Provide(
		NewPVServiceProvider,
		fx.Annotate(
			NewPVControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewPVServiceProvider creates a new PV service for DI
func NewPVServiceProvider(
	pvRepo repository.PVRepository,
	logger *zap.Logger,
) Service {
	return NewPVService(pvRepo, logger)
}

// NewPVControllerProvider creates a new PV controller for DI
func NewPVControllerProvider(service Service) interfaces.Controller {
	return NewPVController(service)
}



