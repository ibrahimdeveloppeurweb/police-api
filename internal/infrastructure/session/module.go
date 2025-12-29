package session

import "go.uber.org/fx"

// Module provides session service for dependency injection
var Module = fx.Module("session",
	fx.Provide(NewService),
)
