package recours

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides Recours service dependencies
var Module = fx.Module("recours",
	fx.Provide(
		NewRecoursServiceProvider,
		fx.Annotate(
			NewRecoursControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewRecoursServiceProvider creates a new Recours service for DI
func NewRecoursServiceProvider(
	recoursRepo repository.RecoursRepository,
	logger *zap.Logger,
) Service {
	return NewRecoursService(recoursRepo, logger)
}

// NewRecoursControllerProvider creates a new Recours controller for DI
func NewRecoursControllerProvider(service Service) interfaces.Controller {
	return NewRecoursController(service)
}
