package app

import (
	"context"

	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/core/router"
	"police-trafic-api-frontend-aligned/internal/core/server"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"
	"police-trafic-api-frontend-aligned/internal/infrastructure/crypto"
	"police-trafic-api-frontend-aligned/internal/infrastructure/database"
	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"
	"police-trafic-api-frontend-aligned/internal/infrastructure/logger"
	"police-trafic-api-frontend-aligned/internal/infrastructure/rbac"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/infrastructure/session"
	"police-trafic-api-frontend-aligned/internal/modules/admin"
	"police-trafic-api-frontend-aligned/internal/modules/alertes"
	"police-trafic-api-frontend-aligned/internal/modules/auth"
	"police-trafic-api-frontend-aligned/internal/modules/commissariat"
	"police-trafic-api-frontend-aligned/internal/modules/competence"
	"police-trafic-api-frontend-aligned/internal/modules/conducteur"
	"police-trafic-api-frontend-aligned/internal/modules/controle"
	"police-trafic-api-frontend-aligned/internal/modules/convocations"
	"police-trafic-api-frontend-aligned/internal/modules/document"
	"police-trafic-api-frontend-aligned/internal/modules/equipe"
	"police-trafic-api-frontend-aligned/internal/modules/infraction"
	"police-trafic-api-frontend-aligned/internal/modules/inspection"
	"police-trafic-api-frontend-aligned/internal/modules/mission"
		"police-trafic-api-frontend-aligned/internal/modules/objectif"
		"police-trafic-api-frontend-aligned/internal/modules/observation"
		"police-trafic-api-frontend-aligned/internal/modules/officers"
		"police-trafic-api-frontend-aligned/internal/modules/paiement"
		"police-trafic-api-frontend-aligned/internal/modules/plainte"
		"police-trafic-api-frontend-aligned/internal/modules/pv"
		"police-trafic-api-frontend-aligned/internal/modules/recours"
		"police-trafic-api-frontend-aligned/internal/modules/vehicule"
		"police-trafic-api-frontend-aligned/internal/modules/verification"
		"police-trafic-api-frontend-aligned/internal/modules/objets-perdus"
		"police-trafic-api-frontend-aligned/internal/modules/objets-retrouves"

	"go.uber.org/fx"
)

// BuildApp builds the application with all dependencies
func BuildApp() fx.Option {
	return fx.Options(
		// Infrastructure
		config.Module,
		logger.Module,
		database.Module,
		jwt.Module,
		crypto.Module,
		rbac.Module,
		middleware.Module,
		repository.Module,
		session.Module,
		
		// Modules
		admin.Module,
		alertes.Module,
		auth.Module,
		commissariat.Module,
		competence.Module,
		conducteur.Module,
		controle.Module,
		convocations.Module,  // ✅ AJOUTÉ ICI
		document.Module,
		equipe.Module,
		infraction.Module,
		inspection.Module,
		mission.Module,
		objectif.Module,
		observation.Module,
		officers.Module,
		paiement.Module,
		plainte.Module,
		pv.Module,
		recours.Module,
		vehicule.Module,
		verification.Module,
		objetsperdus.Module,
		objetsretrouves.Module,
		
		// Core
		fx.Provide(
			fx.Annotate(
				router.NewServer,
				fx.ParamTags(``, ``, ``, `group:"controllers"`),
			),
		),

		// Application lifecycle
		fx.Invoke(func(lc fx.Lifecycle, server *server.Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						if err := server.Start(ctx); err != nil {
							// Log error but don't fail startup
							return
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return server.Shutdown(ctx)
				},
			})
		}),
	)
}
