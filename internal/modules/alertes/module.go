package alertes

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides alertes module dependencies
var Module = fx.Module("alertes",
	fx.Provide(
		NewServiceProvider,
		fx.Annotate(
			NewControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewServiceProvider creates a new alertes service for DI
func NewServiceProvider(
	alerteRepo repository.AlerteRepository,
	userRepo repository.UserRepository,
	commissariatRepo repository.CommissariatRepository,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	return NewService(alerteRepo, userRepo, commissariatRepo, cfg, logger)
}

// NewControllerProvider creates a new alertes controller for DI
func NewControllerProvider(service Service, authMiddleware *middleware.AuthMiddleware) interfaces.Controller {
	return NewController(service, authMiddleware)
}



