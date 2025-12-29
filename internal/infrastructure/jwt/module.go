package jwt

import "go.uber.org/fx"

// Module provides JWT service dependency
var Module = fx.Module("jwt",
	fx.Provide(NewJWTService),
)