package config

import (
	"go.uber.org/fx"
)

// Module provides configuration dependency
var Module = fx.Module("config",
	fx.Provide(LoadConfig),
)


