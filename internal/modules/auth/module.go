package auth

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"

	"go.uber.org/fx"
)

// Module provides auth module dependencies
var Module = fx.Module("auth",
	fx.Provide(
		NewService,
		fx.Annotate(
			NewController,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)


