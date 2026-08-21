package custom_domains

import (
	"context"
	"database/sql"

	customdomainapi "github.com/voxire/onion/internal/modules/custom_domains/internal/api"
	customdomainapp "github.com/voxire/onion/internal/modules/custom_domains/internal/application"
)

type (
	API     = customdomainapi.API
	Service = customdomainapp.Service
)

func NewService(ctx context.Context, db *sql.DB) *Service { return customdomainapp.NewService(ctx, db) }
func NewAPI(service *Service) *API                        { return customdomainapi.New(service) }
