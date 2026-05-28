package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	autoMigrate(context.Background(), db)
	return &Repository{db: db}
}

func autoMigrate(ctx context.Context, db *sql.DB) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxire_website_leads (
			id         BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			email      VARCHAR(255) NOT NULL,
			phone      VARCHAR(50),
			company    VARCHAR(255),
			service    VARCHAR(100),
			project    TEXT,
			budget     VARCHAR(50),
			source     VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxire_website_leads table", logger.Error(err))
	}
}

func (r *Repository) Create(ctx context.Context, input domain.LeadInput) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO voxire_website_leads (name, email, phone, company, service, project, budget, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Email, input.Phone, input.Company, input.Service,
		input.Project, input.Budget, input.Source,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert website lead", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}
