package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides logger dependency
var Module = fx.Module("logger",
	fx.Provide(NewLogger),
	fx.Invoke(func(logger *zap.Logger) {
		// Replace global logger
		zap.ReplaceGlobals(logger)
	}),
)


