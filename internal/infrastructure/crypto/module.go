package crypto

import "go.uber.org/fx"

// Module provides crypto service dependency
var Module = fx.Module("crypto",
	fx.Provide(NewPasswordService),
)