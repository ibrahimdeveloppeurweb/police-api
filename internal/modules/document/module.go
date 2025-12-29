package document

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides document service dependencies
var Module = fx.Module("document",
	fx.Provide(
		NewDocumentServiceProvider,
		fx.Annotate(
			NewDocumentControllerProvider,
			fx.As(new(interfaces.Controller)),
			fx.ResultTags(`group:"controllers"`),
		),
	),
)

// NewDocumentServiceProvider creates a new document service for DI
func NewDocumentServiceProvider(
	documentRepo repository.DocumentRepository,
	logger *zap.Logger,
) Service {
	return NewDocumentService(documentRepo, logger)
}

// NewDocumentControllerProvider creates a new document controller for DI
func NewDocumentControllerProvider(service Service) interfaces.Controller {
	return NewDocumentController(service)
}
