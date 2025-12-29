package conducteur

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides conducteur service dependencies
var Module = fx.Module("conducteur",
	fx.Provide(
		NewConducteurService,
		fx.Annotate(
			NewConducteurController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewConducteurService creates a new conducteur service for DI
func NewConducteurService(repo repository.ConducteurRepository, logger *zap.Logger) Service {
	return NewService(repo, logger)
}

// NewConducteurController creates a new conducteur controller for DI
func NewConducteurController(service Service) interfaces.Controller {
	return NewController(service)
}