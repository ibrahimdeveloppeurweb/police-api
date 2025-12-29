package infraction

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides infraction service dependencies
var Module = fx.Module("infraction",
	fx.Provide(
		NewInfractionService,
		fx.Annotate(
			NewInfractionController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewInfractionService creates a new infraction service for DI
func NewInfractionService(
	infractionRepo repository.InfractionRepository,
	infractionTypeRepo repository.InfractionTypeRepository,
	controleRepo repository.ControleRepository,
	vehiculeRepo repository.VehiculeRepository,
	conducteurRepo repository.ConducteurRepository,
	pvRepo repository.PVRepository,
	logger *zap.Logger,
) Service {
	return NewService(infractionRepo, infractionTypeRepo, controleRepo, vehiculeRepo, conducteurRepo, pvRepo, logger)
}

// NewInfractionController creates a new infraction controller for DI
func NewInfractionController(service Service) interfaces.Controller {
	return NewController(service)
}