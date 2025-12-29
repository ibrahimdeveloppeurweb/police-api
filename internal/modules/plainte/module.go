package plainte

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides plainte service dependencies
var Module = fx.Module("plainte",
	fx.Provide(
		NewPlainteService,
		fx.Annotate(
			NewPlainteController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewPlainteService creates a new plainte service for DI
func NewPlainteService(
	client *ent.Client,
	logger *zap.Logger,
) Service {
	return NewService(client, logger)
}

// NewPlainteController creates a new plainte controller for DI
func NewPlainteController(service Service) interfaces.Controller {
	return NewController(service)
}
