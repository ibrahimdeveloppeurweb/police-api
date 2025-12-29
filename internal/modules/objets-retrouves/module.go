package objetsretrouves

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides objets retrouves module dependencies
var Module = fx.Module("objets-retrouves",
	fx.Provide(
		NewServiceProvider,
		fx.Annotate(
			NewControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewServiceProvider creates a new objets retrouves service for DI
func NewServiceProvider(
	objetRetrouveRepo repository.ObjetRetrouveRepository,
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	return NewService(objetRetrouveRepo, commissariatRepo, userRepo, cfg, logger)
}

// NewControllerProvider creates a new objets retrouves controller for DI
func NewControllerProvider(service Service, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) interfaces.Controller {
	return NewController(service, authMiddleware, logger)
}

