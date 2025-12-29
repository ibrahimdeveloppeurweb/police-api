package vehicule

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides vehicule service dependencies
var Module = fx.Module("vehicule",
	fx.Provide(
		NewVehiculeService,
		fx.Annotate(
			NewVehiculeController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewVehiculeService creates a new vehicule service for DI
func NewVehiculeService(repo repository.VehiculeRepository, logger *zap.Logger) Service {
	return NewService(repo, logger)
}

// NewVehiculeController creates a new vehicule controller for DI
func NewVehiculeController(service Service) interfaces.Controller {
	return NewController(service)
}