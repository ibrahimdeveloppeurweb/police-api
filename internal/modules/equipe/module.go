package equipe

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides equipe module dependencies
var Module = fx.Module("equipe",
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

// NewRepository creates a new equipe repository
func NewRepository(client *ent.Client, logger *zap.Logger) repository.EquipeRepository {
	return repository.NewEquipeRepository(client, logger)
}
