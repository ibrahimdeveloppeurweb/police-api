package mission

import (
	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides mission module dependencies
var Module = fx.Module("mission",
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

// NewRepository creates a new mission repository
func NewRepository(client *ent.Client, logger *zap.Logger) repository.MissionRepository {
	return repository.NewMissionRepository(client, logger)
}

// NewServiceProvider creates a new mission service for DI
func NewServiceProvider(repo repository.MissionRepository, logger *zap.Logger) Service {
	return NewService(repo, logger)
}
