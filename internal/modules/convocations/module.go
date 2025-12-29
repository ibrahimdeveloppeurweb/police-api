package convocations

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides convocations service dependencies
var Module = fx.Module("convocations",
	fx.Provide(
		NewConvocationsService,
		fx.Annotate(
			NewConvocationsController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewConvocationsService creates a new convocations service for DI
func NewConvocationsService(
	client *ent.Client,
	cfg *config.Config,
	logger *zap.Logger,
) Service {
	// Créer les repositories nécessaires
	convocationRepo := repository.NewConvocationRepository(client, logger)
	commissariatRepo := repository.NewCommissariatRepository(client, logger)
	userRepo := repository.NewUserRepository(client, logger)
	
	return NewService(convocationRepo, commissariatRepo, userRepo, cfg, logger)
}

// NewConvocationsController creates a new convocations controller for DI
func NewConvocationsController(
	service Service,
	authMiddleware *middleware.AuthMiddleware,
	logger *zap.Logger,
) interfaces.Controller {
	return NewController(service, authMiddleware, logger)
}
