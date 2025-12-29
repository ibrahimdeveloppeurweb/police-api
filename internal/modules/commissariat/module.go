package commissariat

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides commissariat module dependencies
var Module = fx.Module("commissariat",
	fx.Provide(
		NewServiceProvider,
		fx.Annotate(
			NewControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewServiceProvider creates a new commissariat service for DI
func NewServiceProvider(
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	controleRepo repository.ControleRepository,
	pvRepo repository.PVRepository,
	alerteRepo repository.AlerteRepository,
	logger *zap.Logger,
) Service {
	return NewService(commissariatRepo, userRepo, controleRepo, pvRepo, alerteRepo, logger)
}

// NewControllerProvider creates a new commissariat controller for DI
func NewControllerProvider(service Service) interfaces.Controller {
	return NewController(service)
}



