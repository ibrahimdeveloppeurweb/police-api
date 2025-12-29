package middleware

import "go.uber.org/fx"

// Module provides middleware dependencies
var Module = fx.Module("middleware",
	fx.Provide(NewAuthMiddleware),
)