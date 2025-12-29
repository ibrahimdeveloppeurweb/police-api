package objectif

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides objectif module dependencies
var Module = fx.Module("objectif",
	fx.Provide(
		NewRepository,
		NewService,
		fx.Annotate(
			NewController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewRepository creates a new objectif repository
func NewRepository(client *ent.Client, logger *zap.Logger) repository.ObjectifRepository {
	return repository.NewObjectifRepository(client, logger)
}
