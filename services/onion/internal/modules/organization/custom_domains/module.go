package custom_domains

import (
	"context"
	"database/sql"

	customdomainapi "github.com/voxire/onion/internal/modules/organization/custom_domains/internal/api"
	customdomainapp "github.com/voxire/onion/internal/modules/organization/custom_domains/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/custom_domains/internal/infrastructure"
)

type (
	API     = customdomainapi.API
	Service = customdomainapp.Service
)

func NewService(ctx context.Context, db *sql.DB) *Service {
	return customdomainapp.NewService(ctx, moduleinfra.NewRepository(db))
}
func NewAPI(service *Service) *API { return customdomainapi.New(service) }
