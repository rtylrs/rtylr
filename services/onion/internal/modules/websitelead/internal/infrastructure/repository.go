package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		db: dbclient.Pool(),
	}
}

func (r *Repository) Create(ctx context.Context, input domain.LeadInput) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO voxire_website_leads (name, email, company, service, project, budget, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Email, input.Company, input.Service,
		input.Project, input.Budget, input.Source,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert website lead", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}
