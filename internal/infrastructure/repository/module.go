package repository

import (
	"go.uber.org/fx"
)

// Module provides repository dependencies
var Module = fx.Module("repository",
	fx.Provide(
		NewUserRepository,
		NewVehiculeRepository,
		NewConducteurRepository,
		NewControleRepository,
		NewInfractionRepository,
		NewInfractionTypeRepository,
		NewPaiementRepository,
		NewDocumentRepository,
		NewPVRepository,
		NewRecoursRepository,
		NewCommissariatRepository,
		NewAlerteRepository,
		NewVerificationRepository,
		NewObjetPerduRepository,
		NewObjetRetrouveRepository,
	),
)