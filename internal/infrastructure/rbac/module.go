package rbac

import "go.uber.org/fx"

// Module provides RBAC service dependency
var Module = fx.Module("rbac",
	fx.Provide(NewRBACService),
)