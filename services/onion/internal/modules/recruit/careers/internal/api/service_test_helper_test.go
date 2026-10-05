package api

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure"
)

func newCareersTestService(db *sql.DB) *application.Service {
	return application.NewService(infrastructure.NewRepository(db))
}
