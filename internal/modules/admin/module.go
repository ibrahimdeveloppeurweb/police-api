package admin

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/crypto"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/infrastructure/session"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides admin module dependencies
var Module = fx.Module("admin",
	fx.Provide(
		NewServiceProvider,
		fx.Annotate(
			NewControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewServiceProvider creates a new admin service for DI
func NewServiceProvider(
	commissariatRepo repository.CommissariatRepository,
	userRepo repository.UserRepository,
	controleRepo repository.ControleRepository,
	pvRepo repository.PVRepository,
	alerteRepo repository.AlerteRepository,
	infractionRepo repository.InfractionRepository,
	passwordService crypto.Service,
	sessionService session.Service,
	logger *zap.Logger,
) Service {
	return NewService(commissariatRepo, userRepo, controleRepo, pvRepo, alerteRepo, infractionRepo, passwordService, sessionService, logger)
}

// NewControllerProvider creates a new admin controller for DI
func NewControllerProvider(service Service) interfaces.Controller {
	return NewController(service)
}



