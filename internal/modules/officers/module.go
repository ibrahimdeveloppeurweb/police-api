package officers

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides officers module dependencies for mobile app
var Module = fx.Module("officers",
	fx.Provide(
		NewServiceProvider,
		fx.Annotate(
			NewControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewServiceProvider creates a new officers service for DI
func NewServiceProvider(
	client *ent.Client,
	userRepo repository.UserRepository,
	infractionRepo repository.InfractionRepository,
	logger *zap.Logger,
) Service {
	return NewService(client, userRepo, infractionRepo, logger)
}

// NewControllerProvider creates a new officers controller for DI
func NewControllerProvider(service Service) interfaces.Controller {
	return NewController(service)
}
