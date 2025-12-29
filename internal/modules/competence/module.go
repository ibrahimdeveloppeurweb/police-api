package competence

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides competence module dependencies
var Module = fx.Module("competence",
	fx.Provide(
		NewRepository,
		NewServiceProvider,
		fx.Annotate(
			NewController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewRepository creates a new competence repository
func NewRepository(client *ent.Client, logger *zap.Logger) repository.CompetenceRepository {
	return repository.NewCompetenceRepository(client, logger)
}

// NewServiceProvider creates a new competence service for DI
func NewServiceProvider(repo repository.CompetenceRepository, logger *zap.Logger) Service {
	return NewService(repo, logger)
}
